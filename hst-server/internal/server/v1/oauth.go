package v1

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"hstserver/model"
	"hstserver/pkg/crypto"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/http"
	"hstserver/pkg/jwt"
	"hstserver/pkg/logger"
	"hstserver/pkg/oauth2"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LoginRequest carries the terminal type. The credentials are in the basic
// auth header, never in the body.
type LoginRequest struct {
	ConnectionType int32 `json:"connection_type" validate:"required"`
}

// ViewToken is what a successful login or refresh returns.
type ViewToken struct {
	Login          int64        `json:"login"`
	AccessToken    string       `json:"access_token"`
	RefreshToken   string       `json:"refresh_token"`
	SessionId      string       `json:"session_id"`
	ExpiresIn      int          `json:"expires_in"`
	ConnectionType int32        `json:"connection_type"`
	Code           http.RetCode `json:"code"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// ViewMe is the profile of the caller.
type ViewMe struct {
	Login          int64           `json:"login"`
	ClientId       int64           `json:"client_id"`
	Group          string          `json:"group"`
	ConnectionType int32           `json:"connection_type"`
	IsManager      bool            `json:"is_manager"`
	Restricted     bool            `json:"restricted"`
	Rights         map[string]bool `json:"manager_rights,omitempty"`
}

// Login authenticates a manager and opens a session.
//
// The order of the checks is deliberate: the account state and manager checks
// run after the password is verified, because answering "disabled" or "not a
// manager" first would tell an attacker the login exists.
//
// @Id			Login
// @Description	Login with a manager account using basic auth
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Success		200	{object}	ViewToken
// @Security	BasicAuth
// @Router		/auth/v1/oauth2/login [post]
func (s *HttpServer) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ip := utils.GetRealIP(c)

	var body LoginRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	loginStr, password, ok := basicAuth(c)
	if !ok {
		return s.App.HttpResponseUnauthorized(c, errs.ErrInvalidBasicAuth)
	}

	login, err := strconv.ParseInt(loginStr, 10, 64)
	if err != nil || login <= 0 {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthClientInvalid, errs.ErrInvalidCredentials)
	}

	// stop credential stuffing before it reaches argon2
	if s.OAuth2.IPThrottled(ctx, ip) {
		return s.App.HttpResponseTooManyRequests(c, errs.ErrTooManyRequests)
	}

	u := &model.User{}
	err = s.DB.DB.QueryRow(ctx,
		`SELECT login, COALESCE(client_id, 0), "group", rights,
		        password_main, locked_until
		   FROM hst.users WHERE login = $1`, login).
		Scan(&u.Login, &u.ClientId, &u.Group, &u.Rights, &u.PasswordMain, &u.LockedUntil)

	if errors.Is(err, pgx.ErrNoRows) {
		// burn the same work as a real verify so a missing login and a wrong
		// password cannot be told apart by timing
		s.OAuth2.Hasher.VerifyDummy(password)
		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthClientInvalid, errs.ErrInvalidCredentials)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if u.LockedUntil > time.Now().UnixNano() {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthAccountBlocked, errs.ErrAccountLocked)
	}

	// managers authenticate with the main slot only. The investor and api slots
	// belong to trader login, which this endpoint does not serve.
	valid, needsRehash, err := s.OAuth2.Hasher.VerifyPassword(u.PasswordMain, password)
	if errors.Is(err, crypto.ErrHasherBusy) {
		return s.App.HttpResponseServiceUnavailable(c, errs.ErrServiceUnavailable)
	}
	if err != nil || !valid {
		locked, ferr := s.OAuth2.RegisterFailure(ctx, login, ip)
		if ferr != nil {
			s.Log.Journal(logger.TypeUser, logger.CodeWarn, "failed to record login failure",
				"login", login, "error", ferr.Error())
		}

		s.Log.Journal(logger.TypeUser, logger.CodeAtt, "failed login",
			"login", login, "ip", ip, "locked", locked)

		if locked {
			return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthAccountBlocked, errs.ErrAccountLocked)
		}
		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthAccountInvalid, errs.ErrInvalidCredentials)
	}

	if !u.Rights.CanConnect() {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountDisabled)
	}

	connType := model.UsersConnectionTypes(body.ConnectionType)
	if !connType.IsStaff() {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthTerminalInvalid, errs.ErrTerminalNotPermitted)
	}

	mgr, err := s.selectManager(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthManagerInvalid, errs.ErrNotAManager)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if !mgr.PermitsTerminal(connType) {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthTerminalInvalid, errs.ErrTerminalNotPermitted)
	}

	// upgrading the stored hash must never block the login
	if needsRehash {
		if hash, herr := s.OAuth2.Hasher.HashPassword(password); herr == nil {
			if _, uerr := s.DB.DB.Exec(ctx,
				`UPDATE hst.users SET password_main = $1 WHERE login = $2`, hash, login); uerr != nil {
				s.Log.Journal(logger.TypeUser, logger.CodeWarn, "failed to upgrade password hash",
					"login", login, "error", uerr.Error())
			}
		}
	}

	cfg := &oauth2.Config{
		Login:          login,
		ClientId:       u.ClientId,
		Scope:          int32(model.UsersPasswords_main),
		ConnectionType: body.ConnectionType,
		IpAddress:      ip,
		UserAgent:      utils.GetUserAgent(c),
		// MT5 reset_pass: the session opens, but it may only change the password
		Restricted: u.Rights.MustChangePassword(),
	}

	view, err := s.openSession(ctx, u, mgr, cfg, "", "")
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.OAuth2.ClearFailures(ctx, login, ip)

	s.Log.Journal(logger.TypeUser, logger.CodeLogin, "login",
		"login", login, "ip", ip, "connection_type", body.ConnectionType)

	return s.App.HttpResponseRetCode(c, view.Code, view, nil)
}

// RefreshToken rotates the refresh token and issues a new access token.
// This is the only endpoint besides Login that may read postgres for auth.
//
// @Id			RefreshToken
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Success		200	{object}	ViewToken
// @Router		/auth/v1/oauth2/refresh [post]
func (s *HttpServer) RefreshToken(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body RefreshRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// the record is read before the lock is taken. Doing it the other way round
	// lets a replay inside the lock window return a plain 401, so the theft is
	// never detected.
	rec, err := s.OAuth2.LoadRefresh(ctx, body.RefreshToken)
	if errors.Is(err, oauth2.ErrSessionNotFound) {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrInvalidSession)
	}
	if err != nil {
		return s.App.HttpResponseServiceUnavailable(c, errs.ErrSessionStoreUnavailable)
	}

	// a token that was already rotated is being replayed. Assume the family is
	// compromised and kill all of it, including the session the thief may hold.
	if rec.Used {
		if rerr := s.OAuth2.RevokeFamily(ctx, rec.FamilyId, model.SessionRevokedReuseDetected); rerr != nil {
			s.Log.Journal(logger.TypeUser, logger.CodeErr, "failed to revoke family",
				"family_id", rec.FamilyId, "error", rerr.Error())
		}

		s.Log.Journal(logger.TypeUser, logger.CodeAtt, "refresh token reuse detected",
			"login", rec.Login, "family_id", rec.FamilyId, "ip", utils.GetRealIP(c))

		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrInvalidSession)
	}

	if rec.ExpiresAt <= time.Now().UnixNano() {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrInvalidSession)
	}

	// only now take the lock, so two parallel refreshes of the same unused
	// token do not both rotate it
	locked, err := s.OAuth2.LockRefresh(ctx, body.RefreshToken)
	if err != nil {
		return s.App.HttpResponseServiceUnavailable(c, errs.ErrSessionStoreUnavailable)
	}
	if !locked {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrInvalidSession)
	}

	// re-read the user: rights may have changed since the token was issued
	u := &model.User{}
	err = s.DB.DB.QueryRow(ctx,
		`SELECT login, COALESCE(client_id, 0), "group", rights
		   FROM hst.users WHERE login = $1`, rec.Login).
		Scan(&u.Login, &u.ClientId, &u.Group, &u.Rights)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if !u.Rights.CanConnect() {
		if rerr := s.OAuth2.RevokeFamily(ctx, rec.FamilyId, model.SessionRevokedRightsChanged); rerr != nil {
			s.Log.Journal(logger.TypeUser, logger.CodeErr, "failed to revoke family",
				"family_id", rec.FamilyId, "error", rerr.Error())
		}
		return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountDisabled)
	}

	old, err := s.selectSession(ctx, rec.SessionId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	mgr, err := s.selectManager(ctx, rec.Login)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthManagerInvalid, errs.ErrNotAManager)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	cfg := &oauth2.Config{
		Login:          u.Login,
		ClientId:       u.ClientId,
		Scope:          int32(old.Scope),
		ConnectionType: int32(old.ConnectionType),
		IpAddress:      utils.GetRealIP(c),
		UserAgent:      utils.GetUserAgent(c),
		Restricted:     u.Rights.MustChangePassword(),
	}

	view, err := s.openSession(ctx, u, mgr, cfg, rec.FamilyId, rec.SessionId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// the old token becomes a tombstone rather than disappearing, so a replay
	// is detectable instead of merely failing
	if err := s.OAuth2.MarkRefreshUsed(ctx, body.RefreshToken, rec); err != nil {
		s.Log.Journal(logger.TypeUser, logger.CodeWarn, "failed to tombstone refresh token",
			"login", u.Login, "error", err.Error())
	}

	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.sessions SET revoked_at = $1, revoked_reason = $2
		   WHERE session_id = $3 AND revoked_at = 0`,
		time.Now().UnixNano(), model.SessionRevokedRotated, rec.SessionId); err != nil {
		s.Log.Journal(logger.TypeUser, logger.CodeWarn, "failed to mark session rotated",
			"session_id", rec.SessionId, "error", err.Error())
	}

	if err := s.OAuth2.Redis.Client.Del(ctx, oauth2.KeySession(rec.SessionId)).Err(); err != nil {
		s.Log.Journal(logger.TypeUser, logger.CodeWarn, "failed to drop rotated session",
			"session_id", rec.SessionId, "error", err.Error())
	}
	s.OAuth2.Cache.Invalidate(rec.SessionId)

	return s.App.HttpResponseRetCode(c, view.Code, view, nil)
}

// Logout closes the current session.
//
// @Id			Logout
// @Tags		Auth
// @Produce		json
// @Security	BearerAuth
// @Router		/api/v1/auth/logout [post]
func (s *HttpServer) Logout(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if err := s.OAuth2.RevokeSession(c.UserContext(), snap.SessionId, model.SessionRevokedLogout); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Journal(logger.TypeUser, logger.CodeLogin, "logout", "login", snap.Login)

	return s.App.HttpResponseNoContent(c)
}

// Me returns the caller's own profile and decoded rights.
//
// @Id			Me
// @Tags		Auth
// @Produce		json
// @Success		200	{object}	ViewMe
// @Security	BearerAuth
// @Router		/api/v1/auth/me [get]
func (s *HttpServer) Me(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	view := &ViewMe{
		Login:          snap.Login,
		ClientId:       snap.ClientId,
		Group:          snap.Group,
		ConnectionType: snap.ConnectionType,
		IsManager:      snap.IsManager,
		Restricted:     snap.Restricted,
	}
	if snap.IsManager {
		view.Rights = snap.ManagerRights.Flags()
	}

	return s.App.HttpResponseOK(c, view)
}

// ChangePassword sets a new main password and closes every other session.
// It is the one route a restricted session may reach.
//
// @Id			ChangePassword
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Router		/api/v1/auth/oauth2/change-password [post]
func (s *HttpServer) ChangePassword(c *fiber.Ctx) error {
	ctx := c.UserContext()

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body ChangePasswordRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	var current string
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT password_main FROM hst.users WHERE login = $1`, snap.Login).Scan(&current); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	valid, _, err := s.OAuth2.Hasher.VerifyPassword(current, body.OldPassword)
	if err != nil || !valid {
		return s.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthAccountInvalid, errs.ErrInvalidCredentials)
	}

	hash, err := s.OAuth2.Hasher.HashPassword(body.NewPassword)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	now := time.Now().UnixNano()
	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.users
		    SET password_main = $1, last_pass_change = $2, updated_at = $2,
		        rights = rights & ~$3::bigint, failed_attempts = 0, locked_until = 0
		  WHERE login = $4`,
		hash, now, int64(model.UsersRights_reset_pass), snap.Login); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// every session built before the change is stale, including this one
	if err := s.OAuth2.InvalidateLogin(ctx, snap.Login, model.SessionRevokedRightsChanged); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Journal(logger.TypeUser, logger.CodeLogin, "password changed", "login", snap.Login)

	return s.App.HttpResponseNoContent(c)
}

// openSession writes the session row, the redis state and the tokens.
// familyId empty means a fresh login; otherwise this is a rotation.
func (s *HttpServer) openSession(ctx context.Context, u *model.User, mgr *model.Manager,
	cfg *oauth2.Config, familyId, parentId string) (*ViewToken, error) {

	sid := uuid.NewString()
	if familyId == "" {
		familyId = sid
	}

	refresh, err := crypto.GenerateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(s.Cfg.Auth.RefreshTTL).UnixNano()

	var parent *string
	if parentId != "" {
		parent = &parentId
	}

	if _, err := s.DB.DB.Exec(ctx,
		`INSERT INTO hst.sessions
		   (session_id, login, scope, connection_type, token_hash, ip, user_agent,
		    created_at, last_seen_at, expires_at, family_id, parent_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8,$9,$10,$11)`,
		sid, u.Login, cfg.Scope, cfg.ConnectionType, crypto.HashToken(refresh),
		cfg.IpAddress, cfg.UserAgent, now.UnixNano(), expiresAt, familyId, parent); err != nil {
		return nil, err
	}

	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.users SET last_access = $1, last_ip = $2, failed_attempts = 0, locked_until = 0
		  WHERE login = $3`, now.UnixNano(), cfg.IpAddress, u.Login); err != nil {
		return nil, err
	}

	snap := oauth2.NewSnapshot(sid, u, mgr, cfg, expiresAt)
	if err := s.OAuth2.SaveSnapshot(ctx, snap); err != nil {
		return nil, err
	}

	if err := s.OAuth2.SaveRefresh(ctx, refresh, &oauth2.RefreshRecord{
		SessionId: sid,
		FamilyId:  familyId,
		Login:     u.Login,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, err
	}

	access, err := s.OAuth2.Signer.Sign(u.Login, &jwt.Claims{
		Sid:        sid,
		Cid:        u.ClientId,
		ConnType:   cfg.ConnectionType,
		Restricted: cfg.Restricted,
		Ver:        snap.Version,
	})
	if err != nil {
		return nil, err
	}

	code := http.RetOK
	if cfg.Restricted {
		code = http.RetAuthUpdatePassword
	}

	return &ViewToken{
		Login:          u.Login,
		AccessToken:    access,
		RefreshToken:   refresh,
		SessionId:      sid,
		ExpiresIn:      int(s.Cfg.Auth.AccessTTL.Seconds()),
		ConnectionType: cfg.ConnectionType,
		Code:           code,
	}, nil
}

// selectSession reads the durable session row.
func (s *HttpServer) selectSession(ctx context.Context, sid string) (*model.Session, error) {
	sess := &model.Session{}
	err := s.DB.DB.QueryRow(ctx,
		`SELECT session_id, login, scope, connection_type, expires_at, revoked_at,
		        revoked_reason, family_id
		   FROM hst.sessions WHERE session_id = $1`, sid).
		Scan(&sess.SessionId, &sess.Login, &sess.Scope, &sess.ConnectionType,
			&sess.ExpiresAt, &sess.RevokedAt, &sess.RevokedReason, &sess.FamilyId)

	return sess, err
}

// basicAuth decodes the Authorization header into login and password.
func basicAuth(c *fiber.Ctx) (string, string, bool) {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Basic ") {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(header[6:])
	if err != nil {
		return "", "", false
	}

	login, password, found := strings.Cut(string(decoded), ":")
	if !found || login == "" || password == "" {
		return "", "", false
	}

	return login, password, true
}
