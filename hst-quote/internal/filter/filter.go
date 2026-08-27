package filter

import (
	"fmt"
	"math"
	"time"

	"hstquote/model"
)

type side struct {
	prev     float64
	count    int32
	hardStg  bool
	gapPrev  float64
	gapCount int32
	gapped   bool
}

func (sd *side) reset(price float64) {
	sd.prev = price
	sd.count = 0
	sd.hardStg = false
}

type symbolState struct {
	bid, ask   side
	lastBid    float64
	lastAsk    float64
	lastVolume float64
	lastTime   time.Time
	broken     bool
	seen       bool
	notes      []string // journal lines from the last Apply
	stats      dayStats
}

// State holds filter channels for one feed runner; one goroutine owns it, so no lock.
type State struct {
	symbols map[int64]*symbolState
}

func New() *State {
	return &State{symbols: make(map[int64]*symbolState)}
}

func (s *State) get(symbolID int64) *symbolState {
	st, ok := s.symbols[symbolID]
	if !ok {
		st = &symbolState{broken: true}
		s.symbols[symbolID] = st
	}
	return st
}

// Notes are the journal lines from the last Apply for a symbol (gap mode and one-sided filtering).
func (s *State) Notes(symbolID int64) []string { return s.get(symbolID).notes }

// MarkBreak arms a stream break so filters are skipped on the next tick.
func (s *State) MarkBreak(symbolID int64) {
	s.get(symbolID).broken = true
}

// Reset arms a break on every symbol, for when the whole stream drops and reconnects.
func (s *State) Reset() {
	for _, st := range s.symbols {
		st.broken = true
	}
}

// points is a price difference in symbol points, rounded so an exact bound survives float error.
func points(diff, pt float64) float64 { return math.Round(diff / pt) }

// Apply runs filtration and gap detection, mutating the tick; a non-empty reason means the tick is rejected.
// Notes holds the journal lines the pass produced.
func (s *State) Apply(set model.SymbolSettings, t *model.Tick) (reason string) {
	st := s.get(t.SymbolID)
	st.notes = st.notes[:0]
	pt := set.PointValue()

	if !set.IsExchange() && !set.NegativeAllowed() && (t.Bid <= 0 || t.Ask <= 0) {
		return "non-positive price"
	}

	// min/max spread applies to floating-spread OTC symbols only
	if !set.IsExchange() && set.FloatingSpread() {
		sp := points(t.Ask-t.Bid, pt)
		if set.FilterSpreadMin > 0 && sp < float64(set.FilterSpreadMin) {
			return fmt.Sprintf("spread %d below minimum %d", int(sp), set.FilterSpreadMin)
		}
		if set.FilterSpreadMax > 0 && sp > float64(set.FilterSpreadMax) {
			return fmt.Sprintf("spread %d above maximum %d", int(sp), set.FilterSpreadMax)
		}
	}

	if st.seen && t.Bid == st.lastBid && t.Ask == st.lastAsk && t.Volume == st.lastVolume &&
		t.Time.Truncate(time.Minute).Equal(st.lastTime.Truncate(time.Minute)) {
		return "duplicate"
	}
	st.seen, st.lastBid, st.lastAsk, st.lastVolume, st.lastTime = true, t.Bid, t.Ask, t.Volume, t.Time

	if !set.FiltersEnabled() {
		firstTick := st.broken
		st.broken = false
		// keep the channel centred so turning filters on mid-stream does not measure against a stale price
		st.bid.reset(t.Bid)
		st.ask.reset(t.Ask)
		gapUpdate(st, set, t, pt, firstTick)
		return ""
	}

	if st.broken {
		st.bid.reset(t.Bid)
		st.ask.reset(t.Ask)
		st.broken = false
		gapUpdate(st, set, t, pt, true)
		return ""
	}

	prevBid, prevAsk := st.bid.prev, st.ask.prev
	whyBid := checkSide(&st.bid, t.Bid, set, pt)
	whyAsk := checkSide(&st.ask, t.Ask, set, pt)
	if whyBid != "" && whyAsk != "" {
		return fmt.Sprintf("%s by bid from %v to %v, %s by ask from %v to %v", whyBid, prevBid, t.Bid, whyAsk, prevAsk, t.Ask)
	}
	// a rejected side keeps its last accepted price so the other side can still publish
	if whyBid != "" {
		st.notes = append(st.notes, fmt.Sprintf("%s by bid from %v to %v, bid kept", whyBid, prevBid, t.Bid))
		t.Bid = st.bid.prev
	}
	if whyAsk != "" {
		st.notes = append(st.notes, fmt.Sprintf("%s by ask from %v to %v, ask kept", whyAsk, prevAsk, t.Ask))
		t.Ask = st.ask.prev
	}

	gapUpdate(st, set, t, pt, false)
	return ""
}

// checkSide returns why a side is rejected, "" when it passes; MT5 wording with the diff and level in points.
func checkSide(sd *side, price float64, set model.SymbolSettings, pt float64) string {
	d := points(math.Abs(price-sd.prev), pt)

	if set.FilterDiscard > 0 && d > float64(set.FilterDiscard) {
		return fmt.Sprintf("discard filter [diff:%d, level:%d]", int(d), set.FilterDiscard)
	}

	if set.FilterSoft > 0 && d > float64(set.FilterSoft) {
		if set.FilterHard > 0 && d > float64(set.FilterHard) {
			sd.count++
			if !sd.hardStg {
				// the quote that completes the soft count also opens the hard one
				if sd.count > set.FilterSoftTicks {
					sd.hardStg = true
					sd.count = 1
				}
				return fmt.Sprintf("hard filter [diff:%d, level:%d, soft counter:%d/%d]", int(d), set.FilterHard, sd.count, set.FilterSoftTicks)
			}
			if sd.count > set.FilterHardTicks {
				sd.reset(price)
				return ""
			}
			return fmt.Sprintf("hard filter [diff:%d, level:%d, hard counter:%d/%d]", int(d), set.FilterHard, sd.count, set.FilterHardTicks)
		}
		sd.count++
		if sd.count > set.FilterSoftTicks {
			sd.reset(price)
			return ""
		}
		return fmt.Sprintf("soft filter [diff:%d, level:%d, counter:%d/%d]", int(d), set.FilterSoft, sd.count, set.FilterSoftTicks)
	}

	sd.reset(price)
	return ""
}

func gapUpdate(st *symbolState, set model.SymbolSettings, t *model.Tick, pt float64, firstTick bool) {
	if set.FilterGap <= 0 || set.FilterGapTicks <= 0 {
		st.bid.gapped, st.ask.gapped = false, false
		st.bid.gapPrev, st.ask.gapPrev = t.Bid, t.Ask
		t.Gap = false
		return
	}

	if firstTick {
		st.bid.gapped, st.ask.gapped = true, true
		st.bid.gapCount, st.ask.gapCount = set.FilterGapTicks, set.FilterGapTicks
		st.bid.gapPrev, st.ask.gapPrev = t.Bid, t.Ask
	} else {
		gapStep(st, &st.bid, "bid", t.Bid, set, pt)
		gapStep(st, &st.ask, "ask", t.Ask, set, pt)
	}

	t.Gap = st.bid.gapped || st.ask.gapped
}

func gapStep(st *symbolState, sd *side, name string, price float64, set model.SymbolSettings, pt float64) {
	d := points(math.Abs(price-sd.gapPrev), pt)
	if d > float64(set.FilterGap) {
		sd.gapped, sd.gapCount = true, 1
		st.notes = append(st.notes, fmt.Sprintf("gap by %s from %v to %v [diff:%d, gap level:%d]", name, sd.gapPrev, price, int(d), set.FilterGapTicks))
	} else if sd.gapped {
		sd.gapCount++
		if sd.gapCount > set.FilterGapTicks {
			sd.gapped, sd.gapCount = false, 0
			st.notes = append(st.notes, fmt.Sprintf("gap by %s mode disabled", name))
		} else {
			st.notes = append(st.notes, fmt.Sprintf("tick without gap by %s [tick counter: %d]", name, sd.gapCount))
		}
	}
	sd.gapPrev = price
}
