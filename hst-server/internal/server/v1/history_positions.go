package v1

import (
	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// ViewClosedPosition is one closed position, rebuilt from the deals that made it.
type ViewClosedPosition struct {
	PositionId int64   `json:"position_id"`
	Symbol     string  `json:"symbol"`
	Action     int32   `json:"action"`
	Volume     float64 `json:"volume"`
	VolumeExt  int64   `json:"-"`
	TimeOpen   int64   `json:"time_open"`
	PriceOpen  float64 `json:"price_open"`
	TimeClose  int64   `json:"time_close"`
	PriceClose float64 `json:"price_close"`
	Profit     float64 `json:"profit"`
	Storage    float64 `json:"storage"`
	Commission float64 `json:"commission"`
	Fee        float64 `json:"fee"`
}

// closedPositions folds the deal ledger back into positions, in the database, since it only grows.
const closedPositions = `
	SELECT d.position_id,
	       MIN(d.symbol) AS symbol,
	       (ARRAY_AGG(d.action ORDER BY d.time, d.deal_id) FILTER (WHERE d.entry IN (0, 2)))[1] AS action,
	       SUM(CASE WHEN d.volume_ext > 0 THEN d.volume_ext ELSE d.volume * 10000 END)
	           FILTER (WHERE d.entry IN (0, 2)) AS volume_ext,
	       MIN(d.time) FILTER (WHERE d.entry IN (0, 2)) AS time_open,
	       (ARRAY_AGG(d.price ORDER BY d.time, d.deal_id) FILTER (WHERE d.entry IN (0, 2)))[1] AS price_open,
	       MAX(d.time) FILTER (WHERE d.entry IN (1, 3)) AS time_close,
	       (ARRAY_AGG(d.price ORDER BY d.time DESC, d.deal_id DESC) FILTER (WHERE d.entry IN (1, 3)))[1] AS price_close,
	       SUM(d.profit) AS profit, SUM(d.storage) AS storage,
	       SUM(d.commission) AS commission, SUM(d.fee) AS fee
	  FROM hst.deals d
	 WHERE d.login = $1 AND d.position_id > 0
	 GROUP BY d.position_id
	HAVING COUNT(*) FILTER (WHERE d.entry IN (0, 2)) > 0
	   AND COUNT(*) FILTER (WHERE d.entry IN (1, 3)) > 0
	   AND NOT EXISTS (SELECT 1 FROM hst.positions p WHERE p.position_id = d.position_id)`

// GetMyClosedPositions is the History panel's Closed Positions tab.
//
//	@Id			GetMyClosedPositions
//	@Tags		Trader
//	@Produce	json
//	@Param		limit	query		int	false	"how many, newest first"
//	@Param		page	query		int	false	"which page, zero based"
//	@Param		from	query		int	false	"unix seconds, inclusive"
//	@Param		to		query		int	false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewClosedPosition}
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/history/positions [get]
func (s *HttpServer) GetMyClosedPositions(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	p := readPage(c, 100)
	// the range is judged on the close, so a position appears in the window it was closed in
	where, args := p.bound("h.time_close", "TRUE", []any{snap.Login})

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT h.position_id, h.symbol, h.action, h.volume_ext, h.time_open, h.price_open,
		        h.time_close, h.price_close, h.profit, h.storage, h.commission, h.fee
		   FROM (`+closedPositions+`) h WHERE `+where+p.tail("h.time_close"), args...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewClosedPosition{}
	for rows.Next() {
		var v ViewClosedPosition
		if err := rows.Scan(&v.PositionId, &v.Symbol, &v.Action, &v.VolumeExt, &v.TimeOpen,
			&v.PriceOpen, &v.TimeClose, &v.PriceClose, &v.Profit, &v.Storage,
			&v.Commission, &v.Fee); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.Volume = model.ExtToLots(v.VolumeExt)
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}
