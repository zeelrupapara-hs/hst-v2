package filter

import (
	"hstquote/model"
)

// dayStats is what the platform works out itself when "Receive market statistics from datafeeds" is off:
// High/Low of Bid, the day's Open and the previous day's Close.
type dayStats struct {
	day                   int
	open, high, low, last float64
	prevClose             float64
}

// Stats fills the tick's OHLC from the platform's own running day statistics, unless the symbol takes them from the feed.
// ponytail: stats live in memory per pod and start from the first tick after a restart; seed from tick history if that matters.
func (s *State) Stats(set model.SymbolSettings, t *model.Tick) {
	if set.FeedStats() {
		return
	}
	st := s.get(t.SymbolID)
	d := t.Time.UTC().YearDay() + t.Time.UTC().Year()*1000
	ds := &st.stats
	if ds.day != d {
		if ds.day != 0 {
			ds.prevClose = ds.last
		}
		ds.day, ds.open, ds.high, ds.low = d, t.Bid, t.Bid, t.Bid
	}
	if t.Bid > ds.high {
		ds.high = t.Bid
	}
	if t.Bid < ds.low {
		ds.low = t.Bid
	}
	ds.last = t.Bid
	t.Open, t.High, t.Low, t.Close = ds.open, ds.high, ds.low, ds.prevClose
}
