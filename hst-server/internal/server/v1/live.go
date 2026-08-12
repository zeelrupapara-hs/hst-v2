package v1

import (
	"context"
	"encoding/json"
	"time"

	"hstserver/model"
)

// queryTimeout is short on purpose: this is a read on a request path, and a database answer is
// better than a slow one.
const queryTimeout = 2 * time.Second

// askEngine reads live state from the pod holding the account. It answers nil rather than an
// error when the engine cannot be reached, so a caller falls back to the database instead of
// failing the request.
func (s *HttpServer) askEngine(ctx context.Context, login int64, what model.QueryWhat) *model.QueryResult {
	payload, err := json.Marshal(model.QueryRequest{Login: login, What: what})
	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	msg, err := s.Nats.NC.RequestWithContext(ctx, model.SubjectSystemQuery, payload)
	if err != nil {
		return nil
	}

	var res model.QueryResult
	if err := json.Unmarshal(msg.Data, &res); err != nil || !res.Found {
		return nil
	}

	return &res
}

// AskEngineAccount is the live money state, or nil if the engine did not answer.
func (s *HttpServer) AskEngineAccount(ctx context.Context, login int64) *model.Account {
	res := s.askEngine(ctx, login, model.QueryWhat_account)
	if res == nil {
		return nil
	}
	return res.Account
}

// AskEnginePositions is the live positions, or nil if the engine did not answer.
func (s *HttpServer) AskEnginePositions(ctx context.Context, login int64) []model.Position {
	res := s.askEngine(ctx, login, model.QueryWhat_positions)
	if res == nil {
		return nil
	}
	return res.Positions
}

// AskEngineSymbols is what the caller's group may trade, resolved and priced, or nil if the
// engine did not answer.
func (s *HttpServer) AskEngineSymbols(ctx context.Context, login int64) []model.SymbolInfo {
	res := s.askEngine(ctx, login, model.QueryWhat_symbols)
	if res == nil {
		return nil
	}
	return res.Symbols
}

// overlayLive replaces the values that only the engine knows: what a position is worth right
// now. Silently leaves the database values alone if the engine cannot be reached.
func (s *HttpServer) overlayLive(ctx context.Context, login int64, out []ViewPosition) {
	live := s.AskEnginePositions(ctx, login)
	if len(live) == 0 {
		return
	}

	at := make(map[int64]int, len(out))
	for i := range out {
		at[out[i].PositionId] = i
	}

	for i := range live {
		j, ok := at[live[i].PositionId]
		if !ok {
			continue
		}
		out[j].PriceCurrent = live[i].PriceCurrent
		out[j].Profit = live[i].Profit
		out[j].Storage = live[i].Storage
	}
}

// overlayLiveAll refreshes a mixed list, asking the engine once per login it contains.
func (s *HttpServer) overlayLiveAll(ctx context.Context, out []ViewPosition) {
	seen := make(map[int64]bool)
	for i := range out {
		login := out[i].Login
		if seen[login] {
			continue
		}
		seen[login] = true
		s.overlayLive(ctx, login, out)
	}
}
