package symbolpath

import (
	"regexp"
	"strings"
)

// SymbolRef is a catalog symbol used for path mask matching.
type SymbolRef struct {
	SymbolID int64
	Symbol   string
	Path     string
}

type condition struct {
	exclude bool
	pattern string
}

// MatchMasks expands MT5-style comma-separated path/name masks into symbol ids.
// Conditions apply sequentially: includes add matches, excludes remove them.
func MatchMasks(masks string, symbols []SymbolRef) []int64 {
	conds := parseConditions(masks)
	if len(conds) == 0 {
		return nil
	}

	selected := map[int64]struct{}{}
	for _, cond := range conds {
		for _, sym := range symbols {
			if !matchesSymbol(cond.pattern, sym) {
				continue
			}
			if cond.exclude {
				delete(selected, sym.SymbolID)
			} else {
				selected[sym.SymbolID] = struct{}{}
			}
		}
	}

	out := make([]int64, 0, len(selected))
	for id := range selected {
		out = append(out, id)
	}
	return out
}

func parseConditions(masks string) []condition {
	masks = strings.TrimSpace(masks)
	if masks == "" {
		return nil
	}
	parts := strings.Split(masks, ",")
	out := make([]condition, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		excl := false
		if strings.HasPrefix(part, "!") {
			excl = true
			part = strings.TrimSpace(strings.TrimPrefix(part, "!"))
		}
		if part == "" {
			continue
		}
		out = append(out, condition{exclude: excl, pattern: part})
	}
	return out
}

func matchesSymbol(pattern string, sym SymbolRef) bool {
	return matchPattern(pattern, sym.Path) || matchPattern(pattern, sym.Symbol)
}

func matchPattern(pattern, value string) bool {
	pattern = strings.TrimSpace(pattern)
	value = strings.TrimSpace(value)
	if pattern == "" || value == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	re := globToRegexp(pattern)
	if re == nil {
		return pattern == value
	}
	return re.MatchString(value)
}

func globToRegexp(pattern string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		switch ch {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteString(".")
		case '.', '+', '(', ')', '|', '^', '$', '[', ']', '{', '}', '\\':
			b.WriteByte('\\')
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	return re
}

// ApplyScopeRule adds or removes catalog symbols matching pattern from selected.
func ApplyScopeRule(selected map[int64]struct{}, exclude bool, pattern string, symbols []SymbolRef) {
	for _, sym := range symbols {
		if !matchesSymbol(pattern, sym) {
			continue
		}
		if exclude {
			delete(selected, sym.SymbolID)
		} else {
			selected[sym.SymbolID] = struct{}{}
		}
	}
}

// ApplyScopeSymbol adds or removes one symbol id from selected.
func ApplyScopeSymbol(selected map[int64]struct{}, symbolID int64, exclude bool) {
	if symbolID <= 0 {
		return
	}
	if exclude {
		delete(selected, symbolID)
	} else {
		selected[symbolID] = struct{}{}
	}
}
