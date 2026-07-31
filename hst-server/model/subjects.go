package model

import (
	"strings"
)

// Group scoped websocket subjects.
//
// A group path is a hierarchy separated by a backslash, demo\forex\usd, and a
// nats subject is a hierarchy separated by a dot. Mapping one onto the other
// makes nats's own matcher the access control engine:
//
//	manager access  demo\*        subscribes  ws.g.groups.demo.>
//	group created   demo\forex    publishes   ws.g.groups.demo.forex.created
//
// The manager receives it because "demo.forex.created" matches "demo.>". No
// lookup, no filter, no bookkeeping. A group created tomorrow under demo is
// matched by the same subscription, which is what "access to demo\* covers
// everything under it" has to mean.
//
// The subject carries three things and nothing else:
//
//	ws.g.<family>.<group path>.<action>
//
// family says what kind of record it is and is gated on a manager right, the
// group path says where it lives and is gated on the manager's group access,
// and the action says whether to add, replace or remove it. Fields never
// appear: a change publishes the whole record, so the ui replaces what it has.
const (
	// SubjectGroupRoot prefixes every group scoped subject.
	SubjectGroupRoot = "ws.g"

	// GroupSep separates the segments of a group path.
	GroupSep = `\`
)

// Family is the kind of record an event is about. Each one is gated on a
// manager right; see FamilyRight.
type Family string

const (
	FamilyGroups   Family = "groups"
	FamilyUsers    Family = "users"
	FamilyClients  Family = "clients"
	FamilyAccounts Family = "accounts"
)

// Action is what happened to the record. Four is the whole vocabulary: a
// change publishes the entire record, so the reader never has to know which
// field moved.
type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
	// ActionMoved is for a record whose group path itself changed, which is
	// the one case where the old and the new audience differ.
	ActionMoved Action = "moved"
)

// FamilyRight is the manager right that gates a family. A manager without the
// right never subscribes to the family at all, whatever its group access.
var FamilyRight = map[Family]uint{
	FamilyGroups:   MgrRightCfgGroups,
	FamilyUsers:    MgrRightAccRead,
	FamilyClients:  MgrRightClientsAccess,
	FamilyAccounts: MgrRightAccRead,
}

// Families is every group scoped family, in a stable order.
var Families = []Family{FamilyGroups, FamilyUsers, FamilyClients, FamilyAccounts}

// GroupToken turns a group path into subject tokens: demo\forex\usd becomes
// demo.forex.usd. Empty segments are dropped so a stray separator cannot
// produce an empty token, which nats rejects.
func GroupToken(path string) string {
	parts := strings.Split(path, GroupSep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ".")
}

// Subject is what a publisher sends on. The group path is the record's own
// group, never a mask.
//
//	Subject(FamilyGroups, `demo\forex`, ActionCreated)
//	  -> ws.g.groups.demo.forex.created
func Subject(family Family, groupPath string, action Action) string {
	token := GroupToken(groupPath)
	if token == "" {
		// a record with no group still has to go somewhere its family's
		// subscribers can hear it, and "root" is a segment no group can own
		token = "root"
	}
	return SubjectGroupRoot + "." + string(family) + "." + token + "." + string(action)
}

// MaskSubject turns one of a manager's group masks into the subscription that
// covers it.
//
//	demo\*     -> ws.g.<family>.demo.>
//	demo\9\a   -> ws.g.<family>.demo.9.a.>
//	demo\*\a   -> ws.g.<family>.demo.*.a.>
//	*          -> ws.g.<family>.>
//
// A trailing * becomes >, because "everything below here" is what the mask
// means and > is the token that says so. A * in the middle stays *, which
// matches exactly one segment in both notations.
func MaskSubject(family Family, mask string) string {
	prefix := SubjectGroupRoot + "." + string(family) + "."

	parts := strings.Split(mask, GroupSep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}

	// a bare * covers every group
	if len(out) == 0 || (len(out) == 1 && out[0] == "*") {
		return prefix + ">"
	}

	// a trailing * is "and everything below", which is > in nats
	if out[len(out)-1] == "*" {
		out = out[:len(out)-1]
	}

	return prefix + strings.Join(out, ".") + ".>"
}

// Subscriptions is every subject one manager session should listen on: each
// family it holds the right for, crossed with each group mask it holds.
//
// Overlapping masks are reduced first. A manager holding both demo\* and
// demo\9\a matches twice otherwise, and receives every event of that family
// twice.
func Subscriptions(rights ManagerRights, masks []string) []string {
	masks = ReduceMasks(masks)
	if len(masks) == 0 {
		return nil
	}

	out := make([]string, 0, len(Families)*len(masks))
	for _, family := range Families {
		right, ok := FamilyRight[family]
		if !ok || !rights.Has(right) {
			continue
		}
		for _, mask := range masks {
			out = append(out, MaskSubject(family, mask))
		}
	}

	return out
}

// ReduceMasks drops any mask already covered by a broader one, so no event is
// delivered twice to the same socket.
func ReduceMasks(masks []string) []string {
	cleaned := make([]string, 0, len(masks))
	for _, m := range masks {
		if m = strings.TrimSpace(m); m != "" {
			cleaned = append(cleaned, m)
		}
	}

	out := make([]string, 0, len(cleaned))
	for i, m := range cleaned {
		covered := false
		for j, other := range cleaned {
			if i == j {
				continue
			}
			// on an exact duplicate keep the first and drop the rest
			if m == other {
				if j < i {
					covered = true
					break
				}
				continue
			}
			if MaskCovers(other, m) {
				covered = true
				break
			}
		}
		if !covered {
			out = append(out, m)
		}
	}

	return out
}

// MaskCovers reports whether outer already grants everything inner grants.
func MaskCovers(outer, inner string) bool {
	o := segments(outer)
	i := segments(inner)

	// a bare * covers everything
	if len(o) == 1 && o[0] == "*" {
		return true
	}

	// outer cannot cover a shorter path than itself, unless it ends in *
	trailing := len(o) > 0 && o[len(o)-1] == "*"
	if trailing {
		o = o[:len(o)-1]
	}
	if len(i) < len(o) {
		return false
	}
	if !trailing && len(i) != len(o) {
		return false
	}

	for n := range o {
		if o[n] == "*" || o[n] == i[n] {
			continue
		}
		return false
	}

	return true
}

func segments(mask string) []string {
	parts := strings.Split(mask, GroupSep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
