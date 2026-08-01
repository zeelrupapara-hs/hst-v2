package filter

import (
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
	warned     bool
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

// WarnOnce reports true the first time it is asked about a symbol, so drops are logged once.
func (s *State) WarnOnce(symbolID int64) bool {
	st := s.get(symbolID)
	if st.warned {
		return false
	}
	st.warned = true
	return true
}

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

// Apply runs filtration and gap detection, mutating the tick; false means the tick is rejected.
func (s *State) Apply(set model.SymbolSettings, t *model.Tick) bool {
	st := s.get(t.SymbolID)
	pt := set.PointValue()

	if !set.IsExchange() && (t.Bid <= 0 || t.Ask <= 0) {
		return false
	}

	// min/max spread applies to floating-spread OTC symbols only
	if !set.IsExchange() && set.FloatingSpread() {
		sp := points(t.Ask-t.Bid, pt)
		if set.FilterSpreadMin > 0 && sp < float64(set.FilterSpreadMin) {
			return false
		}
		if set.FilterSpreadMax > 0 && sp > float64(set.FilterSpreadMax) {
			return false
		}
	}

	if st.seen && t.Bid == st.lastBid && t.Ask == st.lastAsk && t.Volume == st.lastVolume &&
		t.Time.Truncate(time.Minute).Equal(st.lastTime.Truncate(time.Minute)) {
		return false
	}
	st.seen, st.lastBid, st.lastAsk, st.lastVolume, st.lastTime = true, t.Bid, t.Ask, t.Volume, t.Time

	if !set.FiltersEnabled() {
		firstTick := st.broken
		st.broken = false
		// keep the channel centred so turning filters on mid-stream does not measure against a stale price
		st.bid.reset(t.Bid)
		st.ask.reset(t.Ask)
		gapUpdate(st, set, t, pt, firstTick)
		return true
	}

	if st.broken {
		st.bid.reset(t.Bid)
		st.ask.reset(t.Ask)
		st.broken = false
		gapUpdate(st, set, t, pt, true)
		return true
	}

	okBid := checkSide(&st.bid, t.Bid, set, pt)
	okAsk := checkSide(&st.ask, t.Ask, set, pt)
	if !okBid && !okAsk {
		return false
	}
	// a rejected side keeps its last accepted price so the other side can still publish
	if !okBid {
		t.Bid = st.bid.prev
	}
	if !okAsk {
		t.Ask = st.ask.prev
	}

	gapUpdate(st, set, t, pt, false)
	return true
}

func checkSide(sd *side, price float64, set model.SymbolSettings, pt float64) bool {
	d := points(math.Abs(price-sd.prev), pt)

	if set.FilterDiscard > 0 && d > float64(set.FilterDiscard) {
		return false
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
				return false
			}
			if sd.count > set.FilterHardTicks {
				sd.reset(price)
				return true
			}
			return false
		}
		sd.count++
		if sd.count > set.FilterSoftTicks {
			sd.reset(price)
			return true
		}
		return false
	}

	sd.reset(price)
	return true
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
		gapStep(&st.bid, t.Bid, set, pt)
		gapStep(&st.ask, t.Ask, set, pt)
	}

	t.Gap = st.bid.gapped || st.ask.gapped
}

func gapStep(sd *side, price float64, set model.SymbolSettings, pt float64) {
	if points(math.Abs(price-sd.gapPrev), pt) > float64(set.FilterGap) {
		sd.gapped, sd.gapCount = true, 1
	} else if sd.gapped {
		sd.gapCount++
		if sd.gapCount > set.FilterGapTicks {
			sd.gapped, sd.gapCount = false, 0
		}
	}
	sd.gapPrev = price
}
