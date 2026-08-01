package v1

import (
	"context"
	"time"

	"hstserver/model"
	nethttp "hstserver/pkg/http"
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
	e.At = time.Now().UnixNano()
	return s.request(ctx, model.SubjectSystemDealing(e.Login), e)
}

func (s *HttpServer) confirmRequest(ctx context.Context, payload *ConfirmRequest, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	return s.sendDealing(ctx, &model.DealingEvent{
		EventType: model.DealingEventConfirm,
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
		EventType: model.DealingEventRequote,
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
		EventType: model.DealingEventReject,
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
		EventType: model.DealingEventCancel,
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
		EventType: model.DealingEventAccept,
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
		EventType: model.DealingEventReturn,
		RequestId: payload.RequestId,
		Login:     payload.Login,
		Dealer:    dealer,
	})
}
