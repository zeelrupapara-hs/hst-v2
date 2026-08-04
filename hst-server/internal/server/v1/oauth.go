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
	"hstserver/pkg/journal"
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

// loginDenied is a refusal with the status and the platform code it should carry.
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

// LoginAs authenticates and opens a session for one panel.
func (s *HttpServer) LoginAs(c *fiber.Ctx, staffOnly bool) error {
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

	if connType.IsStaff() != staffOnly {
		return s.loginFailed(c, denied(http.StatusForbidden, http.RetAuthClientInvalid, errs.ErrWrongPanel))
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
		err = s.checkTraderAccess(ctx, user)
	}
	if err != nil {
		return s.loginFailed(c, err)
	}

	// A trading account holds one session at a time. Closing the earlier one here, before the new
	// one exists, is what tells the terminal already signed in that it has been replaced: it is
	// sent session.revoked and its socket closes, rather than sitting there receiving nothing.
	//
	// Staff are exempt: running the manager and the administrator side by side, or a desk across
	// two screens, is how the platform is actually operated.
	if !connType.IsStaff() {
		if err := s.OAuth2.InvalidateLogin(ctx, login, model.SessionRevokedSuperseded); err != nil {
			return s.loginFailed(c, err)
		}
	}

	// hand over the tokens
	view, err := s.openSession(ctx, user, manager, &oauth2.Config{
		Login:          login,
		ClientId:       user.ClientId,
		Scope:          int32(scope),
		ConnectionType: body.ConnectionType,
		IpAddress:      ip,
		UserAgent:      utils.GetUserAgent(c),
		// reset_pass: the session opens, but it may only change the password
		Restricted: user.Rights.MustChangePassword(),
	}, "", "")
	if err != nil {
		return s.loginFailed(c, err)
	}

	s.OAuth2.ClearFailures(ctx, login, ip)

	s.Log.Log(logger.TypeUser, logger.CodeLogin, "login",
		"login", login, "ip", ip, "connection_type", body.ConnectionType)

	s.WriteJournalFrom(ctx, login, ip, connType.Channel(), utils.OperatingSystem(utils.GetUserAgent(c)),
		logger.TypeUser, logger.CodeLogin, journal.SignedInMsg(login, ip), view)

	return s.App.HttpResponseRetCode(c, view.Code, view)
}

// checkPassword loads the login, verifies the password, and reports which slot answered.
func (s *HttpServer) checkPassword(ctx context.Context, login int64, password, ip string,
	connType model.UsersConnectionTypes) (*model.User, model.UsersPasswords, error) {
	user := &model.User{}

	var groupPermissions int32

	err := s.DB.DB.QueryRow(ctx,
		`SELECT u.login, COALESCE(u.client_id, 0), u."group", u.rights, u.password_main,
		        u.password_investor, u.locked_until, COALESCE(g.permission_flags, 0)
		   FROM hst.users u
		   LEFT JOIN hst.groups g ON g."group" = u."group"
		  WHERE u.login = $1`, login).
		Scan(&user.Login, &user.ClientId, &user.Group, &user.Rights,
			&user.PasswordMain, &user.PasswordInvestor, &user.LockedUntil, &groupPermissions)

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

		// the group can close the door for everyone in it at once, without touching an account
		if groupPermissions&int32(model.PermissionsFlags_enable_connection) == 0 {
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
func (s *HttpServer) checkTraderAccess(ctx context.Context, u *model.User) error {
	// a member of staff is not a trading account, whichever door it knocks on
	if _, err := s.SelectManager(ctx, u.Login); err == nil {
		return denied(http.StatusForbidden, http.RetAuthClientInvalid, errs.ErrWrongPanel)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

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

	manager, err := s.SelectManager(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, denied(http.StatusForbidden, http.RetAuthManagerNoConfig, errs.ErrNotAManager)
	}
	if err != nil {
		return nil, err
	}

	// admin and manager terminals are gated separately in the platform
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

// RefreshSession exchanges a refresh token for a new pair, on the panel that issued it.
func (s *HttpServer) RefreshSession(c *fiber.Ctx, staffOnly bool) error {
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

	old, err := s.selectSession(ctx, rec.SessionId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// a session refreshes on the panel that issued it, and being on the wrong one is a plain refusal.
	if model.UsersConnectionTypes(old.ConnectionType).IsStaff() != staffOnly {
		return s.App.HttpResponseDenied(c, http.StatusForbidden, http.RetAuthClientInvalid, errs.ErrWrongPanel)
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

	// only a staff session carries a manager row, so a trader is not asked for one
	var mgr *model.Manager
	if staffOnly {
		mgr, err = s.SelectManager(ctx, rec.Login)
		if errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseDenied(c, http.StatusForbidden, http.RetAuthManagerNoConfig, errs.ErrNotAManager)
		}
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
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

// LogoutSession closes the calling session.
func (s *HttpServer) LogoutSession(c *fiber.Ctx) error {
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

// CurrentSession describes the caller to itself.
func (s *HttpServer) CurrentSession(c *fiber.Ctx) error {
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

// SetPassword changes the caller's own password.
func (s *HttpServer) SetPassword(c *fiber.Ctx) error {
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

// SelectManager reads the whole right set in one round trip.
func (s *HttpServer) SelectManager(ctx context.Context, login int64) (*model.Manager, error) {
	m := &model.Manager{}

	err := s.DB.DB.QueryRow(ctx,
		`SELECT login, name, mailbox, server, request_limit_logs,
		        request_limit_reports, groups, access,
		        right_admin, right_manager, right_cfg_time, right_cfg_holidays,
		        right_cfg_groups, right_cfg_managers, right_cfg_requests,
		        right_cfg_gateways, right_cfg_datafeeds, right_cfg_reports,
		        right_cfg_symbols, right_cfg_web_services, right_cfg_messengers,
		        right_cfg_kyc, right_cfg_automations, right_cfg_allocations,
		        right_cfg_corporate, right_cfg_payments, right_cfg_mails,
		        right_cfg_streaming, right_srv_journals, right_srv_reports,
		        right_charts, right_email, right_news, right_export,
		        right_techsupport, right_market, right_accountant, right_acc_read,
		        right_acc_details_name, right_acc_details_location,
		        right_acc_details_address, right_acc_details_id,
		        right_acc_details_email, right_acc_details_phone,
		        right_acc_details_general, right_acc_technical,
		        right_acc_tech_modify, right_acc_manager, right_acc_delete,
		        right_acc_online, right_confirm_actions, right_notifications,
		        right_trades_read, right_trades_manager, right_trades_delete,
		        right_trades_dealer, right_trades_supervisor, right_quotes_raw,
		        right_quotes, right_symbol_details, right_risk_manager,
		        right_group_margin, right_group_commission, right_reports,
		        right_clients_access, right_clients_create, right_clients_edit,
		        right_clients_delete, right_clients_kyc,
		        right_clients_details_name, right_clients_details_location,
		        right_clients_details_address, right_clients_details_id,
		        right_clients_details_email, right_clients_details_phone,
		        right_clients_details_general, right_documents_access,
		        right_documents_create, right_documents_edit,
		        right_documents_delete, right_documents_files_add,
		        right_documents_files_delete, right_comments_access,
		        right_comments_create, right_comments_delete
		   FROM hst.managers WHERE login = $1`, login).
		Scan(&m.Login, &m.Name, &m.Mailbox, &m.Server, &m.RequestLimitLogs,
			&m.RequestLimitReports, &m.Groups, &m.Access,
			&m.RightAdmin,
			&m.RightManager,
			&m.RightCfgTime,
			&m.RightCfgHolidays,
			&m.RightCfgGroups,
			&m.RightCfgManagers,
			&m.RightCfgRequests,
			&m.RightCfgGateways,
			&m.RightCfgDatafeeds,
			&m.RightCfgReports,
			&m.RightCfgSymbols,
			&m.RightCfgWebServices,
			&m.RightCfgMessengers,
			&m.RightCfgKyc,
			&m.RightCfgAutomations,
			&m.RightCfgAllocations,
			&m.RightCfgCorporate,
			&m.RightCfgPayments,
			&m.RightCfgMails,
			&m.RightCfgStreaming,
			&m.RightSrvJournals,
			&m.RightSrvReports,
			&m.RightCharts,
			&m.RightEmail,
			&m.RightNews,
			&m.RightExport,
			&m.RightTechsupport,
			&m.RightMarket,
			&m.RightAccountant,
			&m.RightAccRead,
			&m.RightAccDetailsName,
			&m.RightAccDetailsLocation,
			&m.RightAccDetailsAddress,
			&m.RightAccDetailsId,
			&m.RightAccDetailsEmail,
			&m.RightAccDetailsPhone,
			&m.RightAccDetailsGeneral,
			&m.RightAccTechnical,
			&m.RightAccTechModify,
			&m.RightAccManager,
			&m.RightAccDelete,
			&m.RightAccOnline,
			&m.RightConfirmActions,
			&m.RightNotifications,
			&m.RightTradesRead,
			&m.RightTradesManager,
			&m.RightTradesDelete,
			&m.RightTradesDealer,
			&m.RightTradesSupervisor,
			&m.RightQuotesRaw,
			&m.RightQuotes,
			&m.RightSymbolDetails,
			&m.RightRiskManager,
			&m.RightGroupMargin,
			&m.RightGroupCommission,
			&m.RightReports,
			&m.RightClientsAccess,
			&m.RightClientsCreate,
			&m.RightClientsEdit,
			&m.RightClientsDelete,
			&m.RightClientsKyc,
			&m.RightClientsDetailsName,
			&m.RightClientsDetailsLocation,
			&m.RightClientsDetailsAddress,
			&m.RightClientsDetailsId,
			&m.RightClientsDetailsEmail,
			&m.RightClientsDetailsPhone,
			&m.RightClientsDetailsGeneral,
			&m.RightDocumentsAccess,
			&m.RightDocumentsCreate,
			&m.RightDocumentsEdit,
			&m.RightDocumentsDelete,
			&m.RightDocumentsFilesAdd,
			&m.RightDocumentsFilesDelete,
			&m.RightCommentsAccess,
			&m.RightCommentsCreate,
			&m.RightCommentsDelete)

	return m, err
}
