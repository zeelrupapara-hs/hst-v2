package utils

import (
	"fmt"
	"strings"
)

// GroupAccess turns a manager's group masks into a SQL predicate, so a list
// endpoint returns the rows that manager may see and nothing else.
//
// The websocket routes events by the same masks. Without this the two
// disagree: a manager granted demo\* would never be told about a group under
// real, but could still read every one of them from the list endpoint.
//
// column is the quoted column holding the group path, next is the number of
// the first placeholder to use. The returned predicate is always safe to drop
// into a WHERE with AND; it is "TRUE" when the manager may see everything and
// "FALSE" when the manager has no access at all.
func GroupAccess(masks []string, column string, next int) (string, []any) {
	// no masks is no access, which is not the same as unrestricted; a manager
	// nobody granted anything to should see nothing
	if len(masks) == 0 {
		return "FALSE", nil
	}

	var (
		parts []string
		args  []any
	)

	for _, mask := range masks {
		mask = strings.TrimSpace(mask)
		if mask == "" {
			continue
		}

		// a bare * is every group, so the whole predicate collapses
		if mask == "*" {
			return "TRUE", nil
		}

		star := strings.Count(mask, "*")

		switch {
		case star == 0:
			// an exact group and nothing below it
			parts = append(parts, fmt.Sprintf("%s = $%d", column, next))
			args = append(args, mask)
			next++

		case star == 1 && strings.HasSuffix(mask, `\*`):
			// the common shape, demo\*: the group itself and its whole
			// subtree. starts_with keeps this indexable and sidesteps LIKE,
			// where the backslash is itself the escape character
			prefix := strings.TrimSuffix(mask, `*`)
			parts = append(parts, fmt.Sprintf(
				"(%s = $%d OR starts_with(%s, $%d))", column, next, column, next+1))
			args = append(args, strings.TrimSuffix(prefix, `\`), prefix)
			next += 2

		default:
			// anything else, real\*\a among them, needs one segment matched at
			// a time, which only a regex expresses
			parts = append(parts, fmt.Sprintf("%s ~ $%d", column, next))
			args = append(args, maskRegex(mask))
			next++
		}
	}

	if len(parts) == 0 {
		return "FALSE", nil
	}

	return "(" + strings.Join(parts, " OR ") + ")", args
}

// maskRegex builds an anchored regex for a mask with a wildcard in the middle.
// A * stands for exactly one segment, never for a separator, or demo\*\a would
// match demo\x\y\a and grant more than was given.
func maskRegex(mask string) string {
	segments := strings.Split(mask, `\`)

	var b strings.Builder
	b.WriteString("^")
	for i, seg := range segments {
		if i > 0 {
			b.WriteString(`\\`)
		}
		if seg == "*" {
			// one segment, separators excluded
			b.WriteString(`[^\\]+`)
			continue
		}
		b.WriteString(regexQuote(seg))
	}
	b.WriteString("$")

	return b.String()
}

// regexQuote escapes the characters a group name could contain that would
// otherwise be read as regex syntax.
func regexQuote(s string) string {
	const special = `\.+*?()|[]{}^$`

	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(special, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// GroupAccessFor is GroupAccess for a session, with the unrestricted case
// spelled out: a login that is not a manager has no group access at all.
func GroupAccessFor(isManager bool, masks []string, column string, next int) (string, []any) {
	if !isManager {
		return "FALSE", nil
	}
	return GroupAccess(masks, column, next)
}
