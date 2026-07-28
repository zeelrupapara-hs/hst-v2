# MT5 SQL Export — Execution & Data Feed Domain

Complete field catalog extracted from the MetaTrader 5 SQL Export documentation
(`Meta Document - SQL/sql_mt5_*.htm`). Covers order routing, gateways, data feeds,
and price-history synchronization.

---

## mt5_routing

Trade request **routing rules** — each row is one named rule with a match scope
(request types + order types) and a single typed action.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Name | String | PK | — | The name of a routing rule. Referenced by `mt5_routing_conds.Name` and `mt5_routing_dealers.RoutingName`. |
| Timestamp | Integer | — | — | A unique value within the table. Used by MetaTrader 5 servers for internal purposes. If the Timestamp of a record has changed, it means the record has been changed. |
| Mode | Integer | — | — | The state of a routing rule: 0 — disabled, 1 — enabled. |
| Request | Integer | — | EnRouteFlags | Types of requests for which the rule is applicable. Passed as a sum of flags. |
| Type | Integer | — | EnTypeFlags | Types of orders for which the rule is applicable. Passed as a sum of flags. |
| Flags | Integer | — | — | Currently not used. |
| ActionType | Integer | — | — | Discriminator: the value type for `Action`. Determines which `ActionValue*` column holds the value (see "Typed-value pattern" below). |
| Action | Integer | — | EnRouteAction | The type of action applied to a request in accordance with the rule. |
| ActionValueInt | Integer | — | — | An int value for the action applied to the rule. Example: for the rule "pass to online dealers", the value `0` means the additional option "skip this rule if no dealers online" is **disabled**. |
| ActionValueUInt | Integer | — | — | A uint value for the action applied to the rule. |
| ActionValueFloat | Fraction | — | — | A float value for the action applied to the rule. |
| ActionValueString | String | — | — | A string value for the action applied to the rule. |
| Routing_index | Integer | — | — | The index number of the configuration in the list, starting from 0. |

### Typed-value pattern (`ActionType`)

`ActionType` is a discriminator selecting exactly one of the four `ActionValue*`
columns. All others are meaningless for that row.

| ActionType | Holding column | Native type | Notes |
|---|---|---|---|
| 0 | *(none)* | — | The current parameter does not have values (e.g. `ACTION_CLEAR_TP`). |
| 1 | ActionValueString | string | |
| 2 | ActionValueInt | int | |
| 3 | ActionValueUInt | uint | Documented as "the value is located in the UInt field". |
| 4 | ActionValueFloat | float | |

Note the non-obvious ordering: **1 = String, 2 = Int, 3 = UInt, 4 = Float** — it is
*not* Int-first. Readers must branch on `ActionType` before touching any value column.

Documented example: for the "pass to online dealers" action, `ActionType = 2` and
`ActionValueInt = 0` means the sub-option "skip this rule if no dealers online" is
disabled; a non-zero int would enable it.

---

## mt5_routing_conds

**Additional conditions** attached to a routing rule. A rule fires only when all of
its condition rows evaluate true. Uses the same typed-value pattern, keyed on `Type`.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Condition_ID | Integer | PK | — | A unique identifier of the condition. |
| Name | String | FK → mt5_routing.Name | — | Name of the routing rule to which the additional condition applies. |
| Condition | Integer | — | EnRouteCondition | Type of the additional condition for the rule. |
| Rule | Integer | — | EnConditionRule | A method for comparing a condition with the specified value. |
| Type | Integer | — | — | Discriminator: type of the condition value. Determines which `Value*` column holds it (see below). |
| ValueInt | Integer | — | — | An int value for the condition. |
| ValueUInt | Integer | — | — | A uint value for the condition. |
| ValueUInt | Integer | — | — | A uint value for the condition. **Used for volume with extended accuracy.** (The source documentation lists `ValueUInt` twice — a second, extended-accuracy uint column with the same printed name; treat as `ValueUInt`/`ValueUIntExt` when modelling.) |
| ValueFloat | Float | — | — | A float value for the condition. |
| ValueString | String | — | — | A string value for the condition. For example, for the "Symbols" condition, the names of symbols (or symbol groups) are specified here. |

### Typed-value pattern (`Type`)

| Type | Holding column | Native type | Notes |
|---|---|---|---|
| 0 | *(none)* | — | Current parameter does not have values (currently not used). |
| 1 | ValueString | string | |
| 2 | ValueInt | int | |
| 3 | ValueUInt | uint | |
| 4 | ValueFloat | float | |

Identical encoding to `mt5_routing.ActionType` (1=String, 2=Int, 3=UInt, 4=Float),
but selecting from the `Value*` family instead of `ActionValue*`.

---

## mt5_routing_dealers

**Execution destinations** for a routing rule — the dealers and/or gateways to whom
matching trade requests are forwarded. The routing rule owns its destination list.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | — | — | The login of a dealer **OR** a gateway ID. A single numeric column carries both kinds of destination: a manager/dealer account login, or the `ID` of a row in `mt5_gateways`. Disambiguate by checking whether the value matches a known `mt5_gateways.ID`. |
| RoutingName | String | FK → mt5_routing.Name | — | The name of the routing rule in which the dealer/gateway is specified. |
| Name | String | — | — | The name of the dealer/gateway. |

Because this table hangs off `RoutingName`, the routing rule — not the gateway or the
dealer — owns its execution destination list; the same gateway may appear under many rules.

---

## mt5_gateways

**Gateway settings**: bridges between the MT5 trade/history servers and an external
trading system, plus live per-session runtime statistics.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Name | String | PK | — | Gateway configuration name. Referenced by `mt5_gateway_params.GatewayName` and `mt5_gateway_translates.GatewayName`. |
| Timestamp | Integer | — | — | A unique value within the table. Used by MetaTrader 5 servers for internal purposes. A changed Timestamp means the record has been changed. |
| Module | String | — | — | Gateway module name. |
| GatewayServer | *(unspecified)* | — | — | The address at which the gateway accepts connections from the history and trade servers. |
| TradingServer | *(unspecified)* | — | — | Address of the server to which the gateway connects. |
| Enable | Integer | — | — | Gateway operation mode: 0 — disabled, 1 — enabled. |
| Flags | Integer | — | gateway flag bitmask | Gateway operation flags (see bit table below). |
| Gateway | String | — | — | Gateway module name. |
| TimeoutReconnect | Integer | — | — | Timeout between attempts to reconnect to an external server, in seconds. |
| TimeoutSleep | Integer | — | — | Timeout between the series of reconnections to an external server, in seconds. |
| AttempsSleep | Integer | — | — | Number of attempts in a series of reconnections to an external server. (Spelling `Attemps` is as documented.) |
| ID | Integer | — | — | The gateway ID. Referenced by `mt5_routing_dealers.Login` when that row denotes a gateway. |
| Symbols | Array | — | — | The list of symbols for which the gateway provides quotes and processes trading operations. |
| SysConnection | Integer | — | — | The state of gateway connection to an external trading system: 0 — connected, 1 — disconnected. (Note: inverted relative to `mt5_feeders.SysConnection`.) |
| SysLastTime | Integer | — | — | The time of the last successful gateway connection to an external trading system, in `YYYY-MM-DD HH:MM:SS.MS` format. |
| Company | String | — | — | The company by which the gateway executable is signed. |
| Issuer | String | — | — | The certification authority that issued the certificate to the above company. |
| TickStatsCount | Integer | — | — | The number of price statistics changes received by the gateway from the external system for the current session. |
| TicksCount | Integer | — | — | The number of price changes received by the gateway from the external system for the current session. |
| BooksCount | Integer | — | — | The number of Market Depth changes received by the gateway from an external trading system for the current session. |
| TradeAverageTime | Integer | — | — | Average time spent by the gateway to process one trading operation, in milliseconds. |
| TradeRequestsCount | Integer | — | — | The number of trading operations processed by the gateway during the current session. |
| BytesReceived | Integer | — | — | Traffic volume received by the gateway during the current session. |
| BytesSent | Integer | — | — | Traffic volume sent by the gateway during the current session. |
| StateFlags | Integer | — | — | Flags of states. Currently not used. |

### `Flags` bitmask

| Bit | Meaning |
|---|---|
| 0x00000001 | Gateway works as a remote application. |
| 0x00000002 | Gateway is allowed to import symbol settings. |
| 0x00000004 | Do not broadcast quotes from the gateway in the system. |
| 0x00000008 | Gateway can manage clients' balances using `IMTExecution::TE_BALANCE_CHANGE` and `IMTExecution::TE_BALANCE_CORRECT` trade executions. |
| 0x00000010 | Extended logging — the gateway log receives additional operation data, including results of measuring trading-operation handling speed. |
| 0x00000020 | Gateway supports requesting the state of external trading system positions (from the gateway's Positions tab). |
| 0x00000040 | Collect advanced metrics related to request processing by the gateway. |
| 0x00000100 | Gateway is running in demo mode (checked against the license when the module is loaded). |
| 0x00000200 | Gateway is integrated into the platform. |

---

## mt5_gateway_params

Additional key/value settings belonging to a gateway configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ParamID | String | PK | — | The unique identifier of the parameter. |
| GatewayName | String | FK → mt5_gateways.Name | — | The name of the gateway configuration the setting applies to. |
| Type | Integer | — | param type enum | Parameter type (see below). |
| Name | String | — | — | Parameter name. |
| Value | String | — | — | Parameter value (always stored as string; interpret per `Type`). |

`Type`: 0 — string, 1 — integer, 2 — floating-point number, 3 — time, 4 — date,
5 — date and time, 6 — list of groups, 7 — list of symbols, 8 — bool, 9 — color.

---

## mt5_gateway_translates

Per-symbol **price translation settings in the gateway** — maps an external source
symbol onto a platform symbol, with bid/ask markups.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Symbol | String | PK (with GatewayName/Server) | — | The name of the symbol in the trading platform. |
| GatewayName | String | FK → mt5_gateways.Name | — | The name of the gateway configuration the setting applies to. |
| Server | Integer | — | — | The ID of the trade server for which the plugin is configured. |
| Source | String | — | — | The symbol name in the data feed to which the gateway connects. |
| BidMarkup | Integer | — | — | Correction for the Bid price received for a symbol from the data source to which the gateway connects. |
| AskMarkup | Integer | — | — | Correction for the Ask price received for a symbol from the data source to which the gateway connects. |
| Digits | Integer | — | — | The number of digits after the decimal point in the price of the symbol that receives quotes. |

Note: the source page carries a stale heading (`mt5_plugin_params`); the exported
table name is `mt5_gateway_translates`.

---

## mt5_feeders

**Data feed settings**: quote/news sources feeding the history server, plus live
per-session runtime statistics.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Name | String | PK | — | Get and set the data feed name. Referenced by `mt5_feeder_params.Feeder` and `mt5_feeder_translates.Feeder`. |
| Timestamp | Integer | — | — | A unique value within the table. Used by MetaTrader 5 servers for internal purposes. A changed Timestamp means the record has been changed. |
| Module | String | — | — | The name of the data feed module. |
| GatewayServer | String | — | — | The addresses at which the data feed will accept connections from the history server. |
| FeedServer | String | — | — | The addresses of the server to which the data feed is connected. |
| Enable | Integer | — | — | The data feed configuration status: 0 — disabled, 1 — enabled. |
| Mode | Integer | — | EnFeedersMode | Data feed operation mode, passed as a **sum of flags**. E.g. 1 = the feed receives news; 9 = receives news while working in "remote datafeed" mode. |
| Timeout | Integer | — | — | Timeout of a data feed before reconnecting. |
| TimeoutReconnect | Integer | — | — | Timeout between attempts to reconnect to the source server. |
| TimeoutSleep | Integer | — | — | Timeout between the series of reconnections to the source server. |
| AttemptsSleep | Integer | — | — | The number of attempts in the series of reconnections to the source server. |
| Symbols | String | — | — | The list of symbols for which the data feed provides quotes. |
| SysConnection | Integer | — | — | The status of the data feed connection to a source server: 0 — no connection, 1 — connected. (Note: inverted relative to `mt5_gateways.SysConnection`.) |
| SysLastTime | DateTime | — | — | The time of the last reconnection to the source server, in `YYYY-MM-DD HH:MM:SS.MSC` format. |
| Company | String | — | — | The name of the company who signed the executable file of the data feed. |
| Issuer | String | — | — | The certification authority that issued the certificate of the above company. |
| TickStatsCount | Integer | — | — | The amount of price statistics received by the data feed from an external data source during the current session. |
| TicksCount | Integer | — | — | The number of price changes received by the data feed from an external data source during the current session. |
| BooksCount | Integer | — | — | The number of Market Depth changes received by the data feed from an external data source during the current session. |
| NewsCounts | Integer | — | — | The number of news items received by the data feed from an external data source during the current session. |
| BytesReceived | Integer | — | — | The volume of traffic (in bytes) received by the data feed during the current session. |
| BytesSent | Integer | — | — | The volume of traffic (in bytes) sent by the data feed during the current session. |
| StateFlags | Integer | — | — | Flags of states. |

---

## mt5_feeder_params

Additional key/value settings belonging to a data feed configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ParamID | String | PK | — | The unique identifier of the parameter. |
| Feeder | String | FK → mt5_feeders.Name | — | The name of the data feed to which the setting applies. |
| Type | Integer | — | param type enum | Parameter type (see below). |
| Name | String | — | — | The name of the parameter. |
| Value | String | — | — | The value of the parameter (stored as string; interpret per `Type`). |

`Type`: 0 — string, 1 — integer, 2 — floating-point number, 3 — time, 4 — date,
5 — date and time, 6 — list of groups, 7 — list of symbols, 8 — bool, 9 — color.
Identical encoding to `mt5_gateway_params.Type`.

---

## mt5_feeder_translates

Per-symbol **conversion settings of data feeds** — maps a source-server symbol onto a
platform symbol, with bid/ask markups.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Symbol | String | PK (with Feeder) | — | The name of the symbol in the trading platform. |
| Feeder | String | FK → mt5_feeders.Name | — | The name of the data feed to which the conversion setting applies. |
| Source | String | — | — | The name of the symbol on the source server. |
| BidMarkup | Integer | — | — | Markup for the symbol's Bid price received from the data feed. |
| AskMarkup | Integer | — | — | Markup for the symbol's Ask price received from the data feed. |
| Digits | Integer | — | — | The number of digits after the decimal point in the price of the symbol that receives quotes. |

---

## mt5_history_sync

**Price history synchronization** settings — one configuration per external server
from which historical data is pulled.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Server_ID | Integer | PK | — | Primary key. A unique configuration identifier for efficient retrieval from the database; assigned automatically upon export. Referenced by `mt5_history_sync_symbols.Server_ID`. |
| Timestamp | Integer | — | — | A unique value within the table, used by MetaTrader 5 servers for internal purposes. A changed Timestamp means the record has been modified. |
| Enable | Integer | — | — | Synchronization mode: 0 — disabled, 1 — enabled. |
| Server | String | — | — | The IP address or the domain name of the server with which historical data is synchronized. |
| ServerType | Integer | — | EnHistorySyncServer | The type of the server with which historical data is synchronized. |
| Mode | Integer | — | EnHistorySyncMode | History data synchronization mode. |
| From | Integer | — | — | The beginning date of the period for which historical data is synchronized, in seconds elapsed since 01/01/1970. |
| To | Integer | — | — | The ending date of the period for which historical data is synchronized, in seconds elapsed since 01/01/1970. |
| TimeCorrect | Integer | — | — | Time zone correction for the synchronization server relative to the platform's time zone, in minutes. Positive and negative values allowed; 0 means automatic time zone correction. |
| Flags | Integer | — | EnHistorySyncFlags | Data synchronization flags. |
| Data | Integer | — | EnHistoryData | Data types for synchronization. |
| Login | Integer | — | — | The trading account login used to connect to the server with which data is synchronized. |

---

## mt5_history_sync_symbols

Symbol list attached to a price history synchronization configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Symbol_ID | Integer | PK | — | Primary key. A unique record identifier for efficient retrieval from the database; assigned automatically upon export. |
| Server_ID | Integer | FK → mt5_history_sync.Server_ID | — | Identifier of the configuration to which the symbol belongs. |
| Path | String | — | — | Path to the symbol. |

---

## Relationship summary

```
mt5_routing.Name ──< mt5_routing_conds.Name
mt5_routing.Name ──< mt5_routing_dealers.RoutingName
mt5_routing_dealers.Login ──> dealer login  OR  mt5_gateways.ID

mt5_gateways.Name ──< mt5_gateway_params.GatewayName
mt5_gateways.Name ──< mt5_gateway_translates.GatewayName

mt5_feeders.Name  ──< mt5_feeder_params.Feeder
mt5_feeders.Name  ──< mt5_feeder_translates.Feeder

mt5_history_sync.Server_ID ──< mt5_history_sync_symbols.Server_ID
```
