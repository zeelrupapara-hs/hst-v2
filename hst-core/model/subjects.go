package model

import "fmt"

// server -> core, one pod owns a shard and subscribes to its three subjects
var (
	SubjectShardOrders    = func(shard uint32) string { return fmt.Sprintf("system.s%d.orders", shard) }
	SubjectShardPositions = func(shard uint32) string { return fmt.Sprintf("system.s%d.positions", shard) }
	SubjectShardDealing   = func(shard uint32) string { return fmt.Sprintf("system.s%d.dealing", shard) }
	SubjectShardBalance   = func(shard uint32) string { return fmt.Sprintf("system.s%d.balance", shard) }
)

// core -> one trading account, under the ws.t.<login>.> tree the socket already holds
var (
	SubjectAccountOrders    = func(login int64) string { return fmt.Sprintf("ws.t.%d.orders", login) }
	SubjectAccountPositions = func(login int64) string { return fmt.Sprintf("ws.t.%d.positions", login) }
	SubjectAccountDeals     = func(login int64) string { return fmt.Sprintf("ws.t.%d.deals", login) }
	SubjectAccountSummary   = func(login int64) string { return fmt.Sprintf("ws.t.%d.summary", login) }
	SubjectAccountResult    = func(login int64) string { return fmt.Sprintf("ws.t.%d.result", login) }
)

// core -> one dealer's request queue
var SubjectDealerRequests = func(login int64) string { return fmt.Sprintf("ws.m.%d.requests", login) }

const (
	SubjectSystemMarketFeedAll = "hstquote.tick.*"

	SubjectSystemGroups       = "system.group.*"
	SubjectSystemGroupSymbols = "system.group_symbol.*"
	SubjectSystemSymbols      = "system.symbol.*"
	SubjectSystemCommissions  = "system.group_commission.*"
	SubjectSystemRules        = "system.routing.*"
	SubjectSystemAccounts     = "system.user.*"
	SubjectSystemEndOfDay     = "system.core.endofday"
)

func SubjectSystemMarketFeed(symbol string) string { return "hstquote.tick." + symbol }
