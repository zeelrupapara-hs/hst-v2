package middleware

import (
	"errors"
	"strings"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/http"
	"hstserver/pkg/oauth2"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// Protect authenticates the request. It is the hot path, so the order matters:
// a forged token is rejected on signature alone, with zero I/O.
//
// It never reads postgres. A missing session means the client should refresh,
// and refresh is the only path allowed to touch the database.
func (m *Middleware) Protect(c *fiber.Ctx) error {
	token, ok := utils.GetToken(c)
	if !ok {
		return m.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrMissingAuthorizationHeader)
	}

	claims, err := m.OAuth2.Signer.Parse(token)
	if err != nil {
		return m.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrInvalidToken)
	}

	snap, hit := m.OAuth2.Cache.Get(claims.Sid)
	if !hit {
		snap, err = m.OAuth2.LoadSnapshot(c.UserContext(), claims.Sid)
		switch {
		case errors.Is(err, oauth2.ErrSessionNotFound):
			return m.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrInvalidSession)
		case err != nil:
			// the session store is down, not the client's fault. A 401 here
			// would make it burn its refresh token on our outage.
			return m.App.HttpResponseServiceUnavailable(c, errs.ErrSessionStoreUnavailable)
		}
		m.OAuth2.Cache.Put(snap)
	}

	if snap.ExpiresAt <= time.Now().UnixNano() {
		return m.App.HttpResponseRetCodeDenied(c, http.StatusUnauthorized, http.RetAuthSessionExpired, errs.ErrInvalidSession)
	}

	if !model.UsersRights(snap.Rights).CanConnect() {
		return m.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountDisabled)
	}

	// a restricted session may do exactly one thing: change its password
	if snap.Restricted && !isPasswordChangeRoute(c) {
		return m.App.HttpResponseRetCodeDenied(c, http.StatusForbidden, http.RetAuthUpdatePassword, errs.ErrMustChangePassword)
	}

	c.Locals(http.LocalsClient, snap)
	m.OAuth2.TouchAsync(claims.Sid)

	return c.Next()
}

func isPasswordChangeRoute(c *fiber.Ctx) bool {
	return c.Method() == fiber.MethodPost &&
		strings.HasSuffix(c.Path(), "/oauth2/change-password")
}
