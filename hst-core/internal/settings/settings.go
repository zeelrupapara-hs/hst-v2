// Package settings folds the instrument, the group and the group's override of it into one flat answer.
package settings

import (
	"strings"
	"sync"

	"hstcore/model"
)

// Rules is everything a trade needs to know about one symbol for one group, already resolved.
type Rules struct {
	Symbol   *model.Symbol
	Group    *model.Group
	Override *model.GroupSymbol

	TradeMode  model.TradeMode
	ExecMode   model.ExecMode
	FillFlags  int32
	ExpirFlags int32

	Digits       int32
	Point        float64
	ContractSize float64
	TickValue    float64
	TickSize     float64
	CalcMode     model.CalcMode

	SpreadDiff  int32
	StopsLevel  int32
	FreezeLevel int32

	VolumeMin   int64
	VolumeMax   int64
	VolumeStep  int64
	VolumeLimit int64

	MarginInitial     float64
	MarginMaintenance float64
	MarginHedged      float64
	MarginFlags       int32

	SwapMode    int32
	SwapRate    [7]float64
	SwapYearDay int32
	SwapLong    float64
	SwapShort   float64

	SwapFlags int32

	// OrderFlags is the set of order types, and of SL/TP, the group may use on this instrument.
	OrderFlags int32
	// IEFlags and REFlags say whether the dealer confirms again after the client accepts a price.
	IEFlags int32
	REFlags int32
	// RequestTimeout is how long a dealer has to answer a request-execution order, in seconds.
	RequestTimeout int32

	CurrencyBase   string
	CurrencyProfit string
	CurrencyMargin string

	// MaxDeviationTime is how stale a quote may be, in seconds, before instant execution requotes.
	MaxDeviationTime int32
	// Slippage in points either side of the requested price, before instant execution requotes.
	MaxDeviationProfit int32
	MaxDeviationLoss   int32
	// MaxInstantVolume is the largest volume instant execution takes; above it the order becomes a request.
	MaxInstantVolume int64
}

// Store holds the settings for every group and symbol, and caches the resolved answer.
type Store struct {
	mu sync.RWMutex

	groups  map[string]*model.Group  // by group path
	symbols map[string]*model.Symbol // by symbol name
	// overrides are held per group, in the order the group lists them.
	overrides map[int64][]*model.GroupSymbol

	resolved map[string]*Rules // by group path + "\x00" + symbol
}

func New() *Store {
	return &Store{
		groups:    make(map[string]*model.Group, 512),
		symbols:   make(map[string]*model.Symbol, 4096),
		overrides: make(map[int64][]*model.GroupSymbol, 512),
		resolved:  make(map[string]*Rules, 8192),
	}
}

// Load replaces everything at once; a half-loaded store would judge a trade wrongly.
func (s *Store) Load(groups map[string]*model.Group, symbols map[string]*model.Symbol,
	overrides map[int64][]*model.GroupSymbol) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.groups = groups
	s.symbols = symbols
	s.overrides = overrides
	s.resolved = make(map[string]*Rules, 8192)
}

func (s *Store) Group(path string) (*model.Group, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g, ok := s.groups[path]
	return g, ok
}

func (s *Store) Symbol(name string) (*model.Symbol, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sym, ok := s.symbols[name]
	return sym, ok
}

func (s *Store) Groups() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.groups)
}

func (s *Store) Symbols() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.symbols)
}

// For resolves the rules, or reports that the group may not trade the symbol at all.
func (s *Store) For(groupPath, symbol string) (*Rules, bool) {
	key := groupPath + "\x00" + symbol

	s.mu.RLock()
	if r, ok := s.resolved[key]; ok {
		s.mu.RUnlock()
		return r, true
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// another goroutine may have resolved it while the write lock was being taken
	if r, ok := s.resolved[key]; ok {
		return r, true
	}

	g, ok := s.groups[groupPath]
	if !ok {
		return nil, false
	}
	sym, ok := s.symbols[symbol]
	if !ok {
		return nil, false
	}

	over := s.overrideFor(g.GroupId, sym)
	if over == nil {
		return nil, false
	}

	r := resolve(g, sym, over)
	s.resolved[key] = r

	return r, true
}

// overrideFor finds the group's entry covering this symbol; the first match in the group's own order wins.
func (s *Store) overrideFor(groupId int64, sym *model.Symbol) *model.GroupSymbol {
	for _, o := range s.overrides[groupId] {
		if matches(o.Path, sym) {
			return o
		}
	}
	return nil
}

// matches reports whether a group-symbol path covers an instrument.
func matches(path string, sym *model.Symbol) bool {
	switch {
	case path == "" || path == "*":
		return true
	case strings.HasSuffix(path, "*"):
		prefix := strings.TrimSuffix(path, "*")
		return strings.HasPrefix(sym.Path, prefix) || strings.HasPrefix(sym.Symbol, prefix)
	default:
		return path == sym.Symbol || path == sym.Path
	}
}

// resolve folds the override onto the instrument; nil means the instrument's own value stands.
func resolve(g *model.Group, sym *model.Symbol, o *model.GroupSymbol) *Rules {
	r := &Rules{
		Symbol:   sym,
		Group:    g,
		Override: o,

		Digits:       sym.Digits,
		Point:        sym.Point,
		ContractSize: sym.ContractSize,
		TickValue:    sym.TickValue,
		TickSize:     sym.TickSize,
		CalcMode:     model.CalcMode(sym.CalcMode),

		CurrencyBase:   sym.CurrencyBase,
		CurrencyProfit: sym.CurrencyProfit,
		CurrencyMargin: sym.CurrencyMargin,

		TradeMode:  model.TradeMode(pick(o.TradeMode, sym.TradeMode)),
		ExecMode:   model.ExecMode(pick(o.ExecMode, sym.ExecMode)),
		FillFlags:  pick(o.FillFlags, sym.FillFlags),
		ExpirFlags: pick(o.ExpirFlags, sym.ExpirFlags),

		SpreadDiff:  pick(o.SpreadDiff, sym.SpreadDiff),
		StopsLevel:  pick(o.StopsLevel, sym.StopsLevel),
		FreezeLevel: pick(o.FreezeLevel, sym.FreezeLevel),

		VolumeMin:   pick(o.VolumeMin, sym.VolumeMin),
		VolumeMax:   pick(o.VolumeMax, sym.VolumeMax),
		VolumeStep:  pick(o.VolumeStep, sym.VolumeStep),
		VolumeLimit: pick(o.VolumeLimit, sym.VolumeLimit),

		MarginInitial:     pick(o.MarginInitial, sym.MarginInitial),
		MarginMaintenance: pick(o.MarginMaintenance, sym.MarginMaintenance),
		MarginHedged:      pick(o.MarginHedged, sym.MarginHedged),
		MarginFlags:       pick(o.MarginFlags, sym.MarginFlags),

		SwapMode:    pick(o.SwapMode, sym.SwapMode),
		SwapYearDay: pick(o.SwapYearDay, sym.SwapYearDay),
		SwapLong:    pick(o.SwapLong, sym.SwapLong),
		SwapShort:   pick(o.SwapShort, sym.SwapShort),

		SwapFlags: pick(o.SwapFlags, sym.SwapFlags),

		OrderFlags:     pick(o.OrderFlags, sym.OrderFlags),
		IEFlags:        value(o.IEFlags),
		REFlags:        value(o.REFlags),
		RequestTimeout: value(o.RETimeout),

		MaxDeviationTime:   value(o.IETimeout),
		MaxDeviationProfit: pick(o.IESlipProfit, sym.SpreadDiff),
		MaxDeviationLoss:   pick(o.IESlipLosing, sym.SpreadDiff),
		MaxInstantVolume:   value(o.IEVolumeMax),
	}

	for i := range r.SwapRate {
		r.SwapRate[i] = pick(o.SwapRate[i], sym.SwapRate[i])
	}

	if r.SwapYearDay <= 0 {
		r.SwapYearDay = 360
	}

	// a step of zero would make every volume invalid, so fall back to the smallest allowed
	if r.VolumeStep == 0 {
		r.VolumeStep = r.VolumeMin
	}

	return r
}

func pick[T any](override *T, base T) T {
	if override != nil {
		return *override
	}
	return base
}

func value[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
