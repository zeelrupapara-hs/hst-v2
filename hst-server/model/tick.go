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

// SubjectTickAll matches every symbol's quotes.
const SubjectTickAll = "hstquote.tick.*"
