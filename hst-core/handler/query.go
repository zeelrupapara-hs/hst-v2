package handler

import (
	"encoding/json"

	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// Reading live state out of the engine.
//
// Orders, positions and balances are written down, but what a position is worth right now is
// not: it is worked out from the price on every tick and kept in memory. A reader going to the
// database therefore sees the account as it was when it last traded, not as it stands. These
// answer with what the engine actually holds.

func (h *Handler) QuerySystemEventHandler(msg *natscore.Msg) {
	var q model.QueryRequest
	if err := json.Unmarshal(msg.Data, &q); err != nil {
		h.Log.Log(logger.TypeAPI, logger.CodeWarn, "bad query", "error", err.Error())
		return
	}

	if msg.Reply == "" {
		return
	}

	res := &model.QueryResult{Login: q.Login}

	switch q.What {
	case model.QueryAccount:
		res.Account, res.Found = h.GetAccount(q.Login)
	case model.QueryPositions:
		res.Positions, res.Found = h.GetAllPositions(q.Login)
	case model.QueryOrders:
		res.Orders, res.Found = h.GetAllOrders(q.Login)
	case model.QueryState:
		res.Account, res.Found = h.GetAccount(q.Login)
		res.Positions, _ = h.GetAllPositions(q.Login)
		res.Orders, _ = h.GetAllOrders(q.Login)
	}

	body, err := json.Marshal(res)
	if err != nil {
		return
	}

	if err := msg.Respond(body); err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not answer a query",
			"login", q.Login, "error", err.Error())
	}
}

func (h *Handler) GetAccount(login int64) (*model.Account, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	defer e.Unlock()

	a := *e.Account

	return &a, true
}

func (h *Handler) GetAllPositions(login int64) ([]model.Position, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	defer e.Unlock()

	out := make([]model.Position, 0, len(e.Positions))
	for _, p := range e.Positions {
		out = append(out, *p)
	}

	return out, true
}

func (h *Handler) GetPosition(login, positionId int64) (*model.Position, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	defer e.Unlock()

	p, ok := e.Positions[positionId]
	if !ok {
		return nil, false
	}

	out := *p

	return &out, true
}

func (h *Handler) GetAllOrders(login int64) ([]model.Order, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	defer e.Unlock()

	out := make([]model.Order, 0, len(e.Orders))
	for _, o := range e.Orders {
		out = append(out, *o)
	}

	return out, true
}
