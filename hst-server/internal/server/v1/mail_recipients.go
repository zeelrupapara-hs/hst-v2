package v1

import (
	"fmt"
	"strconv"
	"strings"

	"hstserver/utils"
)

// The To field of the mail dialog takes an expression rather than a login list:
//
//	expr    := orGroup (';' orGroup)*      groups deliver to their union
//	orGroup := term (',' term)*
//	term    := ['!'] (login | lo-hi | group:mask | country:text | city:text)
//
// Inside one ';' group, positive terms of the same kind widen each other (any of these logins)
// while different kinds narrow (in this group AND in this country); '!' terms always narrow.
// A group of nothing but negations means everything in the caller's reach except those.

// mailTermKind is what a term filters on.
type mailTermKind int

const (
	mailTermLogin mailTermKind = iota
	mailTermRange
	mailTermGroup
	mailTermCountry
	mailTermCity
)

// mailTerm is one parsed term of the To expression.
type mailTerm struct {
	neg    bool
	kind   mailTermKind
	lo, hi int64
	text   string
}

// parseMailTo splits a To expression into its ';' groups of terms.
func parseMailTo(to string) ([][]mailTerm, error) {
	var groups [][]mailTerm

	for _, rawGroup := range strings.Split(to, ";") {
		var terms []mailTerm

		for _, raw := range strings.Split(rawGroup, ",") {
			tok := strings.TrimSpace(raw)
			if tok == "" {
				continue
			}

			term, err := parseMailTerm(tok)
			if err != nil {
				return nil, err
			}
			terms = append(terms, term)
		}

		if len(terms) > 0 {
			groups = append(groups, terms)
		}
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("the To field names nobody")
	}

	return groups, nil
}

func parseMailTerm(tok string) (mailTerm, error) {
	var t mailTerm

	if strings.HasPrefix(tok, "!") {
		t.neg = true
		tok = strings.TrimSpace(strings.TrimPrefix(tok, "!"))
	}

	if kind, text, ok := prefixedTerm(tok); ok {
		if text == "" {
			return t, fmt.Errorf("recipient term %q names nothing after the colon", tok)
		}
		t.kind, t.text = kind, text
		return t, nil
	}

	if lo, hi, ok := strings.Cut(tok, "-"); ok {
		var err error
		if t.lo, err = strconv.ParseInt(strings.TrimSpace(lo), 10, 64); err != nil {
			return t, fmt.Errorf("recipient range %q does not start with a login", tok)
		}
		if t.hi, err = strconv.ParseInt(strings.TrimSpace(hi), 10, 64); err != nil {
			return t, fmt.Errorf("recipient range %q does not end with a login", tok)
		}
		if t.lo > t.hi {
			t.lo, t.hi = t.hi, t.lo
		}
		t.kind = mailTermRange
		return t, nil
	}

	login, err := strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return t, fmt.Errorf("recipient term %q is not a login, a range, or a group:/country:/city: term", tok)
	}
	t.kind, t.lo = mailTermLogin, login

	return t, nil
}

func prefixedTerm(tok string) (mailTermKind, string, bool) {
	for prefix, kind := range map[string]mailTermKind{
		"group:":   mailTermGroup,
		"country:": mailTermCountry,
		"city:":    mailTermCity,
	} {
		if len(tok) >= len(prefix) && strings.EqualFold(tok[:len(prefix)], prefix) {
			return kind, strings.TrimSpace(tok[len(prefix):]), true
		}
	}

	return 0, "", false
}

// mailToWhere renders the parsed expression as a predicate on the joined user row u.
// Placeholders start at next; the caller advances by len(args).
func mailToWhere(groups [][]mailTerm, next int) (string, []any) {
	var (
		parts []string
		args  []any
	)

	for _, terms := range groups {
		part, groupArgs := mailGroupWhere(terms, next)
		parts = append(parts, part)
		args = append(args, groupArgs...)
		next += len(groupArgs)
	}

	return "(" + strings.Join(parts, " OR ") + ")", args
}

// mailGroupWhere renders one ';' group: same-kind positives union, kinds intersect, '!' excludes.
func mailGroupWhere(terms []mailTerm, next int) (string, []any) {
	var (
		buckets = map[mailTermKind][]string{}
		nots    []string
		args    []any
	)

	for _, t := range terms {
		pred, termArgs := mailTermWhere(t, next)
		args = append(args, termArgs...)
		next += len(termArgs)

		if t.neg {
			nots = append(nots, pred)
			continue
		}

		// logins and ranges name the same field, so they share a bucket
		kind := t.kind
		if kind == mailTermRange {
			kind = mailTermLogin
		}
		buckets[kind] = append(buckets[kind], pred)
	}

	var ands []string
	for _, kind := range []mailTermKind{mailTermLogin, mailTermGroup, mailTermCountry, mailTermCity} {
		if preds := buckets[kind]; len(preds) > 0 {
			ands = append(ands, "("+strings.Join(preds, " OR ")+")")
		}
	}
	for _, pred := range nots {
		ands = append(ands, "NOT ("+pred+")")
	}

	if len(ands) == 0 {
		return "TRUE", nil
	}

	return "(" + strings.Join(ands, " AND ") + ")", args
}

func mailTermWhere(t mailTerm, next int) (string, []any) {
	switch t.kind {
	case mailTermRange:
		return fmt.Sprintf("u.login BETWEEN $%d AND $%d", next, next+1), []any{t.lo, t.hi}

	case mailTermGroup:
		// the same matcher the manager's own masks go through, so demo\* covers the subtree
		return utils.GroupAccess([]string{t.text}, `u."group"`, next)

	case mailTermCountry:
		return fmt.Sprintf("u.country ILIKE $%d", next), []any{likePattern(t.text)}

	case mailTermCity:
		return fmt.Sprintf("u.city ILIKE $%d", next), []any{likePattern(t.text)}

	default:
		return fmt.Sprintf("u.login = $%d", next), []any{t.lo}
	}
}

// likePattern turns the term's * wildcard into %, with LIKE's own syntax neutralised.
func likePattern(text string) string {
	text = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(text)
	return strings.ReplaceAll(text, "*", "%")
}
