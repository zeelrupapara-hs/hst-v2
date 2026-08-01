package model

// Tick is one quote as the feed publishes it.
type Tick struct {
	Symbol     string  `json:"symbol"`
	Digits     int32   `json:"digits"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	Last       float64 `json:"last"`
	Volume     int64   `json:"volume"`
	VolumeReal float64 `json:"volume_real"`
	Time       int64   `json:"time"`
}

func (t *Tick) Spread() float64 { return t.Ask - t.Bid }

func (t *Tick) Mid() float64 { return (t.Ask + t.Bid) / 2 }

// A buy trades at Ask, a sell at Bid.
func (t *Tick) OpenPrice(buy bool) float64 {
	if buy {
		return t.Ask
	}
	return t.Bid
}

func (t *Tick) ClosePrice(buyPosition bool) float64 {
	if buyPosition {
		return t.Bid
	}
	return t.Ask
}

// A missing side is not a price.
func (t *Tick) Ok() bool { return t.Bid > 0 && t.Ask > 0 }
