package v1

import (
	"context"
	"errors"
	"strconv"
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

// LoginRequest carries the terminal type.
type LoginRequest struct {
	// zero is a real terminal type, the trader one, so the value is checked against the enum instead
	ConnectionType int32 `json:"connection_type"`
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

// loginDenied is a refusal with the status and MT5 code it should carry.
type loginDenied struct {
	status int
	code   http.RetCode
	reason error
}

func (d *loginDenied) Error() string { return d.reason.Error() }

func denied(status int, code http.RetCode, reason error) error {
	return &loginDenied{status: status, code: code, reason: reason}
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
//	@Id				Login
//	@Description	Login with a manager account using basic auth
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"terminal type: 32 admin, 33 manager"
//	@Success		200		{object}	Response{data=ViewToken}
//	@Failure		401		{object}	Response
//	@Failure		403		{object}	Response
//	@Failure		500		{object}	Response
//	@Security		BasicAuth
//	@Router			/auth/v1/oauth2/login [post]
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

	login, password, err := basicCredentials(c)
	if err != nil {
		return s.loginFailed(c, err)
	}

	// staff and traders are admitted by different rules, and a terminal says which is asking
	connType := model.UsersConnectionTypes(body.ConnectionType)
	if _, ok := model.UsersConnectionTypes_name[body.ConnectionType]; !ok {
		return s.App.HttpResponseBadRequest(c, errs.ErrInvalidConnectionType)
	}

	// stop credential stuffing before it ever reaches argon2
	if s.OAuth2.IPThrottled(ctx, ip) {
		return s.App.HttpResponseTooManyRequests(c, errs.ErrTooManyRequests)
	}

	// is this really them
	user, scope, err := s.checkPassword(ctx, login, password, ip, connType)
	if err != nil {
		return s.loginFailed(c, err)
	}

	var manager *model.Manager
	if connType.IsStaff() {
		manager, err = s.checkManagerAccess(ctx, login, connType, ip)
	} else {
		err = s.checkTraderAccess(user)
	}
	if err != nil {
		return s.loginFailed(c, err)
	}

	// hand over the tokens
	view, err := s.openSession(ctx, user, manager, &oauth2.Config{
		Login:          login,
		ClientId:       user.ClientId,
		Scope:          int32(scope),
		ConnectionType: body.ConnectionType,
		IpAddress:      ip,
		UserAgent:      utils.GetUserAgent(c),
		// MT5 reset_pass: the session opens, but it may only change the password
		Restricted: user.Rights.MustChangePassword(),
	}, "", "")
	if err != nil {
		return s.loginFailed(c, err)
	}

	s.OAuth2.ClearFailures(ctx, login, ip)

	s.Log.Log(logger.TypeUser, logger.CodeLogin, "login",
		"login", login, "ip", ip, "connection_type", body.ConnectionType)

	return s.App.HttpResponseRetCode(c, view.Code, view)
}

// checkPassword loads the login, verifies the password, and reports which slot answered.
//
// Staff authenticate with the master slot only. A trading terminal may also present the investor
// password, which opens a session that sees everything and trades nothing.
func (s *HttpServer) checkPassword(ctx context.Context, login int64, password, ip string,
	connType model.UsersConnectionTypes) (*model.User, model.UsersPasswords, error) {
	user := &model.User{}

	err := s.DB.DB.QueryRow(ctx,
		`SELECT login, COALESCE(client_id, 0), "group", rights, password_main, password_investor, locked_until
		   FROM hst.users WHERE login = $1`, login).
		Scan(&user.Login, &user.ClientId, &user.Group, &user.Rights,
			&user.PasswordMain, &user.PasswordInvestor, &user.LockedUntil)

	if errors.Is(err, pgx.ErrNoRows) {
		// burn the same work as a real verify.
		s.OAuth2.Hasher.VerifyDummy(password)
		return nil, 0, denied(http.StatusUnauthorized, http.RetAuthAccountUnknown, errs.ErrInvalidCredentials)
	}
	if err != nil {
		return nil, 0, err
	}

	if user.LockedUntil > time.Now().UnixNano() {
		return nil, 0, denied(http.StatusForbidden, http.RetAccountLocked, errs.ErrAccountLocked)
	}

	slots := []struct {
		scope model.UsersPasswords
		hash  string
	}{{model.UsersPasswords_main, user.PasswordMain}}
	if !connType.IsStaff() && user.PasswordInvestor != "" {
		slots = append(slots, struct {
			scope model.UsersPasswords
			hash  string
		}{model.UsersPasswords_investor, user.PasswordInvestor})
	}

	for _, slot := range slots {
		ok, needsRehash, err := s.OAuth2.Hasher.VerifyPassword(slot.hash, password)
		if errors.Is(err, crypto.ErrHasherBusy) {
			return nil, 0, denied(http.StatusServiceUnavailable, http.RetAuthServerBusy, errs.ErrServiceUnavailable)
		}
		if err != nil || !ok {
			continue
		}

		if !user.Rights.CanConnect() {
			return nil, 0, denied(http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountDisabled)
		}

		// only the master slot is rehashed in place, because that is the one the user is asked to change
		if needsRehash && slot.scope == model.UsersPasswords_main {
			s.upgradePasswordHash(ctx, login, password)
		}

		return user, slot.scope, nil
	}

	return nil, 0, s.recordFailedLogin(ctx, login, ip)
}

// checkTraderAccess decides whether this login may use the trader panel.
//
// A trader needs no manager row; what it needs is an account that may connect and that somebody
// has approved. A preliminary account exists but has not been approved, so it may not trade yet.
func (s *HttpServer) checkTraderAccess(u *model.User) error {
	if !u.Rights.CanConnect() {
		return denied(http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountDisabled)
	}

	if IsPreliminaryGroup(u.Group) {
		return denied(http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountPending)
	}

	return nil
}

// checkManagerAccess decides whether this login may use the admin or manager panel.
func (s *HttpServer) checkManagerAccess(ctx context.Context, login int64,
	connType model.UsersConnectionTypes, ip string) (*model.Manager, error) {

	if !connType.IsStaff() {
		return nil, denied(http.StatusForbidden, http.RetAuthManagerType, errs.ErrTerminalNotPermitted)
	}

	manager, err := s.selectManager(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, denied(http.StatusForbidden, http.RetAuthManagerNoConfig, errs.ErrNotAManager)
	}
	if err != nil {
		return nil, err
	}

	// admin and manager terminals are gated separately in MT5
	if !manager.PermitsTerminal(connType) {
		return nil, denied(http.StatusForbidden, http.RetAuthManagerType, errs.ErrTerminalNotPermitted)
	}

	// the allowlist is checked last, so a wrong ip never leaks that the login exists
	if !manager.PermitsIP(ip) {
		return nil, denied(http.StatusForbidden, http.RetAuthManagerIpBlock, errs.ErrManagerIpBlocked)
	}

	return manager, nil
}

// recordFailedLogin counts the attempt and returns what the caller should answer:
func (s *HttpServer) recordFailedLogin(ctx context.Context, login int64, ip string) error {
	locked, err := s.OAuth2.RegisterFailure(ctx, login, ip)
	if err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to record login failure",
			"login", login, "error", err.Error())
	}

	s.Log.Log(logger.TypeUser, logger.CodeAtt, "failed login",
		"login", login, "ip", ip, "locked", locked)

	if locked {
		return denied(http.StatusForbidden, http.RetAccountLocked, errs.ErrAccountLocked)
	}

	return denied(http.StatusUnauthorized, http.RetAuthAccountInvalid, errs.ErrInvalidCredentials)
}

// upgradePasswordHash rewrites a hash made with weaker settings.
func (s *HttpServer) upgradePasswordHash(ctx context.Context, login int64, password string) {
	hash, err := s.OAuth2.Hasher.HashPassword(password)
	if err != nil {
		return
	}

	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.users SET password_main = $1 WHERE login = $2`, hash, login); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to upgrade password hash",
			"login", login, "error", err.Error())
	}
}

// basicCredentials reads what BasicAuthParser put in Locals.
func basicCredentials(c *fiber.Ctx) (int64, string, error) {
	loginStr, _ := c.Locals(http.LocalsUsername).(string)
	password, _ := c.Locals(http.LocalsPassword).(string)

	login, err := strconv.ParseInt(loginStr, 10, 64)
	if err != nil || login <= 0 {
		return 0, "", denied(http.StatusUnauthorized, http.RetAuthAccountUnknown, errs.ErrInvalidCredentials)
	}

	return login, password, nil
}

// loginFailed writes the refusal a step returned.
func (s *HttpServer) loginFailed(c *fiber.Ctx, err error) error {
	var d *loginDenied
	if errors.As(err, &d) {
		return s.App.HttpResponseDenied(c, d.status, d.code, d.reason)
	}

	return s.App.HttpResponseInternalServerErrorRequest(c, err)
}

// RefreshToken rotates the refresh token and issues a new access token.
//
//	@Id			RefreshToken
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		RefreshRequest	true	"the refresh token from login"
//	@Success	200		{object}	Response{data=ViewToken}
//	@Failure	401		{object}	Response
//	@Failure	500		{object}	Response
//	@Router		/auth/v1/oauth2/refresh [post]
func (s *HttpServer) RefreshToken(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body RefreshRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// the record is read before the lock is taken.
	rec, err := s.OAuth2.LoadRefresh(ctx, body.RefreshToken)
	if errors.Is(err, oauth2.ErrSessionNotFound) {
		return s.App.HttpResponseDenied(c, http.StatusUnauthorized, http.RetSessionExpired, errs.ErrInvalidSession)
	}
	if err != nil {
		return s.App.HttpResponseServiceUnavailable(c, errs.ErrSessionStoreUnavailable)
	}

	// a token that was already rotated is being replayed.
	if rec.Used {
		if rerr := s.OAuth2.RevokeFamily(ctx, rec.FamilyId, model.SessionRevokedReuseDetected); rerr != nil {
			s.Log.Log(logger.TypeUser, logger.CodeErr, "failed to revoke family",
				"family_id", rec.FamilyId, "error", rerr.Error())
		}

		s.Log.Log(logger.TypeUser, logger.CodeAtt, "refresh token reuse detected",
			"login", rec.Login, "family_id", rec.FamilyId, "ip", utils.GetRealIP(c))

		return s.App.HttpResponseDenied(c, http.StatusUnauthorized, http.RetSessionExpired, errs.ErrInvalidSession)
	}

	if rec.ExpiresAt <= time.Now().UnixNano() {
		return s.App.HttpResponseDenied(c, http.StatusUnauthorized, http.RetSessionExpired, errs.ErrInvalidSession)
	}

	// lock now, so two parallel refreshes cannot both rotate
	locked, err := s.OAuth2.LockRefresh(ctx, body.RefreshToken)
	if err != nil {
		return s.App.HttpResponseServiceUnavailable(c, errs.ErrSessionStoreUnavailable)
	}
	if !locked {
		return s.App.HttpResponseDenied(c, http.StatusUnauthorized, http.RetSessionExpired, errs.ErrInvalidSession)
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
			s.Log.Log(logger.TypeUser, logger.CodeErr, "failed to revoke family",
				"family_id", rec.FamilyId, "error", rerr.Error())
		}
		return s.App.HttpResponseDenied(c, http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountDisabled)
	}

	old, err := s.selectSession(ctx, rec.SessionId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	mgr, err := s.selectManager(ctx, rec.Login)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseDenied(c, http.StatusForbidden, http.RetAuthManagerNoConfig, errs.ErrNotAManager)
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

	// the old token becomes a tombstone rather than disappearing.
	if err := s.OAuth2.MarkRefreshUsed(ctx, body.RefreshToken, rec); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to tombstone refresh token",
			"login", u.Login, "error", err.Error())
	}

	if _, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.sessions SET revoked_at = $1, revoked_reason = $2
		   WHERE session_id = $3 AND revoked_at = 0`,
		time.Now().UnixNano(), model.SessionRevokedRotated, rec.SessionId); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to mark session rotated",
			"session_id", rec.SessionId, "error", err.Error())
	}

	if err := s.OAuth2.Redis.Client.Del(ctx, oauth2.KeySession(rec.SessionId)).Err(); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to drop rotated session",
			"session_id", rec.SessionId, "error", err.Error())
	}
	s.OAuth2.Cache.Invalidate(rec.SessionId)

	return s.App.HttpResponseRetCode(c, view.Code, view)
}

// Logout closes the current session.
//
//	@Id			Logout
//	@Tags		Auth
//	@Produce	json
//	@Success	204	{object}	Response
//	@Failure	401	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/auth/logout [post]
func (s *HttpServer) Logout(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if err := s.OAuth2.RevokeSession(c.UserContext(), snap.SessionId, model.SessionRevokedLogout); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeUser, logger.CodeLogin, "logout", "login", snap.Login)

	return s.App.HttpResponseNoContent(c)
}

// Me returns the caller's own profile and decoded rights.
//
//	@Id			Me
//	@Tags		Auth
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewMe}
//	@Failure	401	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/auth/me [get]
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
//
//	@Id			ChangePassword
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		ChangePasswordRequest	true	"old and new password"
//	@Success	204		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	401		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/auth/oauth2/change-password [post]
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
		return s.App.HttpResponseDenied(c, http.StatusUnauthorized, http.RetAuthAccountInvalid, errs.ErrInvalidCredentials)
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

	s.Log.Log(logger.TypeUser, logger.CodeLogin, "password changed", "login", snap.Login)

	return s.App.HttpResponseNoContent(c)
}

// openSession writes the session row, the redis state and the tokens.
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

	// a rotation may not outlive the family it belongs to.
	if parentId != "" {
		var familyStart int64
		if err := s.DB.DB.QueryRow(ctx,
			`SELECT min(created_at) FROM hst.sessions WHERE family_id = $1`,
			familyId).Scan(&familyStart); err == nil && familyStart > 0 {
			if cap := familyStart + int64(s.Cfg.Auth.RefreshAbsoluteTTL); cap < expiresAt {
				expiresAt = cap
			}
		}
	}

	if expiresAt <= now.UnixNano() {
		return nil, errs.ErrInvalidSession
	}

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
		return nil, denied(http.StatusServiceUnavailable, http.RetAuthServerBusy, errs.ErrSessionStoreUnavailable)
	}

	if err := s.OAuth2.SaveRefresh(ctx, refresh, &oauth2.RefreshRecord{
		SessionId: sid,
		FamilyId:  familyId,
		Login:     u.Login,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, denied(http.StatusServiceUnavailable, http.RetAuthServerBusy, errs.ErrSessionStoreUnavailable)
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
		code = http.RetAuthResetPassword
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
