# MT5 SQL Export — Automation, Plugins, Reports, Subscriptions & Streaming

Source: MetaTrader 5 Help > Platform Components > Backup Server > SQL Export.

Notes:
- The streaming domain is explicitly **Kafka**: `mt5_streamings.Login` is documented as
  "Username for connecting to the **Kafka cluster**", and `Prefix` is "Prefix added to the
  full topic name". Streaming configs fan out to groups, symbols, and topics; each topic
  fans out to `mt5_streaming_topic_data` rows describing which data types flow on it.
- **Typed-value discriminator pattern** appears twice (mirroring the routing domain):
  - `mt5_automation_conditions`: `Condition` + `Rule` select semantics, and exactly one of
    `ValueString` / `ValueInt` / `ValueUInt` / `ValueFloat` carries the payload.
  - `mt5_automation_params`: `Param` is the discriminator, payload in
    `ParamInt` / `ParamUInt` / `ParamFloat`.
  - `mt5_plugin_params` / `mt5_report_params` use a different variant: a `Type` code
    (0–9) plus a single stringified `Value` column.

---

## mt5_automations

Automation tasks (trigger + schedule definition). Root of the automation hierarchy.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ID | Integer | PK | | Unique configuration identifier, assigned automatically upon export. |
| Timestamp | Integer | | | Unique value within the table; internal MT5 use. Change means record modified. |
| ParentID | Integer | | | ID of the subdirectory in which the automation task is located. |
| Name | String | | | Name of the automation task. |
| Flags | Integer | | EnFlags | Additional automation task settings. |
| Trigger | Integer | | EnTriggers | Platform event that, when fired, executes the automation task. |
| TimeStart | Integer | | | Date/time from which the task trigger is checked (seconds since 01/01/1970). |
| TimeExpire | Integer | | | Date/time until which the task trigger is checked (seconds since 01/01/1970). |
| TimeWeekdays | Integer | | EnTriggerWeekdays | Days of the week on which the trigger may activate. |
| TimeMonths | Integer | | EnTriggerMonths | Months in which the trigger may activate. |
| TimeMonthdays | Integer | | EnTriggerMonthDays | Days of the month on which the trigger may activate. |
| EventPauseMinutes | Integer | | | Minutes in the period over which repetitions are counted. |
| EventPauseHours | Integer | | | Hours in the period over which repetitions are counted. |
| EventPauseDays | Integer | | | Days in the period over which repetitions are counted. |
| EventRepeats | Integer | | | Maximum number of event repetitions. |

## mt5_automation_conditions

Conditions gating an automation task. Typed-value discriminator table.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Condition_ID | Integer | PK | | Unique configuration identifier, assigned automatically upon export. |
| Automation_ID | Integer | FK → mt5_automations.ID | | Automation task to which the condition belongs. |
| Condition | Integer | | EnConditions | Automation task triggering condition (discriminator). |
| Rule | Integer | | EnConditionRule | Method for comparing the condition against the specified value. |
| ValueString | String | | | Condition value of type string. |
| ValueInt | Integer | | | Condition value of type INT. |
| ValueUInt | Integer | | | Condition value of type UINT. |
| ValueFloat | Float | | | Condition value of type double. |
| OrId | Integer | | | Identifier of the "OR" condition group. Same-type conditions sharing an OrId form one OR block. |

## mt5_automation_actions

Actions executed when an automation task fires.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Action_ID | Integer | PK | | Unique configuration identifier, assigned automatically upon export. |
| Automation_ID | Integer | FK → mt5_automations.ID | | Automation task to which the action belongs. |
| Action | Integer | | EnActions | The action performed when the automation task is triggered. |
| Name | String | | | The name of the action in the automation task. |

## mt5_automation_params

Parameters of an automation action. Typed-value discriminator table (parent is the **action**, not the automation).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Param_ID | Integer | PK | | Unique configuration identifier, assigned automatically upon export. |
| Action_ID | Integer | FK → mt5_automation_actions.Action_ID | | Automation action to which the parameter belongs. |
| Param | Integer | | EnParams | Parameter type for the automation action (discriminator). |
| ParamInt | Integer | | | Parameter value of type INT. |
| ParamUInt | Integer | | | Parameter value of type UINT. |
| ParamFloat | Float | | | Parameter value of type double. |

## mt5_plugins

Plugin configurations per trade server. Identity is (Name, Server) — no surrogate key.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Name | String | PK (with Server) | | The name of the plugin configuration. |
| Server | Integer | PK (with Name) | | ID of the trade server for which the plugin is configured. |
| Timestamp | Integer | | | Unique value within the table; internal MT5 use. Change means record modified. |
| Module | String | | | The name of the plugin module. |
| Enable | Integer | | inline | Plugin operation mode: 0 — disabled, 1 — enabled. |
| Flags | Integer | | inline | Plugin operation flags: 0 — no flags; 1 — permission to configure the plugin from a manager terminal; 2 — profiling mode enabled. |

## mt5_plugin_params

Additional plugin settings (typed name/value pairs).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ParamID | String | PK | | The unique identifier of the parameter. |
| Plugin | String | FK → mt5_plugins.Name | | Name of the plugin to which the setting applies. |
| Server | Integer | FK → mt5_plugins.Server | | ID of the trade server for which the plugin is configured. |
| Type | Integer | | inline | Parameter type: 0 — string, 1 — integer, 2 — floating-point number, 3 — time, 4 — date, 5 — date and time, 6 — list of groups, 7 — list of symbols, 8 — bool, 9 — color. |
| Name | String | | | The name of the parameter. |
| Value | String | | | The value of the parameter (stringified per `Type`). |

## mt5_reports

Report configurations per trade server. Identity is (Name, Server).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Name | String | PK (with Server) | | The name of the report configuration. |
| Server | Integer | PK (with Name) | | ID of the trade server for which the report is configured. |
| Timestamp | Integer | | | Unique value within the table; internal MT5 use. Change means record modified. |
| Module | String | | | The name of the report module. |
| Mode | Integer | | inline | Report operation mode: 0 — disabled, 1 — enabled. |

## mt5_report_params

Additional report settings (typed name/value pairs). Same shape as `mt5_plugin_params`.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ParamID | String | PK | | The unique identifier of the parameter. |
| Report | String | FK → mt5_reports.Name | | Name of the report to which the setting applies. |
| Server | Integer | FK → mt5_reports.Server | | ID of the trade server for which the report is configured. |
| Type | Integer | | inline | Parameter type: 0 — string, 1 — integer, 2 — floating-point number, 3 — time, 4 — date, 5 — date and time, 6 — list of groups, 7 — list of symbols, 8 — bool, 9 — color. |
| Name | String | | | The name of the parameter. |
| Value | String | | | The value of the parameter (stringified per `Type`). |

## mt5_subscriptions

Subscription (market-data / news entitlement) configurations. Root of the subscription hierarchy.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ID | Integer | PK | | Unique identifier of the configuration. |
| Timestamp | Integer | | | Unique value within the table; internal MT5 use. Change means record modified. |
| Type | Integer | | EnType | Configuration type. |
| Name | String | | | Subscription name. |
| URL | String | | | Link to an additional subscription description. |

## mt5_subscription_groups

Groups covered by a subscription.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Group_ID | Integer | PK | | Unique identifier of the setting. |
| Subscription_ID | Integer | FK → mt5_subscriptions.ID | | Configuration associated with this setting. |
| Group | String | | | Full path to the group. |

## mt5_subscription_symbols

Symbol/price-data entitlements of a subscription.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Symbols_ID | Integer | PK | | Unique identifier of the setting. |
| Subscription_ID | Integer | FK → mt5_subscriptions.ID | | Configuration associated with this setting. |
| Level | Integer | | EnLevel | Type of price data available by subscription. |
| Symbols | Integer | | | Path to the symbol (group of symbols) for which data is provided by subscription. |
| TickHistory | Integer | | EnTickHistory | Tick data depth available by subscription. |

## mt5_subscription_news

News entitlements of a subscription.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| News_ID | Integer | PK | | Unique identifier of the setting. |
| Subscription_ID | Integer | FK → mt5_subscriptions.ID | | Configuration associated with this setting. |
| Category | String | | | News categories available by subscription. |
| Language | Integer | | | News language available by subscription. |

## mt5_subscription_countries

Countries in which a subscription is available.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Country_ID | Integer | PK | | Unique identifier of the setting. |
| Subscription_ID | Integer | FK → mt5_subscriptions.ID | | Configuration associated with this setting. |
| Country | String | | | Two-letter country code for which the subscription is available. |

## mt5_streamings

Streaming service (Kafka) integration configurations. Root of the streaming hierarchy.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ID | Integer | PK | | Unique identifier of the configuration. |
| Timestamp | Integer | | | Unique value within the table; internal MT5 use. Change means record modified. |
| Name | String | | | Configuration name. |
| Type | Integer | | EnStreamingType | Type of streaming service. |
| Flags | Integer | | EnFlags | Additional streaming settings. |
| Address | String | | | Comma-separated list of streaming service servers (broker bootstrap list). |
| Login | String | | | Username for connecting to the **Kafka cluster**. |
| Prefix | String | | | Prefix added to the full topic name. |
| ProtocolType | Integer | | EnProtocolType | Security protocol and authentication type used to connect to the streaming service. |
| CompressionType | Integer | | EnCompressionType | Data compression type. |

## mt5_streaming_groups

Groups whose data is streamed.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Group_ID | Integer | PK | | Unique identifier of the setting. |
| Sreaming_ID | Integer | FK → mt5_streamings.ID | | Configuration associated with this setting. Note: column name is misspelled ("Sreaming_ID") in the official docs. |
| Group | String | | | Full path to the group for which data will be streamed. |

## mt5_streaming_symbols

Symbols whose data is streamed.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Symbols_ID | Integer | PK | | Unique identifier of the setting. |
| Sreaming_ID | Integer | FK → mt5_streamings.ID | | Configuration associated with this setting (docs misspelling retained). |
| Symbol | Integer | | | Path to the symbol (group of symbols) for which data will be streamed. |

## mt5_streaming_topics

(Source file: `sql_mt5_streaming_topics_enum.htm`.) Kafka topics defined for a streaming
configuration. **This page documents a regular table, not an enumeration** — it carries no
enumerated value list; the only enum it references is `EnFlags`. The published topic name is
`mt5_streamings.Prefix` + `Topic`.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Topic_ID | Integer | PK | | Unique topic identifier. |
| Sreaming_ID | Integer | FK → mt5_streamings.ID | | Configuration associated with this setting (docs misspelling retained). |
| Topic | String | | | Topic name. |
| Description | String | | | Topic description. |
| Flags | Integer | | EnFlags | Additional topic settings. |

## mt5_streaming_topic_data

Data types transmitted on a given topic.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Data_ID | Integer | PK | | Unique identifier of the setting. |
| Topic_ID | Integer | FK → mt5_streaming_topics.Topic_ID | | Topic associated with this setting. |
| Type | Integer | | EnDataTypes | Type of transmitted data. |
