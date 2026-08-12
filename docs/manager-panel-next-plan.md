# Manager panel, round two — the remaining MT5 features, planned

Grounded in the MT5 Manager documentation (`Balance Operations`, `Working with Trading
Positions/Orders`, `Bulk Operations`, `Bulk Closing`, `Quoting and Symbol Management`,
`Dealing`, `Exposure`, `Push Notifications, SMS and Mail`). Each phase names the broker
use case first, because the use case decides the shape. Online Users already shipped.

Order is by value-for-effort: integrity first, then the dealer's hands, then the desk's
automation, then the risk views.

---

## Phase 1 — Check / Fix Balance

**Use case (MT5):** the accountant's integrity audit. Every money movement is a deal, so
the stored balance must equal the sum of the account's deals; a mismatch means a bug, a
crash mid-write, or tampering. MT5 shows both values side by side ("Balance / Checked"),
turns mismatched accounts red, journals `account xxx has invalid balance: A, valid: B`,
and Fix writes the recomputed value.

**Build:**
- `GET /api/v1/balance/check` (Accountant right): one aggregate query over `hst.deals` —
  valid balance = Σ(profit + storage + fee) over all deals; valid credit = Σ(profit) over
  the credit actions (the existing `AffectsCredit` set). Compares with `hst.users`.
  Journals one warning per mismatch, MT5's wording.
- `POST /api/v1/balance/fix {login}`: applies the delta **through the engine** as a
  Correction deal (exempt from the no-money check) — never a direct DB write, so the fix
  is itself a deal, auditable, and the live stream updates every open panel.
- Accounts grid: `Balances ▸ Check Balance` runs over every account in scope; mismatched
  rows go red, Balance cell shows `stored / valid`; `Fix Balance` heals the row and the
  correction lands in the account's Balance-tab history.
- Verify: corrupt `users.balance` by SQL, check flags it, fix heals it, deal recorded.

## Phase 2 — Account Trade tab (trading on the client's behalf)

**Use case (MT5):** the client is on the phone — "close my euro position", "buy me half a
lot". The dealer trades for them from inside the account dialog: market order with
volume/SL/TP/comment, pending placement, close (full or partial), modify SL/TP, cancel.
Every operation carries the dealer's login in the deal record.

**Build (backend already complete — this is UI only):**
- The admin API already acts for a named login: `POST /orders` (market + pending),
  `POST /positions/{id}/close`, `PUT /positions/{id}` (SL/TP), `PUT /orders/{id}`,
  `POST /orders/{id}/cancel`. The engine already stamps the dealer.
- New **Trade** tab in the account dialog (manager panel, `trades_manager` right):
  left, a live bid/ask strip for the picked symbol (from `useMarketFeed`) — right, the
  order form: Symbol (tree select), Volume, SL, TP, Comment (31 chars), red **Sell** /
  blue **Buy** at the live prices; below, the account's open positions with per-row
  Close / Modify actions (reusing the dealing context-pane table).
- Verify: buy 0.01 for a trader from the dialog, position appears live everywhere,
  close half, modify SL, cancel a pending — each lands as a deal with dealer 1000.

## Phase 3 — Bulk operations

**Use case (MT5):** month-end and incident response. Credit 50 accounts a bonus from a
CSV (`login;amount`); close every position on a symbol before a delisting; strip SL/TP
across a group before a data correction. Gated hard: Dealer right for closing, and the
"confirm dangerous actions" prompt.

**Build:**
- `POST /api/v1/balance/bulk` `{operations: [{login, action, amount}], comment}`:
  loops the existing single-operation engine path, answering per-row results (done /
  refused+reason) — three journal lines per op like MT5 (queued / accepted / done).
- Bulk balance dialog on the accounts grid (multi-select rows): operation, one amount
  ("Set" to all) or per-row CSV import, comment, Process → per-row result list.
- `POST /api/v1/positions/bulk-close` `{symbol, group_mask, mode}` where mode ∈
  {close_positions, delete_orders, clear_sltp}: server selects the affected rows
  (the existing filtered queries), then drives each through the engine's close/cancel/
  modify paths. Preview first: the dialog shows the affected rows before Close.
- Verify: bulk credit two accounts from CSV; open three positions on one symbol across
  two groups, bulk-close by mask, engine fills each, journal carries all of it.

## Phase 4 — Throw-in quotes

**Use case (MT5):** the feed died or a symbol needs a manual price — the dealer opens
Quotes (F4), types Bid/Ask (arrows ±1 point, Shift 5 / Ctrl 10), Send injects the price
into the flow as if the feed spoke it. Permission "Throw in quotes".

**Build (no hst-quote changes — the transport is already ours):**
- Every consumer reads ticks from the NATS subject `hstquote.tick.<symbol>`; nothing
  cares who published. `POST /api/v1/quotes` `{symbol, bid, ask}` (right `quotes`):
  validates the symbol and digits, stamps the time, publishes the same JSON tick model
  hst-quote publishes. Engine, charts, liveness, terminals — all consume it identically.
- Market Watch: double-click a symbol (or F4) opens the small Quotes dialog — Bid/Ask
  prefilled from the last tick, ▲▼ point steppers, **Send**. Journal: type Symbols,
  "a quote was thrown in for 'EURUSD': 1.09120 / 1.09122".
- Spread/execution overrides: already served by the admin Symbols dialog (spread,
  spread balance, execution live in the symbol record) — not duplicated here.
- Verify: stop the sim feed for one symbol, throw a quote, watch it tick in Market
  Watch + trader terminal + charts; liveness turns the symbol live again.

## Phase 5 — Dealing auto-processing

**Use case (MT5):** a quiet desk auto-confirms the small stuff. Per request type, a
checkbox + max volume (0 = unlimited); only active while connected as dealer; resets on
every terminal restart (deliberately — automation must be re-armed consciously).

**Build (client-side, exactly like MT5 — it is a terminal feature):**
- DealingModule gains an **Automation** popover: per type (new order / modification /
  cancellation) an enable toggle + max-lots field; state in component memory only, so it
  resets on reload, and it only acts while on the desk — both matching MT5.
- The handler: on a queue row appearing (event or poll) with automation on and
  `volume ≤ max`, auto-`confirmRequest` at the live market price; the row flashes
  "auto" in the queue. Everything else stays manual.
- Verify: arm auto ≤ 0.05, trader sends 0.01 → fills untouched in ~2s; trader sends
  0.10 → sits in the queue with the countdown.

## Phase 6 — Exposure toolbox tab

**Use case (MT5):** the risk manager's currency book. Every position decomposes into its
two currencies (buy EURUSD = long EUR, short USD); summed across all clients this is
what the broker actually holds. Accounts in groups named `coverage*` count on the other
side (hedges at liquidity providers); Net Total is the uncovered risk, converted to one
currency.

**Build:**
- `GET /api/v1/exposure` (Risk-manager right): server-side — positions in scope joined
  with the symbol master (base/profit currency, contract size), each position split into
  ±base and ∓profit amounts; client vs coverage decided by `group LIKE 'coverage%'`;
  rates from the quote cache in Redis (`hstquote:last:*`), majors direct or inverted,
  crosses via USD. Answer: per-asset {clients, coverage, net, rate, net_usd}.
- Toolbox **Exposure** tab (manager panel, risk-manager right): the MT5 columns + the
  long/short bar per asset and the totals line; refreshed every 15s and on position
  events (the Summary tab's exact refresh pattern).
- A `coverage\lp` group seed note in the doc, since none exists yet.
- Verify: open EURUSD and USDJPY client positions, one coverage-account position;
  EUR/JPY/USD rows show clients vs coverage vs net; hand-check one conversion.

## Phase 7 — Notifications / mail to accounts

**Use case (MT5):** the desk writes to a client — a margin-call top-up demand from the
Margin Calls window, or a broadcast to a group. Internal mail lands in the terminal
mailbox; SMTP goes out only if a mail server is configured.

**Build:**
- The trader terminal already has an internal mailbox (`mails.go`, trader mail
  endpoints, `email_*` events). Add the sending side:
  `POST /api/v1/mails` `{logins | group_mask, subject, body}` (Email right) writing one
  mailbox row per recipient + the existing `email_received` event so open terminals pop
  it live; optional SMTP relay through the configured mail server + templates.
- UI: **Send Mail…** in the accounts-grid and Margin-Calls context menus — small dialog
  (To shows the selection, Subject, Body).
- Verify: mail 1001 from the Margin Calls window, the open trader terminal shows it
  arrive; group-mask send hits every account in `demo*`.

---

## Ground rules (unchanged)

Engine stays the single source of money truth — every mutation through the existing
system subjects, never direct DB writes. Scope enforced by group masks server-side.
Existing hooks (`useLiveAccounts`, `useMarketFeed`), table/dialog/context-menu
conventions, cache+listener state — no new dependencies. One phase, one commit, each
verified live with Playwright before the next. No pushes without the word; a push
auto-deploys staging.
