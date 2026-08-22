package handler

import (
	"fmt"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
)

// The ledger rows a fill produces.

// MakeDealIn is the row for volume that opened or grew a position.
func (h *Handler) MakeDealIn(o *model.Order, r *settings.Rules, e *book.Entry, price float64,
	volume, positionId, now int64) *model.Deal {
	return h.dealFor(o, r, e, price, volume, 0, model.DealEntry_in, positionId, now)
}

// MakeDealOut is the row for volume that closed or shrank a position.
func (h *Handler) MakeDealOut(o *model.Order, r *settings.Rules, e *book.Entry,
	p *model.Position, price float64, closed, now int64) *model.Deal {
	entry := model.DealEntry_out
	if o.VolumeCurrent > closed {
		entry = model.DealEntry_inout
	}

	d := h.dealFor(o, r, e, price, o.VolumeCurrent, closed, entry, p.PositionId, now)
	d.PricePosition = p.PriceOpen
	d.Storage = storageShare(p, closed)

	return d
}

// storageShare is the swap a position has accrued on the volume now closing, all of it for a full close.
func storageShare(p *model.Position, closed int64) float64 {
	if p.Volume <= 0 || closed >= p.Volume {
		return p.Storage
	}
	return NormalisePrice(p.Storage*float64(closed)/float64(p.Volume), p.DigitsCurrency)
}

// MakeDealOutBy is the row for one leg of a close against an opposite position.
func (h *Handler) MakeDealOutBy(o *model.Order, r *settings.Rules, e *book.Entry,
	p, by *model.Position, price float64, volume, now int64) *model.Deal {
	d := h.dealFor(o, r, e, price, volume, volume, model.DealEntry_out_by, p.PositionId, now)

	d.PricePosition = p.PriceOpen
	d.Storage = storageShare(p, volume)
	d.Comment = fmt.Sprintf("close hedge by #%d", by.PositionId)

	// the side is the position being taken off, not the order that asked for it
	d.Action = model.DealAction_sell
	if p.IsBuy() {
		d.Action = model.DealAction_buy
	}

	return d
}

// dealFor is the shared row every builder above shapes.
func (h *Handler) dealFor(o *model.Order, r *settings.Rules, e *book.Entry, price float64,
	volume, closed int64, entry model.DealEntry, positionId, now int64) *model.Deal {
	action := model.DealAction_buy
	if !o.Kind().IsBuy() {
		action = model.DealAction_sell
	}

	return &model.Deal{
		Login:          o.Login,
		Dealer:         o.Dealer,
		OrderId:        o.OrderId,
		Action:         action,
		Entry:          entry,
		Digits:         r.Digits,
		DigitsCurrency: e.Account.CurrencyDigits,
		ContractSize:   r.ContractSize,
		Time:           now,
		Symbol:         o.Symbol,
		Price:          price,
		PriceSL:        o.PriceSL,
		PriceTP:        o.PriceTP,
		Volume:         volume,
		VolumeClosed:   closed,
		RateProfit:     h.RateProfit(r, e.Account, o.Kind().IsBuy()),
		RateMargin:     o.RateMargin,
		ExpertId:       o.ExpertId,
		PositionId:     positionId,
		Comment:        o.Comment,
		TickValue:      r.TickValue,
		TickSize:       r.TickSize,
		Reason:         o.Reason,
	}
}
