package model

import (
	"strings"
	"time"
)

// HolidayMode is the Mode column, the platform's Enable checkbox.
type HolidayMode int32

const (
	HolidayMode_disabled HolidayMode = 0
	HolidayMode_enabled  HolidayMode = 1
)

// Enum value maps for HolidayMode.
var (
	HolidayMode_name  = map[int32]string{0: "disabled", 1: "enabled"}
	HolidayMode_value = map[string]int32{"disabled": 0, "enabled": 1}
)

// Holiday narrows the server work time for a set of symbol masks on one date.
type Holiday struct {
	HolidayId   int64       `db:"holiday_id" json:"holiday_id"`
	Year        int32       `db:"year" json:"year"`
	Month       int16       `db:"month" json:"month"`
	Day         int16       `db:"day" json:"day"`
	From        int32       `db:"from" json:"from"`
	To          int32       `db:"to" json:"to"`
	Description string      `db:"description" json:"description"`
	Timestamp   int64       `db:"timestamp" json:"timestamp"`
	Mode        HolidayMode `db:"mode" json:"mode"`
	Symbols     []string    `db:"symbols" json:"symbols"`
	ConfigIndex int32       `db:"config_index" json:"config_index"`
}

func (Holiday) TableName() string { return "hst.holidays" }

// OnDate reports whether the holiday falls on t. Year 0 repeats every year.
func (h Holiday) OnDate(t time.Time) bool {
	if h.Year != 0 && int(h.Year) != t.Year() {
		return false
	}
	return int(h.Month) == int(t.Month()) && int(h.Day) == t.Day()
}

// Covers reports whether one of the holiday masks matches the symbol or its path.
func (h Holiday) Covers(symbol, path string) bool {
	for _, m := range h.Symbols {
		if matchMask(m, symbol, path) {
			return true
		}
	}
	return false
}

// Closed reports whether the record leaves no working time, the platform's "leave zero values in these.
func (h Holiday) Closed() bool { return h.From == 0 && h.To == 0 }

// matchMask matches one symbol mask against a symbol and its path.
// ponytail: only the prefix form of * is honoured, which is every mask the terminal
func matchMask(mask, symbol, path string) bool {
	if mask == "" || mask == "*" {
		return true
	}
	if i := strings.IndexByte(mask, '*'); i >= 0 {
		p := strings.ToLower(mask[:i])
		return strings.HasPrefix(strings.ToLower(symbol), p) ||
			strings.HasPrefix(strings.ToLower(path), p)
	}
	return strings.EqualFold(mask, symbol) || strings.EqualFold(mask, path)
}
