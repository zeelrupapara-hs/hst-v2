package admin

import (
	"github.com/gofiber/fiber/v2"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"
)

// ViewOnlineUser is one live connection; a login on three devices is three rows.
type ViewOnlineUser struct {
	SessionId      string `json:"session_id"`
	ConnectionId   string `json:"connection_id"`
	Login          int64  `json:"login"`
	Name           string `json:"name"`
	Group          string `json:"group"`
	ConnectionType int32  `json:"connection_type"`
	Ip             string `json:"ip"`
	Os             string `json:"os"`
	ConnectedAt    int64  `json:"connected_at"`
}

// ListOnlineUsers returns every live connection whose account the caller's masks cover.
//
//	@Id			ListOnlineUsers
//	@Tags		Online
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewOnlineUser}
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/online [get]
func (s *Server) ListOnlineUsers(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	connections := s.Hub.Snapshot()
	if len(connections) == 0 {
		return s.App.HttpResponseOK(c, []ViewOnlineUser{})
	}

	// one query names every connected login, then the masks decide which rows the caller sees
	logins := make([]int64, 0, len(connections))
	seen := make(map[int64]bool, len(connections))
	for _, conn := range connections {
		if !seen[conn.Login] {
			seen[conn.Login] = true
			logins = append(logins, conn.Login)
		}
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT login, name, "group" FROM hst.users WHERE login = ANY($1)`, logins)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	type identity struct{ name, group string }
	identities := make(map[int64]identity, len(logins))
	for rows.Next() {
		var login int64
		var id identity
		if err := rows.Scan(&login, &id.name, &id.group); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		identities[login] = id
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	out := []ViewOnlineUser{}
	for _, conn := range connections {
		id, known := identities[conn.Login]
		if !known || !model.MasksCover(snap.ManagerGroups, []string{id.group}) {
			continue
		}
		out = append(out, ViewOnlineUser{
			SessionId:      conn.SessionId,
			ConnectionId:   conn.Id,
			Login:          conn.Login,
			Name:           id.name,
			Group:          id.group,
			ConnectionType: conn.ConnectionType,
			Ip:             conn.Ip,
			Os:             conn.Os,
			ConnectedAt:    conn.ConnectedAt.UnixNano(),
		})
	}

	return s.App.HttpResponseOK(c, out)
}

// DisconnectSession revokes one session and closes its sockets; the login's other devices stay.
//
//	@Id			DisconnectSession
//	@Tags		Online
//	@Produce	json
//	@Param		session_id	path	string	true	"the session to disconnect"
//	@Success	204
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/online/{session_id} [delete]
func (s *Server) DisconnectSession(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	sessionId := c.Params("session_id")
	connections := s.Hub.Session(sessionId)
	if len(connections) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	target := connections[0].Login

	// a session outside the caller's masks reads as absent, so ids cannot be probed
	var group string
	if err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT "group" FROM hst.users WHERE login = $1`, target).Scan(&group); err != nil {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if !model.MasksCover(snap.ManagerGroups, []string{group}) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	if err := s.OAuth2.RevokeSession(c.UserContext(), sessionId, model.SessionRevokedLogout); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	s.Hub.CloseSession(sessionId)

	s.Log.Log(logger.TypeUser, logger.CodeWarn, "session disconnected",
		"actor", snap.Login, "target", target, "session_id", sessionId)
	s.JournalEntry(c, model.JournalType_auth, logger.CodeWarn,
		journal.SessionDisconnectedMsg(target), nil)

	return s.App.HttpResponseNoContent(c)
}
