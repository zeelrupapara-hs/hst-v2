// Package book holds the accounts this pod is responsible for.
package book

import (
	"sync"

	"hstcore/model"
)

// Entry is one account and everything open on it.
type Entry struct {
	mu sync.Mutex

	Account   *model.Account
	Orders    map[int64]*model.Order
	Positions map[int64]*model.Position
}

// Lock takes the account for writing.
func (e *Entry) Lock() { e.mu.Lock() }

// Unlock releases it.
func (e *Entry) Unlock() { e.mu.Unlock() }

// Symbols lists every instrument this account has something open on.
func (e *Entry) Symbols() []string {
	seen := make(map[string]bool, len(e.Positions)+len(e.Orders))
	for _, p := range e.Positions {
		seen[p.Symbol] = true
	}
	for _, o := range e.Orders {
		seen[o.Symbol] = true
	}

	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}

	return out
}

// Book is every account this pod holds, with an index from symbol to the accounts that care.
type Book struct {
	mu       sync.RWMutex
	accounts map[int64]*Entry
	bySymbol map[string]map[int64]*Entry
}

// New builds an empty book.
func New() *Book {
	return &Book{
		accounts: make(map[int64]*Entry, 16384),
		bySymbol: make(map[string]map[int64]*Entry, 4096),
	}
}

// Add puts an account in the book and indexes what it holds.
func (b *Book) Add(e *Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.accounts[e.Account.Login] = e
	for _, sym := range e.Symbols() {
		b.indexLocked(sym, e)
	}
}

// Remove drops an account, and any symbol index entries it was the last holder of.
func (b *Book) Remove(login int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.accounts[login]
	if !ok {
		return
	}

	for _, sym := range e.Symbols() {
		if set := b.bySymbol[sym]; set != nil {
			delete(set, login)
			if len(set) == 0 {
				delete(b.bySymbol, sym)
			}
		}
	}

	delete(b.accounts, login)
}

// Get finds one account.
func (b *Book) Get(login int64) (*Entry, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	e, ok := b.accounts[login]
	return e, ok
}

// Len is how many accounts this pod holds.
func (b *Book) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.accounts)
}

// Watching returns the accounts holding something on this symbol.
func (b *Book) Watching(symbol string) []*Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	set := b.bySymbol[symbol]
	if len(set) == 0 {
		return nil
	}

	out := make([]*Entry, 0, len(set))
	for _, e := range set {
		out = append(out, e)
	}

	return out
}

// Watch records that an account now cares about a symbol.
func (b *Book) Watch(symbol string, e *Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.indexLocked(symbol, e)
}

// Unwatch drops the account from a symbol's index once it holds nothing there any more.
func (b *Book) Unwatch(symbol string, login int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	set := b.bySymbol[symbol]
	if set == nil {
		return
	}

	delete(set, login)
	if len(set) == 0 {
		delete(b.bySymbol, symbol)
	}
}

// Each walks every account. Used by the jobs that run over the whole book, like the daily swap.
func (b *Book) Each(f func(*Entry)) {
	b.mu.RLock()
	all := make([]*Entry, 0, len(b.accounts))
	for _, e := range b.accounts {
		all = append(all, e)
	}
	b.mu.RUnlock()

	for _, e := range all {
		f(e)
	}
}

// indexLocked adds to the symbol index. The book's write lock must already be held.
func (b *Book) indexLocked(symbol string, e *Entry) {
	set := b.bySymbol[symbol]
	if set == nil {
		set = make(map[int64]*Entry, 8)
		b.bySymbol[symbol] = set
	}
	set[e.Account.Login] = e
}
