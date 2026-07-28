# MT5 Enumeration Catalog

Complete numeric enumeration reference for the MetaTrader 5 data model, to accompany the ER
diagram in `docs/tables/`.

**Sources (all values verified, none invented):**

- `Meta Document - Admin/Enumerations - mt5_{orders,deals,positions,symbols,groups,users,routing,leverages} - SQL Export - Backup Server.pdf`
- `Meta Document - Admin/Enumerations - IMT{Order,Deal,Position,Request,User,Client,Document,ConGroup,ConSymbol,ConManager,ConCommission,ConCommTier,ConRoute,ConCondition,ConLeverageRule}*.pdf`
- `Meta Document - Admin/{Trade Requests,Dealer,Trade management,Subscriptions,Successful completion} - Return Codes*.pdf`
- `Meta Document - SQL/sql_mt5_{routing,commissions,commissions_tiers}.htm`
- `docs/mt5/01-symbols.md` §3, `docs/mt5/02-groups.md` §3, `docs/mt5/04-orders-deals-positions.md`

**Conventions:** bitmask enums are marked "(bitmask — combine with OR)" and given in hex.
Plain enums are given in decimal. `*_FIRST` / `*_LAST` / `*_END` sentinel members are omitted
except where they carry a distinct numeric limit.

---

# Orders, Deals, Positions

## EnOrderType

Type of a trade order. Stored in `mt5_orders.Type`.

| Value | Name | Meaning |
|---|---|---|
| 0 | OP_BUY | Buy order |
| 1 | OP_SELL | Sell order |
| 2 | OP_BUY_LIMIT | Buy Limit order |
| 3 | OP_SELL_LIMIT | Sell Limit order |
| 4 | OP_BUY_STOP | Buy Stop order |
| 5 | OP_SELL_STOP | Sell Stop order |
| 6 | OP_BUY_STOP_LIMIT | Buy Stop Limit order |
| 7 | OP_SELL_STOP_LIMIT | Sell Stop Limit order |
| 8 | OP_CLOSE_BY | Close-by — closes two oppositely directed positions on one symbol. Hedging accounting only (MARGIN_MODE_RETAIL_HEDGED) |

## EnOrderFilling

Order filling policy. Stored in `mt5_orders.TypeFill`.

| Value | Name | Meaning |
|---|---|---|
| 0 | ORDER_FILL_FOK | Fill or Kill — must fill completely or be canceled. Auto-set for instant and request execution |
| 1 | ORDER_FILL_IOC | Immediate or Cancel — partial fill allowed, remainder canceled. Stock/market execution only |
| 2 | ORDER_FILL_RETURN | Return remainder to the queue. Pending orders only |

## EnOrderTime

Order expiration policy. Stored in `mt5_orders.TypeTime`.

| Value | Name | Meaning |
|---|---|---|
| 0 | ORDER_TIME_GTC | Good till canceled |
| 1 | ORDER_TIME_DAY | Intraday — expires at end of trading day |
| 2 | ORDER_TIME_SPECIFIED | Expires at a specified date/time |
| 3 | ORDER_TIME_SPECIFIED_DAY | Expires at 00:00 of the specified day, or the nearest trading time |

## EnOrderState

Lifecycle state of an order. Stored in `mt5_orders.State`.

| Value | Name | Meaning |
|---|---|---|
| 0 | ORDER_STATE_STARTED | Started |
| 1 | ORDER_STATE_PLACED | Placed |
| 2 | ORDER_STATE_CANCELED | Canceled |
| 3 | ORDER_STATE_PARTIAL | Partially filled |
| 4 | ORDER_STATE_FILLED | Filled |
| 5 | ORDER_STATE_REJECTED | Rejected |
| 6 | ORDER_STATE_EXPIRED | Expired |
| 7 | ORDER_STATE_REQUEST_ADD | Passed to the gateway to be placed; the add request is being processed |
| 8 | ORDER_STATE_REQUEST_MODIFY | Passed to the gateway to be modified; the modify request is being processed |
| 9 | ORDER_STATE_REQUEST_CANCEL | Passed to the gateway to be deleted; the delete request is being processed |

## EnOrderActivation

How/why an order was activated. Stored in `mt5_orders.ActivationMode`.

| Value | Name | Meaning |
|---|---|---|
| 0 | ACTIVATION_NONE | Not activated |
| 1 | ACTIVATION_PENDING | Activation of a pending order |
| 2 | ACTIVATION_STOPLIMIT | Activation of a Stop Limit order |
| 3 | ACTIVATION_EXPIRATION | Order canceled upon expiration |
| 4 | ACTIVATION_STOPOUT | Order removed because of a stop out |

## EnOrderReason

Origin of the order. Stored in `mt5_orders.Reason`.

| Value | Name | Meaning |
|---|---|---|
| 0 | ORDER_REASON_CLIENT | Placed manually by a client from the client terminal |
| 1 | ORDER_REASON_EXPERT | Placed by a client via an Expert Advisor |
| 2 | ORDER_REASON_DEALER | Placed by a dealer through the manager terminal |
| 3 | ORDER_REASON_SL | Placed as a result of Stop Loss activation |
| 4 | ORDER_REASON_TP | Placed as a result of Take Profit activation |
| 5 | ORDER_REASON_SO | Placed when the client reached the Stop-Out level |
| 6 | ORDER_REASON_ROLLOVER | Placed when reopening a position for charging swaps |
| 7 | ORDER_REASON_EXTERNAL_CLIENT | Placed by a client from an external trading system |
| 8 | ORDER_REASON_VMARGIN | Placed for accruing variation margin |
| 9 | ORDER_REASON_GATEWAY | Placed by an MT5 gateway connected to the platform |
| 10 | ORDER_REASON_SIGNAL | Placed by copying a trade signal per a terminal subscription |
| 11 | ORDER_REASON_SETTLEMENT | Settlement of a futures contract/option. Not used at present |
| 12 | ORDER_REASON_TRANSFER | Position transfer at settlement price to a new symbol with the same underlying. Not used at present |
| 13 | ORDER_REASON_SYNC | Synchronization of an account's trade state with an external system |
| 14 | ORDER_REASON_EXTERNAL_SERVICE | Placed from an external trading system for technical reasons |
| 15 | ORDER_REASON_MIGRATION | Created by importing trade operations from a MetaTrader 4 server |
| 16 | ORDER_REASON_MOBILE | Created via the MT5 mobile terminal (Android/iPhone) |
| 17 | ORDER_REASON_WEB | Created via the web terminal |
| 18 | ORDER_REASON_SPLIT | Created as a result of a symbol split |
| 19 | ORDER_REASON_CORPORATE_ACTION | Created by a corporate action (consolidation, renaming, account transfer). Excluded from commission calculations |

## EnTradeActivationFlags

*(bitmask — combine with OR)* Suppression flags for automatic activation handling. Set on
orders; **inherited by the positions created from them**. Same enum used by `mt5_orders` and
`mt5_positions`.

| Value | Name | Meaning |
|---|---|---|
| 0x00 | ACTIV_FLAGS_NONE | No flags |
| 0x01 | ACTIV_FLAGS_NO_LIMIT | Do not handle reaching of the Limit level |
| 0x02 | ACTIV_FLAGS_NO_STOP | Do not handle reaching of the Stop level |
| 0x04 | ACTIV_FLAGS_NO_SLIMIT | Do not handle reaching of the Stop-Limit level |
| 0x08 | ACTIV_FLAGS_NO_SL | Do not handle activation upon Stop Loss |
| 0x10 | ACTIV_FLAGS_NO_TP | Do not handle activation upon Take Profit |
| 0x20 | ACTIV_FLAGS_NO_SO | Do not handle activation upon Stop-Out |
| 0x40 | ACTIV_FLAGS_NO_EXPIRATION | Do not handle order cancellation upon expiration |

## EnTradeModifyFlags

*(bitmask — combine with OR)* Records who modified an order.

| Value | Name | Meaning |
|---|---|---|
| 0x00000000 | (none) | No flags |
| 0x00000001 | MODIFY_FLAGS_ADMIN | Order changed by an administrator |
| 0x00000002 | MODIFY_FLAGS_MANAGER | Open price modified by a manager |
| 0x00000004 | MODIFY_FLAGS_POSITION | Not used for orders |
| 0x00000008 | MODIFY_FLAGS_RESTORE | Order restored |
| 0x00000010 | MODIFY_FLAGS_API_ADMIN | Changed via Manager API administrator interface |
| 0x00000020 | MODIFY_FLAGS_API_MANAGER | Changed via Manager API manager interface |
| 0x00000040 | MODIFY_FLAGS_API_SERVER | Changed via Server API |
| 0x00000080 | MODIFY_FLAGS_API_GATEWAY | Changed via Gateway API |

## EnDealAction

Type of operation a deal represents. Stored in `mt5_deals.Action`.

| Value | Name | Meaning |
|---|---|---|
| 0 | DEAL_BUY | Buy deal |
| 1 | DEAL_SELL | Sell deal |
| 2 | DEAL_BALANCE | Balance operation |
| 3 | DEAL_CREDIT | Credit operation |
| 4 | DEAL_CHARGE | Additional charges/withdrawals |
| 5 | DEAL_CORRECTION | Correcting operation |
| 6 | DEAL_BONUS | Bonus |
| 7 | DEAL_COMMISSION | Commission |
| 8 | DEAL_COMMISSION_DAILY | Daily commission |
| 9 | DEAL_COMMISSION_MONTHLY | Monthly commission |
| 10 | DEAL_AGENT_DAILY | Daily agent commission |
| 11 | DEAL_AGENT_MONTHLY | Monthly agent commission |
| 12 | DEAL_INTERESTRATE | Accrual of annual interest |
| 13 | DEAL_BUY_CANCELED | Canceled Buy deal (via `IMTExecution::TE_DEAL_CANCEL`). P/L cleared; excluded from account state and position recalculation |
| 14 | DEAL_SELL_CANCELED | Canceled Sell deal — same semantics as DEAL_BUY_CANCELED |
| 15 | DEAL_DIVIDEND | Dividend operation |
| 16 | DEAL_DIVIDEND_FRANKED | Franked (non-taxable) dividend — tax paid by the company, not the client |
| 17 | DEAL_TAX | Charging a tax |
| 18 | DEAL_AGENT | Instant agent commission charge (each time the agent's client trades) |
| 19 | DEAL_SO_COMPENSATION | Compensation of a negative account after a Stop Out event |
| 20 | DEAL_SO_COMPENSATION_CREDIT | Withdrawing credit funds after a negative-balance compensation |

## EnDealEntry (a.k.a. EnEntryFlags)

Effect of a deal on the position. Stored in `mt5_deals.Entry`. The SQL-export docs name this
`EnEntryFlags`; the trade API names it `EnDealEntry`. Despite "Flags", it is a plain enum.

| Value | Name | Meaning |
|---|---|---|
| 0 | ENTRY_IN | Entering the market or adding volume |
| 1 | ENTRY_OUT | Exiting the market or partial closure |
| 2 | ENTRY_INOUT | Reversal |
| 3 | ENTRY_OUT_BY | Close-by — simultaneous closure of two opposite positions on the same instrument. Hedging mode only |

## EnDealReason

Origin of a deal. Stored in `mt5_deals.Reason`. Values align 1:1 with EnOrderReason except
for the meaning shifts noted at 11 and 12.

| Value | Name | Meaning |
|---|---|---|
| 0 | DEAL_REASON_CLIENT | Performed manually by a client from the client terminal |
| 1 | DEAL_REASON_EXPERT | Performed by a client via an Expert Advisor |
| 2 | DEAL_REASON_DEALER | Performed by a dealer through the manager terminal |
| 3 | DEAL_REASON_SL | Result of Stop Loss activation |
| 4 | DEAL_REASON_TP | Result of Take Profit activation |
| 5 | DEAL_REASON_SO | Client reached the Stop-Out level |
| 6 | DEAL_REASON_ROLLOVER | Reopening a position for charging swaps |
| 7 | DEAL_REASON_EXTERNAL_CLIENT | Performed by a client from an external trading system. **Commission is charged** (unlike 14) |
| 8 | DEAL_REASON_VMARGIN | Accruing variation margin |
| 9 | DEAL_REASON_GATEWAY | Performed by an MT5 gateway connected to the platform |
| 10 | DEAL_REASON_SIGNAL | Copying a trade signal per a terminal subscription |
| 11 | DEAL_REASON_SETTLEMENT | Compulsory position close due to futures/option settlement |
| 12 | DEAL_REASON_TRANSFER | Position transfer at settlement price to a new symbol with the same underlying |
| 13 | DEAL_REASON_SYNC | Synchronization of an account's trade state with an external system |
| 14 | DEAL_REASON_EXTERNAL_SERVICE | Performed from an external trading system for technical reasons. **Commission is not charged** |
| 15 | DEAL_REASON_MIGRATION | Created by importing trade operations from a MetaTrader 4 server |
| 16 | DEAL_REASON_MOBILE | Conducted via the MT5 mobile terminal (Android/iPhone) |
| 17 | DEAL_REASON_WEB | Conducted via the web terminal |
| 18 | DEAL_REASON_SPLIT | Result of a symbol split |
| 19 | DEAL_REASON_CORPORATE_ACTION | Result of a corporate action. Excluded from commission calculations |

## EnPositionAction

Direction of a position. Stored in `mt5_positions.Action`.

| Value | Name | Meaning |
|---|---|---|
| 0 | POSITION_BUY | Buy (long) |
| 1 | POSITION_SELL | Sell (short) |

## EnActivation

Position activation state. Stored in `mt5_positions.ActivationMode`. **Note the numbering
differs from `EnOrderActivation`** — do not share a lookup table between them.

| Value | Name | Meaning |
|---|---|---|
| 0 | ACTIVATION_NONE | None |
| 1 | ACTIVATION_SL | Stop Loss |
| 2 | ACTIVATION_TP | Take Profit |
| 3 | ACTIVATION_STOPOUT | Stop Out |

## EnPositionReason

Reason a position was opened. Stored in `mt5_positions.Reason`. Value space matches
EnOrderReason; several values are never produced for positions.

| Value | Name | Meaning |
|---|---|---|
| 0 | POSITION_REASON_CLIENT | Opened manually by a client from the client terminal |
| 1 | POSITION_REASON_EXPERT | Opened by a client using an Expert Advisor |
| 2 | POSITION_REASON_DEALER | Opened by a dealer through the manager terminal |
| 3 | POSITION_REASON_SL | Not used for positions |
| 4 | POSITION_REASON_TP | Not used for positions |
| 5 | POSITION_REASON_SO | Not used for positions |
| 6 | POSITION_REASON_ROLLOVER | Reopened to charge swaps |
| 7 | POSITION_REASON_EXTERNAL_CLIENT | Opened from an external trading system |
| 8 | POSITION_REASON_VMARGIN | Not used for positions |
| 9 | POSITION_REASON_GATEWAY | Opened by an MT5 gateway connected to the platform |
| 10 | POSITION_REASON_SIGNAL | Opened by copying a trading signal per a terminal subscription |
| 11 | POSITION_REASON_SETTLEMENT | Futures/option delivery date operations. Not currently used |
| 12 | POSITION_REASON_TRANSFER | Transfer of a position at a calculated price to a new symbol with the same underlying |
| 13 | POSITION_REASON_SYNC | Opened while synchronizing account state with an external system |
| 14 | POSITION_REASON_EXTERNAL_SERVICE | Opened in the external trading system for service purposes |
| 15 | POSITION_REASON_MIGRATION | Result of importing client trading operations from a MetaTrader 4 server |
| 16 | POSITION_REASON_MOBILE | Opened via the MT5 mobile terminal (Android/iPhone) |
| 17 | POSITION_REASON_WEB | Opened via the web terminal |
| 18 | POSITION_REASON_SPLIT | Result of a symbol split |
| 19 | POSITION_REASON_CORPORATE_ACTION | Result of a corporate action. Excluded from commission calculations |

---

# Trade Requests

## EnTradeActions

Type of trade request (`IMTRequest::Action`). Values are banded: client 0–99, server 100–199,
dealer 200–255.

| Value | Name | Meaning |
|---|---|---|
| 0 | TA_PRICE / TA_CLIENT_FIRST | Price request (request execution). Also the start of the client band |
| 1 | TA_REQUEST | Order execution confirmation at dealer level (request execution) |
| 2 | TA_INSTANT | Order placement in instant execution mode |
| 3 | TA_MARKET | Order placement in market execution mode |
| 4 | TA_EXCHANGE | Order placement in exchange execution mode |
| 5 | TA_PENDING | Placing a pending order |
| 6 | TA_SLTP | Modification of position Stop Loss / Take Profit |
| 7 | TA_MODIFY | Modification of a pending order |
| 8 | TA_REMOVE | Deleting a pending order |
| 9 | TA_TRANSFER | Transfer operation |
| 10 | TA_CLOSE_BY | Close-by, performed by a client |
| 99 | TA_CLIENT_LAST | End of the client request band |
| 100 | TA_SERVER_FIRST / TA_ACTIVATE | Start of the server band; activation of a pending order |
| 101 | TA_ACTIVATE_SL | Stop Loss activation |
| 102 | TA_ACTIVATE_TP | Take Profit activation |
| 103 | TA_ACTIVATE_STOPLIMIT | Stop Limit activation |
| 104 | TA_STOPOUT_ORDER | Pending order deletion on stop-out |
| 105 | TA_STOPOUT_POSITION | Position closure on stop-out |
| 106 | TA_EXPIRATION | Order cancellation upon expiration |
| 199 | TA_SERVER_LAST | End of the server band |
| 200 | TA_DEALER_FIRST / TA_DEALER_POS_EXECUTE | Start of the dealer band; position opening/closing by a dealer |
| 201 | TA_DEALER_ORD_PENDING | Pending order placement by a dealer |
| 202 | TA_DEALER_POS_MODIFY | Position modification by a dealer |
| 203 | TA_DEALER_ORD_MODIFY | Order modification by a dealer |
| 204 | TA_DEALER_ORD_REMOVE | Order deletion by a dealer |
| 205 | TA_DEALER_ORD_ACTIVATE | Order activation by a dealer |
| 206 | TA_DEALER_BALANCE | Balance operation by a dealer |
| 207 | TA_DEALER_ORD_SLIMIT | Stop Limit activation by a dealer |
| 208 | TA_DEALER_CLOSE_BY | Close-by performed by a dealer |
| 255 | TA_DEALER_LAST / TA_END | End of the dealer band and of the enumeration |

## EnTradeActionFlags (TA_FLAG_*)

*(bitmask — combine with OR)* Attributes of a trade request (`IMTRequest::Flags`).

| Value | Name | Meaning |
|---|---|---|
| 0x00000000 | TA_FLAG_NONE | No flags |
| 0x00000001 | TA_FLAG_CLOSE | Client requested to close a position |
| 0x00000002 | TA_FLAG_MARKET | Request to conduct a deal at the market price |
| 0x00000004 | TA_FLAG_CHANGED_PRICE | Request modifies position open price or order price |
| 0x00000008 | TA_FLAG_CHANGED_TRIGGER | Request modifies the trigger price of a Stop Limit order |
| 0x00000010 | TA_FLAG_CHANGED_SL | Request modifies the Stop Loss level |
| 0x00000020 | TA_FLAG_CHANGED_TP | Request modifies the Take Profit level |
| 0x00000040 | TA_FLAG_CHANGED_EXP_TYPE | Request modifies order expiration type |
| 0x00000080 | TA_FLAG_CHANGED_EXP_TIME | Request modifies order expiration time |
| 0x00000100 | TA_FLAG_EXPERT | Request sent by an MQL5 program |
| 0x00000200 | TA_FLAG_SIGNAL | Operation copied by the terminal per a trading-signal subscription |
| 0x00000400 | TA_FLAG_SKIP_MARGIN_CHECK | Skip free-margin checks during balance operations and execution |

## Return codes — success

| Value | Name | Meaning |
|---|---|---|
| 0 | MT_RET_OK | Successful completion |
| 1 | MT_RET_OK_NONE | Successful completion with no information returned |

## Return codes — trade requests (10001–10046)

| Value | Name | Meaning |
|---|---|---|
| 10001 | MT_RET_REQUEST_INWAY | Request is on the way |
| 10002 | MT_RET_REQUEST_ACCEPTED | Request accepted |
| 10003 | MT_RET_REQUEST_PROCESS | Request processed |
| 10004 | MT_RET_REQUEST_REQUOTE | Requote in response to the request |
| 10005 | MT_RET_REQUEST_PRICES | Prices in response to the request |
| 10006 | MT_RET_REQUEST_REJECT | Request rejected |
| 10007 | MT_RET_REQUEST_CANCEL | Request canceled |
| 10008 | MT_RET_REQUEST_PLACED | An order was placed as a result of the request |
| 10009 | MT_RET_REQUEST_DONE | Request fulfilled |
| 10010 | MT_RET_REQUEST_DONE_PARTIAL | Request partially fulfilled |
| 10011 | MT_RET_REQUEST_ERROR | Common request error |
| 10012 | MT_RET_REQUEST_TIMEOUT | Request timed out |
| 10013 | MT_RET_REQUEST_INVALID | Invalid request |
| 10014 | MT_RET_REQUEST_INVALID_VOLUME | Invalid volume |
| 10015 | MT_RET_REQUEST_INVALID_PRICE | Invalid price |
| 10016 | MT_RET_REQUEST_INVALID_STOPS | Wrong stop levels or price |
| 10017 | MT_RET_REQUEST_TRADE_DISABLED | Trade is disabled |
| 10018 | MT_RET_REQUEST_MARKET_CLOSED | Market is closed |
| 10019 | MT_RET_REQUEST_NO_MONEY | Not enough money |
| 10020 | MT_RET_REQUEST_PRICE_CHANGED | Price has changed |
| 10021 | MT_RET_REQUEST_PRICE_OFF | No price |
| 10022 | MT_RET_REQUEST_INVALID_EXP | Invalid order expiration |
| 10023 | MT_RET_REQUEST_ORDER_CHANGED | Order has been changed |
| 10024 | MT_RET_REQUEST_TOO_MANY | Too many trade requests (e.g. rate limit on one Manager API instance) |
| 10025 | MT_RET_REQUEST_NO_CHANGES | Request does not contain changes |
| 10026 | MT_RET_REQUEST_AT_DISABLED_SERVER | Autotrading disabled on the server |
| 10027 | MT_RET_REQUEST_AT_DISABLED_CLIENT | Autotrading disabled on the client side |
| 10028 | MT_RET_REQUEST_LOCKED | Request blocked by the dealer |
| 10029 | MT_RET_REQUEST_FROZEN | Modification failed — order or position is frozen |
| 10030 | MT_RET_REQUEST_INVALID_FILL | Fill mode is not supported |
| 10031 | MT_RET_REQUEST_CONNECTION | No connection |
| 10032 | MT_RET_REQUEST_ONLY_REAL | Allowed only for real accounts |
| 10033 | MT_RET_REQUEST_LIMIT_ORDERS | Reached the limit on the number of orders |
| 10034 | MT_RET_REQUEST_LIMIT_VOLUME | Reached the volume limit |
| 10035 | MT_RET_REQUEST_INVALID_ORDER | Invalid or prohibited order type |
| 10036 | MT_RET_REQUEST_POSITION_CLOSED | Position is already closed |
| 10037 | MT_RET_REQUEST_EXECUTION_SKIPPED | Used for internal purposes |
| 10038 | MT_RET_REQUEST_INVALID_CLOSE_VOLUME | Volume to close exceeds the current position volume |
| 10039 | MT_RET_REQUEST_CLOSE_ORDER_EXIST | An order to close the position already exists (hedging mode) |
| 10040 | MT_RET_REQUEST_LIMIT_POSITIONS | Group's open-position limit reached |
| 10041 | MT_RET_REQUEST_REJECT_CANCEL | Request rejected, order canceled — returned when routing action ACTION_CANCEL_ORDER applies |
| 10042 | MT_RET_REQUEST_LONG_ONLY | Rejected — symbol is TRADE_LONGONLY |
| 10043 | MT_RET_REQUEST_SHORT_ONLY | Rejected — symbol is TRADE_SHORTONLY |
| 10044 | MT_RET_REQUEST_CLOSE_ONLY | Rejected — symbol is TRADE_CLOSEONLY |
| 10045 | MT_RET_REQUEST_PROHIBITED_BY_FIFO | Position closure not allowed by the FIFO rule (TRADEFLAGS_FIFO_CLOSE) |
| 10046 | MT_RET_REQUEST_HEDGE_PROHIBITED | Hedge positions prohibited for the group (TRADEFLAGS_HEDGE_PROHIBIT) |

## Return codes — dealer (11000–11002)

| Value | Name | Meaning |
|---|---|---|
| 11000 | MT_RET_REQUEST_RETURN | Request returned to the queue |
| 11001 | MT_RET_REQUEST_DONE_CANCEL | Request partially filled, remainder canceled |
| 11002 | MT_RET_REQUEST_REQUOTE_RETURN | Request requoted and returned to the queue |

## Return codes — trade management (4001–4009)

| Value | Name | Meaning |
|---|---|---|
| 4001 | MT_RET_TRADE_LIMIT_REACHED | Reached the limit on the number of orders or deals |
| 4002 | MT_RET_TRADE_ORDER_EXIST | The order already exists |
| 4003 | MT_RET_TRADE_ORDER_EXHAUSTED | The range of orders has been exhausted |
| 4004 | MT_RET_TRADE_DEAL_EXHAUSTED | The range of deals has been exhausted |
| 4005 | MT_RET_TRADE_MAX_MONEY | Reached the limit on the amount of money |
| 4006 | MT_RET_TRADE_DEAL_EXIST | A deal with this ticket already exists |
| 4007 | MT_RET_TRADE_ORDER_PROHIBITED | Order identifier reserved for another use |
| 4008 | MT_RET_TRADE_DEAL_PROHIBITED | Deal identifier reserved for another use |
| 4009 | MT_RET_TRADE_SPLIT_VOLUME | Position volume would become zero after the split |

## Return codes — subscriptions (15000–15007)

| Value | Name | Meaning |
|---|---|---|
| 15000 | MT_RET_SUBS_NOT_FOUND | Subscription not found |
| 15001 | MT_RET_SUBS_NOT_FOUND_CFG | Subscription configuration not found |
| 15002 | MT_RET_SUBS_NOT_FOUND_USER | User from subscription not found |
| 15003 | MT_RET_SUBS_DISABLED | Subscription disabled |
| 15004 | MT_RET_SUBS_PERMISSION_USER | Subscription not allowed for the user |
| 15005 | MT_RET_SUBS_PERMISSION_SUBSCRIBE | Subscribing not allowed |
| 15006 | MT_RET_SUBS_PERMISSION_UNSUBSCRIBE | Unsubscribing not allowed |
| 15007 | MT_RET_SUBS_REAL_ONLY | Subscription allowed only for real accounts |

> Other return-code bands exist (authentication, configuration management, price data, report
> generation, user management, messengers, common errors). They are documented in
> `Meta Document - Admin/*Return Codes*.pdf` and are outside the data-model scope of this file.

---

# Symbols

## EnCalcMode

Profit and margin calculation model for a symbol. Stored in `mt5_symbols.CalcMode`. Note the
sparse banding: 0–5 OTC, 32–37 exchange, 64 service.

| Value | Name | Meaning |
|---|---|---|
| 0 | TRADE_MODE_FOREX | Forex calculation |
| 1 | TRADE_MODE_FUTURES | Futures calculation |
| 2 | TRADE_MODE_CFD | CFD calculation |
| 3 | TRADE_MODE_CFDINDEX | CFD Index calculation |
| 4 | TRADE_MODE_CFDLEVERAGE | CFD Leverage calculation |
| 5 | TRADE_MODE_FOREX_NO_LEVERAGE | Forex without leverage |
| 32 | TRADE_MODE_EXCH_STOCKS | Exchange Stocks |
| 33 | TRADE_MODE_EXCH_FUTURES | Exchange Futures |
| 34 | TRADE_MODE_EXCH_FORTS | Exchange FORTS (Moscow Exchange Derivatives Market) |
| 35 | TRADE_MODE_EXCH_OPTIONS | Exchange Options |
| 36 | TRADE_MODE_EXCH_OPTIONS_MARGIN | Exchange Margin Options |
| 37 | TRADE_MODE_EXCH_BONDS | Exchange Bonds |
| 64 | TRADE_MODE_SERV_COLLATERAL | Non-tradable collateral instrument — used as client assets to provide margin for other instruments. No margin/profit calculated |

## EnTradeMode

Trading permission for a symbol. Stored in `mt5_symbols.TradeMode`.

| Value | Name | Meaning |
|---|---|---|
| 0 | TRADE_DISABLED | Trade is disabled |
| 1 | TRADE_LONGONLY | Only long positions are allowed |
| 2 | TRADE_SHORTONLY | Only short positions are allowed |
| 3 | TRADE_CLOSEONLY | Only closure is allowed |
| 4 | TRADE_FULL | Full trading access |

## EnExecutionMode

Order execution model. Stored in `mt5_symbols.ExecMode`.

| Value | Name | Meaning |
|---|---|---|
| 0 | EXECUTION_REQUEST | Request execution |
| 1 | EXECUTION_INSTANT | Instant execution |
| 2 | EXECUTION_MARKET | Market execution |
| 3 | EXECUTION_EXCHANGE | Exchange execution |

## EnFillingFlags

*(bitmask — combine with OR)* Filling methods allowed for a symbol.

| Value | Name | Meaning |
|---|---|---|
| 0 | FILL_FLAGS_NONE | All filling methods disabled |
| 1 | FILL_FLAGS_FOK | Fill or Kill allowed |
| 2 | FILL_FLAGS_IOC | Immediate or Cancel allowed |

## EnExpirationFlags

*(bitmask — combine with OR)* Order expiration types allowed for a symbol.

| Value | Name | Meaning |
|---|---|---|
| 0 | TIME_FLAGS_NONE | All expiration types disabled |
| 1 | TIME_FLAGS_GTC | Good till canceled |
| 2 | TIME_FLAGS_DAY | Effective only during the current trading day |
| 4 | TIME_FLAGS_SPECIFIED | Effective until a trader-specified date |
| 8 | TIME_FLAGS_SPECIFIED_DAY | Expires at 00:00 of a specified day, or at the nearest trade time |

## EnOrderFlags

*(bitmask — combine with OR)* Order types permitted for the symbol.

| Value | Name | Meaning |
|---|---|---|
| 0 | ORDER_FLAGS_NONE | No flags |
| 1 | ORDER_FLAGS_MARKET | Market Buy/Sell allowed |
| 2 | ORDER_FLAGS_LIMIT | Buy Limit / Sell Limit allowed |
| 4 | ORDER_FLAGS_STOP | Buy Stop / Sell Stop allowed |
| 8 | ORDER_FLAGS_STOP_LIMIT | Buy Stop Limit / Sell Stop Limit allowed |
| 16 | ORDER_FLAGS_SL | Stop Loss orders allowed |
| 32 | ORDER_FLAGS_TP | Take Profit orders allowed |
| 64 | ORDER_FLAGS_CLOSEBY | Close-by (OP_CLOSE_BY) allowed. Usable only on hedging accounts regardless of this flag |

## EnGTCMode

End-of-day handling of pending orders and stops.

| Value | Name | Meaning |
|---|---|---|
| 0 | ORDERS_GTC | Good till canceled — pending orders survive the day change |
| 1 | ORDERS_DAILY | Good till today, including SL/TP — all SL/TP levels and pending orders are deleted at day end |
| 2 | ORDERS_DAILY_NO_STOPS | Good till today, excluding SL/TP — only pending orders are deleted; SL/TP preserved |

## EnSwapMode

Swap calculation method. Amounts come from SwapLong / SwapShort.

| Value | Name | Meaning |
|---|---|---|
| 0 | SWAP_DISABLED | Swap charging disabled |
| 1 | SWAP_BY_POINTS | In points of the symbol price |
| 2 | SWAP_BY_SYMBOL_CURRENCY | In the symbol base currency |
| 3 | SWAP_BY_MARGIN_CURRENCY | In the symbol margin currency |
| 4 | SWAP_BY_GROUP_CURRENCY | In the group (deposit) currency |
| 5 | SWAP_BY_INTEREST_CURRENT | Percent of the symbol price at the moment of swap calculation |
| 6 | SWAP_BY_INTEREST_OPEN | Percent of the position open price |
| 7 | SWAP_REOPEN_BY_CLOSE_PRICE | Reopen the position next day at close price ± SwapLong/SwapShort points |
| 8 | SWAP_REOPEN_BY_BID | Reopen the position next day at current Bid ± SwapLong/SwapShort points |
| 9 | SWAP_BY_PROFIT_CURRENCY | In the symbol profit currency |

## EnSwapDays

Day on which triple swap is charged.

| Value | Name | Meaning |
|---|---|---|
| 0 | SWAP_DAY_SUNDAY | Sunday |
| 1 | SWAP_DAY_MONDAY | Monday |
| 2 | SWAP_DAY_TUESDAY | Tuesday |
| 3 | SWAP_DAY_WEDNESDAY | Wednesday |
| 4 | SWAP_DAY_THURSDAY | Thursday |
| 5 | SWAP_DAY_FRIDAY | Friday |
| 6 | SWAP_DAY_SATURDAY | Saturday |
| 7 | SWAP_DAY_DISABLED | Triple swaps disabled |

## EnSwapFlags

*(bitmask — combine with OR)* Additional swap settings.

| Value | Name | Meaning |
|---|---|---|
| 0 | SWAP_FLAGS_NONE | Additional settings not used |
| 1 | SWAP_FLAGS_CONSIDER_HOLIDAYS | Account for holidays: swap is doubled the day before a holiday and not charged on the holiday |

## EnMarginFlags (symbol-level)

*(bitmask — combine with OR)* Additional margin checks configured per symbol. **Distinct from
the group-level `EnMarginFlags`** below — same name, different enum.

| Value | Name | Meaning |
|---|---|---|
| 0 | MARGIN_FLAGS_NONE | Standard checking — margin checked on order placement and pending-order trigger |
| 1 | MARGIN_FLAGS_CHECK_PROCESS | Additionally check margin before executing an order after server/dealer/gateway confirmation |
| 2 | MARGIN_FLAGS_CHECK_SLTP | Additionally check margin before closing a position by SL/TP; skip the close if margin would become insufficient |
| 4 | MARGIN_FLAGS_HEDGE_LARGE_LEG | Calculate hedged margin using the larger leg |

## EnMarginRateTypes (a.k.a. EnMarginTypes)

Index into the symbol's initial/maintenance margin-rate arrays. Named `EnMarginTypes` in the
SQL-export docs, `IMTConSymbol::EnMarginRateTypes` in the API docs.

| Value | Name | Meaning |
|---|---|---|
| 0 | MARGIN_BUY | Market Buy order |
| 1 | MARGIN_SELL | Market Sell order |
| 2 | MARGIN_BUY_LIMIT | Buy Limit pending order |
| 3 | MARGIN_SELL_LIMIT | Sell Limit pending order |
| 4 | MARGIN_BUY_STOP | Buy Stop pending order |
| 5 | MARGIN_SELL_STOP | Sell Stop pending order |
| 6 | MARGIN_BUY_STOP_LIMIT | Buy Stop Limit pending order |
| 7 | MARGIN_SELL_STOP_LIMIT | Sell Stop Limit pending order |

## EnTickFlags

*(bitmask — combine with OR)* Tick-data handling options.

| Value | Name | Meaning |
|---|---|---|
| 0 | TICK_NONE | No rights / no options |
| 1 | TICK_REALTIME | Allow real-time quotes from data feeds |
| 2 | TICK_COLLECTRAW | Keep raw prices |
| 4 | TICK_FEED_STATS | Take market statistics (Ask High, Bid Low, …) directly from the feed instead of computing them on the history server. Requires TICK_REALTIME |

## EnChartMode

Which price series drives the symbol chart.

| Value | Name | Meaning |
|---|---|---|
| 0 | CHART_MODE_BID_PRICE | Chart based on Bid prices |
| 1 | CHART_MODE_LAST_PRICE | Chart based on Last prices |
| 255 | CHART_MODE_OLD | Service value for internal use |

## EnOptionMode

Option type and exercise style.

| Value | Name | Meaning |
|---|---|---|
| 0 | OPTION_MODE_EUROPEAN_CALL | European Call option |
| 1 | OPTION_MODE_EUROPEAN_PUT | European Put option |
| 2 | OPTION_MODE_AMERICAN_CALL | American Call option |
| 3 | OPTION_MODE_AMERICAN_PUT | American Put option |

## EnSpliceType

Futures continuous-contract splicing method.

| Value | Name | Meaning |
|---|---|---|
| 0 | SPLICE_NONE | No splicing |
| 1 | SPLICE_UNADJUSTED | Splice "as is" — the previous contract's price level is not adjusted to the front contract |
| 2 | SPLICE_ADJUSTED | Shift all previous-contract quotes by the gap between its last quote and the front contract's first quote, producing a smooth chart |

## EnSpliceTimeType

Date on which splicing occurs.

| Value | Name | Meaning |
|---|---|---|
| 0 | SPLICE_TIME_EXPIRATION | Splice at the moment of instrument expiration (`TimeExpiration`) |

## EnInstantMode

Instant-execution check type.

| Value | Name | Meaning |
|---|---|---|
| 0 | INSTANT_CHECK_NORMAL | Normal mode of instant execution |

## EnRequestFlags (symbol-level)

*(bitmask — combine with OR)* Request-execution-mode options.

| Value | Name | Meaning |
|---|---|---|
| 0 | REQUEST_FLAGS_NONE | No flags |
| 1 | REQUEST_FLAGS_ORDER | Additional confirmation mode |

## EnTradeFlags (symbol-level)

*(bitmask — combine with OR)* Per-symbol trade options. **Distinct from the group-level
`EnTradeFlags`** — same name, different enum and different value space.

| Value | Name | Meaning |
|---|---|---|
| 0 | TRADE_FLAGS_NONE | No flags |
| 1 | TRADE_FLAGS_PROFIT_BY_MARKET | Forex symbols only. Convert P/L to deposit currency using current Bid (profitable deals) or Ask (losing deals) rather than the deal price |
| 2 | TRADE_FLAGS_ALLOW_SIGNALS | Allow clients to copy operations on this symbol via the Signals service |

## EnSectors

Economic sector a trading instrument belongs to.

| Value | Name | Meaning |
|---|---|---|
| 0 | SECTOR_UNDEFINED | Undefined |
| 1 | SECTOR_BASIC_MATERIALS | Basic materials |
| 2 | SECTOR_COMMUNICATION_SERVICES | Communication services |
| 3 | SECTOR_CONSUMER_CYCLICAL | Consumer cyclical |
| 4 | SECTOR_CONSUMER_DEFENSIVE | Consumer defensive |
| 5 | SECTOR_ENERGY | Energy |
| 6 | SECTOR_FINANCIAL | Finance |
| 7 | SECTOR_HEALTHCARE | Healthcare |
| 8 | SECTOR_INDUSTRIALS | Industrials |
| 9 | SECTOR_REAL_ESTATE | Real estate |
| 10 | SECTOR_TECHNOLOGY | Technology |
| 11 | SECTOR_UTILITIES | Utilities |
| 12 | SECTOR_CURRENCY | Currency |
| 13 | SECTOR_CURRENCY_CRYPTO | Crypto currency |
| 14 | SECTOR_INDEXES | Indices |
| 15 | SECTOR_COMMODITIES | Commodities |

## EnIndustries (band map)

Industry branch of a trading instrument. The enumeration has ~180 members; each sector owns a
50-value band. Full member list is in
`Meta Document - Admin/Enumerations - mt5_symbols - SQL Export - Backup Server.pdf`. The band
boundaries are reproduced here because they are what schema range-checks need:

| First | Last | Band end | Sector band |
|---|---|---|---|
| 0 | 0 | — | INDUSTRY_UNDEFINED |
| 1 | 14 | 50 | Basic materials |
| 51 | 57 | 100 | Communication services |
| 101 | 123 | 150 | Consumer cyclical |
| 151 | 162 | 200 | Consumer defensive |
| 201 | 208 | 250 | Energy |
| 251 | 269 | 300 | Finance |
| 301 | 311 | 350 | Healthcare |
| 351 | 375 | 400 | Industrials |
| 401 | 412 | 450 | Real estate |
| 451 | 462 | 500 | Technology |
| 501 | 506 | 550 | Utilities |
| 551 | 554 | 600 | Commodities |

---

# Groups

## EnMarginMode

Risk-management model — determines pre-trade control and the position accounting system.
Stored in `mt5_groups.MarginMode`. This is the single most load-bearing group enum.

| Value | Name | Meaning |
|---|---|---|
| 0 | MARGIN_MODE_RETAIL | OTC market. Margin from instrument type + group settings. **Netting** position accounting |
| 1 | MARGIN_MODE_EXCHANGE_DISCOUNT | Exchange market. Margin from discounts in symbol settings, floored by exchange-set values |
| 2 | MARGIN_MODE_RETAIL_HEDGED | OTC market. Margin from instrument type + group settings. **Hedging** position accounting |

## EnStopOutMode

Units in which Margin Call and Stop Out levels are expressed.

| Value | Name | Meaning |
|---|---|---|
| 0 | STOPOUT_PERCENT | Levels in percentage terms |
| 1 | STOPOUT_MONEY | Levels in money terms |

## EnFreeMarginMode

Whether unrealized P/L counts toward free margin.

| Value | Name | Meaning |
|---|---|---|
| 0 | FREE_MARGIN_NOT_USE_PL | Do not use unrealized profit/loss |
| 1 | FREE_MARGIN_USE_PL | Use unrealized profit and loss |
| 2 | FREE_MARGIN_PROFIT | Use unrealized profit only |
| 3 | FREE_MARGIN_LOSS | Use unrealized loss only |

## EnMarginFreeProfitFlags

How intraday realized P/L feeds into free margin.

| Value | Name | Meaning |
|---|---|---|
| 0 | FREE_MARGIN_PROFIT_PL | Include both profit and loss fixed during the day |
| 1 | FREE_MARGIN_PROFIT_LOSS | Include only loss fixed during the day. Intraday profits accumulate in `IMTAccount::BlockedProfit` and are credited to balance at end of day |

## EnMarginFlags (group-level)

*(bitmask — combine with OR)* Group margin calculation flags. **Distinct from the symbol-level
`EnMarginFlags`.**

| Value | Name | Meaning |
|---|---|---|
| 0 | MARGIN_FLAGS_NONE | No flags |
| 1 | MARGIN_FLAGS_CLEAR_ACC | Only meaningful in FREE_MARGIN_PROFIT_LOSS mode. Release accumulated profit into free margin at end of trade day. If disabled, only an external application can release it |

## EnTradeFlags (group-level)

*(bitmask — combine with OR)* Group trade options. **Distinct from the symbol-level
`EnTradeFlags`.**

| Value | Name | Meaning |
|---|---|---|
| 0x00000000 | TRADEFLAGS_NONE | Options disabled |
| 0x00000001 | TRADEFLAGS_SWAPS | Allow charging of swaps |
| 0x00000002 | TRADEFLAGS_TRAILING | Enable trailing stop |
| 0x00000004 | TRADEFLAGS_EXPERTS | Enable trading via Expert Advisors |
| 0x00000008 | TRADEFLAGS_EXPIRATION | Enable order expiration |
| 0x00000010 | TRADEFLAGS_SIGNALS_ALL | Allow the Signals service in client terminals |
| 0x00000020 | TRADEFLAGS_SIGNALS_OWN | Allow only signals sourced from this broker's own accounts |
| 0x00000040 | TRADEFLAGS_SO_COMPENSATION | Auto-run the "so compensation" operation to zero a negative balance after a Stop Out close |
| 0x00000080 | TRADEFLAGS_SO_FULLY_HEDGED | Stop out accounts with open positions, zero margin (fully covered) and negative equity. Hedging accounts only |
| 0x00000100 | TRADEFLAGS_FIFO_CLOSE | FIFO close mode — positions per instrument must be closed oldest-first. Hedging accounts only. Applies to manual, SL/TP, and Stop-Out closes |
| 0x00000200 | TRADEFLAGS_HEDGE_PROHIBIT | Prohibit opposite positions/orders on the same instrument. Hedging groups only |
| 0x00000400 | TRADEFLAGS_DEAL_COST | Calculate and display deal execution costs in client terminals (required for NFA-regulated brokers) |
| 0x00000800 | TRADEFLAGS_SO_COMPENSATION_CREDIT | Addition to TRADEFLAGS_SO_COMPENSATION — zero out credit funds after a negative-balance compensation, via a separate "so credit compensation" operation |

## EnPermissionsFlags

*(bitmask — combine with OR)* Group-level client permissions.

| Value | Name | Meaning |
|---|---|---|
| 0x00000000 | PERMISSION_NONE | No permissions (default) |
| 0x00000001 | PERMISSION_CERT_CONFIRM | Enable confirmation of certificates |
| 0x00000002 | PERMISSION_ENABLE_CONNECTION | Allow client connections |
| 0x00000004 | PERMISSION_RESET_PASSWORD | Force users to change their master password at first login |
| 0x00000008 | PERMISSION_FORCED_OTP_USAGE | Require one-time passwords for all clients in the group |
| 0x00000010 | PERMISSION_RISK_WARNING | Show a risk warning on connect; block trading until the client acknowledges it. Shown once per terminal session |
| 0x00000020 | PERMISSION_REGULATION_PROTECT | Enforce country-specific regulatory restrictions for retail clients |
| 0x00000040 | PERMISSION_NOTIFY_DEALS | Allow subscribing to server push notifications about deals |
| 0x00000080 | PERMISSION_NOTIFY_ORDERS | Allow subscribing to server push notifications about orders |
| 0x00000100 | PERMISSION_NOTIFY_BALANCES | Allow subscribing to server push notifications about balance operations |

## EnAuthMode

Client authorization mode for the group.

| Value | Name | Meaning |
|---|---|---|
| 0 | AUTH_STANDARD | Standard authorization |
| 1 | AUTH_RSA1024 | Extended authorization, 1024-bit encryption |
| 2 | AUTH_RSA2048 | Extended authorization, 2048-bit encryption |

## EnAuthOTPMode

One-time-password authentication mode.

| Value | Name | Meaning |
|---|---|---|
| 0 | AUTH_OTP_DISABLED | OTP authentication disabled |
| 1 | AUTH_OTP_TOTP_SHA256 | TOTP SHA-256 generator; OTP required for all connection types |
| 2 | AUTH_OTP_TOTP_SHA256_WEB | TOTP SHA-256 generator; OTP required only for web terminal connections |

## EnHistoryLimit

Trading history window available to clients in the group.

| Value | Name | Meaning |
|---|---|---|
| 0 | TRADE_HISTORY_ALL | The entire history |
| 1 | TRADE_HISTORY_MONTHS_1 | One month |
| 2 | TRADE_HISTORY_MONTHS_3 | Three months |
| 3 | TRADE_HISTORY_MONTHS_6 | Six months |
| 4 | TRADE_HISTORY_YEAR_1 | One year |
| 5 | TRADE_HISTORY_YEAR_2 | Two years |
| 6 | TRADE_HISTORY_YEAR_3 | Three years |

## EnTransferMode

Rules for transferring money between accounts.

| Value | Name | Meaning |
|---|---|---|
| 0 | TRANSFER_MODE_DISABLED | Transfer of funds is disabled |
| 1 | TRANSFER_MODE_NAME | Allowed only between accounts registered under the same name |
| 2 | TRANSFER_MODE_GROUP | Allowed only between accounts in the same group |
| 3 | TRANSFER_MODE_NAME_GROUP | Allowed only when both conditions hold — same name and same group |

## EnReportsMode

End-of-day / end-of-month report data generation.

| Value | Name | Meaning |
|---|---|---|
| 0 | REPORTS_DISABLED | Reports disabled |
| 1 | REPORTS_FULL | Save both end-of-day and end-of-month account states to `bases\daily\daily_*.dat` |
| 2 | REPORTS_DAY_ONLY | Generate report data only at end of day |
| 3 | REPORTS_MONTH_ONLY | Generate report data only at end of month |

## EnReportsFlags

*(bitmask — combine with OR)* Report delivery options.

| Value | Name | Meaning |
|---|---|---|
| 0 | REPORTSFLAGS_NONE | No additional options |
| 1 | REPORTSFLAGS_EMAIL | Email generated HTML reports to clients, using the account email address |
| 2 | REPORTSFLAGS_SUPPORT | Send report copies to the technical support email address |
| 4 | REPORTSFLAGS_STATEMENTS | Generate account state reports from `\templates\confirmation\` and `\templates\statement\`. Requires the standard reports mode to be enabled |

## EnNewsMode

News delivery to clients in the group.

| Value | Name | Meaning |
|---|---|---|
| 0 | NEWS_MODE_DISABLED | News sending is disabled |
| 1 | NEWS_MODE_HEADERS | Only news headers |
| 2 | NEWS_MODE_FULL | Full news package |

## EnMailMode

Internal mail system availability.

| Value | Name | Meaning |
|---|---|---|
| 0 | MAIL_MODE_DISABLED | Internal mail disabled |
| 1 | MAIL_MODE_FULL | Internal mail enabled |

---

# Users, Clients, Managers

## EnUsersRights

*(bitmask — combine with OR)* Per-account permissions. Stored as a 64-bit value in
`mt5_users.Rights`.

| Value | Name | Meaning |
|---|---|---|
| 0x0000000000000000 | USER_RIGHT_NONE | No permissions |
| 0x0000000000000001 | USER_RIGHT_ENABLED | The user is allowed to connect |
| 0x0000000000000002 | USER_RIGHT_PASSWORD | The user is allowed to change the password |
| 0x0000000000000004 | USER_RIGHT_TRADE_DISABLED | Trading is disabled for the user |
| 0x0000000000000008 | USER_RIGHT_INVESTOR | Service value for internal use |
| 0x0000000000000010 | USER_RIGHT_CONFIRMED | User's certificate is confirmed |
| 0x0000000000000020 | USER_RIGHT_TRAILING | Allowed to use trailing stop |
| 0x0000000000000040 | USER_RIGHT_EXPERT | Allowed to use Expert Advisors |
| 0x0000000000000080 | USER_RIGHT_OBSOLETE | Obsolete; not used |
| 0x0000000000000100 | USER_RIGHT_REPORTS | Allowed to receive daily reports. If off, daily reports are neither generated nor sent |
| 0x0000000000000200 | USER_RIGHT_READONLY | Service value for internal use |
| 0x0000000000000400 | USER_RIGHT_RESET_PASS | Must change password on next connection |
| 0x0000000000000800 | USER_RIGHT_OTP_ENABLED | May use OTP authentication |
| 0x0000000000002000 | USER_RIGHT_SPONSORED_HOSTING | Broker-sponsored VPS available to this account; controls whether the payment plan is offered in the terminal |
| 0x0000000000004000 | USER_RIGHT_API_ENABLED | Allowed to connect via the Web API. *(The `IMTUser` page marks this flag obsolete; the SQL-export page does not.)* |
| 0x0000000000008000 | USER_RIGHT_PUSH_NOTIFICATION | Push notifications from the trade server are enabled in the terminal. Subscribe-ability is gated by `PERMISSION_NOTIFY_*` on the group |
| 0x0000000000010000 | USER_RIGHT_TECHNICAL | Marks a technical account. Hides it from managers lacking `RIGHT_ACC_TECHNICAL`, in account lists and the online-accounts list |
| 0x0000000000020000 | USER_RIGHT_EXCLUDE_REPORTS | Exclude the account from server reports |

> Bit `0x0000000000001000` is not documented on either source page. Do not assume it is free.

## EnUsersPasswords

Password slot type.

| Value | Name | Meaning |
|---|---|---|
| 0 | USER_PASS_MAIN | Master password |
| 1 | USER_PASS_INVESTOR | Investor (read-only) password |
| 2 | USER_PASS_API | API password |

## EnUsersConnectionTypes

Client/terminal type of a connection. Banded: clients 0–11, staff 32+.

| Value | Name | Meaning |
|---|---|---|
| 0 | USER_TYPE_CLIENT | Client terminal |
| 1 | USER_TYPE_CLIENT_WINMOBILE | Windows Mobile terminal (not used) |
| 2 | USER_TYPE_CLIENT_WINPHONE | Windows Phone 7 terminal (not used) |
| 3 | USER_TYPE_CLIENT_API_WEB | Client Web API |
| 4 | USER_TYPE_CLIENT_IPHONE | iPhone mobile terminal |
| 5 | USER_TYPE_CLIENT_ANDROID | Android mobile terminal |
| 6 | USER_TYPE_CLIENT_BLACKBERRY | BlackBerry mobile terminal (not used) |
| 11 | USER_TYPE_CLIENT_WEB | WebTerminal |
| 32 | USER_TYPE_ADMIN | Administrator terminal |
| 33 | USER_TYPE_MANAGER | Manager terminal |
| 34 | USER_TYPE_MANAGER_API | Manager API, manager interface |
| 36 | USER_TYPE_ADMIN_API | Manager API, administrator interface |
| 37 | USER_TYPE_MANAGER_API_WEB | Manager Web API |

> Value 11 appears in `IMTUser` but not on the SQL-export page. Value 35 is not documented.

## EnSoActivation

Account margin status. Stored in `mt5_users.SOActivation`.

| Value | Name | Meaning |
|---|---|---|
| 0 | ACTIVATION_NONE | None |
| 1 | ACTIVATION_MARGIN_CALL | Margin call |
| 2 | ACTIVATION_STOP_OUT | Stop out |

## EnManagerRights

Manager/administrator permission identifiers. These are **bit indices, not bit values** — the
manager config stores a per-right state (see `EnManagerRightFlags`) addressed by these indices.
`RIGHT_LAST` = 128 bounds the space.

| Value | Name | Meaning |
|---|---|---|
| 0 | RIGHT_ADMIN | Connect using the administrator terminal |
| 1 | RIGHT_MANAGER | Connect using the manager terminal |
| 10 | RIGHT_CFG_SERVERS | Network configuration |
| 11 | RIGHT_CFG_ACCESS | Configuration of the IP access list |
| 12 | RIGHT_CFG_TIME | Configuration of server working time |
| 13 | RIGHT_CFG_HOLIDAYS | Configuration of holidays |
| 14 | RIGHT_CFG_HST_SYNC | Configuration of history synchronization |
| 15 | RIGHT_CFG_SYMBOLS | Configuration of symbols |
| 16 | RIGHT_CFG_GROUPS | Configuration of groups |
| 17 | RIGHT_CFG_MANAGERS | Configuration of manager rights |
| 18 | RIGHT_CFG_DATAFEEDS | Configuration of data feeds |
| 19 | RIGHT_CFG_REQUESTS | Configuration of the routing table |
| 20 | RIGHT_SRV_JOURNALS | Access to server journals |
| 21 | RIGHT_SRV_REPORTS | Receive automatic server reports |
| 22 | RIGHT_CHARTS | Edit history data on the server (also needs RIGHT_QUOTES) |
| 23 | RIGHT_EMAIL | Send internal emails |
| 24 | RIGHT_ACCOUNTANT | Work with funds on accounts |
| 25 | RIGHT_ACC_READ | Access to accounts |
| 26 | RIGHT_ACC_DETAILS_NAME | Access to name details in accounts |
| 27 | RIGHT_ACC_MANAGER | Edit accounts |
| 28 | RIGHT_ACC_ONLINE | Get current client connections |
| 29 | RIGHT_TRADES_READ | View orders, deals and positions. Gates RIGHT_TRADES_MANAGER and RIGHT_TRADES_DEALER |
| 30 | RIGHT_TRADES_MANAGER | Modify any orders/deals/positions; access the Exposure tab; place requests on behalf of clients |
| 31 | RIGHT_QUOTES | Add quotes to the stream |
| 32 | RIGHT_RISK_MANAGER | Receive client position and company coverage information |
| 33 | RIGHT_REPORTS | Request and receive reports |
| 34 | RIGHT_NEWS | Send news (account must be in a group configured for it) |
| 35 | RIGHT_CFG_GATEWAYS | Configure gateways |
| 36 | RIGHT_CFG_PLUGINS | Configure plugins |
| 37 | RIGHT_TRADES_DEALER | Trading and dealing activities in the manager terminal |
| 38 | RIGHT_CFG_REPORTS | Configuration of reports |
| 39 | RIGHT_EXPORT | Export data (orders, accounts, …) from admin/manager terminals |
| 40 | RIGHT_SYMBOL_DETAILS | Change spread and execution settings of symbols |
| 41 | RIGHT_TECHSUPPORT | Technical Support tab (obsolete, no longer used) |
| 42 | RIGHT_TRADES_SUPERVISOR | View requests forwarded from client groups and observe processing by other dealers, without connecting as a dealer |
| 43 | RIGHT_QUOTES_RAW | Enables "Show raw quotes" in Market Watch — view quotes before the group spread markup |
| 44 | RIGHT_MARKET | Access the Market of applications (obsolete) |
| 45 | RIGHT_GRP_DETAILS_MARGIN | Change group margin settings (`IMTConGroupSymbol::Margin*`) |
| 46 | RIGHT_NOTIFICATIONS | Send push notifications to clients by MetaQuotes ID |
| 47 | RIGHT_ACC_DELETE | Delete client accounts (also needs RIGHT_ACC_MANAGER) |
| 48 | RIGHT_TRADES_DELETE | Delete orders/positions (also needs RIGHT_TRADES_MANAGER) |
| 49 | RIGHT_CONFIRM_ACTIONS | Skip the confirmation-code dialog for balance operations and bulk order closes |
| 50 | RIGHT_CFG_ECN | ECN settings |
| 51 | RIGHT_GRP_DETAILS_COMMISSION | Change commission settings (`IMTConCommission`) |
| 52 | RIGHT_SUBSCRIPTIONS_VIEW | View subscription settings and statistics |
| 53 | RIGHT_SUBSCRIPTIONS_EDIT | Create/change/remove subscription settings |
| 54 | RIGHT_CFG_FUNDS | Configure investment funds |
| 55 | RIGHT_CFG_MAILS | Configure email service integration |
| 56 | RIGHT_CFG_MESSENGERS | Configure SMS provider integration |
| 57 | RIGHT_CFG_KYC | Configure KYC service integration |
| 58 | RIGHT_CFG_AUTOMATIONS | Configure automatic actions |
| 59 | RIGHT_CFG_ALLOCATIONS | Access the Allocations section (groups/servers for demo and preliminary real accounts) |
| 60 | RIGHT_CFG_VPS | Configure Sponsored VPS |
| 61 | RIGHT_CFG_PAYMENTS | Configure payment system integration |
| 62 | RIGHT_ADMIN_COMPUTER | Server machine administration (Network section) |
| 63 | RIGHT_CFG_WEB_SERVICES | Configure web service integration and callback addresses |
| 64 | RIGHT_FINTEZA_ACCESS | Access the Finteza Analytics section |
| 65 | RIGHT_FINTEZA_WEBSITES | View Finteza website data |
| 66 | RIGHT_FINTEZA_CAMPAIGNS | View Finteza marketing campaign data |
| 67 | RIGHT_FINTEZA_REPORTS | Not currently used |
| 70 | RIGHT_ACC_TECHNICAL | Access accounts flagged `USER_RIGHT_TECHNICAL` |
| 71 | RIGHT_ACC_TECHNICAL_MODIFY | Toggle `USER_RIGHT_TECHNICAL` / `USER_RIGHT_EXCLUDE_REPORTS`. Without it, those options are read-only |
| 72 | RIGHT_PAYMENTS_PROCESS | Confirm and reject payments via the manager terminal |
| 73 | RIGHT_CFG_CORPORATES | Access corporate link settings |
| 74 | RIGHT_CFG_LEVERAGES | Configuration of floating leverage |
| 75 | RIGHT_CFG_STREAMING | Configure data streaming to external systems (e.g. Kafka) |
| 76 | RIGHT_ACC_DETAILS_LOCATION | Access location data in accounts (country, region, code) |
| 77 | RIGHT_ACC_DETAILS_ADDRESS | Access address details in accounts |
| 78 | RIGHT_ACC_DETAILS_ID | Access document numbers in accounts |
| 79 | RIGHT_ACC_DETAILS_EMAIL | Access email details in accounts |
| 80 | RIGHT_ACC_DETAILS_PHONE | Access phone details in accounts |
| 81 | RIGHT_ACC_DETAILS_GENERAL | Access other account data (language, MetaQuotes ID, …) |
| 82 | RIGHT_CLIENTS_DETAILS_NAME | Access name details in clients |
| 83 | RIGHT_CLIENTS_DETAILS_LOCATION | Access location data in clients |
| 84 | RIGHT_CLIENTS_DETAILS_ADDRESS | Access address details in clients |
| 85 | RIGHT_CLIENTS_DETAILS_ID | Access document numbers in clients |
| 86 | RIGHT_CLIENTS_DETAILS_EMAIL | Access email details in clients |
| 87 | RIGHT_CLIENTS_DETAILS_PHONE | Access phone details in clients |
| 88 | RIGHT_CLIENTS_DETAILS_GENERAL | Access other client data (language, Lead Campaign, …) |
| 96 | RIGHT_CLIENTS_ACCESS | Access the Clients section |
| 97 | RIGHT_CLIENTS_CREATE | Create new client entries manually |
| 98 | RIGHT_CLIENTS_EDIT | Edit client data |
| 99 | RIGHT_CLIENTS_DELETE | Delete client entries |
| 100 | RIGHT_DOCUMENTS_ACCESS | View client documents |
| 101 | RIGHT_DOCUMENTS_CREATE | Add document metadata to client entries |
| 102 | RIGHT_DOCUMENTS_EDIT | Edit document metadata |
| 103 | RIGHT_DOCUMENTS_DELETE | Delete document metadata |
| 104 | RIGHT_DOCUMENTS_FILES_ADD | Add document files |
| 105 | RIGHT_DOCUMENTS_FILES_DELETE | Delete document files |
| 106 | RIGHT_COMMENTS_ACCESS | Read comments on clients and documents |
| 107 | RIGHT_COMMENTS_CREATE | Write comments on clients and documents |
| 108 | RIGHT_COMMENTS_DELETE | Delete comments |
| 109 | RIGHT_CLIENTS_KYC | Launch automated KYC validation (`IMTAdminAPI::KYCStart`, `IMTServerAPI::KYCStart`) |
| 110 | RIGHT_PAYMENTS_ACCESS | View current and processed payments and payment accounts |
| 111 | RIGHT_PAYMENTS_EDIT | Edit payments and payment accounts |
| 112 | RIGHT_PAYMENTS_DELETE | Delete payments and payment accounts |
| 113 | RIGHT_ULTENCY_ACCESS | View Ultency settings (providers, provider symbols, aggregated symbols) |
| 114 | RIGHT_ULTENCY_EDIT | Edit Ultency settings and add liquidity providers |
| 115 | RIGHT_ULTENCY_DELETE | Delete Ultency settings |
| 116 | RIGHT_ULTENCY_SERVICEDESK | Use the built-in Ultency communication/service desk |
| 128 | RIGHT_LAST | End of enumeration |

## EnManagerRightFlags

State of an individual manager right.

| Value | Name | Meaning |
|---|---|---|
| 0 | RIGHT_FLAGS_DENIED / RIGHT_FLAGS_NONE | The right is not granted |
| 1 | RIGHT_FLAGS_GRANTED | The right is granted |

## EnManagerLimit

Lookback window a manager may query for reports and system logs.

| Value | Name | Meaning |
|---|---|---|
| 0 | MANAGER_LIMIT_ALL | Unlimited |
| 1 | MANAGER_LIMIT_MONTHS_1 | One month |
| 2 | MANAGER_LIMIT_MONTHS_3 | Three months |
| 3 | MANAGER_LIMIT_MONTHS_6 | Six months |
| 4 | MANAGER_LIMIT_YEAR_1 | One year |
| 5 | MANAGER_LIMIT_YEAR_2 | Two years |
| 6 | MANAGER_LIMIT_YEAR_3 | Three years |

## EnClientType

| Value | Name | Meaning |
|---|---|---|
| 0 | CLIENT_TYPE_UNDEFINED | Not set |
| 1 | CLIENT_TYPE_INDIVIDUAL | Private individual |
| 2 | CLIENT_TYPE_CORPORATE | Corporate |
| 3 | CLIENT_TYPE_FUND | Fund |

## EnClientStatus

CRM lifecycle status of a client record. Note the 100-step spacing, which leaves room for
intermediate statuses.

| Value | Name | Meaning |
|---|---|---|
| 0 | CLIENT_STATUS_UNREGISTERED | Not registered — anonymous record created from a demo account |
| 100 | CLIENT_STATUS_REGISTERED | Registered — contact details filled |
| 200 | CLIENT_STATUS_NOTINTERESTED | Not interested — has contact data but did not open a real account |
| 300 | CLIENT_STATUS_APPLICATION_INCOMPLETED | Application for a real account not completed |
| 400 | CLIENT_STATUS_APPLICATION_COMPLETED | Application completed and documents submitted |
| 500 | CLIENT_STATUS_APPLICATION_INFORMATION | Further information needed from the client |
| 600 | CLIENT_STATUS_APPLICATION_REJECTED | Application rejected |
| 700 | CLIENT_STATUS_APPROVED | Approved — a real account may be opened |
| 800 | CLIENT_STATUS_FUNDED | Client has funded the account |
| 900 | CLIENT_STATUS_ACTIVE | Active |
| 1000 | CLIENT_STATUS_INACTIVE | Inactive |
| 1100 | CLIENT_STATUS_SUSPENDED | Suspended |
| 1200 | CLIENT_STATUS_CLOSED | Closed |
| 1300 | CLIENT_STATUS_TERMINATED | Terminated at the company's initiative |

## EnGender

| Value | Name | Meaning |
|---|---|---|
| 0 | GENDER_UNSPECIFIED | Not specified |
| 1 | GENDER_MALE | Male |
| 2 | GENDER_FEMALE | Female |

## EnClientOrigin

How the client record was created.

| Value | Name | Meaning |
|---|---|---|
| 0 | CLIENT_ORIGIN_MANUAL | Created manually |
| 1 | CLIENT_ORIGIN_DEMO | Created automatically from a demo account |
| 2 | CLIENT_ORIGIN_CONTEST | Created automatically from a contest account |
| 3 | CLIENT_ORIGIN_PRELIMINARY | Created automatically from a preliminary account |

> `CLIENT_ORIGIN_REAL` exists and is documented as `CLIENT_ORIGIN_LAST`, but its numeric value
> is blank in the source table. Almost certainly 4 — **not verified, do not rely on it.**

## EnKYCStatus

| Value | Name | Meaning |
|---|---|---|
| 0 | KYC_STATUS_UNDEFINED | Not determined / verification not yet performed |
| 1 | KYC_STATUS_APPROVED | Client approved |
| 2 | KYC_STATUS_DECLINED | Client rejected |

## EnEmployment

| Value | Name | Meaning |
|---|---|---|
| 0 | EMPLOY_UNEMPLOYED | Unemployed |
| 1 | EMPLOY_EMPLOYED | Employed |
| 2 | EMPLOY_SELF_EMPLOYED | Entrepreneur or self-employed |
| 3 | EMPLOY_RETIRED | Retired |
| 4 | EMPLOY_STUDENT | Student |
| 5 | EMPLOY_OTHER | Other |

## EnEmploymentIndustry

Client's employment industry. **Distinct from the symbol-level `EnIndustries`.**

| Value | Name | Meaning |
|---|---|---|
| 0 | INDUSTRY_NONE | None |
| 1 | INDUSTRY_AGRICULTURE | Agriculture and food |
| 2 | INDUSTRY_CONSTRUCTION | Architecture and construction |
| 3 | INDUSTRY_MANAGEMENT | Administration and business management |
| 4 | INDUSTRY_COMMUNICATION | Art, audio/video technology and communications |
| 5 | INDUSTRY_EDUCATION | Education and training |
| 6 | INDUSTRY_GOVERNMENT | State and administrative services |
| 7 | INDUSTRY_HEALTHCARE | Health care |
| 8 | INDUSTRY_TOURISM | Tourism and hospitality |
| 9 | INDUSTRY_IT | Information technology |
| 10 | INDUSTRY_SECURITY | Legal and public safety services |
| 11 | INDUSTRY_MANUFACTURING | Production |
| 12 | INDUSTRY_MARKETING | Marketing and sales |
| 13 | INDUSTRY_SCIENCE | Science and technology |
| 14 | INDUSTRY_ENGINEERING | Engineering and mathematics |
| 15 | INDUSTRY_TRANSPORT | Transportation and distribution |
| 16 | INDUSTRY_OTHER | Other |

## EnEducationLevel

| Value | Name | Meaning |
|---|---|---|
| 0 | EDUCATION_LEVEL_NONE | None |
| 1 | EDUCATION_LEVEL_HIGH_SCHOOL | Secondary |
| 2 | EDUCATION_LEVEL_BACHELOR | Bachelor's degree or equivalent |
| 3 | EDUCATION_LEVEL_MASTER | Master's degree or equivalent |
| 4 | EDUCATION_LEVEL_PHD | PhD or equivalent |
| 5 | EDUCATION_LEVEL_OTHER | Other |

## EnWealthSource

| Value | Name | Meaning |
|---|---|---|
| 0 | WEALTH_SOURCE_EMPLOYMENT | Employment or business |
| 1 | WEALTH_SOURCE_SAVINGS | Savings or investments |
| 2 | WEALTH_SOURCE_INHERITANCE | Gift or inheritance |
| 3 | WEALTH_SOURCE_OTHER | Other |

## EnPreferredCommunication

| Value | Name | Meaning |
|---|---|---|
| 0 | PREFERRED_COMMUNICATION_UNDEFINED | Not specified |
| 1 | PREFERRED_COMMUNICATION_EMAIL | Email |
| 2 | PREFERRED_COMMUNICATION_PHONE | Phone |
| 3 | PREFERRED_COMMUNICATION_PHONE_SMS | SMS |
| 4 | PREFERRED_COMMUNICATION_MESSENGER | Instant messenger |

## EnTradingExperience

Applied to `ExperienceCFD`, `ExperienceFutures`, `ExperienceFX`, `ExperienceStocks`.

| Value | Name | Meaning |
|---|---|---|
| 0 | EXPERIENCE_LESS_1_YEAR | Less than 1 year |
| 1 | EXPERIENCE_1_3_YEAR | 1–3 years |
| 2 | EXPERIENCE_ABOVE_3_YEAR | More than 3 years |

## EnDocumentTypes

Client document category. Individual documents 0–2, corporate 1000+.

| Value | Name | Meaning |
|---|---|---|
| 0 | DOCUMENT_TYPE_OTHER | Another document |
| 1 | DOCUMENT_TYPE_PERSONAL_IDENTITY | ID document |
| 2 | DOCUMENT_TYPE_PERSONAL_ADDRESS | Proof of address |
| 1000 | DOCUMENT_TYPE_REGISTERED_ADDRESS | Document confirming legal address |
| 1001 | DOCUMENT_TYPE_DIRECTORS_PASSPORT | Company CEO's ID document |
| 1002 | DOCUMENT_TYPE_CERTIFICATE_OF_INCORPORATION | Certificate of incorporation |
| 1003 | DOCUMENT_TYPE_CERTIFICATE_OF_DIRECTORS | Certificate of directors |
| 1004 | DOCUMENT_TYPE_CERTIFICATE_OF_GOOD_STANDING | Certificate of good standing |

## EnDocumentSubtype

| Value | Name | Meaning |
|---|---|---|
| 0 | DOCUMENT_SUBTYPE_OTHER | Another document |
| 1 | DOCUMENT_SUBTYPE_ID_CARD | ID card |
| 2 | DOCUMENT_SUBTYPE_PASSPORT | Passport |
| 3 | DOCUMENT_SUBTYPE_DRIVERS | Driver's license |
| 4 | DOCUMENT_SUBTYPE_BANK_CARD | Bank card |
| 5 | DOCUMENT_SUBTYPE_UTILITY_BILL | Utility bill |
| 6 | DOCUMENT_SUBTYPE_BANK_STATEMENT | Bank statement |
| 7 | DOCUMENT_SUBTYPE_TAX_STATEMENT | Tax return statement |
| 8 | DOCUMENT_SUBTYPE_SELFIE | Selfie |
| 9 | DOCUMENT_SUBTYPE_PROFILE_IMAGE | Profile photo |
| 10 | DOCUMENT_SUBTYPE_ID_DOC_PHOTO | Photo from an identity document |
| 11 | DOCUMENT_SUBTYPE_AGREEMENT | Agreement |
| 12 | DOCUMENT_SUBTYPE_CONTRACT | Contract |
| 13 | DOCUMENT_SUBTYPE_RESIDENCE_PERMIT | Residence permit |
| 14 | DOCUMENT_SUBTYPE_EMPLOYMENT_CERTIFICATE | Employment agreement |
| 15 | DOCUMENT_SUBTYPE_DRIVERS_TRANSLATION | Translation of a driver's license |
| 16 | DOCUMENT_SUBTYPE_INVESTOR_DOC | Investor document |
| 17 | DOCUMENT_SUBTYPE_VEHICLE_REG_CERTIFICATE | Vehicle registration |
| 18 | DOCUMENT_SUBTYPE_INCOME_SOURCE | Proof of income |
| 19 | DOCUMENT_SUBTYPE_PAYMENT_METHOD | Proof of payment |

## EnDocumentStatus

| Value | Name | Meaning |
|---|---|---|
| 0 | DOCUMENT_STATUS_NEW | New |
| 100 | DOCUMENT_STATUS_APPROVED | Approved |
| 200 | DOCUMENT_STATUS_REJECTED | Rejected |
| 300 | DOCUMENT_STATUS_ARCHIVED | Archived |
| 400 | DOCUMENT_STATUS_DELETED | Deleted |

---

# Commissions

Terminology note: the task-level names `EnCommTierMode` / `EnCommTierType` correspond to the
documented `IMTConCommTier::EnCommissionMode` and `IMTConCommTier::EnCommissionVolumeType`.

## EnCommMode

Commission type. `mt5_commissions.Mode`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_STANDARD | Standard commission charged by the broker on client trades |
| 1 | COMM_AGENT | Agent commission — routed to the account in `IMTUser::Agent` |
| 2 | COMM_FEE | Fee — same settings and calculation as standard, but only `COMM_CHARGE_INSTANT` is allowed and the amount lands in `IMTDeal::Fee` rather than `IMTDeal::Commission` |

> The SQL-export page for `mt5_commissions` documents only 0 and 1. Value 2 (`COMM_FEE`) comes
> from `IMTConCommission`.

## EnCommRangeMode

How commission tier boundaries are measured. `mt5_commissions.ModeRange`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_RANGE_VOLUME | Levels by volume |
| 1 | COMM_RANGE_TURNOVER_MONEY | Levels by turnover in money. Period set by ChargeMode; currency defaults to the group deposit currency, overridable via `TurnoverCurrency` |
| 2 | COMM_RANGE_TURNOVER_VOLUME | Levels by turnover in lots. Period set by ChargeMode |

## EnCommChargeMode

When the commission is debited. `mt5_commissions.ModeCharge`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_CHARGE_DAILY | End of day. Accrues in `IMTAccount::BlockedCommission`, then debited as a `DEAL_COMMISSION_DAILY` (agent: `DEAL_AGENT_DAILY`) operation |
| 1 | COMM_CHARGE_MONTHLY | End of month. Same accrual, debited as `DEAL_COMMISSION_MONTHLY` (agent: `DEAL_AGENT_MONTHLY`) |
| 2 | COMM_CHARGE_INSTANT | Immediately on execution. Written to `IMTDeal::Commission`; agent commissions become `DEAL_AGENT` operations. Requires `COMM_RANGE_VOLUME` tiers |

## EnCommEntryMode

Charge by deal direction relative to the position. `mt5_commissions.ModeEntry`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_ENTRY_ALL | Charged regardless of direction. Not charged on Close-By deals (the two original deals were already charged) |
| 1 | COMM_ENTRY_IN | Only on entry deals. For in/out deals, only on the newly opened volume |
| 2 | COMM_ENTRY_OUT | Only on exit deals. For in/out deals, only on the closed volume. For Close-By, charged for both deals with the total written to the exit deal |

## EnCommActionMode

Charge by deal type. `mt5_commissions.ModeAction`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_ACTION_ALL | All transactions regardless of type |
| 1 | COMM_ACTION_BUY | Only Buy deals |
| 2 | COMM_ACTION_SELL | Only Sell deals |

## EnCommProfitMode

Charge by deal profitability. `mt5_commissions.ModeProfit`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_PROFIT_ALL | All deals |
| 1 | COMM_PROFIT_PROFIT | Only profitable deals |
| 2 | COMM_PROFIT_LOSS | Only losing deals |

## EnCommReasonFlags (ModeReason)

*(bitmask — combine with OR)* Charge only for deals whose reason matches one of these flags.
`mt5_commissions.ModeReason`.

| Value | Name | Meaning |
|---|---|---|
| 0x00000000 | COMM_REASON_FLAG_NONE | No commission charged for any transaction |
| 0x00000001 | COMM_REASON_FLAG_CLIENT | Deal performed manually via the client terminal |
| 0x00000002 | COMM_REASON_FLAG_EXPERT | Deal performed via an Expert Advisor |
| 0x00000004 | COMM_REASON_FLAG_DEALER | Deal performed by a dealer via the manager terminal |
| 0x00000008 | COMM_REASON_FLAG_EXTERNAL_CLIENT | Deal performed from an external trading system |
| 0x00000010 | COMM_REASON_FLAG_MOBILE | Deal performed via the MT5 mobile terminal (Android/iPhone) |
| 0x00000020 | COMM_REASON_FLAG_WEB | Deal performed via the web terminal |
| 0x00000040 | COMM_REASON_FLAG_SIGNAL | Deal performed by copying a trading signal per a subscription |
| 0x00000080 | (gateway) | Deal performed from a gateway |
| 0x00000100 | (Ultency) | Deal received from Ultency — used when acting as a liquidity provider |

> Flags `0x80` and `0x100` are documented on `sql_mt5_commissions.htm` but not on the
> `IMTConCommission` API page, which stops at `0x40`. The SQL page does not give them symbolic
> names; the descriptions above are its wording.

## EnCommTierMode (IMTConCommTier::EnCommissionMode)

Unit in which a commission tier's amount is expressed. `mt5_commissions_tiers.Mode`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_MONEY_DEPOSIT | Money, in the group deposit currency |
| 1 | COMM_MONEY_SYMBOL_BASE | Money, in the symbol base currency |
| 2 | COMM_MONEY_SYMBOL_PROFIT | Money, in the symbol profit currency |
| 3 | COMM_MONEY_SYMBOL_MARGIN | Money, in the symbol margin currency |
| 4 | COMM_PIPS | In pips (points) |
| 5 | COMM_PERCENT | Percent of the deal/turnover value (volume × contract size × price). Conversion currency overridable via `IMTConCommission::TurnoverCurrency` |
| 6 | COMM_MONEY_SPECIFIED | Money, in the currency named in `IMTConCommTier::Currency` |
| 7 | COMM_PERCENT_PROFIT | Percent of deal profit (instant) or of total period profit (daily/monthly). On a loss the amount is **credited** to the trader instead of debited |

> The `sql_mt5_commissions_tiers.htm` page documents only 0–6. Value 7 comes from
> `IMTConCommTier`. After calculation the amount is converted to the group deposit currency.

## EnCommTierType (IMTConCommTier::EnCommissionVolumeType)

Whether the tier amount is per deal or per unit of volume. `mt5_commissions_tiers.Type`.

| Value | Name | Meaning |
|---|---|---|
| 0 | COMM_TYPE_DEAL | Commission per deal |
| 1 | COMM_TYPE_VOLUME | Commission per volume |

---

# Routing

## EnRouteFlags

*(bitmask — combine with OR)* Request types a routing rule applies to.

| Value | Name | Meaning |
|---|---|---|
| 0x00000000 | REQUEST_NONE | No request-type condition |
| 0x00000001 | REQUEST_PRICE | Price request (request execution) |
| 0x00000002 | REQUEST_REQUEST | Dealer confirmation of order execution in request-execution mode |
| 0x00000004 | REQUEST_INSTANT | Order placement in instant execution mode |
| 0x00000008 | REQUEST_MARKET | Order placement in market execution mode |
| 0x00000010 | REQUEST_EXCHANGE | Order placement in exchange execution mode |
| 0x00000020 | REQUEST_PENDING | Placing a pending order |
| 0x00000040 | REQUEST_SLTP | Modification of position Stop Loss / Take Profit |
| 0x00000080 | REQUEST_MODIFY | Modification of a pending order |
| 0x00000100 | REQUEST_REMOVE | Deleting a pending order |
| 0x00000200 | REQUEST_ACTIVATE | Activation (triggering) of a pending order |
| 0x00000400 | REQUEST_STOPLIMIT | Activation of a Stop Limit order |
| 0x00000800 | REQUEST_SL | Triggering of a Stop Loss |
| 0x00001000 | REQUEST_TP | Triggering of a Take Profit |
| 0x00002000 | REQUEST_STOPOUT_ORDER | Delete a pending order on stop-out (when pending orders carry margin) |
| 0x00004000 | REQUEST_STOPOUT_POSITION | Close a position on stop-out |
| 0x00008000 | REQUEST_EXPIRATION | Cancellation of an order upon expiration |
| 0x00010000 | REQUEST_DEALER_POS_EXECUTE | Position opening/closing by a dealer |
| 0x00020000 | REQUEST_DEALER_ORD_PENDING | Pending order placement by a dealer |
| 0x00040000 | REQUEST_DEALER_POS_MODIFY | Position modification by a dealer |
| 0x00080000 | REQUEST_DEALER_ORD_MODIFY | Order modification by a dealer |
| 0x00100000 | REQUEST_DEALER_ORD_REMOVE | Order deletion by a dealer |
| 0x00200000 | REQUEST_DEALER_ORD_ACTIVATE | Order activation by a dealer |
| 0x00400000 | REQUEST_DEALER_ORD_SLIMIT | Stop Limit activation by a dealer (order becomes a limit order) |
| 0x00800000 | REQUEST_DEALER_CLOSE_BY | Close-by performed by a dealer |
| 0x01000000 | REQUEST_CLOSE_BY | Close-by performed by a client |

## EnTypeFlags

*(bitmask — combine with OR)* Order types a routing rule applies to.

| Value | Name | Meaning |
|---|---|---|
| 0x0000 | TYPE_NONE | No order-type condition |
| 0x0001 | TYPE_BUY | Buy order |
| 0x0002 | TYPE_SELL | Sell order |
| 0x0004 | TYPE_BUY_LIMIT | Buy Limit order |
| 0x0008 | TYPE_SELL_LIMIT | Sell Limit order |
| 0x0010 | TYPE_BUY_STOP | Buy Stop order |
| 0x0020 | TYPE_SELL_STOP | Sell Stop order |
| 0x0040 | TYPE_BUY_STOP_LIMIT | Buy Stop Limit order |
| 0x0080 | TYPE_SELL_STOP_LIMIT | Sell Stop Limit order |

## EnRouteAction

Action a routing rule performs. Banded: 0–4 modify-and-continue, 1001+ terminal dispositions.

| Value | Name | Meaning |
|---|---|---|
| 0 | ACTION_DELAY_TIME | Delay execution by N milliseconds (N in a separate parameter), then continue with the rules below |
| 1 | ACTION_DELAY_TICK | Delay execution by N ticks, then continue with the rules below |
| 2 | ACTION_CLEAR_TP | Clear the Take Profit level set in the order |
| 3 | ACTION_CLEAR_SL | Clear the Stop Loss level set in the order |
| 4 | ACTION_CLEAR_SLTP | Clear both Stop Loss and Take Profit |
| 1001 | ACTION_DEALER | Enqueue for the specified dealer. Skip-if-no-dealer-online is a separate parameter |
| 1002 | ACTION_DEALER_ONLINE | Pass to dealers currently online. Skip-if-none-online is a separate parameter |
| 1003 | ACTION_REJECT | Reject the request |
| 1004 | ACTION_REQUOTE | Send current market prices in response |
| 1005 | ACTION_CONFIRM_CLIENT | Confirm execution at the price requested in the order |
| 1006 | ACTION_CONFIRM_MARKET | Confirm execution at the current market price |
| 1007 | ACTION_CANCEL_ORDER | Cancel a pending order during activation or modification. Adds "deleted [by dealer]" to the comment, logs the rule, and returns `MT_RET_REQUEST_REJECT_CANCEL` (10041). Applies only to pending-order activation/modification |

## EnRouteCondition

Parameter a routing rule condition tests. Banded: request 0–12, account 1000s, financial 2000s,
activity 3000s, position/order 4000s.

| Value | Name | Meaning |
|---|---|---|
| 0 | CONDITION_DATETIME | Compare request date and time with the Value field |
| 1 | CONDITION_SYMBOL | Symbol or symbol group the rule applies to |
| 2 | CONDITION_VOLUME | Requested deal volume in lots |
| 3 | CONDITION_MARKET_DEVIATION | Instant execution only. Difference between client request price and current market price, in points |
| 4 | CONDITION_TIME | Request arrival time, in minutes since 00:00 |
| 5 | CONDITION_WEEKDAY | Day of the week |
| 6 | CONDITION_COMMENT | Request comment. `=` is exact match; `>`/`>=` search the given substring in the comment; `<`/`<=` search the comment in the given string |
| 7 | CONDITION_EXPERT | Requests placed by MQL5 programs |
| 8 | CONDITION_SIGNAL | Operations copied per a trading-signal subscription |
| 9 | CONDITION_DEALER_LOGIN | Dealer/gateway identifier on the order or position. Only usable on modify/delete — new orders have no dealer id yet |
| 10 | CONDITION_SOURCE_LOGIN | Login of the dealer who placed the request on a client's behalf |
| 11 | CONDITION_MARKET_DEVIATION_SPREAD | Like CONDITION_MARKET_DEVIATION but measured in spreads; applies to Instant, Market, pending, and SL/TP triggers. Positive deviation favours the client |
| 12 | CONDITION_GAP | Boolean — whether the symbol's gap mode is active during the request check |
| 1000 | CONDITION_LOGIN | Client account number |
| 1001 | CONDITION_GROUP | Client's account group |
| 1002 | CONDITION_COUNTRY | Client's country |
| 1003 | CONDITION_CITY | Client's city |
| 1004 | CONDITION_COLOR | Client's marker colour |
| 1005 | CONDITION_LEVERAGE | Client's leverage |
| 1006 | CONDITION_COMMENT_CLIENT | Client's comment field |
| 2000 | CONDITION_MARGIN | Currently reserved margin, in deposit currency |
| 2001 | CONDITION_MARGIN_LEVEL | Current margin level, in percent |
| 2002 | CONDITION_MARGIN_FREE | Current free margin, in deposit currency |
| 2003 | CONDITION_EQUITY | Current account equity, in deposit currency |
| 2004 | CONDITION_BALANCE | Current balance, in deposit currency |
| 2005 | CONDITION_PROFIT | Current floating profit |
| 3000 | CONDITION_DAILY_DEALS | Number of client deals for the current and previous days (weekends/holidays included) |
| 3001 | CONDITION_DAILY_DEALS_PERIOD | Deal frequency per day, from the average time between the last 8 deals |
| 3002 | CONDITION_DAILY_PROFIT | Client profit for the current and previous days |
| 4000 | CONDITION_POSITION_VOLUME | Current position volume for the request's symbol |
| 4001 | CONDITION_POSITION_PROFIT | Current position profit for the request's symbol |
| 4002 | CONDITION_POSITION_AGE | Seconds since position opening for the request's symbol |
| 4003 | CONDITION_POSITION_MODIFY_TIME | Seconds since last position modification (volume increase, partial close, SL/TP change) |
| 4004 | CONDITION_POSITION_AVERAGE_TIME | Average position age = current time − ((open time + modification time) / 2) |
| 4005 | CONDITION_POSITION_TOTAL | Total number of the client's open positions across all symbols |
| 4006 | CONDITION_POSITION_TOTAL_SYMBOL | Number of positions on the request's symbol |
| 4007 | CONDITION_ORDER_TOTAL | Total number of the client's orders across all symbols, including history |
| 4008 | CONDITION_ORDER_TOTAL_SYMBOL | Number of orders (active and history) on the request's symbol |
| 4009 | CONDITION_POSITION_SL_TOUCHED | Boolean — market price touched a position's stop loss (may not have activated yet) |
| 4010 | CONDITION_POSITION_TP_TOUCHED | Boolean — market price touched a position's take profit |
| 4011 | CONDITION_ORDER_SL_TOUCHED | Boolean — market price touched a pending order's stop loss. Combined with REQUEST_ACTIVATE, detects simultaneous trigger and SL breach |
| 4012 | CONDITION_ORDER_TP_TOUCHED | Boolean — market price touched a pending order's take profit. Combined with REQUEST_ACTIVATE, detects simultaneous trigger and TP breach |

## EnConditionRule

Comparison operator used by a routing condition.

| Value | Name | Meaning |
|---|---|---|
| 0 | RULE_EQ | Equal |
| 1 | RULE_NOT_EQ | Not equal |
| 2 | RULE_GREATER | Greater than |
| 3 | RULE_NOT_LESS | Not less than (≥) |
| 4 | RULE_LESS | Less than |
| 5 | RULE_NOT_GREATER | Not greater than (≤) |

## Routing rule state (`mt5_routing.Mode`)

| Value | Name | Meaning |
|---|---|---|
| 0 | (disabled) | The routing rule is disabled |
| 1 | (enabled) | The routing rule is enabled |

> Documented inline on `sql_mt5_routing.htm` without a symbolic enum name.

## Routing action value type (`mt5_routing.ActionValueType`)

Selects which typed column holds the action's parameter.

| Value | Name | Meaning |
|---|---|---|
| 0 | — | The parameter has no value (e.g. `ACTION_CLEAR_TP`) |
| 1 | — | Value is in `ActionValueString`, type string |
| 2 | — | Value is in `ActionValueInt`, type int |
| 3 | — | Value is in the UInt field, type uint |
| 4 | — | Value is in `ActionValueFloat`, type float |

> Documented inline on `sql_mt5_routing.htm` without symbolic enum names.

---

# Leverage (Floating Margin)

## EnRangeMode

Level type for a floating-leverage rule (`IMTConLeverageRule::EnRangeMode`). Determines what
quantity the rule's tiers are measured against.

| Value | Name | Meaning |
|---|---|---|
| 0 | RANGE_VOLUME | Total volume of open positions across all instruments in the rule's `Path` |
| 1 | RANGE_VOLUME_PER_SYMBOL | Volume of open positions for each individual instrument in `Path` |
| 2 | RANGE_VALUE | Total value of open positions across all instruments in `Path` |
| 3 | RANGE_VALUE_PER_SYMBOL | Value of open positions for each individual symbol in `Path` |

---

# Unverified / Not Found

The following enum names were requested but **could not be verified from any available source**.
No values are given for them, and none should be guessed.

- **EnUsersLoginFlags** — no page in either document set defines this name, and no
  `LOGIN_FLAG*` constants appear anywhere in the corpus. The nearest documented concept is
  `EnUsersConnectionTypes` (connection/terminal type, above), which is a plain enum rather than
  a flags set. It is possible the intended enum is `EnUsersConnectionTypes` under an
  alternate name, but that identification is not confirmed.
- **EnClientRights** — not defined in `IMTClient`, which exposes only the descriptive
  enumerations catalogued above (type, status, gender, employment, education, wealth source,
  preferred communication, trading experience, origin, KYC status). Per-account permissions
  live on `IMTUser::EnUsersRights` (above), not on the client record.

Two further partial gaps, flagged inline above and repeated here so they are not lost:

- **EnUsersRights bit `0x1000`** — undocumented on both the `IMTUser` and `mt5_users` pages.
  It sits between `USER_RIGHT_OTP_ENABLED` (0x800) and `USER_RIGHT_SPONSORED_HOSTING` (0x2000).
  Treat as reserved, not free.
- **EnClientOrigin `CLIENT_ORIGIN_REAL`** — the member is documented and is stated to be the
  last member, but its numeric cell is blank in the source table. By position it should be 4;
  this is not confirmed.
- **EnCommReasonFlags `0x80` / `0x100`** — present on the SQL-export page with descriptions but
  no symbolic names; absent from the `IMTConCommission` API page.
