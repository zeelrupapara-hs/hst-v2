package model

import (
	"fmt"
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
//	ws.g.<family>.<group path>
//
// family says what kind of record it is and is gated on a manager right; the
// group path says where it lives and is gated on the manager's group access.
// Nothing else. Whether the record was added, replaced or removed is a property
// of the event, not of the address it was sent to, so it travels with the
// event and the subject stays the answer to one question: who may see this.
//
// Fields never appear either: a change publishes the whole record, so the
// reader replaces what it has.
const (
	// SubjectGroupRoot prefixes every group scoped subject.
	SubjectGroupRoot = "ws.g"

	// GroupSep separates the segments of a group path.
	GroupSep = `\`
)

// family is the kind of record an event is about: a group, a user, a client, an account or a group symbol.
type family string

const (
	familyGroups   family = "groups"
	familyUsers    family = "users"
	familyClients  family = "clients"
	familyAccounts family = "accounts"
	// familyGroupSymbols is a symbol override attached to a group, so it is
	// scoped by the group it hangs off rather than by the symbol.
	familyGroupSymbols     family = "group_symbols"
	familyGroupCommissions family = "group_commissions"
)

// familyRight is the manager right that gates a family. A manager without the
// right never subscribes to the family at all, whatever its group access.
var familyRight = map[family]uint{
	familyGroups:           MgrRightCfgGroups,
	familyUsers:            MgrRightAccRead,
	familyClients:          MgrRightClientsAccess,
	familyAccounts:         MgrRightAccRead,
	familyGroupSymbols:     MgrRightCfgGroups,
	familyGroupCommissions: MgrRightCfgGroups,
}

// families is every group scoped family, in a stable order.
var families = []family{
	familyGroups, familyUsers, familyClients, familyAccounts,
	familyGroupSymbols, familyGroupCommissions,
}

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

// From the api to the outside world "websocket". One name per subject, so a
// publisher says what happened rather than assembling a string, and a family
// can never be paired with the wrong action by accident.
//
//	s.NotifyWS(model.SubjectGroup(v.Group), model.EventCreated, v)
//	  -> ws.g.groups.demo.forex
var (
	SubjectGroup           = func(path string) string { return subject(familyGroups, path) }
	SubjectGroupSymbol     = func(path string) string { return subject(familyGroupSymbols, path) }
	SubjectUser            = func(path string) string { return subject(familyUsers, path) }
	SubjectClient          = func(path string) string { return subject(familyClients, path) }
	SubjectAccount         = func(path string) string { return subject(familyAccounts, path) }
	SubjectGroupCommission = func(path string) string { return subject(familyGroupCommissions, path) }

	// SubjectJournal carries a manager's own journal lines: it records what that manager did, so nobody else is listening.
	SubjectJournal = func(login int64) string { return fmt.Sprintf("websocket.%d.journal", login) }
)

// Records with no group of their own. Access to them is a right, not a path, so the subject carries no group.
const (
	SubjectSymbol   = "ws.right.symbols"
	SubjectHoliday  = "ws.right.holidays"
	SubjectLeverage = "ws.right.leverages"
	SubjectManager  = "ws.right.managers"
)

// From the api to the other services. A system subject is not a websocket
// subject: nobody's group access decides who reads it, because the reader is
// another service reloading its own state rather than a person watching a
// screen. It carries the same payload as the websocket event.
const (
	SubjectSystemGroupCreated = "system.group.created"
	SubjectSystemGroupUpdated = "system.group.updated"
	SubjectSystemGroupDeleted = "system.group.deleted"

	SubjectSystemGroupSymbolCreated = "system.group_symbol.created"
	SubjectSystemGroupSymbolUpdated = "system.group_symbol.updated"
	SubjectSystemGroupSymbolDeleted = "system.group_symbol.deleted"

	SubjectSystemUserCreated = "system.user.created"
	SubjectSystemUserUpdated = "system.user.updated"
	SubjectSystemUserDeleted = "system.user.deleted"

	SubjectSystemClientCreated = "system.client.created"
	SubjectSystemClientUpdated = "system.client.updated"
	SubjectSystemClientDeleted = "system.client.deleted"

	SubjectSystemGroupCommissionCreated = "system.group_commission.created"
	SubjectSystemGroupCommissionUpdated = "system.group_commission.updated"
	SubjectSystemGroupCommissionDeleted = "system.group_commission.deleted"

	SubjectSystemSymbolCreated = "system.symbol.created"
	SubjectSystemSymbolUpdated = "system.symbol.updated"
	SubjectSystemSymbolDeleted = "system.symbol.deleted"

	SubjectSystemHolidayCreated   = "system.holiday.created"
	SubjectSystemHolidayUpdated   = "system.holiday.updated"
	SubjectSystemHolidayDeleted   = "system.holiday.deleted"
	SubjectSystemHolidayReordered = "system.holiday.reordered"

	SubjectSystemLeverageCreated   = "system.leverage.created"
	SubjectSystemLeverageUpdated   = "system.leverage.updated"
	SubjectSystemLeverageDeleted   = "system.leverage.deleted"
	SubjectSystemLeverageReordered = "system.leverage.reordered"

	SubjectSystemManagerCreated = "system.manager.created"
	SubjectSystemManagerUpdated = "system.manager.updated"
	SubjectSystemManagerDeleted = "system.manager.deleted"
)

// The event types a record change carries. They ride in the event rather than
// in the subject, so a reader that wants everything about a group subscribes
// once instead of once per verb.
//
// Each one names the record as well as the verb. A reader handling group_created
// and user_created off the same connection should not have to look at where a
// message arrived to know which of the two it is holding.
const (
	EventGroupCreated = "group_created"
	EventGroupUpdated = "group_updated"
	EventGroupDeleted = "group_deleted"

	EventGroupSymbolCreated = "group_symbol_created"
	EventGroupSymbolUpdated = "group_symbol_updated"
	EventGroupSymbolDeleted = "group_symbol_deleted"

	EventUserCreated = "user_created"
	EventUserUpdated = "user_updated"
	EventUserDeleted = "user_deleted"
	EventUserMoved   = "user_moved"

	EventClientCreated = "client_created"
	EventClientUpdated = "client_updated"
	EventClientDeleted = "client_deleted"

	EventAccountUpdated = "account_updated"

	EventGroupCommissionCreated = "group_commission_created"
	EventGroupCommissionUpdated = "group_commission_updated"
	EventGroupCommissionDeleted = "group_commission_deleted"

	EventSymbolCreated = "symbol_created"
	EventSymbolUpdated = "symbol_updated"
	EventSymbolDeleted = "symbol_deleted"

	EventHolidayCreated   = "holiday_created"
	EventHolidayUpdated   = "holiday_updated"
	EventHolidayDeleted   = "holiday_deleted"
	EventHolidayReordered = "holiday_reordered"

	EventLeverageCreated     = "leverage_created"
	EventLeverageUpdated     = "leverage_updated"
	EventLeverageDeleted     = "leverage_deleted"
	EventLeverageRuleCreated = "leverage_rule_created"
	EventLeverageRuleUpdated = "leverage_rule_updated"
	EventLeverageRuleDeleted = "leverage_rule_deleted"
	EventLeverageReordered   = "leverage_reordered"

	EventManagerCreated = "manager_created"
	EventManagerUpdated = "manager_updated"
	EventManagerDeleted = "manager_deleted"

	EventJournal = "journal"
)

// subject builds a group scoped subject. The named builders above are the only
// way in; this is the one place that knows the shape.
//
//	Subject(familyGroups, `demo\forex`, actionCreated)
//	  -> ws.g.groups.demo.forex.created
func subject(f family, groupPath string) string {
	token := GroupToken(groupPath)
	if token == "" {
		// a record with no group still has to go somewhere its family's
		// subscribers can hear it, and "root" is a segment no group can own
		token = "root"
	}
	return SubjectGroupRoot + "." + string(f) + "." + token
}

// maskSubject turns one of a manager's group masks into the subscription that
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
func buildSubjectPatterns(f family, mask string) []string {
	prefix := SubjectGroupRoot + "." + string(f) + "."

	segs := segments(mask)

	// a bare * covers every group
	if len(segs) == 0 || (len(segs) == 1 && segs[0] == "*") {
		return []string{prefix + ">"}
	}

	// a trailing * is "and everything below": the group itself, and the
	// subtree under it
	if segs[len(segs)-1] == "*" {
		path := strings.Join(segs[:len(segs)-1], ".")
		return []string{prefix + path, prefix + path + ".>"}
	}

	// anything else names one group and grants only that one. The subject
	// carries no action, so this is a single exact address rather than a
	// prefix, and a child group is no longer swept up with its parent.
	return []string{prefix + strings.Join(segs, ".")}
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

	out := make([]string, 0, len(families)*len(masks)*2)
	for _, f := range families {
		right, ok := familyRight[f]
		if !ok || !rights.Has(right) {
			continue
		}
		for _, mask := range masks {
			out = append(out, buildSubjectPatterns(f, mask)...)
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

// MasksCover reports whether every mask in inner is already granted by outer.
//
// This is what confines delegation: a manager may hand out a slice of its own
// access and never more, so the region it was given is the whole world it can
// carve up.
func MasksCover(outer, inner []string) bool {
	for _, want := range inner {
		if strings.TrimSpace(want) == "" {
			continue
		}

		granted := false
		for _, have := range outer {
			if MaskCovers(have, want) {
				granted = true
				break
			}
		}
		if !granted {
			return false
		}
	}

	return true
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
