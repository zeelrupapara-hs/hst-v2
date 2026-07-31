package model

// Tick is one quote, as hst-quote publishes it on hstquote.tick.<symbol>.
//
// Bid and Ask are the only two prices a fill needs: a buy trades at Ask, a sell at Bid. Last and
// Volume are carried for exchange instruments and reporting.
type Tick struct {
	Symbol     string  `json:"symbol"`
	Digits     int32   `json:"digits"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	Last       float64 `json:"last"`
	Volume     int64   `json:"volume"`
	VolumeReal float64 `json:"volume_real"`
	// Time is when the quote was produced, epoch nanoseconds
	Time int64 `json:"time"`
}

// Spread is the difference between the two sides.
func (t *Tick) Spread() float64 { return t.Ask - t.Bid }

// Mid is the midpoint, for the places that want a side-neutral price.
func (t *Tick) Mid() float64 { return (t.Ask + t.Bid) / 2 }

// OpenPrice is what a trade of this side pays to open: buy at Ask, sell at Bid.
func (t *Tick) OpenPrice(buy bool) float64 {
	if buy {
		return t.Ask
	}
	return t.Bid
}

// ClosePrice is what closes a position of this side, the opposite of opening it.
func (t *Tick) ClosePrice(buyPosition bool) float64 {
	if buyPosition {
		return t.Bid
	}
	return t.Ask
}

// Ok reports whether the quote can be traded on. A missing side is not a price.
func (t *Tick) Ok() bool { return t.Bid > 0 && t.Ask > 0 }

// TickSubject is where hst-quote publishes one symbol's quotes.
func TickSubject(symbol string) string { return "hstquote.tick." + symbol }

// TickSubjectAll matches every symbol. Every pod listens to all of them, because any pod may
// hold an account trading any symbol.
const TickSubjectAll = "hstquote.tick.*"
