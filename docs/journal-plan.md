# Symbol journal: what MT5 shows, what we show, what is missing

Reference: MT5 Administrator docs, `network_journal.htm`, `journal_datafeed.htm`, `admin_symbols.htm`.
Trace: Symbols → right-click → Journal in the admin web, end to end.

## How MT5 does it

Symbols → right-click → **Journal** opens the *trade server* journal with the symbol name in the
query field. Pressing Request returns every server line that mentions the symbol, across all
subsystems, for the chosen period. The screenshot for `UKOUSD.c` returned **577 lines over 2 months**.

Request form: keyword · mode (Full / Without logins / Errors only) · event type (All, Configuration,
System, Network, History, Accounts, Trades, API …) · period (Today / Yesterday / Custom from–to) ·
Request button. Result columns: Time · IP · Message. Export CSV/HTML, Find, Copy.

Also: the administrator terminal keeps its own local journal and writes one line per request
(`'1000': 577 journal records for 'UKOUSD.c' … have been requested`).

## How ours does it today

Flow: `SymbolsModule` → `toolbox.openJournal(symbol)` → `Toolbox.jsx` calls
`GET /api/v1/journal?search=<symbol>&limit=200` → `admin/journal.go MyJournal`.

What actually comes back for EURUSD on this DB: **~80 rows, 58 of them "signed in"** (they match
because the message text is filtered by LIKE and the manager's own login rows dominate).

### Root causes, in order of damage

1. **`MyJournal` is scoped to `login = caller`** (`admin/journal.go:73`). A symbol request
   shows only what *this manager* did, never what the server, other managers, hst-core or
   hst-quote did. This alone hides most of the journal. MT5's symbol journal is server-wide.
2. **hst-core writes nothing to `hst.journal`.** Every trade event on a symbol — order placed,
   pending filled, position closed, stop out, margin call, trade refused, dealer confirm/requote,
   swaps, activation not admitted — exists only as a `logger.Log(TypeTrade, …)` line in the pod
   log (50 distinct log points, see `hst-core`). None of them is queryable by symbol from the admin.
3. **hst-quote journals per feed, not per symbol.** Only 4 lines exist (`connected`,
   `disconnected`, `feed stopped`, `<symbol> activation`) and they land under `channel=datafeed:N`,
   which the symbol search never queries (the `channel` filter replaces the login filter, it does
   not add to it).
4. **Configuration lines carry no diff.** `symbol 'EURUSD' was updated` is all a manager can read;
   the full record is in `detail` JSON but the message does not say what changed. The drag-and-drop
   move therefore reads as a generic update, not `symbol 'EURUSD' was moved from 'Forex\Crosses' to
   'Forex\Major'`. MT5 names the changed fields.
5. **No request UI.** The toolbox journal has a read-only query box and no Request button, no
   period, no mode, no type filter, no export; `limit=200`, newest-first. MT5 shows ~30 000 lines
   with time bounds. The `from/to/type/mode` query params already exist on the API and are unused.
6. **The journal request itself is not journaled.** MT5 writes one admin-terminal line per request;
   we write nothing, so there is no audit that someone pulled a symbol's history.

## Status (2026-08-27)

Done: server-scoped search (`scope=server`) with the request journaled; hst-core trade lines over
`hstcore.journal.<login>` into `hst.journal` (logger sink, `symbol=` on every line that has one);
hst-quote journals `<symbol> tick dropped: <reason>` once per symbol and `deactivation` when a source
loses the symbol; `symbol 'X' was moved from 'A' to 'B'` and `was updated: k=v, …`; toolbox request
form (keyword, mode, type, Today/Yesterday/Custom, Find, Columns, Copy, Export CSV, day separators,
paged to 5000); trigram index; `JOURNAL_RETENTION_DAYS` purge (0 = keep all).

Skipped: session open/close scheduler lines (a closed session shows as the first dropped tick
instead), "no ticks for N minutes" (the arbiter's `deactivation` covers a silent source),
group diff messages (groups have no `changes` helper yet).

## Plan

Priority order. Each step is independently shippable.

### P0 — make the symbol search return the server's trail (backend, small)

- `admin/journal.go`: add `scope=server` (or drop the login clause when the caller holds
  `right_srv_journal`, name to confirm against `managerRights.js`). Keep `MyJournal` behaviour as
  the default so the personal trail is unchanged.
- Include `channel = datafeed:*` rows in the same result; the `channel` param becomes an extra
  filter, not a replacement.
- `Toolbox.jsx`: symbol/group/account journal requests use the server scope.
- Journal the request: `journal.JournalRequestedMsg(query, from, to, n)` → `type=system`.

### P0 — trade events into the journal (hst-core → hst.journal)

hst-core already has `EventJournalCreate` on the wire but never sends it. Add one journal emit
next to each existing `logger.Log(TypeTrade, …)` — same message, same fields, no new text:

| core log line | journal type | code | message includes symbol |
|---|---|---|---|
| order placed / modified / canceled | trade | ok | yes |
| pending order filled, stop limit became a limit | trade | ok | yes |
| position closed / closed by the client / by an opposite one | trade | ok | yes |
| trade refused, money check refused | trade | warn | yes + retcode name |
| request queued for a dealer / confirmed / passed / requote | trade | ok | yes + dealer login |
| an activation / close / expiry not admitted by a rule | trade | warn | yes + rule |
| margin call, stop out, negative balance compensated | trade | warn | yes |
| swaps charged, period commissions charged | trade | ok | yes |
| could not save … (any) | trade | err | yes |

Transport: publish on NATS as `journal_create`; hst-server's `workerstatus` subscriber already
does exactly this for hst-quote (`subscriber.go:201`) — extend it with a `hstcore.journal.>`
subject, `login = trader`, `channel = system`. No new table, no new package.

### P1 — quote/feed events per symbol (hst-quote)

Add journal lines where hst-quote already knows the symbol, on the existing
`status.Journal(id, code, msg)` path:
- symbol subscribed / unsubscribed on a feed, source switched (arbiter)
- tick filtered (spike, negative spread, digits mismatch) — `Filter` keyword like MT5
- no ticks for N minutes / feed resumed for the symbol
- session opened / closed, holiday applied (from the sessions scheduler)
- quote thrown in manually (already exists on hst-server, keep)

### P1 — configuration diffs in the message

- `admin/symbols.go` update already computes `changes`; put it in the message:
  `symbol 'EURUSD' was updated: path 'Forex\Crosses' → 'Forex\Major', spread 10 → 12`.
  Add `SymbolMovedMsg(symbol, from, to)` for the path-only case (drag-and-drop, bulk, clone).
- Same pattern for group symbols and groups (one helper, reused).
- Symbol folder create / rename / delete are currently unjournaled — add.

### P1 — request UI in the toolbox

- Query box editable + **Request** button, period (Today / Yesterday / Custom from–to), mode
  (Full / Without logins / Errors only), type dropdown (`JournalType_name`), rows count.
- Columns Time · Source · IP · Message; Type/Channel behind "Columns" toggle.
- Export CSV (client-side, from the rows already fetched), Copy row, Find in results.
- `limit` up to 5000 with paging, oldest-first for a time window.

### P2 — housekeeping

- `hst.journal` index on `(created_at)` alone and a trigram index on `message` for the LIKE
  search once volume grows (trade rows will dominate).
- Retention job (MT5 keeps files per day; we keep rows) — a purge older than N days, configurable.
- Day-separator rows in the panel as MT5 does (client-side only).

## Not in scope

Market Depth journal (`ecn_dom_journal.htm`), gateway timing lines, LiveUpdate — no equivalents
in this platform.
