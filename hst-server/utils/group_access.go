package utils

import (
	"fmt"
	"strings"
)

// GroupAccess turns a manager's group masks into a SQL predicate for a list endpoint
func GroupAccess(masks []string, column string, next int) (string, []any) {
	// no masks is no access, which is not the same as unrestricted; a manager nobody granted anything to should see nothing
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
			// the common shape, demo\*: the group itself and its whole subtree.
			prefix := strings.TrimSuffix(mask, `*`)
			parts = append(parts, fmt.Sprintf(
				"(%s = $%d OR starts_with(%s, $%d))", column, next, column, next+1))
			args = append(args, strings.TrimSuffix(prefix, `\`), prefix)
			next += 2

		default:
			// anything else, real\*\a among them, needs one segment matched at a time, which only a regex expresses
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

// regexQuote escapes the characters a group name could contain that would otherwise be read as regex syntax.
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

// regexQuote escapes what a group name could contain that reads as regex syntax
func GroupAccessFor(isManager bool, masks []string, column string, next int) (string, []any) {
	if !isManager {
		return "FALSE", nil
	}
	return GroupAccess(masks, column, next)
}
