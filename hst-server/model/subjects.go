package model

import (
	wire "hstmodel"

	"strings"
)

// GroupSep separates the segments of a group path.
const GroupSep = `\`

// SubjectGroupRoot prefixes every group scoped subject, under the websocket root so the delivery
// rule holds: anything beginning websocket reaches a client.
const SubjectGroupRoot = wire.RootGroupScoped

// family is the kind of record an event is about: a group, a user, a client, an account or a group symbol.
type family string

const (
	familyGroups   family = "groups"
	familyUsers    family = "users"
	familyClients  family = "clients"
	familyAccounts family = "accounts"
	// family is gated on a right, the group path on the manager's access masks
	familyGroupSymbols     family = "group_symbols"
	familyGroupCommissions family = "group_commissions"
)

// familyRight is the manager right that gates a family.
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

// GroupToken turns a group path into subject tokens: demo\forex\usd becomes demo.forex.usd.
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

// From the api to the outside world "websocket".
var (
	SubjectGroup           = func(path string) string { return subject(familyGroups, path) }
	SubjectGroupSymbol     = func(path string) string { return subject(familyGroupSymbols, path) }
	SubjectUser            = func(path string) string { return subject(familyUsers, path) }
	SubjectClient          = func(path string) string { return subject(familyClients, path) }
	SubjectAccount         = func(path string) string { return subject(familyAccounts, path) }
	SubjectGroupCommission = func(path string) string { return subject(familyGroupCommissions, path) }

	// SubjectTrader is everything one trading account hears: its own and nothing else.
	SubjectTrader = wire.SubjectAccountAll
	// SubjectTraderProfile carries a change to one account's own record.
	SubjectTraderProfile = wire.SubjectAccountProfile

	// SubjectJournal is a manager's own journal lines, on the operations topic of their account.
	SubjectJournal = wire.SubjectAccountOperations
)

// server -> core. One subject per command; a queue group picks the pod, and that pod forwards to
// the owner if it is not holding the account itself.
const (
	SubjectSystemOrders    = wire.SubjectSystemOrders
	SubjectSystemPositions = wire.SubjectSystemPositions
	SubjectSystemDealing   = wire.SubjectSystemDealing
	SubjectSystemBalance   = wire.SubjectSystemBalance
	SubjectSystemQuery     = wire.SubjectSystemQuery
)

// core -> one trading account, under the tree the socket already holds
var (
	SubjectAccountOrders     = wire.SubjectAccountOrders
	SubjectAccountPositions  = wire.SubjectAccountPositions
	SubjectAccountDeals      = wire.SubjectAccountDeals
	SubjectAccountSummary    = wire.SubjectAccountSummary
	SubjectAccountMarginCall = wire.SubjectAccountMarginCall
	SubjectAccountOperations = wire.SubjectAccountOperations

	// core -> one dealer's request queue
	SubjectDealerRequests = wire.SubjectBrokerRequests
)

const (
	// ticks, produced outside both modules
	SubjectSystemMarketFeedAll = wire.SubjectQuoteTickAll

	// config reload, broadcast to every pod
	SubjectSystemRoutingCreated = "system.routing.created"
	SubjectSystemRoutingUpdated = "system.routing.updated"
	SubjectSystemRoutingDeleted = "system.routing.deleted"
)

// Records with no group of their own. Access to them is a right, not a path, so the subject carries no group.
const (
	SubjectSymbol   = "websocket.broker.symbols_update"
	SubjectHoliday  = "websocket.broker.holidays_update"
	SubjectLeverage = "websocket.broker.leverages_update"
	SubjectManager  = "websocket.broker.managers_update"
)

// From the api to the other services.
const (
	SubjectSystemEndOfDay     = "system.core.endofday"
	SubjectSystemEndOfDayTime = "system.core.endofday.time"

	SubjectSystemGroupCreated = "system.groups.created"
	SubjectSystemGroupUpdated = "system.groups.updated"
	SubjectSystemGroupDeleted = "system.groups.deleted"

	SubjectSystemGroupSymbolCreated = "system.group_symbols.created"
	SubjectSystemGroupSymbolUpdated = "system.group_symbols.updated"
	SubjectSystemGroupSymbolDeleted = "system.group_symbols.deleted"

	SubjectSystemUserCreated = "system.accounts.created"
	SubjectSystemUserUpdated = "system.accounts.updated"
	SubjectSystemUserDeleted = "system.accounts.deleted"

	SubjectSystemClientCreated = "system.clients.created"
	SubjectSystemClientUpdated = "system.clients.updated"
	SubjectSystemClientDeleted = "system.clients.deleted"

	SubjectSystemGroupCommissionCreated = "system.commissions.created"
	SubjectSystemGroupCommissionUpdated = "system.commissions.updated"
	SubjectSystemGroupCommissionDeleted = "system.commissions.deleted"

	SubjectSystemSymbolCreated = "system.symbols.created"
	SubjectSystemSymbolUpdated = "system.symbols.updated"
	SubjectSystemSymbolDeleted = "system.symbols.deleted"

	SubjectSystemHolidayCreated   = "system.holidays.created"
	SubjectSystemHolidayUpdated   = "system.holidays.updated"
	SubjectSystemHolidayDeleted   = "system.holidays.deleted"
	SubjectSystemHolidayReordered = "system.holidays.reordered"

	SubjectSystemLeverageCreated   = "system.leverages.created"
	SubjectSystemLeverageUpdated   = "system.leverages.updated"
	SubjectSystemLeverageDeleted   = "system.leverages.deleted"
	SubjectSystemLeverageReordered = "system.leverages.reordered"

	SubjectSystemManagerCreated = "system.managers.created"
	SubjectSystemManagerUpdated = "system.managers.updated"
	SubjectSystemManagerDeleted = "system.managers.deleted"
)

// The event types a record change carries.
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

// subject builds a group scoped subject.
func subject(f family, groupPath string) string {
	token := GroupToken(groupPath)
	if token == "" {
		// a trailing * is the group itself and the subtree under it
		token = "root"
	}

	return wire.SubjectGroupScoped(string(f), token)
}

// maskSubject turns one of a manager's group masks into the subscription that covers it.
func buildSubjectPatterns(f family, mask string) []string {
	prefix := SubjectGroupRoot + "." + string(f) + "."

	segs := segments(mask)

	// a bare * covers every group
	if len(segs) == 0 || (len(segs) == 1 && segs[0] == "*") {
		return []string{prefix + ">"}
	}

	// a trailing * is "and everything below": the group itself, and the subtree under it
	if segs[len(segs)-1] == "*" {
		path := strings.Join(segs[:len(segs)-1], ".")
		return []string{prefix + path, prefix + path + ".>"}
	}

	// anything else names one group and grants only that one.
	return []string{prefix + strings.Join(segs, ".")}
}

// overlapping masks are reduced first, or an event is delivered twice
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

// ReduceMasks drops any mask already covered by a broader one, so no event is delivered twice to the same socket.
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
