package v1

import (
	"context"
	"errors"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"

	"github.com/jackc/pgx/v5"
)

// ConfirmRequest fills the queued request.
type ConfirmRequest struct {
	RequestId string  `json:"request_id" validate:"required,max=64"`
	Login     int64   `json:"login" validate:"required,gt=0"`
	Price     float64 `json:"price" validate:"gte=0"`
}

// RequoteRequest offers the client a new price.
type RequoteRequest struct {
	RequestId string  `json:"request_id" validate:"required,max=64"`
	Login     int64   `json:"login" validate:"required,gt=0"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}

// RejectRequest refuses with a reason the client sees.
type RejectRequest struct {
	RequestId string `json:"request_id" validate:"required,max=64"`
	Login     int64  `json:"login" validate:"required,gt=0"`
	Reason    string `json:"reason" validate:"max=31"`
}

// CancelRequest removes the pending order the request was for.
type CancelRequest struct {
	RequestId string `json:"request_id" validate:"required,max=64"`
	Login     int64  `json:"login" validate:"required,gt=0"`
	Reason    string `json:"reason" validate:"max=31"`
}

// AcceptRequote is the client taking the price a dealer offered them back.
type AcceptRequote struct {
	RequestId string `json:"request_id" validate:"required,max=64"`
	Login     int64  `json:"login" validate:"required,gt=0"`
}

// ReturnRequestBody is a dealer putting the request back on the desk.
type ReturnRequestBody struct {
	RequestId string `json:"request_id" validate:"required,max=64"`
	Login     int64  `json:"login" validate:"required,gt=0"`
}

// sendDealing hands the dealer's answer to the pod holding the account.
func (s *HttpServer) sendDealing(ctx context.Context, e *model.DealingEvent) (*model.TradeResult, int, error) {
	if status, err := s.bindRequest(ctx, e.RequestId, e.Login); err != nil {
		return nil, status, err
	}
	e.At = time.Now().UnixNano()
	return s.request(ctx, model.SubjectSystemDealing, e)
}

// bindRequest refuses a request id that is not a waiting request of the named login, since the engine looks it up by id alone.
func (s *HttpServer) bindRequest(ctx context.Context, requestId string, login int64) (int, error) {
	var one int
	err := s.DB.DB.QueryRow(ctx,
		`SELECT 1 FROM hst.orders WHERE (order_id::text = $1 OR external_id = $1) AND login = $2 AND state = ANY($3)`,
		requestId, login, []int32{
			int32(model.OrderState_request_add),
			int32(model.OrderState_request_modify),
			int32(model.OrderState_request_cancel),
		}).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return nethttp.StatusNotFound, errs.ErrNotFound
	}
	if err != nil {
		return nethttp.StatusInternalServerError, err
	}
	return 0, nil
}

func (s *HttpServer) confirmRequest(ctx context.Context, payload *ConfirmRequest, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	return s.sendDealing(ctx, &model.DealingEvent{
		EventType: model.DealingEvent_confirm,
		RequestId: payload.RequestId,
		Login:     payload.Login,
		Dealer:    dealer,
		Price:     payload.Price,
	})
}

func (s *HttpServer) requoteRequest(ctx context.Context, payload *RequoteRequest, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	return s.sendDealing(ctx, &model.DealingEvent{
		EventType: model.DealingEvent_requote,
		RequestId: payload.RequestId,
		Login:     payload.Login,
		Dealer:    dealer,
		Price:     payload.Price,
	})
}

func (s *HttpServer) rejectRequest(ctx context.Context, payload *RejectRequest, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	return s.sendDealing(ctx, &model.DealingEvent{
		EventType: model.DealingEvent_reject,
		RequestId: payload.RequestId,
		Login:     payload.Login,
		Dealer:    dealer,
		Reason:    payload.Reason,
	})
}

func (s *HttpServer) cancelRequest(ctx context.Context, payload *CancelRequest, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	return s.sendDealing(ctx, &model.DealingEvent{
		EventType: model.DealingEvent_cancel,
		RequestId: payload.RequestId,
		Login:     payload.Login,
		Dealer:    dealer,
		Reason:    payload.Reason,
	})
}

func (s *HttpServer) acceptRequote(ctx context.Context, payload *AcceptRequote) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	return s.sendDealing(ctx, &model.DealingEvent{
		EventType: model.DealingEvent_accept,
		RequestId: payload.RequestId,
		Login:     payload.Login,
	})
}

func (s *HttpServer) returnRequest(ctx context.Context, payload *ReturnRequestBody,
	dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	return s.sendDealing(ctx, &model.DealingEvent{
		EventType: model.DealingEvent_return,
		RequestId: payload.RequestId,
		Login:     payload.Login,
		Dealer:    dealer,
	})
}
