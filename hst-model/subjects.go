package model

import "fmt"

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

// SubjectQuoteTick is one instrument's price stream.
func SubjectQuoteTick(symbol string) string { return fmt.Sprintf("hstquote.tick.%s", symbol) }

// SubjectSystemChange is a reference record changing, as system.<family>.<verb>.
func SubjectSystemChange(family, verb string) string {
	return fmt.Sprintf("system.%s.%s", family, verb)
}

// SubjectOwner is the private inbox of the pod holding a shard.
//
// Only the engine uses it, and only to hand a command to whichever pod owns the account. It is
// the one subject that mentions a shard, and no client can reach it.
func SubjectOwner(shard uint32, topic string) string {
	return fmt.Sprintf("system.owner.%d.%s", shard, topic)
}
