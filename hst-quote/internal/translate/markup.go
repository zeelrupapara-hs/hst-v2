package translate

import (
	"math"

	"hstquote/internal/provider"
	"hstquote/model"
)

// ApplyMarkup maps a raw LP tick to a platform tick using translate rules.
func ApplyMarkup(feed model.QuoteFeed, raw provider.RawTick) (*model.Tick, bool) {
	for _, tr := range feed.Translates {
		if tr.ExternalSymbol() != raw.SourceSymbol {
			continue
		}
		if tr.SymbolID <= 0 {
			continue
		}
		digits := tr.Digits
		if digits <= 0 {
			digits = 5
		}
		point := math.Pow10(-int(digits))
		bid := round(raw.Bid+float64(tr.BidMarkup)*point, digits)
		ask := round(raw.Ask+float64(tr.AskMarkup)*point, digits)

		return &model.Tick{
			DatafeedID: feed.Datafeed.DatafeedID,
			SymbolID:   tr.SymbolID,
			Symbol:     tr.Symbol,
			Source:     raw.SourceSymbol,
			Bid:        bid,
			Ask:        ask,
			High:       raw.High,
			Low:        raw.Low,
			Open:       raw.Open,
			Close:      raw.Close,
			Volume:     raw.Volume,
			Time:       raw.Time,
		}, true
	}
	return nil, false
}

func round(v float64, digits int16) float64 {
	pow := math.Pow10(int(digits))
	return math.Round(v*pow) / pow
}
