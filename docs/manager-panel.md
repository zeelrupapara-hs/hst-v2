# Manager Panel — how the live system works

The manager panel is the running of the platform: live money, the dealing desk, margin
calls, the market. Everything live rides one idea: **the engine already computes every
number on every tick — the panel just subscribes to the right subjects.** Nothing is
polled that can be streamed; nothing is computed twice.

## The one picture

```mermaid
flowchart LR
    SIM[hst-sim<br/>FIX acceptor] -->|FIX 4.4 ticks| QUOTE[hst-quote]
    QUOTE -->|"hstquote.tick.&lt;symbol&gt;"| NATS[(NATS)]

    NATS --> CORE[hst-core engine]
    CORE -->|"recalculates equity, margin,<br/>P/L per affected account"| CORE

    CORE -->|"websocket.accounts.&lt;login&gt;.summary<br/>(the trader's own line)"| NATS
    CORE -->|"websocket.groups.accounts.&lt;group&gt;<br/>(the SAME line, for managers)"| NATS

    NATS --> SERVER[hst-server ws hub]
    SERVER -->|"only the subjects this manager's<br/>group masks cover"| PANEL[Manager panel]

    PANEL --> GRID[Accounts grid<br/>Equity / Level / Profit]
    PANEL --> MC[Margin Calls window]
    PANEL --> POS[Positions P/L]
    PANEL --> BAL[Balance desk totals]
    PANEL --> SUM[Summary tab]
```

## Scope is the subject, not a filter

A manager socket is subscribed by the server (`hst-server ws.go subjectsFor`) to
`websocket.groups.accounts.<token>` **only for the group tokens their masks cover**
(`model.Subscriptions` + `ReduceMasks`, gated by the `acc_read` right). A manager with
`demo*` masks is physically never delivered a `real` account's frame — the isolation
happens in NATS, before any application code runs. There is no client-side filtering to
get wrong.

## The summary line — one format everywhere

The engine builds one comma line per account per tick (throttled to 500ms per account
per symbol, `hst-core/handler/publish.go SummaryFor`):

```
summary,<login>,<balance>,<credit>,<equity>,<margin>,<free>,<level>%,<profit>[,<positionId>,<positionProfit>]...
```

The trader terminal reads it for its own account; the manager panel reads the same line
for every account in scope. `Equity = Balance + Credit ± Floating P/L`,
`Free = Equity − Margin`, `Level = Equity / Margin × 100` — computed once, in the engine,
under the account's lock.

## Frontend: two hooks feed everything

```mermaid
flowchart TD
    WS[socket.js<br/>one shared websocket] -->|"summary, lines"| LA[useLiveAccounts<br/>login → money + per-position P/L]
    WS -->|tick lines| MF[useMarketFeed<br/>symbol → bid/ask/high/low/dir]
    WS -->|json events| EV[onEvent listeners]

    LA --> A[AccountsModule<br/>live columns]
    LA --> M[MarginCallsModule]
    LA --> B[BalanceOps totals line]
    LA --> P[PositionsModule P/L]
    LA --> S[SummaryPanel profit]
    LA --> D[Dealing context pane]

    MF --> W[MarketWatchModule]
    MF --> P
    MF --> D2[Dealing price prefill]
```

- `socket.js` decodes every frame: `summary,` lines → `account_summary` events, other
  CSV → `market_feed`, JSON stays JSON. `sendEvent("start_market_feed")` opens the tick
  stream (managers receive every symbol, raw).
- Both hooks are module-level caches with listeners (the `useGroups` pattern) and batch
  their notifications (100–150ms), so a burst of ticks paints once.
- The tick map mutates in place — **never memoize on it** (that froze the market watch
  once already); read it fresh each render.

## Margin calls

```mermaid
flowchart LR
    CORE[engine checkStopOut] -->|"level ≤ group margin_call"| WARN[margin_call event<br/>login + level + call_level]
    CORE -->|"level ≤ group margin_stop_out"| SO[stop_out event +<br/>forced closing]
    WARN --> PANEL[Margin Calls window]
    SO --> PANEL
    LIVE[live summary lines] --> PANEL
    PANEL -->|"row while level ≤ call,<br/>red when ≤ stop out,<br/>clears when money returns"| VIEW[View]
```

The window itself needs no request: it joins the live account stream with the group
thresholds (`margin_call` / `margin_stop_out` from `useGroups`) and lists any account
whose margin is in use with a level at or under the call line.

## Balance operations

```mermaid
flowchart LR
    UI[BalanceOps form] -->|"POST /api/v1/balance<br/>login, action, ±amount, comment"| API[hst-server]
    API -->|system.balance| CORE[engine NewBalance]
    CORE -->|"refuses if it would go negative<br/>(Correction is exempt)"| CORE
    CORE -->|deal + money_change + fresh summary| NATS[(NATS)]
    NATS --> UI2[history refreshes,<br/>totals line moves]
```

Operations: Balance, Credit, Charge, Correction, Bonus, Commission — blue button
deposits (+), red withdraws (−). Every operation lands as a **deal**, so the history is
just the account's balance-type deals.

## Dealing room

```mermaid
sequenceDiagram
    participant T as Trader
    participant C as Engine
    participant D as Dealer (manager)
    T->>C: order (routing rule says "dealer")
    C->>D: dealer_request (only to the current holder)
    Note over D: queue row + 30s countdown +<br/>client context pane (live money, positions)
    alt dealer confirms
        D->>C: confirm at price (market price prefilled)
        C->>T: filled
    else dealer requotes / rejects
        D->>C: requote(price) / reject(reason)
        C->>T: requote / refused
    else nobody answers in 30s
        C->>T: "timed out" — order canceled
    end
```

The desk is a Redis lease (90s TTL, 15s heartbeat). Only the dealers named on the
routing rule — and whose group masks cover the client — are ever offered the request.

## Where things live

| Piece | File |
|---|---|
| Group fan-out publishes | `hst-core/handler/ticks.go`, `publish.go`, `stopout.go`, `accounts.go` |
| Group-scoped subject | `hst-core/model/subjects.go SubjectGroupAccounts` |
| Frame decoding + send | `hst-admin-web/src/api/socket.js` |
| Live money cache | `hst-admin-web/src/hooks/useLiveAccounts.js` |
| Live prices cache | `hst-admin-web/src/hooks/useMarketFeed.js` |
| Accounts grid | `hst-admin-web/src/modules/accounts/AccountsModule.jsx` |
| Balance desk | `hst-admin-web/src/modules/balance/BalanceOps.jsx`, `BalanceModule.jsx` |
| Margin calls | `hst-admin-web/src/modules/margincalls/MarginCallsModule.jsx` |
| Market watch | `hst-admin-web/src/modules/market/MarketWatchModule.jsx` |
| Summary tab | `hst-admin-web/src/components/layout/SummaryPanel.jsx` |
| Dealing room | `hst-admin-web/src/modules/dealing/DealingModule.jsx` |

## Still open (the next plan)

Exposure tab (needs `coverage*` groups), throw-in quotes and spread overrides (needs a
quote-injection path in hst-quote), Online Users, Check/Fix Balance, account Trade tab
(trading on the client's behalf), bulk operations, notifications/e-mail, dealing
auto-processing.
