package model

// Tick is one quote, as hst-quote publishes it on hstquote.tick.<symbol>.
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

// Spread is the raw difference between the two sides.
func (t *Tick) Spread() float64 { return t.Ask - t.Bid }

// Mid is the midpoint, used where a side-neutral price is wanted.
func (t *Tick) Mid() float64 { return (t.Ask + t.Bid) / 2 }

// PriceFor is the price a trade of this side fills at: buy pays Ask, sell receives Bid.
func (t *Tick) PriceFor(buy bool) float64 {
	if buy {
		return t.Ask
	}
	return t.Bid
}

// ClosePriceFor is the price that closes a position of this side, the opposite of opening it.
func (t *Tick) ClosePriceFor(buyPosition bool) float64 {
	if buyPosition {
		return t.Bid
	}
	return t.Ask
}

// SubjectTick is where hst-quote publishes a symbol's quotes.
func SubjectTick(symbol string) string { return "hstquote.tick." + symbol }

// SubjectTickAll matches every symbol's quotes.
const SubjectTickAll = "hstquote.tick.*"
