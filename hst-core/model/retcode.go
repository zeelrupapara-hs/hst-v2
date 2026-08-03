package model

// status code of return response
type RetCode int32

const (
	RetOK RetCode = 0

	RetError        RetCode = 1
	RetInvalidData  RetCode = 3
	RetNotFound     RetCode = 5
	RetPermissions  RetCode = 8
	RetAuthDisabled RetCode = 1002

	// trade refusals
	RetTradeDisabled        RetCode = 10001
	RetTradeMarketClosed    RetCode = 10002
	RetTradeNoMoney         RetCode = 10003
	RetTradePrices          RetCode = 10004
	RetTradeVolumeLimit     RetCode = 10005
	RetTradeInvalidStops    RetCode = 10006
	RetTradeTooManyOrder    RetCode = 10007
	RetTradeInvalidVolume   RetCode = 10008
	RetTradeOrderExist      RetCode = 10009
	RetTradeBadSymbol       RetCode = 10010
	RetTradeNoQuotes        RetCode = 10011
	RetTradeRejected        RetCode = 10012
	RetTradeRequote         RetCode = 10013
	RetTradeTimeout         RetCode = 10014
	RetTradeHedgeProhibited RetCode = 10015
	RetTradeCloseOnly       RetCode = 10016
	RetTradeFillPolicy      RetCode = 10017
	RetTradeExpiration      RetCode = 10018
	RetTradeFrozen          RetCode = 10019
	RetTradeNotProcessed    RetCode = 10020
	RetTradeMaxVolume       RetCode = 10021
	RetTradeAccountNotFound RetCode = 10022
	RetTradeDealerQueued    RetCode = 10023
	RetTradeDealerReturned  RetCode = 10024
	RetTradeCloseOrderExist RetCode = 10025
)

// retNames is what each refusal means, for the journal and the log.
var retNames = map[RetCode]string{
	RetOK:                   "done",
	RetError:                "internal error",
	RetInvalidData:          "invalid request",
	RetNotFound:             "not found",
	RetPermissions:          "not permitted",
	RetAuthDisabled:         "account disabled",
	RetTradeDisabled:        "trading is disabled",
	RetTradeMarketClosed:    "market is closed",
	RetTradeNoMoney:         "not enough money",
	RetTradePrices:          "price is off the market",
	RetTradeVolumeLimit:     "volume limit reached",
	RetTradeInvalidStops:    "stops are too close",
	RetTradeTooManyOrder:    "too many orders",
	RetTradeInvalidVolume:   "invalid volume",
	RetTradeOrderExist:      "order already exists",
	RetTradeBadSymbol:       "unknown symbol",
	RetTradeNoQuotes:        "no price for this symbol",
	RetTradeRejected:        "rejected by a routing rule",
	RetTradeRequote:         "requote",
	RetTradeTimeout:         "timed out",
	RetTradeHedgeProhibited: "hedging is not allowed",
	RetTradeCloseOnly:       "closing only",
	RetTradeFillPolicy:      "fill policy not allowed",
	RetTradeExpiration:      "expiry not allowed",
	RetTradeFrozen:          "order is frozen",
	RetTradeNotProcessed:    "no routing rule admitted this request",
	RetTradeMaxVolume:       "position volume limit reached",
	RetTradeAccountNotFound: "account not found",
	RetTradeDealerQueued:    "request placed in a dealer queue",
	RetTradeDealerReturned:  "request rejected, due all assigned dealers returned request in queue",
	RetTradeCloseOrderExist: "an older position on this symbol must be closed first",
}

func (r RetCode) String() string {
	if s, ok := retNames[r]; ok {
		return s
	}
	return "unknown"
}

func (r RetCode) OK() bool { return r == RetOK }
