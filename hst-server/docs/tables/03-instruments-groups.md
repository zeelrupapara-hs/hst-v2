# MT5 SQL Export — Instruments & Group Config Domain

Field catalog extracted from the MetaTrader 5 SQL Export documentation
(`Meta Document - SQL/sql_mt5_*.htm`). Types are as documented by MetaQuotes
(Integer / Float / String / COLORREF / Array).

Domain tables: `mt5_symbols`, `mt5_symbols_sessions`, `mt5_groups`,
`mt5_groups_symbols`, `mt5_commissions`, `mt5_commissions_tiers`,
`mt5_leverages`, `mt5_leverage_rules`, `mt5_leverage_tiers`, `mt5_holidays`,
`mt5_time`, `mt5_time_weekdays`, `mt5_spreads`, `mt5_spread_legs`.

---

## mt5_symbols

Purpose: base (server-wide) configuration of every financial instrument — identity,
quoting/filtering, trade limits, margin rates, swaps, execution and option/bond data.
All group-level overrides in `mt5_groups_symbols` layer on top of these values.

### Identity & classification

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Symbol_ID | Integer | **PK** | | Unique symbol ID for efficient querying. Assigned automatically during export. |
| Timestamp | Integer | | | Unique value within the table, used internally by MT5 servers. A changed Timestamp means the record changed. |
| Symbol | String | | | Symbol name. |
| Path | String | | | Path to a symbol (symbol tree location). |
| ISIN | String | | | International Securities Identification Number of the symbol. |
| Description | String | | | Symbol description. |
| International | String | | | International symbol name. |
| Category | String | | | Category or sector name the symbol belongs to. |
| Exchange | String | | | Name of the exchange where the security is traded. |
| CFI | String | | | Instrument classification per ISO 10962. |
| Sector | Integer | | EnSectors | Economic sector the instrument belongs to. |
| Industry | Integer | | EnIndustries | Industry branch the instrument belongs to. |
| Country | String | | | Country of the company whose shares are traded. |
| Basis | String | | | Underlying asset of a derivative instrument. |
| Source | String | | | Name of the source symbol whose quotes feed this instrument. |
| Page | String | | | Web page address of the symbol. |

### Currencies & precision

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| CurrencyBase | String | | | Base currency of the symbol. |
| CurrencyBaseDigits | Integer | | | Accuracy of conversion into the base currency. |
| CurrencyProfit | String | | | Profit currency of the symbol. |
| CurrencyProfitDigits | Integer | | | Accuracy of conversion into the profit currency. |
| CurrencyMargin | String | | | Margin currency of the symbol. |
| CurrencyMarginDigits | Integer | | | Accuracy of conversion into the margin currency. |
| Color | COLORREF | | | Symbol color in the terminals' Market Watch window. |
| ColorBackground | COLORREF | | | Symbol background color in Market Watch. |
| Digits | Integer | | | Number of decimal places in the symbol price. |
| Point | Float | | | Point size, calculated as 1/10^Digits. |
| Multiply | Float | | | Multiplier converting price to points, calculated as 10^Digits. |

### Quoting, ticks & price filtering

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| TickFlags | Integer | | EnTicksFlags (bitmask) | Options for working with tick data. |
| TickBookDepth | Integer | | | Range of the Depth of Market. |
| FilterSoft | Integer | | | Soft level of price filtering. |
| FilterSoftTicks | Integer | | | Ticks counter value for soft filtering. |
| FilterHard | Integer | | | Hard level of price filtering. |
| FilterHardTicks | Integer | | | Ticks counter value for hard filtering. |
| FilterDiscard | Integer | | | Discard level of price filtering. |
| FilterSpreadMax | Integer | | | Maximum allowed spread value. |
| FilterSpreadMin | Integer | | | Minimum allowed spread. |
| SubscriptionsDelay | Integer | | | Delivery delay in minutes for quotes provided by subscription. |
| FilterGap | Integer | | | Difference between previous and next quote from which a gap is considered formed. |
| FilterGapTicks | Integer | | | Number of ticks after which gap mode is disabled if no new gap occurs. |
| TickChartMode | Integer | | EnChartMode | Chart creation mode: 0 — using Bid price, 1 — using Last price. |

### Trade mode & execution

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| TradeMode | Integer | | EnTradeMode | Symbol trading mode. |
| CalcMode | Integer | | EnCalcMode | Mode of margin and profit calculation. |
| ExecMode | Integer | | EnExecutionMode | Execution mode of the symbol. |
| GTCMode | Integer | | EnGTCMode (bitmask) | Types of orders that can be set for the symbol. |
| FillFlags | Integer | | EnFillingFlags (bitmask) | Filling types allowed for the symbol. |
| ExpirFlags | Integer | | EnExpirationFlags (bitmask) | Available order expiration types for the symbol. |
| OrderFlags | Integer | | EnOrderFlags (bitmask) | Flags of order types allowed for the symbol. |
| TradeFlags | Integer | | EnTradeFlags | Trade flags of the symbol. |

### Spread & tick pricing

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Spread | Integer | | | Symbol spread size. |
| SpreadBalance | Integer | | | Spread balance — shift from equal distribution of the spread between Bid and Ask (e.g. spread 10 as -5 Bid/+5 Ask = 0; -6/+4 = -1; -4/+6 = 1). |
| SpreadDiff | Integer | | | Symbol spread difference. Base value, effectively 0; use the group parameter for per-group spread difference. |
| SpreadDiffBalance | Integer | | | Spread difference balance. Base value, effectively 0; use the group parameter for per-group value. |
| TickValue | Float | | | Price of one tick of the symbol. |
| TickSize | Float | | | Size of one tick of the symbol. |
| ContractSize | Float | | | Contract size for the symbol. |

### Trade limits & volumes

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| StopsLevel | Integer | | | Price band within which placing stop orders is not allowed. |
| FreezeLevel | Integer | | | Price band within which modifying orders and positions is not allowed. |
| QuotesTimeout | Integer | | | Seconds to wait for quotes, after which trading is automatically disabled for the symbol. |
| VolumeMin | Integer | | | Minimum trade volume. 1 unit = 1/10000 lot. |
| VolumeMinExt | Integer | | | Minimum trade volume, extended accuracy. 1 unit = 1/100000000 lot. |
| VolumeMax | Integer | | | Maximum trade volume. 1 unit = 1/10000 lot. |
| VolumeMaxExt | Integer | | | Maximum trade volume, extended accuracy. 1 unit = 1/100000000 lot. |
| VolumeStep | Integer | | | Volume change step. 1 unit = 1/10000 lot. |
| VolumeStepExt | Integer | | | Volume change step, extended accuracy. 1 unit = 1/100000000 lot. |
| VolumeLimit | Integer | | | Maximum aggregate volume of positions and orders in one direction. 1 unit = 1/10000 lot. |
| VolumeLimitExt | Integer | | | Maximum aggregate volume in one direction, extended accuracy. 1 unit = 1/100000000 lot. |

### Margin

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| MarginFlags | Integer | | EnMarginFlags | Additional margin checking modes. |
| MarginInitial | Float | | | Size of the initial margin. |
| MarginMaintenance | Float | | | Size of the maintenance margin. |
| MarginInitialBuy | Float | | EnMarginRateTypes | Initial margin rate for market Buy orders. |
| MarginInitialSell | Float | | EnMarginRateTypes | Initial margin rate for market Sell orders. |
| MarginInitialBuyLimit | Float | | EnMarginRateTypes | Initial margin rate for Buy Limit orders. |
| MarginInitialSellLimit | Float | | EnMarginRateTypes | Initial margin rate for Sell Limit orders. |
| MarginInitialBuyStop | Float | | EnMarginRateTypes | Initial margin rate for Buy Stop orders. |
| MarginInitialSellStop | Float | | EnMarginRateTypes | Initial margin rate for Sell Stop orders. |
| MarginInitialBuyStopLimit | Float | | EnMarginRateTypes | Initial margin rate for Buy Stop Limit orders. |
| MarginInitialSellStopLimit | Float | | EnMarginRateTypes | Initial margin rate for Sell Stop Limit orders. |
| MarginMaintenanceBuy | Float | | EnMarginRateTypes | Maintenance margin rate for market Buy orders. |
| MarginMaintenanceSell | Float | | EnMarginRateTypes | Maintenance margin rate for market Sell orders. |
| MarginMaintenanceBuyLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Buy Limit orders. |
| MarginMaintenanceSellLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Sell Limit orders. |
| MarginMaintenanceBuyStop | Float | | EnMarginRateTypes | Maintenance margin rate for Buy Stop orders. |
| MarginMaintenanceSellStop | Float | | EnMarginRateTypes | Maintenance margin rate for Sell Stop orders. |
| MarginMaintenanceBuyStopLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Buy Stop Limit orders. |
| MarginMaintenanceSellStopLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Sell Stop Limit orders. |
| MarginHedged | Float | | | Hedged margin value. |
| MarginRateLiquidity | Float | | | Liquidity rate — share of the asset's current value counted as collateral in client equity. |
| MarginRateCurrency | Float | | | Margin currency rate (rate change radius of the currency the futures contract is denominated in). |

### Swaps

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| SwapMode | Integer | | EnSwapMode | Swap calculation mode for the symbol. |
| SwapLong | Float | | | Swap size for long positions. |
| SwapShort | Float | | | Swap size for short positions. |
| SwapYearDay | Integer | | EnSwapDays | Number of days in a year used when calculating swap percent. |
| SwapFlags | Integer | | EnSwapFlags | Additional swap settings. |
| SwapRateSunday | Float | | | Swap multiplier for Sundays. |
| SwapRateMonday | Float | | | Swap multiplier for Mondays. |
| SwapRateTuesday | Float | | | Swap multiplier for Tuesdays. |
| SwapRateWednesday | Float | | | Swap multiplier for Wednesdays. |
| SwapRateThursday | Float | | | Swap multiplier for Thursdays. |
| SwapRateFriday | Float | | | Swap multiplier for Fridays. |
| SwapRateSaturday | Float | | | Swap multiplier for Saturdays. |

### Trading period, request & instant execution

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| TimeStart | Integer | | | Start date of trading for the symbol. TimeStart = TimeExpiration = 0 means no time limitation. |
| TimeExpiration | Integer | | | Trading expiration date for the symbol. TimeStart = TimeExpiration = 0 means no time limitation. |
| REFlags | Integer | | EnRequestsFlags (bitmask) | Request execution flags. |
| RETimeout | Integer | | | Seconds during which the dealer-issued price in Request Execution mode stays valid. |
| IECheckMode | Integer | | EnInstantMode | Check mode for instant execution. |
| IETimeout | Integer | | | Max allowed difference between the arrival time of the order price and the time of the last price. |
| IESlipProfit | Integer | | | Max allowed slippage in the profitable direction during instant execution. |
| IESlipLosing | Integer | | | Max allowed slippage in the loss direction during instant execution. |
| IEVolumeMax | Integer | | | Max trade volume executable in instant execution mode. 1 unit = 1/10000 lot. |
| IEVolumeMaxExt | Integer | | | Max trade volume in instant execution mode, extended accuracy. 1 unit = 1/100000000 lot. |

### Prices, bonds, futures splicing & options

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| PriceSettle | Float | | | Clearing price of the previous session. |
| PriceLimitMax | Float | | | Maximum allowed price of the symbol. |
| PriceLimitMin | Float | | | Minimum allowed price of the symbol. |
| FaceValue | Float | | | Face value of a bond. |
| AccruedInterest | Float | | | Accrued interest of a bond. |
| SpliceType | Integer | | | Futures contract splicing type. |
| SpliceTimeType | Integer | | | Date of splicing of the futures contracts. |
| SpliceTimeDays | Integer | | | Offset of splicing of the futures contracts. |
| OptionMode | Integer | | | Option type/style: 0 — European call, 1 — European put, 2 — American call, 3 — American put. |
| PriceStrike | Float | | | Strike price — price at which the option gives the right to buy or sell the asset. |

---

## mt5_symbols_sessions

Purpose: trade and quoting sessions per symbol per weekday.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Session_ID | Integer | **PK** | | Unique session ID, assigned automatically during export. |
| Symbol_ID | Integer | **FK → mt5_symbols.Symbol_ID** | | ID of the symbol the session applies to. |
| Type | Integer | | | Session type: 0 — quoting, 1 — trade. |
| Day | Integer | | | Day of the week, 0–6 (0 = Sunday, 6 = Saturday). |
| Open | Integer | | | Session opening time in minutes since 00:00 (e.g. 100 = 01:40). |
| Close | Integer | | | Session closing time in minutes since 00:00. |

---

## mt5_groups

Purpose: client group configuration — permissions, company branding, reports/news/mail,
trade and margin policy, demo defaults and account limits.

### Identity

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Group_ID | Integer | **PK** | | Unique group ID, assigned automatically during export. |
| Timestamp | Integer | | | Unique value within the table, used internally by MT5 servers. Changed value = changed record. |
| Group | String | | | Group name including its hierarchical path. |
| Server | Integer | | | ID of the trade server the group is linked to. |

### Permissions & authorization

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| PermissionFlags | Integer | | EnPermissionsFlags (bitmask) | Flags of group permissions. |
| AuthMode | Integer | | EnAuthMode | Authorization mode for accounts in the group. |
| AuthPasswordMin | Integer | | | Minimum password length for accounts in the group. |

### Company & currency

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Company | String | | | Name of the company servicing the group. |
| CompanyPage | String | | | Website address of the servicing company. |
| CompanyEmail | String | | | Email address of the servicing company. |
| CompanySupportPage | String | | | Technical support website address. |
| CompanySupportEmail | String | | | Technical support email address. |
| CompanyCatalog | String | | | Subdirectory storing report/email templates for the servicing company. |
| Currency | String | | | Group deposit currency. |
| CurrencyDigits | Integer | | | Number of digits after the decimal point in the deposit currency. |

### Reports, news & mail

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ReportsMode | Integer | | EnReportsMode | Report generation mode. |
| ReportsFlags | Integer | | EnReportFlags (bitmask) | Report sending options. |
| ReportsEmail | String | | | Mail server used to send reports to clients of the group. |
| ReportsSMTP | String | | | SMTP server address for sending reports. Obsolete, not updated. |
| ReportsSMTPLogin | String | | | Login for SMTP authorization when sending reports. Obsolete, not updated. |
| NewsMode | Integer | | EnNewsMode | Mode of sending news to clients of the group. |
| NewsCategory | String | | | News categories received by the group; use "\\" for subcategories. |
| NewsLangs | Array of integers | | | Languages in which the group receives news, in MS Windows LANGID format. |
| MailMode | Integer | | EnMailMode | Operation mode of the internal mail system for the group. |

### Trading & margin policy

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| TradeFlags | Integer | | EnTradeFlags (bitmask) | Trade options of the group. |
| TradeInterestrate | Float | | | Annual interest rate on deposits of the group accounts. |
| TradeVirtualCredit | Float | | | Additional funds the broker provides to open positions larger than the client's own funds allow. |
| MarginFreeMode | Integer | | EnFreeMarginMode | Mode of using floating profit/loss in free margin. |
| MarginSOMode | Integer | | EnStopOutMode | Mode of checking Stop Out and Margin Call levels. |
| MarginCall | Float | | | Margin Call level; units determined by MarginSOMode. |
| MarginStopOut | Float | | | Stop Out level; units determined by MarginSOMode. |
| MarginFreeProfitMode | Integer | | | Mode of using profit/loss fixed during a trade day in free margin. |
| MarginMode | Integer | | EnMarginMode | Risk management model of the group. |
| MarginFlags | Integer | | EnMarginFlags | Margin calculation flags. |

### Demo defaults & limits

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| DemoLeverage | Integer | | | Default credit leverage for demo accounts opened in the group. |
| DemoDeposit | Float | | | Default deposit amount for demo accounts opened in the group. |
| LimitHistory | Integer | | EnHistoryLimit | Maximum number of days of trade history the group can request. |
| LimitOrders | Integer | | | Maximum number of orders an account of this group may have placed simultaneously. |
| LimitSymbols | Integer | | | Maximum number of symbols an account can receive quotes for simultaneously. |
| LimitPositions | Integer | | | Maximum number of simultaneously open positions per account. |
| LimitPositionsVolume | Float | | | Currently not used. |
| TradeTransferMode | Integer | | | Mode of transferring funds between accounts. |

---

## mt5_groups_symbols

Purpose: per-group symbol overrides. Each record applies to a **symbol path mask**
(`Path`) rather than a single symbol, and layers over the base `mt5_symbols` record.

**Mask / inheritance semantics**
- `Path` is a mask: a single symbol (`EURUSD`) or a group of symbols (`Forex\*`, `*`).
  Records are evaluated in `Config_Index` order (0-based position in the group's
  configuration list); later entries refine earlier ones.
- A **NULL** value in any field from `TradeMode` onward means the setting is
  **inherited from the base symbol** in `mt5_symbols` (the "Default" state in the
  Manager UI). Only non-NULL fields are actual overrides.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Symbol_ID | Integer | **PK** | | Unique symbol ID for the group. |
| Group_ID | Integer | **FK → mt5_groups.Group_ID** | | Group the symbol settings are applied to. |
| Timestamp | Integer | | | Unique value within the table, used internally by MT5 servers. Changed value = changed record. |
| Path | String | **FK (mask) → mt5_symbols.Path** | | Path/mask of the symbol or symbol group subject to these group settings. |
| TradeMode | Integer | | EnTradeMode | Symbol trading mode for the group. |
| ExecMode | Integer | | EnExecutionMode | Symbol execution mode for the group. |
| FillFlags | Integer | | EnFillingFlags (bitmask) | Filling types allowed for the symbol in this group. |
| ExpirFlags | Integer | | EnExpirationFlags (bitmask) | Order expiration types allowed for the symbol in this group. |
| SpreadDiff | Integer | | | Difference between the group's symbol spread and the default spread. |
| SpreadDiffBalance | Integer | | | Balance of spread difference — shift from equal Bid/Ask distribution (diff 4 as -2/+2 = 0; -3/+1 = -1; -1/+3 = 1). |
| StopsLevel | Integer | | | Price band within which the group may not place stop orders for the symbol. |
| FreezeLevel | Integer | | | Price band within which the group may not modify orders and positions. |
| VolumeMin | Integer | | | Minimum trade volume for the group. 1 unit = 1/10000 lot. |
| VolumeMinExt | Integer | | | Minimum trade volume, extended accuracy. 1 unit = 1/100000000 lot. |
| VolumeMax | Integer | | | Maximum trade volume for the group. 1 unit = 1/10000 lot. |
| VolumeMaxExt | Integer | | | Maximum trade volume, extended accuracy. 1 unit = 1/100000000 lot. |
| VolumeStep | Integer | | | Volume change step for the group. 1 unit = 1/10000 lot. |
| VolumeStepExt | Integer | | | Volume change step, extended accuracy. 1 unit = 1/100000000 lot. |
| VolumeLimit | Integer | | | Max aggregate volume of positions and orders in one direction for the group. 1 unit = 1/10000 lot. |
| VolumeLimitExt | Integer | | | Max aggregate volume in one direction, extended accuracy. 1 unit = 1/100000000 lot. |
| MarginFlags | Integer | | EnMarginFlags | Additional symbol margin checking modes for the group. |
| MarginInitial | Float | | | Initial symbol margin for the group. |
| MarginMaintenance | Float | | | Maintenance symbol margin for the group. |
| MarginInitialBuyLimit | Float | | EnMarginRateTypes | Initial margin rate for Buy Limit orders. |
| MarginInitialSellLimit | Float | | EnMarginRateTypes | Initial margin rate for Sell Limit orders. |
| MarginInitialBuyStop | Float | | EnMarginRateTypes | Initial margin rate for Buy Stop orders. |
| MarginInitialSellStop | Float | | EnMarginRateTypes | Initial margin rate for Sell Stop orders. |
| MarginInitialBuyStopLimit | Float | | EnMarginRateTypes | Initial margin rate for Buy Stop Limit orders. |
| MarginInitialSellStopLimit | Float | | EnMarginRateTypes | Initial margin rate for Sell Stop Limit orders. |
| MarginMaintenanceBuy | Float | | EnMarginRateTypes | Maintenance margin rate for market Buy orders. |
| MarginMaintenanceSell | Float | | EnMarginRateTypes | Maintenance margin rate for market Sell orders. |
| MarginMaintenanceBuyLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Buy Limit orders. |
| MarginMaintenanceSellLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Sell Limit orders. |
| MarginMaintenanceBuyStop | Float | | EnMarginRateTypes | Maintenance margin rate for Buy Stop orders. |
| MarginMaintenanceSellStop | Float | | EnMarginRateTypes | Maintenance margin rate for Sell Stop orders. |
| MarginMaintenanceBuyStopLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Buy Stop Limit orders. |
| MarginMaintenanceSellStopLimit | Float | | EnMarginRateTypes | Maintenance margin rate for Sell Stop Limit orders. |
| MarginCurrency | String | | | Currency margin rate. |
| MarginLiquidity | Float | | | Liquidity rate of the symbol for the group — share of asset value counted as collateral in client equity. |
| MarginHedged | Float | | | Hedged margin value. |
| SwapMode | Integer | | EnSwapMode | Swap calculation mode for the symbol in this group. |
| SwapLong | Float | | | Long position swap for the symbol in the group. |
| SwapShort | Float | | | Short position swap for the symbol in the group. |
| SwapYearDay | Integer | | EnSwapDays | Days in a year used for swap percent calculation for the group. |
| SwapFlags | Integer | | EnSwapFlags | Additional swap settings for the symbol in the group. |
| SwapRateSunday | Float | | | Sunday swap multiplier for the group. |
| SwapRateMonday | Float | | | Monday swap multiplier for the group. |
| SwapRateTuesday | Float | | | Tuesday swap multiplier for the group. |
| SwapRateWednesday | Float | | | Wednesday swap multiplier for the group. |
| SwapRateThursday | Float | | | Thursday swap multiplier for the group. |
| SwapRateFriday | Float | | | Friday swap multiplier for the group. |
| SwapRateSaturday | Float | | | Saturday swap multiplier for the group. |
| RETimeout | Integer | | | Seconds during which the dealer-issued price in Request Execution mode is valid. |
| IECheckMode | Integer | | EnInstantMode | Check mode during instant execution set for the group. |
| IETimeout | Integer | | | Max allowed difference (seconds) between arrival time of the order price and the last price time. |
| IESlipProfit | Integer | | | Max allowed slippage in the profitable direction during instant execution. |
| IESlipLosing | Integer | | | Max allowed slippage in the loss direction during instant execution. |
| IEVolumeMax | Integer | | | Max trade volume executable in instant execution mode. 1 unit = 1/10000 lot. |
| IEVolumeMaxExt | Integer | | | Max trade volume in instant execution mode, extended accuracy. 1 unit = 1/100000000 lot. |
| IEFlags | Integer | | EnInstantFlags | Instant execution flags. |
| OrderFlags | Integer | | EnOrderFlags (bitmask) | Flags of order types allowed for the symbol. |
| PermissionsFlags | Integer | | EnPermissionsFlags (bitmask) | Permission flags for the group's symbols. |
| PermissionsBookDepth | Integer | | | Number of orders in the Market Depth window allowed for this group. |
| REFlags | Integer | | EnRequestsFlags | Request execution flags for the group. |
| Config_Index | Integer | | | Index of the configuration in the group's list, starting from 0. |

---

## mt5_commissions

Purpose: commission settings attached to a group and a symbol path mask.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Commission_ID | Integer | **PK** | | Unique commission setting ID. |
| Group_ID | Integer | **FK → mt5_groups.Group_ID** | | Group the commission setting belongs to. |
| Name | String | | | Commission setting name (up to 64 characters). |
| Description | String | | | Commission setting description (up to 64 characters). |
| Path | String | **FK (mask) → mt5_symbols.Path** | | Path to a symbol or group of symbols covered by the commission setting. |
| Mode | Integer | | EnCommissionMode | Commission type: 0 — standard, 1 — agent. |
| ModeRange | Integer | | EnCommRangeMode | Commission levels type: 0 — volume, 1 — turnover in money, 2 — turnover in volume. |
| ModeCharge | Integer | | EnCommChargeMode | Commission charge mode: 0 — daily, 1 — monthly, 2 — instant. |
| TurnoverCurrency | String | | | Currency in which the money turnover is calculated. |
| ModeEntry | Integer | | EnCommEntryMode | By trade direction: 0 — all trades, 1 — entry deals only, 2 — exit deals only. |
| ModeAction | Integer | | EnCommActionMode | By trade type: 0 — all trades, 1 — Buy deals only, 2 — Sell deals only. |
| ModeProfit | Integer | | EnCommProfitMode | By deal profit: 0 — all deals, 1 — profitable only, 2 — losing only. |
| ModeReason | Integer | | EnCommReasonMode (bitmask) | By deal reason: 0x00000000 no commission; 0x01 manual via client terminal; 0x02 via Expert Advisor; 0x04 by a dealer via Manager terminal; 0x08 from an external trading system; 0x10 via mobile terminal (Android/iPhone); 0x20 via web terminal; 0x40 from copying a trading signal per subscription; 0x80 from a gateway; 0x100 received from Ultency (liquidity-provider mode). |

---

## mt5_commissions_tiers

Purpose: tiered levels (ranges) belonging to a commission setting.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Tier_ID | Integer | **PK** | | Unique commission level ID. |
| Commission_ID | Integer | **FK → mt5_commissions.Commission_ID** | | Commission setting the level belongs to. |
| Mode | Integer | | EnCommTierMode | Commission calculation unit: 0 — account currency, 1 — base currency, 2 — profit currency, 3 — margin currency, 4 — points, 5 — percentage, 6 — specified currency. |
| Type | Integer | | EnCommTierType | Commission charge type: 0 — per trade, 1 — per volume. |
| Value | Float | | | Commission sum; units depend on Mode. |
| RangeFrom | Float | | | Minimum deal volume (turnover) from which the commission is charged. |
| RangeTo | Float | | | Maximum deal volume (turnover) up to which the commission is charged. |
| Minimal | Float | | | Minimum commission amount, in the group deposit currency. |
| Currency | String | | | Commission calculation currency (used when Mode = 6, "Specified currency"). |

---

## mt5_leverages

Purpose: floating leverage configurations (headers).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Leverage_ID | Integer | **PK** | | Unique identifier of the floating leverage configuration. |
| Name | String | | | Name of the floating leverage configuration. |
| Timestamp | Integer | | | Unique value within the table, used internally by MT5 servers. Changed value = changed record. |
| Flags | Integer | | | Currently not used. |

---

## mt5_leverage_rules

Purpose: rules inside a floating leverage configuration, scoped to a symbol path mask.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Rule_ID | Integer | **PK** | | Unique identifier of the rule. |
| Leverage_ID | Integer | **FK → mt5_leverages.Leverage_ID** | | Floating leverage configuration this rule belongs to. |
| Name | String | | | Name of the rule. |
| Description | String | | | Description of the rule. |
| Path | String | **FK (mask) → mt5_symbols.Path** | | Path to a symbol or group of symbols the rule applies to. |
| RangeMode | Integer | | **EnRangeMode** | Level type for the rule (e.g. volume vs. notional-value RANGE_VALUE* modes). |
| RangeValueCurrency | String | | | Currency the notional value of positions is converted to in RANGE_VALUE* modes. |
| RangeValueCurrencyDigits | Integer | | | Number of decimal places in RangeValueCurrency. |

---

## mt5_leverage_tiers

Purpose: levels (ranges) within a floating leverage rule. **Each tier carries its own
`MarginRateInitial` and `MarginRateMaintenance`** — this is where the effective leverage
per range is expressed.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Tier_ID | Integer | **PK** | | Unique level identifier, assigned automatically upon export. |
| Rule_ID | Integer | **FK → mt5_leverage_rules.Rule_ID** | | Rule the level belongs to. |
| RangeFrom | Float | | | Minimum range value for the level. |
| RangeTo | Float | | | Maximum range value for the level. |
| MarginRateInitial | Float | | | Initial margin rate for the level. |
| MarginRateMaintenance | Float | | | Maintenance margin rate for the level. |

---

## mt5_holidays

Purpose: holiday calendar applied to symbols/symbol groups.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Year | Integer | | | Year of the holiday; 0 means a yearly (recurring) holiday. |
| Month | Integer | | | Month of the holiday (1 — January … 12 — December). |
| Day | Integer | | | Day of the holiday. |
| From | Integer | | | Holiday start time in minutes since 00:00 (600 = 10:00). |
| To | Integer | | | Holiday end time in minutes since 00:00 (1200 = 20:00). |
| Description | String | | | Holiday description (max 128 characters). |
| Timestamp | Integer | | | Unique value within the table, used internally by MT5 servers. Changed value = changed record. |
| Mode | Integer | | | Holiday mode: 0 — disabled, 1 — enabled. |
| Symbols | String | **FK (mask) → mt5_symbols.Path** | | Comma-separated list of instruments or instrument groups the holiday applies to, e.g. `EURUSD,CFD\*`. |

---

## mt5_time

Purpose: platform-wide trading time / time-zone settings.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| TimeZone | Integer | **PK** (natural key; referenced by mt5_time_weekdays) | | Server time zone in minutes from GMT (0 = GMT, -60 = GMT-1, 60 = GMT+1). |
| Timestamp | Integer | | | Unique value within the table, used internally by MT5 servers. Changed value = changed record. |
| TimeServer | String | | | Address of the current time synchronization server. |
| Daylight | Integer | | | Daylight Saving Time mode: 0 — off, 1 — on. |
| DaylightState | Integer | | | Presence of DST in the platform time zone: 0 = no DST applied; any non-zero value otherwise. |

---

## mt5_time_weekdays

Purpose: platform working-time schedule, one row per weekday, one column per hour.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| TimeZone | Integer | **FK → mt5_time.TimeZone** | | Server time zone in minutes from GMT; matches the TimeZone value in `mt5_time`. |
| Day | Integer | | | Ordinal number of the weekday (0 = Sunday, 6 = Saturday). |
| 00 | Integer | | | Working-time flag for 00:00–00:59. 0 — non-working, 1 — working. |
| 01 | Integer | | | Working-time flag for 01:00–01:59. 0 — non-working, 1 — working. |
| 02 … 22 | Integer | | | One further column per hour of the day, same 0/1 semantics. |
| 23 | Integer | | | Working-time flag for 23:00–23:59. 0 — non-working, 1 — working. |

---

## mt5_spreads

Purpose: spread (multi-leg synthetic instrument) configurations.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ID | Integer | **PK** | | Unique configuration identifier, assigned automatically upon export. |
| Timestamp | Integer | | | Unique value within the table, used internally by MT5 servers. Changed value = changed record. |
| Flags | Integer | | | Spread configuration flags. Currently not used. |
| MarginInitial | Float | | | Parameter value used to configure the initial margin. |
| MarginMaintenance | Float | | | Parameter value used to configure the maintenance margin. |
| MarginType | Integer | | EnSpreadMarginType | Margin charging type. |

---

## mt5_spread_legs

Purpose: individual legs of a spread configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Leg_ID | Integer | **PK** | | Unique record identifier, assigned automatically upon export. |
| Spread_ID | Integer | **FK → mt5_spreads.ID** | | Spread configuration the leg belongs to. |
| Side | Integer | | | Spread leg: 0 — leg A, 1 — leg B. |
| Mode | Integer | | EnLegMode | Symbol indication mode for the spread leg. |
| Flags | Integer | | | Spread leg flags. Currently not used. |
| Symbol | String | **FK → mt5_symbols.Symbol** | | Symbol or basic asset for the spread leg. |
| TimeFrom | Integer | | | Start of the period for filtering symbols by expiration time when a basic asset is specified. |
| TimeTo | Integer | | | End of the period for filtering symbols by expiration time when a basic asset is specified. |
| Ratio | Float | | | Weight of the specified symbol. |
