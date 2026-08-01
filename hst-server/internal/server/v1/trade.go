package v1

import (
	"context"
	"encoding/json"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
)

// Sending a trade to the engine.
//
// The API server does not trade. It checks who is asking, hands the request to the pod holding
// that account, and waits for the answer. Everything about whether the trade is allowed — the
// group's limits, the instrument's, the routing rules — belongs to the engine, which is the
// only thing holding the account's live money state.
//
// This file is the transport-neutral core, in the same shape as the rest of the server: the
// HTTP and websocket adapters call it and neither knows about the other.

// engineTimeout is how long to wait for the engine. A trade that has not been answered in this
// long is not going to be, and the client needs to be told rather than left hanging.
//
// It is deliberately longer than a routing delay rule is likely to ask for.
const engineTimeout = 10 * time.Second

// TradeRequest is what the engine expects. It mirrors hst-core's own type; the two are joined
// by the subject and the JSON, not by a shared package.
type TradeRequest struct {
	RequestId string  `json:"request_id"`
	Login     int64   `json:"login"`
	Symbol    string  `json:"symbol"`
	Type      int32   `json:"type"`
	Volume    int64   `json:"volume"`
	Price     float64 `json:"price"`
	PriceSL   float64 `json:"price_sl"`
	PriceTP   float64 `json:"price_tp"`
	TypeFill  int32   `json:"type_fill"`
	TypeTime  int32   `json:"type_time"`
	Expiry    int64   `json:"expiry"`
	Comment   string  `json:"comment"`
	ExpertId  int64   `json:"expert_id"`
	Reason    int32   `json:"reason"`
	Dealer    int64   `json:"dealer"`
}

// TradeResult is what comes back.
type TradeResult struct {
	RequestId  string  `json:"request_id"`
	Login      int64   `json:"login"`
	RetCode    int32   `json:"retcode"`
	Message    string  `json:"message"`
	OrderId    int64   `json:"order_id,omitempty"`
	PositionId int64   `json:"position_id,omitempty"`
	DealId     int64   `json:"deal_id,omitempty"`
	Price      float64 `json:"price,omitempty"`
	Volume     int64   `json:"volume,omitempty"`
	Profit     float64 `json:"profit,omitempty"`
	Rule       string  `json:"rule,omitempty"`
}

// CrtTrade is what a client asks for. Volume is in lots, the unit a trader thinks in; it is
// converted to the engine's integer units here so no client has to know about them.
type CrtTrade struct {
	Symbol   string  `json:"symbol" validate:"required,max=32"`
	Type     int32   `json:"type" validate:"gte=0,lte=7"`
	Volume   float64 `json:"volume" validate:"required,gt=0"`
	Price    float64 `json:"price" validate:"gte=0"`
	PriceSL  float64 `json:"price_sl" validate:"gte=0"`
	PriceTP  float64 `json:"price_tp" validate:"gte=0"`
	TypeFill int32   `json:"type_fill" validate:"gte=0,lte=3"`
	TypeTime int32   `json:"type_time" validate:"gte=0,lte=3"`
	Expiry   int64   `json:"expiry"`
	Comment  string  `json:"comment" validate:"max=64"`
	ExpertId int64   `json:"expert_id"`
}

// SendTrade hands a request to the pod holding this account and waits for the answer.
//
// The shard is worked out from the login, so the subject alone routes the request. Nothing here
// asks which pod owns what.
func (s *HttpServer) SendTrade(ctx context.Context, login int64, body *CrtTrade,
	reason model.OrderReason, dealer int64) (*TradeResult, int, error) {
	req := TradeRequest{
		RequestId: newRequestId(),
		Login:     login,
		Symbol:    body.Symbol,
		Type:      body.Type,
		Volume:    model.LotsToVolume(body.Volume),
		Price:     body.Price,
		PriceSL:   body.PriceSL,
		PriceTP:   body.PriceTP,
		TypeFill:  body.TypeFill,
		TypeTime:  body.TypeTime,
		Expiry:    body.Expiry,
		Comment:   body.Comment,
		ExpertId:  body.ExpertId,
		Reason:    int32(reason),
		Dealer:    dealer,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, nethttp.StatusInternalServerError, err
	}

	ctx, cancel := context.WithTimeout(ctx, engineTimeout)
	defer cancel()

	msg, err := s.Nats.NC.RequestWithContext(ctx, model.SubjectShardRequest(login), payload)
	if err != nil {
		// no pod answered: either none owns this shard or the engine is down. Either way the
		// client must not be told the trade went through.
		return nil, nethttp.StatusServiceUnavailable, errs.ErrEngineUnavailable
	}

	var res TradeResult
	if err := json.Unmarshal(msg.Data, &res); err != nil {
		return nil, nethttp.StatusInternalServerError, err
	}

	// the engine's refusal is the answer, not an error in this server
	if res.RetCode != 0 {
		return &res, nethttp.StatusBadRequest, nil
	}

	return &res, nethttp.StatusOK, nil
}

// newRequestId is a handle for one request, so a client can match an answer to what it asked.
func newRequestId() string {
	return time.Now().Format("20060102150405.000000000")
}
