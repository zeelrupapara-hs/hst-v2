package model

import wire "hstmodel"

// Every string comes from the shared contract, so the api and the engine cannot drift into two
// names for one wire string.

// server -> core. Every pod listens on one subject per command under a queue group, so exactly one
// receives it; that pod forwards to the owner when it is not holding the account itself.
const (
	SubjectSystemOrders    = wire.SubjectSystemOrders
	SubjectSystemPositions = wire.SubjectSystemPositions
	SubjectSystemDealing   = wire.SubjectSystemDealing
	SubjectSystemBalance   = wire.SubjectSystemBalance
	SubjectSystemQuery     = wire.SubjectSystemQuery

	QueueOrders    = wire.QueueOrders
	QueuePositions = wire.QueuePositions
	QueueDealing   = wire.QueueDealing
	QueueBalance   = wire.QueueBalance
	QueueQuery     = wire.QueueQuery
)

// SubjectOwner is the private inbox of the pod holding a shard, the one place a shard is named.
var SubjectOwner = wire.SubjectOwner

// core -> one trading account, under the tree the socket already holds
var (
	SubjectAccountOrders      = wire.SubjectAccountOrders
	SubjectAccountPositions   = wire.SubjectAccountPositions
	SubjectAccountDeals       = wire.SubjectAccountDeals
	SubjectAccountSummary     = wire.SubjectAccountSummary
	SubjectAccountMarginCall  = wire.SubjectAccountMarginCall
	SubjectAccountMoneyChange = wire.SubjectAccountMoneyChange
)

// core -> one dealer's request queue
var SubjectDealerRequests = wire.SubjectBrokerRequests

const (
	SubjectSystemMarketFeedAll = wire.SubjectQuoteTickAll
	// SubjectSystemSnapshot asks the feed for the last price of every instrument.
	SubjectSystemSnapshot = wire.SubjectQuoteSnapshot

	SubjectSystemGroups       = wire.SubjectSystemGroups
	SubjectSystemGroupSymbols = wire.SubjectSystemGroupSymbols
	SubjectSystemSymbols      = wire.SubjectSystemSymbols
	SubjectSystemCommissions  = wire.SubjectSystemCommissions
	SubjectSystemRules        = wire.SubjectSystemRouting
	SubjectSystemAccounts     = wire.SubjectSystemAccounts
	SubjectSystemEndOfDay     = wire.SubjectSystemEndOfDay
	SubjectSystemEndOfDayTime = wire.SubjectSystemEndOfDayTime
)
