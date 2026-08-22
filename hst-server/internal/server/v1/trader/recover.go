package trader

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/mailer"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// A recovery code lives 15 minutes and survives 5 wrong guesses.
const (
	resetCodeTTL      = 15 * time.Minute
	resetCodeAttempts = 5
)

// CrtForgotPassword asks for a code, by login or by the email on the account.
type CrtForgotPassword struct {
	Login int64  `json:"login"`
	Email string `json:"email" validate:"omitempty,email,max=255"`
}

// CrtVerifyCode checks a code without spending it.
type CrtVerifyCode struct {
	Login int64  `json:"login"`
	Email string `json:"email" validate:"omitempty,email,max=255"`
	Code  string `json:"code" validate:"required,max=16"`
}

// CrtResetPassword spends the code and sets the new password.
type CrtResetPassword struct {
	Login    int64  `json:"login"`
	Email    string `json:"email" validate:"omitempty,email,max=255"`
	Code     string `json:"code" validate:"required,max=16"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// ForgotPassword issues a one time recovery code.
//
//	@Id			TraderForgotPassword
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtForgotPassword	true	"the login or the email on the account"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	429		{object}	Response
//	@Router		/auth/trader/v1/forgot-password [post]
func (s *Server) ForgotPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ip := utils.GetRealIP(c)

	if s.OAuth2.IPThrottled(ctx, ip) || s.OAuth2.PublicCallThrottled(ctx, ip) {
		return s.App.HttpResponseTooManyRequests(c, errs.ErrTooManyRequests)
	}

	var body CrtForgotPassword
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.Login == 0 && body.Email == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	// the answer is the same whether or not the account exists, so this cannot enumerate logins
	login, err := s.findRecoverable(ctx, body)
	if err != nil {
		return s.App.HttpResponseOK(c, nil)
	}

	code, err := newResetCode()
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	hash, err := s.OAuth2.Hasher.HashPassword(code)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	now := time.Now()

	// an older code stops working the moment a new one is asked for
	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.password_reset_codes SET used_at = $1 WHERE login = $2 AND used_at = 0`,
		now.UnixNano(), login); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if _, err := s.DB.DB.Exec(ctx,
		`INSERT INTO hst.password_reset_codes (login, code_hash, expires_at, used_at, attempts, created_at)
		 VALUES ($1,$2,$3,0,0,$4)`,
		login, []byte(hash), now.Add(resetCodeTTL).UnixNano(), now.UnixNano()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.SendResetCode(ctx, login, code)

	s.Log.Log(logger.TypeUser, logger.CodeOK, "password reset code issued", "login", login, "ip", ip)

	// a recovery code is a credential, so it reaches the log only when a developer asks for it
	if s.Cfg.Auth.LogResetCodes {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "password reset code", "login", login, "code", code)
	}

	return s.App.HttpResponseOK(c, nil)
}

// VerifyCode confirms a code is valid and unexpired, without spending it.
//
//	@Id			TraderVerifyResetCode
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtVerifyCode	true	"the login and the code"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	401		{object}	Response
//	@Failure	429		{object}	Response
//	@Router		/auth/trader/v1/verify-code [post]
func (s *Server) VerifyCode(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if ip := utils.GetRealIP(c); s.OAuth2.IPThrottled(ctx, ip) || s.OAuth2.PublicCallThrottled(ctx, ip) {
		return s.App.HttpResponseTooManyRequests(c, errs.ErrTooManyRequests)
	}

	var body CrtVerifyCode
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	login, err := s.recoverLogin(ctx, body.Login, body.Email)
	if err != nil {
		return s.resetDenied(c, err)
	}

	if _, err := s.claimResetCode(ctx, login, body.Code, false); err != nil {
		return s.resetDenied(c, err)
	}

	return s.App.HttpResponseOK(c, nil)
}

// ResetPassword spends the code, sets the password and drops every session of that login.
//
//	@Id			TraderResetPassword
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtResetPassword	true	"the login, the code and the new password"
//	@Success	204		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	401		{object}	Response
//	@Failure	429		{object}	Response
//	@Router		/auth/trader/v1/reset-password [post]
func (s *Server) ResetPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if ip := utils.GetRealIP(c); s.OAuth2.IPThrottled(ctx, ip) || s.OAuth2.PublicCallThrottled(ctx, ip) {
		return s.App.HttpResponseTooManyRequests(c, errs.ErrTooManyRequests)
	}

	var body CrtResetPassword
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	login, err := s.recoverLogin(ctx, body.Login, body.Email)
	if err != nil {
		return s.resetDenied(c, err)
	}

	if err := s.RequirePasswordLength(ctx, login, "", body.Password); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	if _, err := s.claimResetCode(ctx, login, body.Code, true); err != nil {
		return s.resetDenied(c, err)
	}

	hash, err := s.OAuth2.Hasher.HashPassword(body.Password)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	now := time.Now().UnixNano()
	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.users
		    SET password_main = $1, last_pass_change = $2, updated_at = $2,
		        rights = rights & ~$3::bigint, failed_attempts = 0, locked_until = 0
		  WHERE login = $4`,
		hash, now, int64(model.UsersRights_reset_pass), login); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// whoever was signed in with the old password is signed out
	if err := s.OAuth2.InvalidateLogin(ctx, login, model.SessionRevokedRightsChanged); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeUser, logger.CodeLogin, "password reset", "login", login)

	return s.App.HttpResponseOK(c, nil)
}

// recoverLogin is the login behind whichever identifier the caller holds. A person recovering a
// password knows their email, not their number, so every step of the flow takes either.
func (s *Server) recoverLogin(ctx context.Context, login int64, email string) (int64, error) {
	if login != 0 {
		return login, nil
	}

	return s.findRecoverable(ctx, CrtForgotPassword{Email: email})
}

// findRecoverable resolves the request to a trading login that may recover a password.
func (s *Server) findRecoverable(ctx context.Context, body CrtForgotPassword) (int64, error) {
	var (
		login  int64
		rights model.UsersRights
	)

	// the email path takes the oldest match, so a duplicate address cannot be used to pick an account
	query := `SELECT login, rights FROM hst.users WHERE login = $1`
	arg := any(body.Login)
	if body.Login == 0 {
		query = `SELECT login, rights FROM hst.users WHERE email = $1 ORDER BY login LIMIT 1`
		arg = any(body.Email)
	}

	if err := s.DB.DB.QueryRow(ctx, query, arg).Scan(&login, &rights); err != nil {
		return 0, err
	}

	if !rights.CanConnect() {
		return 0, errs.ErrAccountDisabled
	}

	// staff recover their password through the back office, not through the terminal
	if _, err := s.SelectManager(ctx, login); err == nil {
		return 0, errs.ErrWrongPanel
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	return login, nil
}

// claimResetCode verifies the live code for a login, counting the attempt, and spends it when asked.
func (s *Server) claimResetCode(ctx context.Context, login int64, code string, spend bool) (int64, error) {
	now := time.Now().UnixNano()

	var (
		codeId   int64
		hash     []byte
		attempts int32
	)

	err := s.DB.DB.QueryRow(ctx,
		`SELECT code_id, code_hash, attempts
		   FROM hst.password_reset_codes
		  WHERE login = $1 AND used_at = 0 AND expires_at > $2
		  ORDER BY code_id DESC LIMIT 1`, login, now).
		Scan(&codeId, &hash, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		s.OAuth2.Hasher.VerifyDummy(code)
		return 0, errs.ErrInvalidCredentials
	}
	if err != nil {
		return 0, err
	}

	if attempts >= resetCodeAttempts {
		return 0, errs.ErrInvalidCredentials
	}

	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.password_reset_codes SET attempts = attempts + 1
		  WHERE code_id = $1`, codeId); err != nil {
		return 0, err
	}

	ok, _, err := s.OAuth2.Hasher.VerifyPassword(string(hash), code)
	if err != nil || !ok {
		return 0, errs.ErrInvalidCredentials
	}

	if spend {
		// single use: the same code cannot reset a password twice
		tag, err := s.DB.DB.Exec(ctx,
			`UPDATE hst.password_reset_codes SET used_at = $1
			  WHERE code_id = $2 AND used_at = 0`, now, codeId)
		if err != nil {
			return 0, err
		}
		if tag.RowsAffected() == 0 {
			return 0, errs.ErrInvalidCredentials
		}
	}

	return codeId, nil
}

// resetDenied answers a bad code as a plain refusal, and anything else as a fault.
func (s *Server) resetDenied(c *fiber.Ctx, err error) error {
	if errors.Is(err, errs.ErrInvalidCredentials) {
		return s.App.HttpResponseDenied(c, nethttp.StatusUnauthorized,
			nethttp.RetAuthAccountInvalid, errs.ErrInvalidCredentials)
	}

	return s.App.HttpResponseInternalServerErrorRequest(c, err)
}

// newResetCode is six digits, drawn from the crypto source.
func newResetCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return string([]byte{
		byte('0' + n.Int64()/100000%10),
		byte('0' + n.Int64()/10000%10),
		byte('0' + n.Int64()/1000%10),
		byte('0' + n.Int64()/100%10),
		byte('0' + n.Int64()/10%10),
		byte('0' + n.Int64()%10),
	}), nil
}

// SendResetCode mails the recovery code to the address on the account. The code is never
// returned by the API, so this is how a trader who has lost their password gets back in.
func (s *Server) SendResetCode(ctx context.Context, login int64, code string) {
	var name, email, group string
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT name, email, "group" FROM hst.users WHERE login = $1`, login).
		Scan(&name, &email, &group); err != nil || email == "" {
		return
	}

	company, catalog := s.CompanyOf(ctx, group)

	body, err := s.Mailer.Render(mailer.KindVerifyEmail, catalog, map[string]string{
		mailer.MacroName:             name,
		mailer.MacroConfirmationCode: code,
		mailer.MacroLogin:            strconv.FormatInt(login, 10),
		mailer.MacroCompany:          company,
	})
	if err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeErr, "reset code mail not rendered",
			"login", login, "error", err.Error())
		return
	}

	if err := s.Mailer.Queue(ctx, email, "Your confirmation code", body); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeErr, "reset code mail not queued",
			"login", login, "error", err.Error())
	}
}
