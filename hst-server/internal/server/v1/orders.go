package v1

import (
	"context"
	"errors"
	"time"

	"hstserver/model"
	"hstserver/pkg/cache"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
	"hstserver/utils"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// TradeResult is the engine's answer, aliased for swagger.
type TradeResult = model.TradeResult

// engineTimeout is longer than any routing delay rule is likely to ask for.
const engineTimeout = 10 * time.Second

// CrtOrder is a manager placing an order on behalf of a named login. Volume is in lots.
type CrtOrder struct {
	Login        int64   `json:"login" validate:"required,gt=0"`
	Symbol       string  `json:"symbol" validate:"required,max=32"`
	Type         int32   `json:"type" validate:"gte=0,lte=7"`
	Volume       float64 `json:"volume" validate:"required,gt=0"`
	Price        float64 `json:"price" validate:"gte=0"`
	PriceTrigger float64 `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64 `json:"price_sl" validate:"gte=0"`
	PriceTP      float64 `json:"price_tp" validate:"gte=0"`
	TypeFill     int32   `json:"type_fill" validate:"gte=0,lte=3"`
	TypeTime     int32   `json:"type_time" validate:"gte=0,lte=3"`
	Expiry       int64   `json:"expiry"`
	Deviation    int64   `json:"deviation" validate:"gte=0"`
	Comment      string  `json:"comment" validate:"max=64"`
	ExpertId     int64   `json:"expert_id"`
}

// CrtMyOrder is the same order placed by the account itself.
type CrtMyOrder struct {
	Symbol       string  `json:"symbol" validate:"required,max=32"`
	Type         int32   `json:"type" validate:"gte=0,lte=7"`
	Volume       float64 `json:"volume" validate:"required,gt=0"`
	Price        float64 `json:"price" validate:"gte=0"`
	PriceTrigger float64 `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64 `json:"price_sl" validate:"gte=0"`
	PriceTP      float64 `json:"price_tp" validate:"gte=0"`
	TypeFill     int32   `json:"type_fill" validate:"gte=0,lte=3"`
	TypeTime     int32   `json:"type_time" validate:"gte=0,lte=3"`
	Expiry       int64   `json:"expiry"`
	Deviation    int64   `json:"deviation" validate:"gte=0"`
	Comment      string  `json:"comment" validate:"max=64"`
	ExpertId     int64   `json:"expert_id"`
}

// UptOrder modifies one working pending order.
type UptOrder struct {
	Login        int64   `json:"login" validate:"required,gt=0"`
	OrderId      int64   `json:"order_id" validate:"required,gt=0"`
	Price        float64 `json:"price" validate:"gte=0"`
	PriceTrigger float64 `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64 `json:"price_sl" validate:"gte=0"`
	PriceTP      float64 `json:"price_tp" validate:"gte=0"`
	TypeTime     int32   `json:"type_time" validate:"gte=0,lte=3"`
	Expiry       int64   `json:"expiry"`
	Comment      string  `json:"comment" validate:"max=64"`
}

// UptMyOrder is the account's own modification.
type UptMyOrder struct {
	OrderId      int64   `json:"order_id" validate:"required,gt=0"`
	Price        float64 `json:"price" validate:"gte=0"`
	PriceTrigger float64 `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64 `json:"price_sl" validate:"gte=0"`
	PriceTP      float64 `json:"price_tp" validate:"gte=0"`
	TypeTime     int32   `json:"type_time" validate:"gte=0,lte=3"`
	Expiry       int64   `json:"expiry"`
	Comment      string  `json:"comment" validate:"max=64"`
}

// CancelOrder removes one working pending order.
type CancelOrder struct {
	Login   int64  `json:"login" validate:"required,gt=0"`
	OrderId int64  `json:"order_id" validate:"required,gt=0"`
	Comment string `json:"comment" validate:"max=64"`
}

// CancelMyOrder is the account's own removal.
type CancelMyOrder struct {
	OrderId int64  `json:"order_id" validate:"required,gt=0"`
	Comment string `json:"comment" validate:"max=64"`
}

// ViewOrder is one order as a panel renders it. Volume is in lots.
type ViewOrder struct {
	OrderId        int64   `json:"order_id"`
	Login          int64   `json:"login"`
	Symbol         string  `json:"symbol"`
	Type           int32   `json:"type"`
	State          int32   `json:"state"`
	Reason         int32   `json:"reason"`
	Volume         float64 `json:"volume"`
	VolumeInitial  int64   `json:"-"`
	VolumeCurrent  int64   `json:"-"`
	VolumeExt      int64   `json:"-"`
	PriceOrder     float64 `json:"price_order"`
	PriceTrigger   float64 `json:"price_trigger"`
	PriceCurrent   float64 `json:"price_current"`
	PriceSL        float64 `json:"price_sl"`
	PriceTP        float64 `json:"price_tp"`
	TimeSetup      int64   `json:"time_setup"`
	TimeExpiration int64   `json:"time_expiration"`
	Comment        string  `json:"comment"`
}

const orderColumns = `o.order_id, o.login, o.symbol, o.type, o.state, o.reason,
	o.volume_initial, o.volume_current, o.volume_current_ext, o.price_order, o.price_trigger, o.price_current,
	o.price_sl, o.price_tp, o.time_setup, o.time_expiration, o.comment`

const orderFrom = ` FROM hst.orders o JOIN hst.users u ON u.login = o.login WHERE `

// makeOrder validates a new order and hands it to the engine.
func (s *HttpServer) makeOrder(ctx context.Context, payload *CrtOrder, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	if _, ok := model.OrderType_name[payload.Type]; !ok {
		return nil, nethttp.StatusBadRequest, errs.ErrInvalidOrderType
	}

	if model.OrderType(payload.Type).IsPending() && payload.Price <= 0 {
		return nil, nethttp.StatusBadRequest, errs.ErrPendingNeedsPrice
	}

	if model.OrderTime(payload.TypeTime) != model.OrderTime_gtc &&
		model.OrderTime(payload.TypeTime) != model.OrderTime_day && payload.Expiry <= 0 {
		return nil, nethttp.StatusBadRequest, errs.ErrExpiryRequired
	}

	req := &model.TradeRequest{
		RequestId:    s.newRequestId(),
		Login:        payload.Login,
		Symbol:       payload.Symbol,
		Type:         payload.Type,
		Volume:       model.LotsToVolume(payload.Volume),
		Price:        payload.Price,
		PriceTrigger: payload.PriceTrigger,
		PriceSL:      payload.PriceSL,
		PriceTP:      payload.PriceTP,
		TypeFill:     payload.TypeFill,
		TypeTime:     payload.TypeTime,
		Expiry:       payload.Expiry,
		Deviation:    payload.Deviation,
		Comment:      payload.Comment,
		ExpertId:     payload.ExpertId,
		Reason:       reasonFor(dealer),
		Dealer:       dealer,
	}

	return s.sendOrder(ctx, payload.Login, &model.OrderEvent{EventType: model.OrderEventNew, Data: req})
}

// updateOrder moves a working pending order's price, levels and expiry.
func (s *HttpServer) updateOrder(ctx context.Context, payload *UptOrder, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	typ, state, symbol, err := s.orderState(ctx, payload.OrderId, payload.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nethttp.StatusNotFound, errs.ErrNotFound
		}
		return nil, nethttp.StatusInternalServerError, err
	}

	if !model.OrderType(typ).IsPending() {
		return nil, nethttp.StatusBadRequest, errs.ErrMarketOrderNotModifiable
	}
	if !model.OrderState(state).IsLive() {
		return nil, nethttp.StatusBadRequest, errs.ErrOrderNotWorking
	}

	req := &model.TradeRequest{
		RequestId:    s.newRequestId(),
		Login:        payload.Login,
		Symbol:       symbol,
		Type:         typ,
		OrderId:      payload.OrderId,
		Price:        payload.Price,
		PriceTrigger: payload.PriceTrigger,
		PriceSL:      payload.PriceSL,
		PriceTP:      payload.PriceTP,
		TypeTime:     payload.TypeTime,
		Expiry:       payload.Expiry,
		Comment:      payload.Comment,
		Reason:       reasonFor(dealer),
		Dealer:       dealer,
	}

	return s.sendOrder(ctx, payload.Login, &model.OrderEvent{EventType: model.OrderEventUpdate, Data: req})
}

// cancelOrder removes a working pending order.
func (s *HttpServer) cancelOrder(ctx context.Context, payload *CancelOrder, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	typ, state, symbol, err := s.orderState(ctx, payload.OrderId, payload.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nethttp.StatusNotFound, errs.ErrNotFound
		}
		return nil, nethttp.StatusInternalServerError, err
	}

	if !model.OrderState(state).IsLive() {
		return nil, nethttp.StatusBadRequest, errs.ErrOrderNotWorking
	}

	req := &model.TradeRequest{
		RequestId: s.newRequestId(),
		Login:     payload.Login,
		Symbol:    symbol,
		Type:      typ,
		OrderId:   payload.OrderId,
		Comment:   payload.Comment,
		Reason:    reasonFor(dealer),
		Dealer:    dealer,
	}

	return s.sendOrder(ctx, payload.Login, &model.OrderEvent{EventType: model.OrderEventCancel, Data: req})
}

// sendOrder hands the envelope to the pod holding this account and waits for the answer.
func (s *HttpServer) sendOrder(ctx context.Context, login int64, e *model.OrderEvent) (*model.TradeResult, int, error) {
	return s.request(ctx, model.SubjectSystemOrders(login), e)
}

// request is the one call into the engine: marshal, ask, decode, map the retcode to a status.
func (s *HttpServer) request(ctx context.Context, subject string, e any) (*model.TradeResult, int, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, nethttp.StatusInternalServerError, err
	}

	ctx, cancel := context.WithTimeout(ctx, engineTimeout)
	defer cancel()

	msg, err := s.Nats.NC.RequestWithContext(ctx, subject, payload)
	if err != nil {
		// no pod answered: the client must not be told the request went through
		return nil, nethttp.StatusServiceUnavailable, errs.ErrEngineUnavailable
	}

	var res model.TradeResult
	if err := json.Unmarshal(msg.Data, &res); err != nil {
		return nil, nethttp.StatusInternalServerError, err
	}

	// the engine's refusal is the answer, not an error in this server
	if res.RetCode != 0 {
		return &res, nethttp.StatusBadRequest, nil
	}

	return &res, nethttp.StatusOK, nil
}

// orderState is what the modify and cancel paths need before they may ask.
func (s *HttpServer) orderState(ctx context.Context, orderId, login int64) (typ, state int32, symbol string, err error) {
	err = s.DB.DB.QueryRow(ctx,
		`SELECT type, state, symbol FROM hst.orders WHERE order_id = $1 AND login = $2`,
		orderId, login).Scan(&typ, &state, &symbol)
	return typ, state, symbol, err
}

// readOrders is the one query behind every order read handler.
func (s *HttpServer) readOrders(ctx context.Context, where string, args []any, p pageOpts) ([]ViewOrder, error) {
	where, args = p.bound("o.time_setup", where, args)

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+orderColumns+orderFrom+where+p.tail("o.order_id"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewOrder{}
	for rows.Next() {
		var v ViewOrder
		if err := rows.Scan(&v.OrderId, &v.Login, &v.Symbol, &v.Type, &v.State, &v.Reason,
			&v.VolumeInitial, &v.VolumeCurrent, &v.VolumeExt, &v.PriceOrder, &v.PriceTrigger, &v.PriceCurrent,
			&v.PriceSL, &v.PriceTP, &v.TimeSetup, &v.TimeExpiration, &v.Comment); err != nil {
			return nil, err
		}
		v.Volume = model.ExtToLots(model.ExtendedVolume(v.VolumeCurrent, v.VolumeExt))
		out = append(out, v)
	}

	return out, rows.Err()
}

// newRequestId is the handle a client matches an answer to.
func (s *HttpServer) newRequestId() string {
	return time.Now().Format("20060102150405.000000000")
}

// crtFromMy fills the login from the session, which is the only difference between the two forms.
func crtFromMy(p *CrtMyOrder, login int64) *CrtOrder {
	return &CrtOrder{
		Login:        login,
		Symbol:       p.Symbol,
		Type:         p.Type,
		Volume:       p.Volume,
		Price:        p.Price,
		PriceTrigger: p.PriceTrigger,
		PriceSL:      p.PriceSL,
		PriceTP:      p.PriceTP,
		TypeFill:     p.TypeFill,
		TypeTime:     p.TypeTime,
		Expiry:       p.Expiry,
		Deviation:    p.Deviation,
		Comment:      p.Comment,
		ExpertId:     p.ExpertId,
	}
}

func uptFromMy(p *UptMyOrder, login int64) *UptOrder {
	return &UptOrder{
		Login:        login,
		OrderId:      p.OrderId,
		Price:        p.Price,
		PriceTrigger: p.PriceTrigger,
		PriceSL:      p.PriceSL,
		PriceTP:      p.PriceTP,
		TypeTime:     p.TypeTime,
		Expiry:       p.Expiry,
		Comment:      p.Comment,
	}
}

// answer writes the engine's reply, carrying the result even when it is a refusal.
func (s *HttpServer) answer(c *fiber.Ctx, res *model.TradeResult, status int, err error) error {
	if err != nil {
		if status == nethttp.StatusServiceUnavailable {
			return s.App.HttpResponseServiceUnavailable(c, err)
		}
		return s.App.HttpResponseStatus(c, status, err)
	}

	// a requote or a refusal carries the prices the client needs, so the body is the result either way
	if res.RetCode != 0 {
		return c.Status(status).JSON(&nethttp.HttpResponse{
			Code:    nethttp.RetCode(res.RetCode),
			Data:    res,
			Error:   nethttp.ErrBadRequest,
			Message: res.Message,
		})
	}

	return s.App.HttpResponseOK(c, res)
}

// inReach answers false and writes the refusal itself when the account is outside the manager's masks.
func (s *HttpServer) inReach(c *fiber.Ctx, snap *cache.Snapshot, login int64) (bool, error) {
	if login <= 0 {
		return false, s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	ok, err := s.accountInReach(c.UserContext(), snap.IsManager, snap.ManagerGroups, login)
	if err != nil {
		return false, s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if !ok {
		return false, s.App.HttpResponseForbidden(c, errs.ErrAccountOutOfReach)
	}

	return true, nil
}

// groupWhere is the manager's group masks as a predicate on the joined user row.
func groupWhere(isManager bool, masks []string, next int) (string, []any) {
	return utils.GroupAccessFor(isManager, masks, `u."group"`, next)
}

// AccountInReach reports whether the caller's masks cover the login, and writes nothing.
//
// The by id handlers need the answer, not a response: a helper that writes the refusal returns
// nil for having written it, and a caller testing that error would read the refusal as success.
func (s *HttpServer) AccountInReach(c *fiber.Ctx, login int64) (bool, error) {
	snap, ok := utils.GetClient(c)
	if !ok {
		return false, errs.ErrCouldNotParseClientCfg
	}

	return s.accountInReach(c.UserContext(), snap.IsManager, snap.ManagerGroups, login)
}

// accountInReach reports whether the masks cover the account they named.
func (s *HttpServer) accountInReach(ctx context.Context, isManager bool, masks []string, login int64) (bool, error) {
	where, args := groupWhere(isManager, masks, 2)

	var ok bool
	err := s.DB.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM hst.users u WHERE u.login = $1 AND `+where+`)`,
		append([]any{login}, args...)...).Scan(&ok)
	return ok, err
}

// writable refuses a session that authenticated with an investor password.
func writable(snap *cache.Snapshot) error {
	if snap.Scope == int32(model.UsersPasswords_investor) {
		return errs.ErrReadOnlySession
	}
	return nil
}

// reasonFor records how the request arrived, which the routing rules can key on.
func reasonFor(dealer int64) int32 {
	if dealer != 0 {
		return int32(model.OrderReason_dealer)
	}
	return int32(model.OrderReason_client)
}
