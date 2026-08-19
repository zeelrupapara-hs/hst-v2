# hst-quote

Live quote ingestion for hst-v2. Connects to configured FIX feeds, applies symbol translates/markups, caches latest ticks in Redis, stores tick history in InfluxDB, and publishes on NATS.

MetaTrader5Feeder (copy quotes from another MT5 server) is intentionally **not** supported.

## Phase 1 scope

- **FIX 4.3 / 4.4** connector (QuickFIX Go initiator)
- Config from `hst-server` internal API + NATS config snapshots (no Postgres)
- NATS publish: `hstquote.tick.{symbol}`
- InfluxDB tick history (shared `marketwatch` bucket, measurement per platform symbol ID)
- NATS subscribe: `hstserver.datafeed.{created,updated,deleted}`

## Run locally

```sh
cp .env.example .env
make up
make run
```

## Configuration: feed fields vs params vs translates

hst-quote reads config from hst-server. Three layers matter:

| Layer | Where | Purpose |
|-------|--------|---------|
| **Feed row** | `POST /api/v1/datafeeds` | Module, enable/mode, connection host/credentials |
| **Params** | `POST /api/v1/datafeeds/{id}/params` | FIX session tuning (Comp IDs, dialect, depth, …) |
| **Feed symbols** | `POST /api/v1/datafeeds/{id}/symbols` | MT5 Symbols tab — explicit symbol or path mask rows |
| **Translates** | `POST /api/v1/datafeeds/{id}/translates` | LP symbol → platform symbol + markups (prefer `symbol_id`) |

Params are stored in `hst.datafeed_params`. Each row needs **`param_key`** and **`value`**. `type` defaults to `0` (string) if omitted. Rows include **`priority`** (0-based order **within that datafeed only** — feed 1 and feed 2 each have their own 0..n-1 sequence; the DB enforces `UNIQUE (datafeed_id, priority)`). `GET` returns params sorted by priority. Reorder via **`PATCH /api/v1/datafeeds/{id}/params/{paramId}`** with `{ "priority": N }` — always scoped to the `{id}` datafeed in the URL; the server swaps with the row already at that priority on the same feed (frontend Up/Down sends one PATCH per click).

```json
POST /api/v1/datafeeds/{id}/params
{
  "param_key": "SenderCompID",
  "value": "HS2"
}
```

New params append at the bottom (`priority = max + 1`). Deleting a param compacts priorities. The worker config snapshot returns params in priority order.

After any param or translate change, hst-server publishes a NATS config snapshot; hst-quote reloads the feed automatically. You still need **`POST …/activate`** once when the feed is first created.

### Feed row fields (FIX)

Supported **`module`** values depend on **`mode`**. Use `GET /api/v1/datafeeds/modules?mode=1` (quotes) or `?mode=2` (news) for the allowed list. Creating or updating a feed with an unsupported module (e.g. `MetaTrader5Feeder.exe`) returns **400**.

| `mode` | Supported modules |
|--------|-------------------|
| `1` (quotes) | `fix44`, `fix43`, `dde` |
| `2` (news) | `RSSNewsFeeder` |

| Field | Example | Used for |
|-------|---------|----------|
| `module` | `fix44`, `fix43` | Connector type + dialect defaults |
| `mode` | `1` | Quotes flag (`FeederFlags_quotes`) |
| `enable` | `1` | Must be enabled |
| `feed_server` | `fixsim.hstrader.com:6002` | `host:port` → QuickFIX `SocketConnectHost/Port` |
| `feed_password` | `HSFIXPass` | FIX logon password (MsgType A) |
| `timeout_reconnect` | `5` | QuickFIX `ReconnectInterval` (seconds) |

`feed_server` can also be a path ending in `.cfg` to use an existing QuickFIX config file instead of auto-generating one.

Optional aliases on the feed row: none for username — use param **`Feed login`** (MT5 naming) or param **`Username`**.

### FIX params (`param_key` → hst-quote)

All keys are case-sensitive in the DB, but hst-quote also accepts lowercase aliases where noted.

| Param key | Aliases | Default | Notes |
|-----------|---------|---------|-------|
| `SenderCompID` | `sendercompid` | — | **Required** unless `FixConfigPath` / `.cfg` feed_server |
| `TargetCompID` | `targetcompid` | — | **Required** unless using external cfg |
| `BeginString` | — | from `module` | e.g. `FIX.4.4`, `FIX.4.3`; overrides dialect |
| `MDUpdateType` | `md_update_type` | from `module` | `0`/`FULL_REFRESH` or `1`/`INCREMENTAL_REFRESH` |
| `MarketDepth` | `market_depth` | `1` | Sent on Market Data Request |
| `Feed login` | `Username`, `username` | — | FIX logon username |
| `Password` | `password` | — | Overrides `feed_password` if set |
| `FixConfigPath` | — | — | Absolute path to existing QuickFIX `.cfg` |
| `SocketConnectHost` | `host` | from `feed_server` | Override host |
| `SocketConnectPort` | `port` | from `feed_server` | Override port |
| `HeartBtInt` | — | `30` | Heartbeat interval (seconds) |
| `ResetOnLogon` | — | `Y` | QuickFIX session flag |
| `ReconnectInterval` | `timeout_reconnect` | `5` | Override reconnect seconds |

Module defaults when params are omitted:

| Module | BeginString | MDUpdateType |
|--------|-------------|----------------|
| `fix44`, `fix`, `FIX44`, … | `FIX.4.4` | `INCREMENTAL_REFRESH` |
| `fix43`, `fix_43`, `fix4.3` | `FIX.4.3` | `FULL_REFRESH` |

Auto-generated QuickFIX cfg is written to `{FIX_CONFIG_DIR}/feed_{datafeed_id}/session.cfg` (default `logs/fix/`).

### DDE module (`module = dde`)

A port of the v1 vfxmarket currency-server client: plain TCP, one raw auth-token write on connect (no ACK), then a push stream of `!`-terminated `MRKTDATAs?<G)_` frames — comma-separated, 16 fields per symbol. The server pushes **all** symbols; there is no subscription, so translates decide which source symbols become platform ticks (unmapped ones are dropped).

Feed row: `feed_server` = `host:port`; `feed_login`/`feed_password` are combined into the v1 handshake token (`&&fBI…?<G)_{user}G)_{pass}G)_+3!` with `)`,`_`,`G` percent-escaped). The connector reports **Connected only when the first frame arrives** — the server never acknowledges the token, so data is the only proof of auth.

| Param key | Aliases | Default | Notes |
|-----------|---------|---------|-------|
| `Token` | `token` | — | Raw pre-built handshake token; skips `feed_login`/`feed_password` |
| `Feed login` | `Username`, `username` | `feed_login` | Token username |
| `Password` | `password` | `feed_password` | Token password |
| `SocketConnectHost` | `host` | from `feed_server` | Override host |
| `SocketConnectPort` | `port` | from `feed_server` | Override port |
| `ReconnectInterval` | `timeout_reconnect` | `5` | Reconnect delay (seconds) |
| `FirstFrameTimeout` | — | `30` | Seconds to wait for the first frame after auth before reconnecting |
| `IdleTimeout` | — | `90` | Seconds without data before the stream is considered dead |
| `VolumeIndex` | — | off | Chunk field index (0–15) to read volume from |

Known chunk fields: 0=source symbol, 1=bid, 2=ask, 3=high, 4=low, 6=open, 8=close. The rest of the 16-field layout is undocumented (v1 published field 8 as both close and volume); volume stays `0` unless `VolumeIndex` names a field.

### Translates (not params)

| Field | Example | Notes |
|-------|---------|-------|
| `symbol_id` | `42` | Preferred stable key from `hst.symbols` |
| `symbol` | `EURUSD` | Alternative lookup; stored denormalized on the translate row |
| `source` | `EURUSD` or `EUR/USD` | LP symbol subscribed on FIX / matched on wire |
| `bid_markup` | `0` | Added in points (× 10^-digits) |
| `ask_markup` | `0` | Subtracted in points |
| `digits` | `5` | Price rounding (default 5) |

Preview mask expansion: `GET /api/v1/datafeeds/{id}/symbols/resolve` returns `{ "count": N, "symbols": [...] }`.

Symbol scope is configured via `feed_symbols` rows only (explicit symbol or path mask). Example:

```json
POST /api/v1/datafeeds/{id}/symbols
{ "path": "Forex\\*" }
```

`GET /api/v1/datafeeds/{id}` returns configured rows (`params`, `feed_symbols`, `translates`) — same pattern as the other tabs.

### Quote sessions

Quote session windows come from `hst.symbols_sessions` (type `quote`) in the worker config snapshot. hst-quote **always** writes Redis cache and Influx; **NATS tick publish** is gated when the quote session is closed for that `symbol_id` (vfxmarket-style).

## FIX example

Supported modules: `fix`, `fix44`, `fix43`, `FIXFeeder`, `QuickFIXFeeder`, etc.

### Example: Hybrid Solutions FIX simulator (vfxmarket-style)

Create feed:

```json
POST /api/v1/datafeeds
{
  "name": "fix-sim",
  "module": "fix44",
  "mode": 1,
  "enable": 1,
  "feed_server": "fixsim.hstrader.com:6002",
  "feed_password": "HSFIXPass"
}
```

Add params (repeat per row):

```json
{ "param_key": "SenderCompID", "value": "HS2" }
{ "param_key": "TargetCompID", "value": "HSSRVR" }
{ "param_key": "Feed login",   "value": "HSFIXUser" }
```

Add translate (LP symbol must match what the simulator sends):

```json
POST /api/v1/datafeeds/{id}/translates
{
  "symbol": "EURUSD",
  "source": "EURUSD",
  "bid_markup": 0,
  "ask_markup": 0,
  "digits": 5
}
```

Activate: `POST /api/v1/datafeeds/{id}/activate`

Confirm in hst-quote logs: `fix session logged on`. FIX wire log: `logs/fix/feed_{id}/fixlog/`.

### Example: real LP

Same shape — set `feed_server`, `feed_password`, Comp IDs, and translates from the LP’s symbol list. Ask the LP for sandbox `SenderCompID`, `TargetCompID`, host, port, username, and password.

## NATS subjects

| Subject | Direction |
|---------|-----------|
| `hstserver.datafeed.*` | hst-server → hst-quote (events + config snapshots) |
| `hstquote.status.{id}` | hst-quote → hst-server |
| `hstquote.tick.{symbol}` | hst-quote → consumers |

## Redis keys

| Key | Purpose |
|-----|---------|
| `hstquote:tick:{datafeed_id}:{symbol}` | Latest tick JSON (5m TTL) |

## InfluxDB tick history

Set `INFLUX_ENABLED=true` (see `.env.example`). Local stack includes InfluxDB 2.x on port `8086`.

| Setting | Default |
|---------|---------|
| `INFLUX_URL` | `http://localhost:8086` |
| `INFLUX_ORG` | `HybridSolutions` |
| `INFLUX_BUCKET` | `marketwatch` |

Each tick is written as:

| Element | Value |
|---------|--------|
| Measurement | Platform symbol ID (`42`, …) |
| Tags | none |
| Fields | `bid`, `ask`, `high`, `low`, `open`, `close`, `volume` |
| Timestamp | Tick time |

Multiple pods share the same bucket; each pod writes symbols from feeds it owns (`QUOTE_INSTANCE_INDEX/COUNT`).

Example Flux query:

```flux
from(bucket: "marketwatch")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "42")
```
