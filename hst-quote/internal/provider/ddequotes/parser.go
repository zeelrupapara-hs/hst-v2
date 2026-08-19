package ddequotes

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"hstquote/internal/provider"
)

// The currency server pushes '!'-terminated frames of comma-separated values:
// a header, the source time, then 16 fields per symbol. Only the indices below
// have a known meaning; the rest of the layout is undocumented. Close lives at
// index 8, and no index is known to carry volume, so volume stays 0 unless the
// VolumeIndex param names a field.
const (
	chunkSize = 16

	idxSymbol = 0
	idxBid    = 1
	idxAsk    = 2
	idxHigh   = 3
	idxLow    = 4
	idxOpen   = 6
	idxClose  = 8
)

const frameHeader = "MRKTDATAs?<G)_"

// parseFrame decodes one wire frame into raw ticks. A frame without the market
// data header is not an error, just not quote data.
func parseFrame(frame string, volumeIndex int, now time.Time) ([]provider.RawTick, error) {
	frame = strings.TrimSuffix(frame, "!")
	if !strings.HasPrefix(frame, frameHeader) {
		return nil, nil
	}
	parts := strings.Split(frame[len(frameHeader):], ",")
	// parts[0] is the source time; the server clock is not trusted, receive time is used
	if len(parts) < 1+chunkSize || (len(parts)-1)%chunkSize != 0 {
		return nil, fmt.Errorf("malformed market data frame, %d values", len(parts))
	}

	num := func(s string) float64 {
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return 0
		}
		return v
	}

	ticks := make([]provider.RawTick, 0, (len(parts)-1)/chunkSize)
	for i := 1; i < len(parts); i += chunkSize {
		chunk := parts[i : i+chunkSize]
		symbol := strings.TrimSpace(chunk[idxSymbol])
		if symbol == "" {
			continue
		}
		tick := provider.RawTick{
			SourceSymbol: symbol,
			Bid:          num(chunk[idxBid]),
			Ask:          num(chunk[idxAsk]),
			High:         num(chunk[idxHigh]),
			Low:          num(chunk[idxLow]),
			Open:         num(chunk[idxOpen]),
			Close:        num(chunk[idxClose]),
			Time:         now.UTC(),
		}
		if volumeIndex >= 0 && volumeIndex < chunkSize {
			tick.Volume = num(chunk[volumeIndex])
		}
		ticks = append(ticks, tick)
	}
	return ticks, nil
}
