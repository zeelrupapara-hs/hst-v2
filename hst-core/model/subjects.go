package model

import (
	"fmt"
	"strings"
)

// Every string comes from the shared contract, so the api and the engine cannot drift into two
// names for one wire string.

// server -> core. Every pod listens on one subject per command under a queue group, so exactly one
// receives it; that pod forwards to the owner when it is not holding the account itself.

// core -> one trading account, under the tree the socket already holds

// core -> one dealer's request queue
var SubjectDealerRequests = SubjectBrokerRequests

const (
	SubjectSystemMarketFeedAll = SubjectQuoteTickAll
	// SubjectSystemSnapshot asks the feed for the last price of every instrument.
	SubjectSystemSnapshot = SubjectQuoteSnapshot

	SubjectSystemRules = SubjectSystemRouting
)

// Whether an event reaches a client is decided by its subject and nothing else.
//
// websocket.accounts.<login>.<topic>   the account holding that login is subscribed to the
//                                      wildcard below it, so any topic named here is delivered
// websocket.clients.<session>.<topic>  one browser tab rather than one account
// websocket.broker.<topic>             staff, and the dealing desk
// system.<noun>                        service to service, never delivered to anyone
//
// Adding a client-visible event means naming a topic under websocket. and publishing to it. No
// registration, no subscription list to edit.

const (
	RootWebsocket = "websocket"
	RootSystem    = "system"
	RootQuote     = "hstquote"
)

// What a trading account is subscribed to.
var (
	SubjectAccountAll         = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.>", login) }
	SubjectAccountOrders      = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.orders", login) }
	SubjectAccountPositions   = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.positions", login) }
	SubjectAccountDeals       = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.deals", login) }
	SubjectAccountSummary     = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.summary", login) }
	SubjectAccountMoneyChange = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.money_change", login) }
	SubjectAccountMarginCall  = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.margin_call", login) }
	SubjectAccountOperations  = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.operations", login) }
	SubjectAccountSymbols     = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.symbols", login) }
	SubjectAccountProfile     = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.profile", login) }
	SubjectAccountAlerts      = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.alerts", login) }
	SubjectAccountMarketFeed  = func(login int64) string { return fmt.Sprintf("websocket.accounts.%d.market_feed", login) }
)

// What one connection is subscribed to, as opposed to one account.
var (
	SubjectClientAll    = func(session string) string { return fmt.Sprintf("websocket.clients.%s.>", session) }
	SubjectClientLogout = func(session string) string { return fmt.Sprintf("websocket.clients.%s.session_logout", session) }
)

// What a manager is subscribed to, scoped to the group tree their masks cover.
//
// The tree is what makes this different from vfx, which sends every manager everything: a group
// path becomes subject tokens, so demo\\forex is websocket.groups.users.demo.forex and a mask
// subscribes to the subtree under it.
var (
	SubjectGroupScoped = func(family, token string) string {
		return fmt.Sprintf("%s.%s.%s", RootGroupScoped, family, token)
	}
)

// RootGroupScoped prefixes every group scoped subject.
const RootGroupScoped = "websocket.groups"

// groupToken turns a group path into subject tokens: demo\forex becomes demo.forex.
func groupToken(group string) string {
	parts := strings.Split(group, "\\")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	token := strings.Join(out, ".")
	if token == "" {
		token = "root"
	}
	return token
}

// SubjectGroupAccounts is the manager-scoped live account stream: only sockets whose group
// masks cover this path are subscribed, so scope is enforced by the subject itself.
func SubjectGroupAccounts(group string) string {
	return SubjectGroupScoped("accounts", groupToken(group))
}

// The manager-scoped trade streams, so a blotter hears a trade the moment it happens
// without polling. Same scoping rule as the account stream.
func SubjectGroupPositions(group string) string {
	return SubjectGroupScoped("positions", groupToken(group))
}

func SubjectGroupOrders(group string) string {
	return SubjectGroupScoped("orders", groupToken(group))
}

func SubjectGroupDeals(group string) string {
	return SubjectGroupScoped("deals", groupToken(group))
}

// What staff and the dealing desk are subscribed to.
var (
	SubjectBrokerAll        = "websocket.broker.>"
	SubjectBrokerOperations = "websocket.broker.operations"
	SubjectBrokerRequests   = func(dealer int64) string { return fmt.Sprintf("websocket.broker.%d.orders.approval", dealer) }
	SubjectBrokerUpdate     = func(topic string) string { return fmt.Sprintf("websocket.broker.%s_update", topic) }
)

// Service to service. Nothing here is ever delivered to a client.
const (
	SubjectSystemOrders    = "system.orders"
	SubjectSystemPositions = "system.positions"
	SubjectSystemDealing   = "system.dealing"
	SubjectSystemBalance   = "system.balance"
	SubjectSystemQuery     = "system.query"

	SubjectSystemAccounts     = "system.accounts.>"
	SubjectSystemGroups       = "system.groups.>"
	SubjectSystemGroupSymbols = "system.group_symbols.>"
	SubjectSystemSymbols      = "system.symbols.>"
	SubjectSystemCommissions  = "system.commissions.>"
	SubjectSystemRouting      = "system.routing.>"
	SubjectSystemHolidays     = "system.holidays.>"
	SubjectSystemLeverages    = "system.leverages.>"
	SubjectSystemManagers     = "system.managers.>"
	SubjectSystemClients      = "system.clients.>"

	SubjectSystemEndOfDay     = "system.core.endofday"
	SubjectSystemEndOfDayTime = "system.core.endofday.time"

	SubjectQuoteTickAll  = "hstquote.tick.*"
	SubjectQuoteSnapshot = "hstquote.snapshot"
)

// The queue group that makes one command reach exactly one engine pod.
const (
	QueueOrders    = "engine_orders"
	QueuePositions = "engine_positions"
	QueueDealing   = "engine_dealing"
	QueueBalance   = "engine_balance"
	QueueQuery     = "engine_query"
)

// SubjectOwner is the private inbox of the pod holding a shard.
//
// Only the engine uses it, and only to hand a command to whichever pod owns the account. It is
// the one subject that mentions a shard, and no client can reach it.
func SubjectOwner(shard uint32, topic string) string {
	return fmt.Sprintf("system.owner.%d.%s", shard, topic)
}
