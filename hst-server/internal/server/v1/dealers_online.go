package v1

import (
	"context"
	"strconv"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// Being signed in is not the same as being on the desk.
//
// A manager holding the dealing right can watch without taking work: that is the supervisor,
// and it is what they are until they connect as a dealer. Only once connected do routing rules
// hand them requests. The connection is a key with a lease, so a dealer whose terminal dies
// stops being offered work rather than swallowing it.

const dealerOnlineTTL = 90 * time.Second

func dealerOnlineKey(login int64) string {
	return "dealers:online:" + strconv.FormatInt(login, 10)
}

// ConnectDealer puts the signed-in manager on the desk.
//
//	@Id			ConnectDealer
//	@Tags		Dealing
//	@Produce	json
//	@Success	200	{object}	Response
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/connect [post]
func (s *HttpServer) ConnectDealer(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if err := s.Redis.Client.Set(c.UserContext(), dealerOnlineKey(snap.Login),
		"1", dealerOnlineTTL).Err(); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, fiber.Map{
		"login": snap.Login, "online": true, "expires_in": int(dealerOnlineTTL.Seconds()),
	})
}

// HeartbeatDealer keeps the lease alive while the terminal is open.
//
//	@Id			HeartbeatDealer
//	@Tags		Dealing
//	@Produce	json
//	@Success	200	{object}	Response
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/heartbeat [post]
func (s *HttpServer) HeartbeatDealer(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	// only a dealer already on the desk is kept there; a lapsed one has to connect again
	held, err := s.Redis.Client.Expire(c.UserContext(),
		dealerOnlineKey(snap.Login), dealerOnlineTTL).Result()
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, fiber.Map{"login": snap.Login, "online": held})
}

// DisconnectDealer takes the signed-in manager off the desk, back to watching.
//
//	@Id			DisconnectDealer
//	@Tags		Dealing
//	@Produce	json
//	@Success	200	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/disconnect [post]
func (s *HttpServer) DisconnectDealer(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if err := s.Redis.Client.Del(c.UserContext(), dealerOnlineKey(snap.Login)).Err(); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, fiber.Map{"login": snap.Login, "online": false})
}

// GetDealerState says whether the caller is on the desk, and in what capacity.
//
//	@Id			GetDealerState
//	@Tags		Dealing
//	@Produce	json
//	@Success	200	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/state [get]
func (s *HttpServer) GetDealerState(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	online, err := s.dealerOnline(c.UserContext(), snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// a supervisor watches the whole desk; a connected dealer sees only their own work
	supervising := snap.ManagerRights.Has(model.MgrRightTradesSupervisor) && !online

	return s.App.HttpResponseOK(c, fiber.Map{
		"login": snap.Login, "online": online, "supervising": supervising,
	})
}

func (s *HttpServer) dealerOnline(ctx context.Context, login int64) (bool, error) {
	n, err := s.Redis.Client.Exists(ctx, dealerOnlineKey(login)).Result()
	return n > 0, err
}
