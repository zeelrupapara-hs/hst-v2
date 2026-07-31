package simulator

import (
	"context"
	"math"
	"math/rand"
	"time"

	"hstquote/internal/provider"
	"hstquote/model"
)

// Connector emits synthetic ticks for local development and testing.
type Connector struct {
	feed  model.QuoteFeed
	ticks chan<- provider.RawTick

	// symbol -> last mid price
	state map[string]float64
}

func NewConnector(feed model.QuoteFeed, ticks chan<- provider.RawTick) *Connector {
	state := make(map[string]float64, len(feed.Translates))
	for i, tr := range feed.Translates {
		state[tr.ExternalSymbol()] = 1.1000 + float64(i)*0.01
	}
	return &Connector{feed: feed, ticks: ticks, state: state}
}

func (c *Connector) Type() provider.ConnectorType { return provider.TypeSimulator }

func (c *Connector) Run(ctx context.Context) error {
	if len(c.feed.Translates) == 0 {
		<-ctx.Done()
		return ctx.Err()
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for _, tr := range c.feed.Translates {
				ext := tr.ExternalSymbol()
				mid := c.state[ext]
				mid += (rng.Float64() - 0.5) * 0.0002
				if mid <= 0 {
					mid = 1.0
				}
				c.state[ext] = mid
				spread := 0.00010
				bid := roundPrice(mid-spread/2, tr.Digits)
				ask := roundPrice(mid+spread/2, tr.Digits)
				high := roundPrice(mid+spread, tr.Digits)
				low := roundPrice(mid-spread, tr.Digits)

				select {
				case c.ticks <- provider.RawTick{
					SourceSymbol: ext,
					Bid:          bid,
					Ask:          ask,
					High:         high,
					Low:          low,
					Open:         bid,
					Close:        ask,
					Volume:       100,
					Time:         time.Now().UTC(),
				}:
				default:
				}
			}
		}
	}
}

func (c *Connector) Close() error { return nil }

func roundPrice(v float64, digits int16) float64 {
	if digits <= 0 {
		digits = 5
	}
	pow := math.Pow10(int(digits))
	return math.Round(v*pow) / pow
}
