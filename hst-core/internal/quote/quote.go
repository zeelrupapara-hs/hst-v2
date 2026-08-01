// Package quote keeps the latest price for every symbol.
package quote

import (
	"sync"

	"hstcore/model"
)

// Book holds the last quote per symbol.
type Book struct {
	mu     sync.RWMutex
	ticks  map[string]model.Tick
	gapped map[string]bool
}

// New builds an empty book.
func New() *Book {
	return &Book{
		ticks:  make(map[string]model.Tick, 4096),
		gapped: make(map[string]bool, 64),
	}
}

// Set records a new quote and reports the one it replaced, so the caller can tell what moved.
func (b *Book) Set(t model.Tick) (previous model.Tick, had bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	previous, had = b.ticks[t.Symbol]
	b.ticks[t.Symbol] = t

	return previous, had
}

// Get returns the last quote for a symbol.
func (b *Book) Get(symbol string) (model.Tick, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	t, ok := b.ticks[symbol]
	return t, ok
}

// Len is how many symbols have ever quoted.
func (b *Book) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.ticks)
}

// Symbols lists every symbol that has quoted.
func (b *Book) Symbols() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]string, 0, len(b.ticks))
	for s := range b.ticks {
		out = append(out, s)
	}

	return out
}

// SetGap marks a symbol as having jumped.
func (b *Book) SetGap(symbol string, gapped bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gapped {
		b.gapped[symbol] = true
		return
	}
	delete(b.gapped, symbol)
}

// Gapped reports whether the symbol is currently in a gap.
func (b *Book) Gapped(symbol string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.gapped[symbol]
}
