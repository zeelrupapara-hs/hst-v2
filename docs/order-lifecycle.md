# Order lifecycle — buy to sell, open to close, end to end

Every flow below is traced from the current code (`hst-server/internal/server/v1/orders.go`,
`hst-core/handler/{orders,validate,routing,positions,pending,ticks,stopout,dealing}.go`).
Diagrams first, words only where a diagram cannot say it.

---

## 1. The big picture

```mermaid
flowchart LR
    T[Trader terminal /\nManager Trade tab] -->|POST /orders| S[hst-server]
    S -->|validate payload\n202 Accepted| N[(NATS\nsystem.orders)]
    N --> E[hst-core engine\nowning shard]
    E -->|validate 14 gates| R{Routing rules}
    R -->|confirm| F[Fill / book]
    R -->|dealer| D[Dealing desk queue]
    R -->|reject / requote| X[Refused]
    D -->|confirm| F
    F --> DB[(hst.orders\nhst.positions\nhst.deals\nhst.accounts)]
    F --> WS[websocket events\norder/position/deal/summary]
    WS --> T
```

The server answer is always **202 Accepted** — the real outcome (`order_create`,
`order_rejected`, `position_create`, …) arrives on the account's websocket.
The server checks only the payload shape; **all trading validation lives in the engine**.

---

## 2. Market order: open (buy or sell)

```mermaid
flowchart TD
    A[system.orders → NewOrder] --> G1{account exists?}
    G1 -- no --> R1[10022 account not found]
    G1 --> G2{symbol rules for group?}
    G2 -- no --> R2[10010 unknown symbol]
    G2 --> G3{quote exists?}
    G3 -- no --> R3[10011 no price]
    G3 --> V[ValidateOrder — 14 gates, first refusal wins]
    V -- refused --> RV[retcode to the terminal]
    V --> CE{checkExecution\ninstant mode only}
    CE -- volume > MaxInstantVolume --> KIND[kind downgraded to 'request'\nnot refused]
    CE -- stale quote / slip beyond deviation --> RQ[10013 requote + bid/ask echoed]
    CE --> RT[Route — walk the rules]
    KIND --> RT
    RT -- no rule admitted --> R20[10020 not processed]
    RT -- reject rule --> R12[10012 rejected + reason]
    RT -- requote rule --> R13[10013 requote]
    RT -- dealer rule --> DQ[dealer queue → §5]
    RT -- confirm_client / confirm_market --> P{pending type?}
    P -- yes --> PL[placeOrder → on the book → §6]
    P -- no --> M2{MarginFlagCheckProcess?}
    M2 -- yes, free margin short --> R03[10003 not enough money]
    M2 --> PR[price: market if AtMarket or none,\nelse requested; normalised to digits]
    PR --> EX[Execute → netting §4]
    EX --> BF[bookFill: commission, swap settle,\nrealised P/L, margins recalc]
    BF --> SV[(one tx: order+position+deals+account)]
    SV --> EV[events: order_create, position_*,\ndeal_create, summary]
```

## 3. ValidateOrder — the 14 gates, in exact order

```mermaid
flowchart TD
    A[1 checkAccount\nenabled? trade-disabled? investor/read-only?] --> B[2 checkSymbol\ntrade mode off / close-only / side allowed]
    B --> C[3 checkMarket\nholiday + trade sessions\ndirection OUT is always open]
    C --> D[4 checkQuote\nquote fresh within QuotesTime]
    D --> E[5 checkOrderFlags\ntype allowed, SL/TP flags allowed]
    E --> F[6 checkExpert\ngroup allows expert trading]
    F --> G[7 checkVolume\nmin / max / step]
    G --> H[8 checkHedging\nhedge-prohibit groups]
    H --> I[9 checkLimits\nmax orders, max positions,\nsymbol volume cap incl. pendings,\nbook value cap, distinct symbols cap]
    I --> J[10 checkFilling\nfill policy in symbol FillFlags]
    J --> K[11 checkExpiry\ntype_time allowed, expiry in future]
    K --> L[12 checkStops\nSL/TP ≥ StopsLevel points from close price\npending price ≥ StopsLevel from open price]
    L --> M[13 checkFreeze\nexisting tickets only —\nnew orders are exempt]
    M --> N[14 checkMoney\nmargin delta vs free margin\nnetting offsets reduce the need]
    N --> OK[pass → execution check → routing]
```

| Gate | Refusal |
|---|---|
| checkAccount | 1002 disabled / 10001 trading disabled |
| checkSymbol | 10010 / 10001 / 10016 close-only |
| checkMarket | 10002 market closed |
| checkQuote | 10011 no price |
| checkOrderFlags | 10001 / 10006 stops |
| checkExpert | 10001 |
| checkVolume | 10008 invalid volume |
| checkHedging | 10015 hedging not allowed |
| checkLimits | 10007 too many orders / 10021 volume limit |
| checkFilling | 10017 fill policy |
| checkExpiry | 10018 expiry |
| checkStops | 10006 stops too close |
| checkFreeze | 10019 frozen |
| checkMoney | 10003 not enough money |

---

## 4. Fill → netting (what a fill does to the book)

```mermaid
flowchart TD
    F[Execute fill] --> T{order names a\nposition ticket?}
    T -- yes, volume < position --> RP[reduce position\nout deal, profit realised on part]
    T -- yes, volume ≥ position --> CI[close position\nout deal, profit realised]
    T -- no --> HM{hedging margin mode?}
    HM -- yes --> NP2[always a NEW position\nno netting — offset via close-by]
    HM -- no --> NET{position on symbol?}
    NET -- none --> NP[new position\nin deal]
    NET -- same direction --> GP[grow position\nweighted avg open price\nin deal]
    NET -- opposite, smaller --> RD[reduce\nout deal, partial profit]
    NET -- opposite, equal --> CL[full close\nout deal]
    NET -- opposite, larger --> RV[reverse:\nclose old with out deal,\nopen remainder the other way]
```

Deal entries: `in` opens/grows, `out` closes/reduces, `inout` marks a reversal's closing
deal, `out_by` marks both legs of a close-by. Every deal carries its `position_id`, which
is what Check Positions / Check Balance recompute from.

After every fill, `bookFill` runs in one transaction: commission → swap settle on closed
positions → realised P/L into balance (or BlockedProfit under day-P/L mode) → margins
recalculated → order/position/deal/account rows written → events published.

---

## 5. Routing rules — the walk

```mermaid
flowchart TD
    RQ[request: kind + order + account + tick] --> W[walk rules top to bottom,\nrule order = the list order]
    W --> M{rule enabled AND matches?\nrequest-kind mask, type mask,\nALL conditions AND-ed}
    M -- no --> W
    M -- soft action --> ACC[accumulate:\ndelay_time / delay_tick /\nclear_sl / clear_tp / clear_sltp] --> W
    M -- terminal --> TA{which terminal action?}
    TA -- 1005 confirm_client --> XC[execute at the requested price]
    TA -- 1006 confirm_market --> XM[execute at current market]
    TA -- 1001 dealer / 1002 dealer_online --> DL[to the desk → below]
    TA -- 1003 reject --> RJ[10012 + rule's reason text]
    TA -- 1004 requote --> RQ2[10013 requote]
    TA -- 1007 cancel_order --> CO[activation paths: delete the order\nelse 10012 order cancelled]
    W -- end of list, no terminal --> NP[10020 no rule admitted\nwhitelist semantics]
```

Rule conditions cover symbol/group masks (`!` negates, `*` globs), volume, spread, gap,
deviation, login, balance/equity/margin/level, position and order totals, time windows,
country and more — all conditions on a rule must hold at once.

### The dealer desk

```mermaid
sequenceDiagram
    participant E as Engine
    participant Q as Dealing queue
    participant D as Dealer(s)
    participant C as Client
    E->>Q: rule 1001/1002 → eligible dealers by group masks<br/>(none online + skip-flag → walk continues)
    E-->>C: 10023 queued
    Q->>D: offered to the current holder, in rule order
    alt confirm
        D->>E: price (dealer → client → market) → optional money re-check → fill
    else requote
        D->>C: new price; client accepts → fill
    else reject
        D->>C: 10012 + reason (max 31 chars)
    else return
        Q->>D: next dealer; all returned → 10024
    else timeout (30s, or symbol RequestTimeout)
        Q->>C: 10014 timed out (1s sweep)
    end
```

---

## 6. Pending orders: buy/sell limit, stop, stop-limit

### Placement

Placement runs the same 14 gates as a market order (checkStops also polices the pending
price distance, checkMoney reserves at placement), then routing; a confirming rule puts
the order on the book as `placed` — nothing fills yet.

### Trigger — on every tick

Trigger price is the **opening side**: Ask for buys, Bid for sells.

```mermaid
flowchart TD
    TK[tick] --> P{order type vs price}
    P -- "buy_limit: Ask ≤ price" --> ACT
    P -- "sell_limit: Bid ≥ price" --> ACT
    P -- "buy_stop: Ask ≥ price" --> ACT
    P -- "sell_stop: Bid ≤ price" --> ACT
    P -- "buy_stop_limit: Ask ≥ trigger" --> SLM[becomes a buy_limit\nat its limit price]
    P -- "sell_stop_limit: Bid ≤ trigger" --> SLM2[becomes a sell_limit]
    SLM --> BOOK[stays on the book,\nwaits for its limit trigger]
    SLM2 --> BOOK
    ACT[CookOrder — activation] --> E1{expired?}
    E1 -- yes --> DEL[order removed 'expired']
    E1 --> R2[Route kind=activate]
    R2 -- cancel_order rule --> DEL2[order deleted by rule]
    R2 -- not admitted --> HOLD[left on the book, logged]
    R2 -- confirmed --> M3{checkMoney again}
    M3 -- short --> DEL3[canceled, not enough money]
    M3 --> FILL[limits fill at their own price,\nstops at market → netting §4]
```

Expiry is swept first on every tick (`type_time` day/specified); a `reject` routing rule
on the expiration kind can deliberately hold an order past its time.

---

## 7. Close paths

### Manual close (full or partial)

```mermaid
flowchart TD
    C[POST close volume V] --> G1{position exists?} -- no --> N4[not found]
    G1 --> G2{quote?} -- no --> N11[10011]
    G2 --> G3{FIFO: an older same-side\nposition on the symbol?}
    G3 -- yes --> N25[10025 close the older one first]
    G3 --> VOL["V ≤ 0 or V > position ⇒ full close"]
    VOL --> CE2[checkExecution: deviation / requote]
    CE2 --> RT2[Route close kind] -- refused --> RR[by rule]
    RT2 --> PX[price = requested, or close side of market:\nBid for a long, Ask for a short]
    PX --> EX2[Execute → reduce or close → §4]
```

A close runs **no** ValidateOrder: no market-open, stops, freeze or money gate — only the
quote must exist, and direction *out* bypasses sessions and holidays by design.

### Close-by (hedging accounts only)

Two opposite positions on the same symbol settle against each other at the older leg's
open price — two `out_by` deals, min volume of the two, no FIFO or market-open check.
Requires hedging margin mode (netting groups get 10015).

### SL / TP — on every tick

Tested on the **closing side**: Bid for a long, Ask for a short.

```mermaid
flowchart TD
    TK[tick] --> H{long: Bid ≤ SL or Bid ≥ TP\nshort: Ask ≥ SL or Ask ≤ TP}
    H -- no --> Z[nothing]
    H --> F1{FIFO: oldest first?} -- blocked --> SKIP[skipped this tick]
    F1 --> F2{MarginFlagCheckSLTP:\nbook still covered after close?} -- no --> SKIP
    F2 --> RT3[Route kind=sl / tp] -- refused --> SKIP
    RT3 --> CL2["closed at market, comment [sl] / [tp]"]
```

### The tick sweep order

```mermaid
flowchart LR
    T[tick on symbol] --> A[reprice + margins + summary]
    A --> B[1 expire pendings]
    B --> C[2 SL/TP closes]
    C --> D[3 pending activations]
    D --> E[4 stop-out check — always last]
```

---

## 8. Margin call and stop out

```mermaid
flowchart TD
    T[after every tick] --> L["level = margin level %\n(or equity, in money mode)"]
    L --> S{level ≤ stop-out?}
    S -- yes --> SO[stop out]
    S --> Cq{level ≤ margin call?}
    Cq -- yes --> MC[margin_call event — warned once,\nclears after level > call × 1.05]
    Cq -- no --> OK2[clear]
    SO --> P1[1 delete margined pendings,\nbiggest reservation first,\nre-check after each]
    P1 --> P2[2 close positions,\nbiggest loser first\nFIFO: oldest per symbol,\none at a time until recovered]
    P2 --> P3["3 negative balance compensation\n(group flag): so_compensation deal\nonce flat, credit zeroed if flagged"]
    P2 --> LOG[(hst.stopout_log)]
```

Every stop-out close is itself routed (`stop_out_order` / `stop_out_position` kinds), so a
rule can exempt a book from forced closing.

---

## 9. Retcodes — the full trading set

| Code | Meaning | | Code | Meaning |
|---|---|---|---|---|
| 0 | done | | 10012 | rejected by a routing rule |
| 1 | internal error | | 10013 | requote |
| 3 | invalid request | | 10014 | timed out (dealer queue) |
| 5 | not found | | 10015 | hedging is not allowed |
| 1002 | account disabled | | 10016 | closing only |
| 10001 | trading is disabled | | 10017 | fill policy not allowed |
| 10002 | market is closed | | 10018 | expiry not allowed |
| 10003 | not enough money | | 10019 | order is frozen |
| 10006 | stops are too close | | 10020 | no routing rule admitted this request |
| 10007 | too many orders | | 10021 | position volume limit reached |
| 10008 | invalid volume | | 10022 | account not found |
| 10010 | unknown symbol | | 10023 | placed in a dealer queue |
| 10011 | no price for this symbol | | 10024 | all dealers returned the request |
| | | | 10025 | an older position must be closed first (FIFO) |

---

## 10. Three subtleties worth knowing

1. **Instant mode oversize is not refused** — a market order above `MaxInstantVolume` is
   silently reclassified as a *request* kind and walks the routing rules under that kind.
2. **New orders skip the freeze gate** — `checkFreeze` only guards changes to existing
   tickets and positions.
3. **Closing is always allowed by the clock** — direction *out* bypasses sessions and
   holidays, so a trader can always get out; only a routing rule can refuse a close.
