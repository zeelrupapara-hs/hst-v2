// Package settings answers one question: for this group and this symbol, what are the rules?
//
// A trade is checked against three things at once — the instrument, the group, and what the
// group overrides about that instrument. Reading all three at every check would be slow and,
// worse, easy to get subtly wrong in one place and not another. So the three are folded into a
// single flat answer once, and everything downstream reads that.
package settings

import (
	"strings"
	"sync"

	"hstcore/model"
)

// Rules is the flat answer: everything a trade needs to know about one symbol for one group.
//
// Every field is already resolved. Nothing downstream needs to know whether a value came from
// the symbol or from the group's override of it.
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

	SwapMode  int32
	SwapLong  float64
	SwapShort float64

	CurrencyBase   string
	CurrencyProfit string
	CurrencyMargin string

	// Instant execution slippage, in points, either side of the requested price.
	SlipProfit int32
	SlipLosing int32
	// IEVolumeMax is the largest volume instant execution will take; above it MT5 falls back to
	// the request mode and a dealer.
	IEVolumeMax int64
}

// Store holds the settings for every group and symbol, and resolves them on demand.
//
// Groups and symbols change rarely — a manager editing them — while lookups happen on every
// tick and every trade. So the resolved answer is cached and thrown away wholesale whenever
// anything underneath it changes.
type Store struct {
	mu sync.RWMutex

	groups  map[string]*model.Group  // by group path
	symbols map[string]*model.Symbol // by symbol name
	// overrides are held per group, in the order the group lists them: the first path that
	// matches a symbol wins, which is how MT5 resolves `Forex\*` against `Forex\EURUSD`.
	overrides map[int64][]*model.GroupSymbol

	resolved map[string]*Rules // by group path + "\x00" + symbol
}

// New builds an empty store.
func New() *Store {
	return &Store{
		groups:    make(map[string]*model.Group, 512),
		symbols:   make(map[string]*model.Symbol, 4096),
		overrides: make(map[int64][]*model.GroupSymbol, 512),
		resolved:  make(map[string]*Rules, 8192),
	}
}

// Load replaces everything at once. Anything half-loaded would let a trade be checked against
// a group that has symbols but no overrides yet.
func (s *Store) Load(groups map[string]*model.Group, symbols map[string]*model.Symbol,
	overrides map[int64][]*model.GroupSymbol) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.groups = groups
	s.symbols = symbols
	s.overrides = overrides
	s.resolved = make(map[string]*Rules, 8192)
}

// Group returns one group by its path.
func (s *Store) Group(path string) (*model.Group, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g, ok := s.groups[path]
	return g, ok
}

// Symbol returns one instrument by name.
func (s *Store) Symbol(name string) (*model.Symbol, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sym, ok := s.symbols[name]
	return sym, ok
}

// Groups is how many groups are loaded.
func (s *Store) Groups() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.groups)
}

// Symbols is how many instruments are loaded.
func (s *Store) Symbols() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.symbols)
}

// For resolves the rules for a group and symbol, or reports that the group may not trade it.
//
// A symbol the group has no override for is not tradable by that group at all: MT5 groups list
// what they can trade, so absence is a refusal rather than a default.
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

// overrideFor finds the group's entry covering this symbol. The list is in the group's own
// order and the first match wins, so a group can put `Forex\EURUSD` above `Forex\*` and have
// the specific one take effect.
func (s *Store) overrideFor(groupId int64, sym *model.Symbol) *model.GroupSymbol {
	for _, o := range s.overrides[groupId] {
		if matches(o.Path, sym) {
			return o
		}
	}
	return nil
}

// matches reports whether a group-symbol path covers an instrument. `*` alone covers
// everything, a trailing `*` covers a subtree, and anything else is the symbol or its path.
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

// resolve folds the override onto the instrument.
//
// An override that was never set is nil, and the instrument's own value stands. Setting a value
// to zero on purpose is a different thing from not setting it, and the pointers keep the two
// apart.
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

		SwapMode:  pick(o.SwapMode, sym.SwapMode),
		SwapLong:  pick(o.SwapLong, sym.SwapLong),
		SwapShort: pick(o.SwapShort, sym.SwapShort),

		SlipProfit:  pick(o.IESlipProfit, sym.SpreadDiff),
		SlipLosing:  pick(o.IESlipLosing, sym.SpreadDiff),
		IEVolumeMax: value(o.IEVolumeMax),
	}

	// a step of zero would make every volume invalid, so fall back to the smallest allowed
	if r.VolumeStep == 0 {
		r.VolumeStep = r.VolumeMin
	}

	return r
}

// pick is the override if the group set one, otherwise the instrument's own value.
func pick[T any](override *T, base T) T {
	if override != nil {
		return *override
	}
	return base
}

// value is an optional number as a plain one, zero when unset.
func value[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
