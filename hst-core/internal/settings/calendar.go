package settings

import (
	"strings"
	"sync"
	"time"
)

type Window struct {
	Day   int16
	Type  int16
	Open  int32
	Close int32
}

type Sessions struct {
	mu sync.RWMutex
	// keyed by symbol id, then weekday, then session type
	windows map[int64]map[int16]map[int16][]Window
}

func NewSessions() *Sessions {
	return &Sessions{windows: make(map[int64]map[int16]map[int16][]Window, 4096)}
}

func (s *Sessions) Load(windows map[int64][]Window) {
	next := make(map[int64]map[int16]map[int16][]Window, len(windows))

	for symbolId, list := range windows {
		byDay := make(map[int16]map[int16][]Window, 7)
		for _, w := range list {
			if byDay[w.Day] == nil {
				byDay[w.Day] = make(map[int16][]Window, 2)
			}
			byDay[w.Day][w.Type] = append(byDay[w.Day][w.Type], w)
		}
		next[symbolId] = byDay
	}

	s.mu.Lock()
	s.windows = next
	s.mu.Unlock()
}

func (s *Sessions) For(symbolId int64, day, kind int16) []Window {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.windows[symbolId][day][kind]
}

func (s *Sessions) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.windows)
}

type Holiday struct {
	Year    int32
	Month   int16
	Day     int16
	From    int32
	To      int32
	Symbols []string
	Enable  bool
}

type Holidays struct {
	mu   sync.RWMutex
	days []Holiday
}

func NewHolidays() *Holidays { return &Holidays{} }

func (h *Holidays) Load(days []Holiday) {
	h.mu.Lock()
	h.days = days
	h.mu.Unlock()
}

func (h *Holidays) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.days)
}

func (h *Holidays) Covers(path, symbol string, at time.Time) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	minutes := int32(at.Hour()*60 + at.Minute())

	for i := range h.days {
		d := &h.days[i]

		if !d.Enable {
			continue
		}
		if d.Year != 0 && d.Year != int32(at.Year()) {
			continue
		}
		if d.Month != int16(at.Month()) || d.Day != int16(at.Day()) {
			continue
		}
		if !holidayCovers(d.Symbols, path, symbol) {
			continue
		}
		if d.From == 0 && d.To == 0 {
			return true
		}
		if minutes >= d.From && minutes < d.To {
			return true
		}
	}

	return false
}

func holidayCovers(masks []string, path, symbol string) bool {
	if len(masks) == 0 {
		return true
	}

	for _, target := range masks {
		target = strings.TrimSpace(target)
		if target == "" || target == "*" {
			return true
		}
		if maskHits(target, path, symbol) {
			return true
		}
	}

	return false
}

func maskHits(target, path, symbol string) bool {
	switch {
	case target == "" || target == "*":
		return true
	case strings.HasSuffix(target, "*"):
		prefix := strings.TrimSuffix(target, "*")
		return strings.HasPrefix(path, prefix) || strings.HasPrefix(symbol, prefix)
	default:
		return target == symbol || target == path
	}
}
