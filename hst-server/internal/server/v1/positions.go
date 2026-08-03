package v1

import (
	"context"
	"errors"
	"fmt"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"

	"github.com/jackc/pgx/v5"
)

// UptPosition moves a position's stop loss and take profit, and nothing else.
type UptPosition struct {
	Login      int64   `json:"login" validate:"required,gt=0"`
	PositionId int64   `json:"position_id" validate:"required,gt=0"`
	PriceSL    float64 `json:"price_sl" validate:"gte=0"`
	PriceTP    float64 `json:"price_tp" validate:"gte=0"`
	Comment    string  `json:"comment" validate:"max=64"`
}

// UptMyPosition is the account's own change of levels.
type UptMyPosition struct {
	PositionId int64   `json:"position_id" validate:"required,gt=0"`
	PriceSL    float64 `json:"price_sl" validate:"gte=0"`
	PriceTP    float64 `json:"price_tp" validate:"gte=0"`
	Comment    string  `json:"comment" validate:"max=64"`
}

// ClosePosition closes a position whole or in part. Volume is in lots, zero closes all of it.
type ClosePosition struct {
	Login      int64   `json:"login" validate:"required,gt=0"`
	PositionId int64   `json:"position_id" validate:"required,gt=0"`
	Volume     float64 `json:"volume" validate:"gte=0"`
	Price      float64 `json:"price" validate:"gte=0"`
	Deviation  int64   `json:"deviation" validate:"gte=0"`
	Comment    string  `json:"comment" validate:"max=64"`
}

// CloseMyPosition is the account's own close.
type CloseMyPosition struct {
	PositionId int64   `json:"position_id" validate:"required,gt=0"`
	Volume     float64 `json:"volume" validate:"gte=0"`
	Price      float64 `json:"price" validate:"gte=0"`
	Deviation  int64   `json:"deviation" validate:"gte=0"`
	Comment    string  `json:"comment" validate:"max=64"`
}

// CloseByPosition settles two opposite positions on one instrument against each other.
type CloseByPosition struct {
	Login        int64  `json:"login" validate:"required,gt=0"`
	PositionId   int64  `json:"position_id" validate:"required,gt=0"`
	PositionById int64  `json:"position_by_id" validate:"required,gt=0"`
	Comment      string `json:"comment" validate:"max=64"`
}

// CloseByMyPosition is the account settling its own pair, which needs no dealer.
type CloseByMyPosition struct {
	PositionId   int64  `json:"position_id" validate:"required,gt=0"`
	PositionById int64  `json:"position_by_id" validate:"required,gt=0"`
	Comment      string `json:"comment" validate:"max=64"`
}

// ViewPosition is one open position as a panel renders it.
type ViewPosition struct {
	PositionId   int64                `json:"position_id"`
	Login        int64                `json:"login"`
	Symbol       string               `json:"symbol"`
	Action       model.PositionAction `json:"action"`
	Reason       model.OrderReason    `json:"reason"`
	Volume       float64              `json:"volume"`
	VolumeUnits  int64                `json:"-"`
	VolumeExt    int64                `json:"-"`
	PriceOpen    float64              `json:"price_open"`
	PriceCurrent float64              `json:"price_current"`
	PriceSL      float64              `json:"price_sl"`
	PriceTP      float64              `json:"price_tp"`
	Profit       float64              `json:"profit"`
	Storage      float64              `json:"storage"`
	TimeCreate   int64                `json:"time_create"`
	TimeUpdate   int64                `json:"time_update"`
	Comment      string               `json:"comment"`
}

const positionColumns = `p.position_id, p.login, p.symbol, p.action, p.reason, p.volume, p.volume_ext,
	p.price_open, p.price_current, p.price_sl, p.price_tp, p.profit, p.storage,
	p.time_create, p.time_update, p.comment`

const positionFrom = ` FROM hst.positions p JOIN hst.users u ON u.login = p.login WHERE `

// updatePosition changes the levels of one open position.
func (s *HttpServer) updatePosition(ctx context.Context, payload *UptPosition, dealer int64) (*Accepted, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	symbol, _, err := s.positionState(ctx, payload.PositionId, payload.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nethttp.StatusNotFound, errs.ErrNotFound
		}
		return nil, nethttp.StatusInternalServerError, err
	}

	req := &model.TradeRequest{
		RequestId:  s.newRequestId(),
		Login:      payload.Login,
		Symbol:     symbol,
		PositionId: payload.PositionId,
		PriceSL:    payload.PriceSL,
		PriceTP:    payload.PriceTP,
		Comment:    payload.Comment,
		Reason:     reasonFor(dealer),
		Dealer:     dealer,
	}

	return s.sendPosition(
		&model.PositionEvent{EventType: model.PositionEvent_update, Data: req})
}

// closePosition closes one position, whole or by volume.
func (s *HttpServer) closePosition(ctx context.Context, payload *ClosePosition, dealer int64) (*Accepted, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	symbol, volume, err := s.positionState(ctx, payload.PositionId, payload.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nethttp.StatusNotFound, errs.ErrNotFound
		}
		return nil, nethttp.StatusInternalServerError, err
	}

	// zero means all of it, which is what a client asking to close a position means
	closing := model.LotsToVolume(payload.Volume)
	if closing <= 0 || closing > volume {
		closing = volume
	}

	req := &model.TradeRequest{
		RequestId:  s.newRequestId(),
		Login:      payload.Login,
		Symbol:     symbol,
		PositionId: payload.PositionId,
		Volume:     closing,
		Price:      payload.Price,
		Deviation:  payload.Deviation,
		Comment:    payload.Comment,
		Reason:     reasonFor(dealer),
		Dealer:     dealer,
	}

	return s.sendPosition(
		&model.PositionEvent{EventType: model.PositionEvent_close, Data: req})
}

// closeByPosition settles a position against an opposite one on the same instrument.
func (s *HttpServer) closeByPosition(ctx context.Context, payload *CloseByPosition, dealer int64) (*Accepted, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}
	if payload.PositionId == payload.PositionById {
		return nil, nethttp.StatusBadRequest, errs.ErrSamePosition
	}

	symbol, _, err := s.positionState(ctx, payload.PositionId, payload.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nethttp.StatusNotFound, errs.ErrNotFound
		}
		return nil, nethttp.StatusInternalServerError, err
	}

	req := &model.TradeRequest{
		RequestId:    s.newRequestId(),
		Login:        payload.Login,
		Symbol:       symbol,
		Type:         model.OrderType_close_by,
		PositionId:   payload.PositionId,
		PositionById: payload.PositionById,
		Comment:      payload.Comment,
		Reason:       reasonFor(dealer),
		Dealer:       dealer,
	}

	return s.sendPosition(
		&model.PositionEvent{EventType: model.PositionEvent_close_by, Data: req})
}

// sendPosition hands the envelope to the pod holding this account and waits for the answer.
func (s *HttpServer) sendPosition(e *model.PositionEvent) (*Accepted, int, error) {
	status, err := s.publish(model.SubjectSystemPositions, e)
	if err != nil {
		return nil, status, err
	}

	return acceptedPosition(e), status, nil
}

// acceptedPosition says what was asked of a position.
func acceptedPosition(e *model.PositionEvent) *Accepted {
	return &Accepted{
		RequestId: e.Data.RequestId,
		Login:     e.Data.Login,
		Message: fmt.Sprintf("position %s requested, #%d",
			model.PositionEvent_name[int32(e.EventType)], e.Data.PositionId),
	}
}

// positionState is what the write paths need before they may ask.
func (s *HttpServer) positionState(ctx context.Context, positionId, login int64) (symbol string, volume int64, err error) {
	err = s.DB.DB.QueryRow(ctx,
		`SELECT symbol, GREATEST(volume_ext, volume * 10000) FROM hst.positions WHERE position_id = $1 AND login = $2`,
		positionId, login).Scan(&symbol, &volume)
	return symbol, volume, err
}

// readPositions is the one query behind every position read handler.
func (s *HttpServer) readPositions(ctx context.Context, where string, args []any) ([]ViewPosition, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+positionColumns+positionFrom+where+` ORDER BY p.position_id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewPosition{}
	for rows.Next() {
		var v ViewPosition
		if err := rows.Scan(&v.PositionId, &v.Login, &v.Symbol, &v.Action, &v.Reason, &v.VolumeUnits, &v.VolumeExt,
			&v.PriceOpen, &v.PriceCurrent, &v.PriceSL, &v.PriceTP, &v.Profit, &v.Storage,
			&v.TimeCreate, &v.TimeUpdate, &v.Comment); err != nil {
			return nil, err
		}
		v.Volume = model.ExtToLots(model.ExtendedVolume(v.VolumeUnits, v.VolumeExt))
		out = append(out, v)
	}

	return out, rows.Err()
}

func uptPositionFromMy(p *UptMyPosition, login int64) *UptPosition {
	return &UptPosition{
		Login:      login,
		PositionId: p.PositionId,
		PriceSL:    p.PriceSL,
		PriceTP:    p.PriceTP,
		Comment:    p.Comment,
	}
}

func closeByFromMy(p *CloseByMyPosition, login int64) *CloseByPosition {
	return &CloseByPosition{
		Login:        login,
		PositionId:   p.PositionId,
		PositionById: p.PositionById,
		Comment:      p.Comment,
	}
}

func closeFromMy(p *CloseMyPosition, login int64) *ClosePosition {
	return &ClosePosition{
		Login:      login,
		PositionId: p.PositionId,
		Volume:     p.Volume,
		Price:      p.Price,
		Deviation:  p.Deviation,
		Comment:    p.Comment,
	}
}
