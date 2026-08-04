package v1

import (
	"context"
	"errors"
	"fmt"
	"hstserver/pkg/logger"
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
	Login        int64              `json:"login" validate:"required,gt=0"`
	Symbol       string             `json:"symbol" validate:"required,max=32"`
	Type         model.OrderType    `json:"type"`
	Volume       float64            `json:"volume" validate:"required,gt=0"`
	Price        float64            `json:"price" validate:"gte=0"`
	PriceTrigger float64            `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64            `json:"price_sl" validate:"gte=0"`
	PriceTP      float64            `json:"price_tp" validate:"gte=0"`
	TypeFill     model.OrderFilling `json:"type_fill"`
	TypeTime     model.OrderTime    `json:"type_time"`
	ExpiryAt     *int64             `json:"expiry_at"`
	Deviation    int64              `json:"deviation" validate:"gte=0"`
	Comment      string             `json:"comment" validate:"max=64"`
	ExpertId     int64              `json:"expert_id"`
}

// CrtMyOrder is the same order placed by the account itself.
type CrtMyOrder struct {
	Symbol       string             `json:"symbol" validate:"required,max=32"`
	Type         model.OrderType    `json:"type"`
	Volume       float64            `json:"volume" validate:"required,gt=0"`
	Price        float64            `json:"price" validate:"gte=0"`
	PriceTrigger float64            `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64            `json:"price_sl" validate:"gte=0"`
	PriceTP      float64            `json:"price_tp" validate:"gte=0"`
	TypeFill     model.OrderFilling `json:"type_fill"`
	TypeTime     model.OrderTime    `json:"type_time"`
	ExpiryAt     *int64             `json:"expiry_at"`
	Deviation    int64              `json:"deviation" validate:"gte=0"`
	Comment      string             `json:"comment" validate:"max=64"`
	ExpertId     int64              `json:"expert_id"`
}

// UptOrder modifies one working pending order.
type UptOrder struct {
	Login        int64           `json:"login" validate:"required,gt=0"`
	OrderId      int64           `json:"order_id" validate:"required,gt=0"`
	Price        float64         `json:"price" validate:"gte=0"`
	PriceTrigger float64         `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64         `json:"price_sl" validate:"gte=0"`
	PriceTP      float64         `json:"price_tp" validate:"gte=0"`
	TypeTime     model.OrderTime `json:"type_time"`
	ExpiryAt     *int64          `json:"expiry_at"`
	Comment      string          `json:"comment" validate:"max=64"`
}

// UptMyOrder is the account's own modification.
type UptMyOrder struct {
	OrderId      int64           `json:"order_id" validate:"required,gt=0"`
	Price        float64         `json:"price" validate:"gte=0"`
	PriceTrigger float64         `json:"price_trigger" validate:"gte=0"`
	PriceSL      float64         `json:"price_sl" validate:"gte=0"`
	PriceTP      float64         `json:"price_tp" validate:"gte=0"`
	TypeTime     model.OrderTime `json:"type_time"`
	ExpiryAt     *int64          `json:"expiry_at"`
	Comment      string          `json:"comment" validate:"max=64"`
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
	OrderId        int64             `json:"order_id"`
	Login          int64             `json:"login"`
	Symbol         string            `json:"symbol"`
	Type           model.OrderType   `json:"type"`
	State          model.OrderState  `json:"state"`
	Reason         model.OrderReason `json:"reason"`
	Volume         float64           `json:"volume"`
	VolumeInitial  int64             `json:"-"`
	VolumeCurrent  int64             `json:"-"`
	VolumeExt      int64             `json:"-"`
	PriceOrder     float64           `json:"price_order"`
	PriceTrigger   float64           `json:"price_trigger"`
	PriceCurrent   float64           `json:"price_current"`
	PriceSL        float64           `json:"price_sl"`
	PriceTP        float64           `json:"price_tp"`
	TimeSetup      int64             `json:"time_setup"`
	TimeExpiration int64             `json:"time_expiration"`
	Comment        string            `json:"comment"`
}

const orderColumns = `o.order_id, o.login, o.symbol, o.type, o.state, o.reason,
	o.volume_initial, o.volume_current, o.volume_current_ext, o.price_order, o.price_trigger, o.price_current,
	o.price_sl, o.price_tp, o.time_setup, o.time_expiration, o.comment`

const orderFrom = ` FROM hst.orders o JOIN hst.users u ON u.login = o.login WHERE `

// checkOrderEnums refuses a value that is not a member of its enum, which a numeric bound cannot do.
func checkOrderEnums(t model.OrderType, f model.OrderFilling, tt model.OrderTime) (int, error) {
	if !model.Valid(t, model.OrderType_name) {
		return nethttp.StatusBadRequest, errs.ErrInvalidOrderType
	}
	if !model.Valid(f, model.OrderFilling_name) {
		return nethttp.StatusBadRequest, errs.ErrInvalidOrderFilling
	}
	if !model.Valid(tt, model.OrderTime_name) {
		return nethttp.StatusBadRequest, errs.ErrInvalidOrderTime
	}

	return 0, nil
}

// expiryOf reads an optional moment; absent means the order does not expire on its own.
func expiryOf(at *int64) int64 {
	if at == nil {
		return 0
	}

	return *at
}

// makeOrder validates a new order and hands it to the engine.
func (s *HttpServer) makeOrder(ctx context.Context, payload *CrtOrder, dealer int64) (*Accepted, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	if code, err := checkOrderEnums(payload.Type, payload.TypeFill, payload.TypeTime); err != nil {
		return nil, code, err
	}

	if payload.Type.IsPending() && payload.Price <= 0 {
		return nil, nethttp.StatusBadRequest, errs.ErrPendingNeedsPrice
	}

	if payload.TypeTime.NeedsExpiry() && payload.ExpiryAt == nil {
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
		ExpiryAt:     expiryOf(payload.ExpiryAt),
		Deviation:    payload.Deviation,
		Comment:      payload.Comment,
		ExpertId:     payload.ExpertId,
		Reason:       reasonFor(dealer),
		Dealer:       dealer,
	}

	return s.sendOrder(&model.OrderEvent{EventType: model.OrderEvent_new_order, Data: req})
}

// updateOrder moves a working pending order's price, levels and expiry.
func (s *HttpServer) updateOrder(ctx context.Context, payload *UptOrder, dealer int64) (*Accepted, int, error) {
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

	if code, err := checkOrderEnums(typ, model.OrderFilling_fok, payload.TypeTime); err != nil {
		return nil, code, err
	}

	if payload.TypeTime.NeedsExpiry() && payload.ExpiryAt == nil {
		return nil, nethttp.StatusBadRequest, errs.ErrExpiryRequired
	}

	if !typ.IsPending() {
		return nil, nethttp.StatusBadRequest, errs.ErrMarketOrderNotModifiable
	}
	if !state.IsLive() {
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
		ExpiryAt:     expiryOf(payload.ExpiryAt),
		Comment:      payload.Comment,
		Reason:       reasonFor(dealer),
		Dealer:       dealer,
	}

	return s.sendOrder(&model.OrderEvent{EventType: model.OrderEvent_update_order, Data: req})
}

// cancelOrder removes a working pending order.
func (s *HttpServer) cancelOrder(ctx context.Context, payload *CancelOrder, dealer int64) (*Accepted, int, error) {
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

	if !state.IsLive() {
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

	return s.sendOrder(&model.OrderEvent{EventType: model.OrderEvent_cancel_order, Data: req})
}

// sendOrder hands the envelope to the engine. The outcome arrives on the account's socket.
func (s *HttpServer) sendOrder(e *model.OrderEvent) (*Accepted, int, error) {
	status, err := s.publish(model.SubjectSystemOrders, e)
	if err != nil {
		return nil, status, err
	}

	return acceptedOrder(e.Data), status, nil
}

// Accepted is what a caller gets back when the engine has been handed a request.
type Accepted struct {
	RequestId string `json:"request_id"`
	Login     int64  `json:"login"`
	Message   string `json:"message"`
}

// acceptedOrder says what was asked for, in the words the journal uses.
func acceptedOrder(req *model.TradeRequest) *Accepted {
	return &Accepted{
		RequestId: req.RequestId,
		Login:     req.Login,
		Message: fmt.Sprintf("order requested, type %s volume %.2f symbol %s",
			model.OrderType_name[int32(req.Type)], model.ExtToLots(req.Volume), req.Symbol),
	}
}

// publish hands a command to the engine without waiting for it.
//
// A trade is not a question with an answer, it is a request that the engine will accept or refuse
// in its own time, and the outcome arrives on the socket as order_create or order_rejected. The
// api only reports whether it managed to hand the request over.
func (s *HttpServer) publish(subject string, e any) (int, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nethttp.StatusInternalServerError, err
	}

	if err := s.Nats.NC.Publish(subject, payload); err != nil {
		// nothing took the request, so the client must not be told it went through
		return nethttp.StatusServiceUnavailable, errs.ErrEngineUnavailable
	}

	// a publish only reaches the connection buffer, and an order lost there is a silent loss
	if err := s.Nats.NC.Flush(); err != nil {
		return nethttp.StatusServiceUnavailable, errs.ErrEngineUnavailable
	}

	return nethttp.StatusAccepted, nil
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
func (s *HttpServer) orderState(ctx context.Context, orderId, login int64) (typ model.OrderType, state model.OrderState, symbol string, err error) {
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
		ExpiryAt:     p.ExpiryAt,
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
		ExpiryAt:     p.ExpiryAt,
		Comment:      p.Comment,
	}
}

// accepted writes the acknowledgement, and journals what was asked for.
//
// The engine has the request; whether it fills is told on the socket, so there is no retcode here
// and a caller that needs the outcome must listen for it.
func (s *HttpServer) accepted(c *fiber.Ctx, res *Accepted, status int, err error) error {
	s.journalAsked(c, res, err)

	if err != nil {
		if status == nethttp.StatusServiceUnavailable {
			return s.App.HttpResponseServiceUnavailable(c, err)
		}

		return s.App.HttpResponseStatus(c, status, err)
	}

	return s.App.HttpResponseAccepted(c, res)
}

// journalAsked records the request, since the outcome is journalled by the engine's own events.
func (s *HttpServer) journalAsked(c *fiber.Ctx, res *Accepted, err error) {
	code, outcome := logger.CodeOK, ""
	if res != nil {
		outcome = res.Message
	}
	if err != nil {
		code, outcome = logger.CodeErr, err.Error()
	}

	login := int64(0)
	if res != nil {
		login = res.Login
	}

	s.JournalEntry(c, logger.TypeTrade, code, fmt.Sprintf("%s %s for #%d: %s", c.Method(), c.Path(), login, outcome), res)
}

// answer writes the engine's reply, carrying the result even when it is a refusal.
//
// Every trading route ends here, so this is where the journal line is written: an order, a
// position, a dealer's answer and a balance operation are all audited by one call rather than by
// remembering to add one to each handler.
func (s *HttpServer) answer(c *fiber.Ctx, res *model.TradeResult, status int, err error) error {
	s.journalTrade(c, res, err)

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

// journalTrade records what was asked for and what came back.
func (s *HttpServer) journalTrade(c *fiber.Ctx, res *model.TradeResult, err error) {
	code := logger.CodeOK
	outcome := "done"

	switch {
	case err != nil:
		code, outcome = logger.CodeErr, err.Error()
	case res == nil:
		return
	case res.RetCode != 0:
		code, outcome = logger.CodeAtt, res.Message
	}

	login := int64(0)
	if res != nil {
		login = res.Login
	}

	s.JournalEntry(c, logger.TypeTrade, code, fmt.Sprintf("%s %s for #%d: %s",
		c.Method(), c.Path(), login, outcome), res)
}

// inReach answers false and writes the refusal itself when the account is outside the manager's masks.
func (s *HttpServer) inReach(c *fiber.Ctx, snap *cache.Session, login int64) (bool, error) {
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

// isReadOnlyScope reports whether the session authenticated with the investor password, which may
// see everything and change nothing.
func isReadOnlyScope(scope int32) bool {
	return scope == int32(model.UsersPasswords_investor)
}

// reasonFor records how the request arrived, which the routing rules can key on.
func reasonFor(dealer int64) model.OrderReason {
	if dealer != 0 {
		return model.OrderReason_dealer
	}

	return model.OrderReason_client
}
