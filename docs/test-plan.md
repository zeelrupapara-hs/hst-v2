# Backend end-to-end test plan

Scope: the whole backend (hst-server, hst-core, hst-quote, hst-news) as it runs in production — real
services, real NATS, real Postgres/Redis, prices through a fake DDE feed into hst-quote. Tests are Go
(`tests/e2e`, build tag `e2e`), organised by **who acts**: admin, managers with different rights and
group masks, traders doing every order kind and every way of closing, and system flows (sharding,
restarts, feeds). No unit tests of functions, no UI. Each scenario: steps → HTTP code → NATS/WS event
→ DB rows.

How to run: `make -C tests/e2e test-e2e` (stack up, migrate, start services, test, stop). CI runs
the same on every PR (`.github/workflows/e2e.yml`).

## How to add a scenario
1. Pick the row below; its id goes in a one-line comment above the test, the test name says what it does (`TestClosePartial`).
2. Put it in the role package (`admin/`, `manager/`, `trader/`, `system/`), one file per feature.
3. Use only `harness` helpers (actors, seed, feed, market, db). Create your own data under the run prefix, clean it in `t.Cleanup`.
4. Trader scenarios run as sub-tests `via=rest` and `via=ws`.
5. Never sleep — wait on an event with a timeout; then assert the DB row.
6. Plain names, one-line comments, no helpers that hide an assertion.

## Harness — Go module `tests/e2e` (stdlib `testing` + net/http + nats.go + pgx; build tag `e2e`)
```
tests/e2e/
  go.mod                      module hste2e (own module, not part of the services)
  harness/
    harness.go   Env: BaseURL, NATS, DB, Redis from env/.env; TestMain helper Boot() checks stack is up
    client.go    Client{token}: Get/Post/Put/Patch/Delete(path, body, &out) → status; BasicLogin(login,pw,connType)
    actors.go    Admin() ; Manager(t, Rights, Masks) creates+logs in ; Trader(t, Persona) ; Persona{Group,Balance,Leverage,Investor,Disabled}
    seed.go      Run prefix e2e<ts>; Group(hedging|netting), GroupSymbol override, Symbol, RoutingRule, Cleanup(t) via t.Cleanup
    feed.go      fake DDE server: Start(port), Replay(dir, speed), Step(sym,bid,ask), Path(sym, []Quote)
    market.go    Tick()=feed.Step(); AwaitEvent(t, login, kind, pred, timeout) on websocket.accounts.<login>.> ; BrokerEvents(dealer)
    db.go        Row/Rows/Count helpers on pgx
    pods.go      StartCore(name, healthPort), StopCore(name); RedisKeys(pattern)
    assert.go    Status(t, got, want), Retcode(t, ev, want), Approx(t, got, want, eps)
  data/ticks/    EURUSD.csv GBPUSD.csv USDJPY.csv XAUUSD.csv BTCUSD.csv  (15–30 min recorded, time,bid,ask)
  admin/        admin_test.go bootstrap_test.go accounts_test.go money_test.go config_test.go feeds_test.go eod_test.go
  manager/      personas_test.go scope_test.go dealer_test.go accountant_test.go readonly_test.go sessions_test.go
  trader/       login_test.go market_test.go pending_test.go stops_test.go close_test.go netting_test.go scope_test.go history_test.go
  system/       sharding_test.go restart_test.go quote_test.go news_test.go shutdown_test.go
  Makefile      test-e2e: stack up, migrate, start services, `go test -tags e2e -count=1 -p 1 ./...`, stop
```
Conventions: one test per scenario, plain name saying what it does (`TestCloseFull`, id in a comment above), table-driven sub-tests for variants, `t.Cleanup` for every row created, no sleeps — wait on events with timeout, no globals except the `Env` from `TestMain`, one-line comments, helpers never assert silently (they take `t *testing.T` and `t.Helper()`).
**Runtime under test (nothing mocked inside services):** docker postgres/nats/redis/influx; real processes hst-server, hst-core (pod A, pod B spawned by SYS-1), hst-quote, hst-news. Tests use only public HTTP/WS, NATS and the DB.

**Prices (hybrid, decided):**
- `tests/e2e/data/ticks/<SYMBOL>.csv` — recorded 15–30 min of real ticks for 5 symbols (EURUSD, GBPUSD, USDJPY, XAUUSD, BTCUSD), `time,bid,ask`.
- `harness/feed.go` — fake **DDE** server (Go net.Listener, frames per hst-quote/README DDE section): accepts the auth token, pushes `!`-terminated `MRKTDATAs?<G)_` frames. Registered in hst-server as datafeed `module=dde`, `feed_server=127.0.0.1:<port>`, translates for the 5 symbols + `E2EUSD`/`E2EMKT`. So every price goes feed → hst-quote → translate/markup → Influx + `hstquote.tick.*` → hst-core → trader WS.
- Two modes on the same server: `Replay(dir, speed)` streams the csv files in the background for the whole run (realistic market, hst-quote under load); `Step(sym,bid,ask)` / `Path(sym, quotes)` pushes scripted quotes for a scenario (deterministic SL/TP/stop-out/requote). `market.Tick()` below = `feed.Step()`; direct NATS publish is kept only as a fallback flag for debugging.

Baseline per run (from `docs/qa-runbook.md`): group `e2e\real` USD 1:100 hedging MC100/SO50; group `e2e\net` netting; symbol `E2EUSD` (5 digits, 100k contract, min 0.01 step 0.01, stops 10, freeze 5, sessions 24×7, fill FOK|IOC, exec instant) + `E2EMKT` (exec market); rule "confirm at market"; ticks injected 1.10000/1.10002. Every scenario = steps → expected HTTP code → expected NATS event → expected DB rows. Engine replies are async (202): assert on `websocket.accounts.<login>.*` events, then DB.

## API surface the scenarios drive (verified from swagger + ws registry)
- Auth: `POST /auth/v1/login` (Basic + `{connection_type}`), `/auth/v1/refresh`, `POST /api/v1/auth/logout`, `/api/v1/auth/me`, `/api/v1/auth/change-password`, `/api/v1/auth/switch-terminal`; trader `/auth/trader/v1/{login,register,refresh,forgot-password,reset-password,verify-code}`.
- Admin REST `/api/v1/`: users(6) clients(5) managers(6) groups(15) symbols(12) leverages(9) holidays(7) routing(12) datafeeds(21) mail-servers(5) balance(5) orders(6) positions(6) deals(3) dealing(11) journal(2) system(9) navigation(1) history(1); trader REST `/api/trader/v1/` (45 ops: account, symbols/tree, orders, positions/close/close-by/net, deals, history, watchlists, alerts, profile, news, mails, requotes accept).
- WebSocket `GET /ws` (subprotocol `bearer`, message `{type,payload}`; no correlation id → tests match replies by order/position id in payload). Request types under test: `start_market_feed/stop_market_feed`, `order_create/update/cancel`, `position_update/close/close_by`, `balance_create`, `order_dealer_*`, `position_dealer_*`, `dealer_confirm/requote/reject/cancel`. Events asserted: `welcome`, `market_feed`, `order_create/update/cancel/rejected/expired`, `position_create/update/close`, `deal_create`, `account_summary`, `money_change`, `margin_call`, `stop_out`, `dealer_request`, `dealer_request_done`, `session.revoked`, `bad_request/forbidden/not_found/unauthorized`.
- Every trader scenario runs **twice**: REST variant and WS variant (table-driven sub-test `via=rest|ws`), because both surfaces exist in production.
- Internal/service: `/internal/v1/datafeeds` (hst-quote/news config), NATS `hstquote.tick.>`, `hstnews.item.*`, `system.*`, `websocket.accounts.<login>.>`, `websocket.broker.>`.

## CI (GitHub Actions — remote is github.com/zeelrupapara-hs/hst-v2; current workflows only build+deploy, no tests)
New `.github/workflows/e2e.yml`:
- Triggers: `pull_request` to `develop`/`main`, push to `develop`, nightly cron, `workflow_dispatch`.
- Job `e2e` on `ubuntu-latest`: `services:` postgres:18-alpine, nats:2.12-alpine (`--jetstream`), redis:8-alpine, influxdb:2.7-alpine (same images/env as `hst-server/stack.yaml`).
- Steps: checkout → setup-go (cache 5 modules) → `make -C hst-server migrate-up` (installs `migrate`) → build binaries (`go build -o bin/ ./cmd` in hst-server, hst-core, hst-quote, hst-news) → start them in background with an `.env.ci` per service (`AUTH_ARGON2_MEMORY_KIB=8192`, `FIRST_MANAGER_PASSWORD`, `POD_NAME=core-a`, `HEALTH_PORT` distinct) → wait on `/api/v1/system/monitor/health` + `/readyz` → `cd tests/e2e && go test -tags e2e -count=1 -p 1 -timeout 40m ./... -json | tee e2e.json` → `go-test-report`/summary to the job, upload `logs/` of all services + `e2e.json` as artifacts on failure.
- Branch protection: `e2e` required on PRs to `main`; `develop` runs it but is not blocked.
- Makefile at `tests/e2e` has the same steps for local (`make test-e2e`), so CI = local.
- Small backend enablers (tiny, in this plan): env overrides `AUTH_MAX_FAILED_ATTEMPTS`, `AUTH_MAX_FAILED_PER_IP`, `AUTH_LOCKOUT_SECONDS` in `hst-server/config/config.go` (currently hardcoded 10/50/15m) so throttle scenarios finish in seconds; existing `make check` added as a `lint` job in the same workflow for the four Go modules.

## Scenario matrix

### ROLE: ADMIN (login 1000, all rights) — package `admin`
| ID | Flow | Steps | Expected end state |
|---|---|---|---|
| ADM-1 bootstrap | login → me → navigation | 200; `can[*]` all true; journal login row has no token |
| ADM-2 build a desk | create group → group-symbol override → symbol → leverage profile → commission → holiday → routing rule + dealer | each 201; `system.<table>.created` received; core reloaded (a trade on the new symbol works) |
| ADM-3 onboard client | create client → create user in group → deposit 10 000 → see account | account row, `money_change` event, balance deal, journal actor=1000 target=login |
| ADM-4 manager lifecycle | create manager with masks `e2e\real\*` + rights subset → login as it → admin edits its rights → its next request refused/re-evaluated → admin deletes it (409 if last `*`) | InvalidateLogin honoured; manager's refresh token dead |
| ADM-5 account control | reset trader password (main/investor) → old trader session 401; disable trade on account → trader's order refused; set group connection off → all traders in group disconnected | matching retcodes and events |
| ADM-6 money ops | credit, correction, bonus, withdraw beyond free margin (refused), bulk balance with one bad row | deals per op, refused row not "done", balances reconcile (`CheckBalances`) |
| ADM-7 trade ops on behalf | open/close/modify/delete position for a trader; fix position; delete order | engine retcodes honoured (no 200 on refusal), journal actor/target |
| ADM-8 destructive guards | delete symbol with open position → 409; delete user with balance → 409; delete group with users → 409; rename group with users → cascades, trader still trades | |
| ADM-9 datafeeds | create quote feed (bad enable → 400), params add/reorder/delete ×5, translates, activate → snapshot published; news feed same | priorities always 0..n-1; `system.datafeeds.*` |
| ADM-10 mail/journal | mail server + template + send to local sink; journal filters | creds never echoed; message received; audited rows complete |
| ADM-11 end of day | set EOD time → trigger EOD with open positions → swap on position persisted; restart core → still there; second EOD same day → no double charge | `hst.positions.storage` |

### ROLE: MANAGERS with different rights + masks — package `manager`
Build 5 manager personas once per run (`harness/actors.go`):
- **DeskManager** masks `e2e\real\*`, rights: Manager, AccRead, AccManager, TradesRead, TradesManager, ClientsAccess/Create/Edit
- **Dealer** masks `e2e\*`, rights: Manager, TradesRead, TradesDealer
- **ConfigManager** masks `e2e\real\*`, rights: Manager, CfgGroups, CfgSymbols, CfgRequests(routing)
- **Accountant** masks `e2e\*`, rights: Manager, AccRead, Accountant, SrvJournals
- **ReadOnlyManager** masks `e2e\real\*`, rights: Manager, AccRead, TradesRead only

| ID | Persona | Flow | Expected |
|---|---|---|---|
| MGR-1 | all | login with `connection_type` admin panel vs manager panel; me/navigation `can[]` equals granted rights | 200, rights exact |
| MGR-2 | DeskManager | list/get/update user inside mask ok; user in `e2e\net` → 404; create user in `e2e\net` → 403; move user out of mask → 403; reset password in mask ok, out → 404 | |
| MGR-3 | DeskManager | open/close position for own-mask trader ok; for out-of-mask trader → 404; delete order in mask ok | engine result events |
| MGR-4 | DeskManager | tries: deposit (no Accountant) → 403; create symbol (no CfgSymbols) → 403; read journal (no SrvJournals) → 403; edit manager (no CfgManagers) → 403 | rights bits enforced per route |
| MGR-5 | ConfigManager | group CRUD inside masks ok, get/update/delete `e2e\net` by id → 404, create group under `other\x` → 403; symbol create ok; routing rule create ok | |
| MGR-6 | ConfigManager | edit group-symbol spread for `e2e\real` → next trader trade reflects it; edit for `e2e\net` → 404 | config → core reload |
| MGR-7 | Dealer | connect to desk; receives `dealer_request` for rule→dealer orders from any `e2e\*` group; confirm / reject / requote / modify-price; confirm by order_id; foreign (non-mask) request → refused | dealing events, order states |
| MGR-8 | Accountant | deposit/withdraw/credit for `e2e\*` accounts ok; open position → 403; journal read ok and contains no tokens; bulk balance with one refused row | |
| MGR-9 | ReadOnlyManager | every mutating call → 403 (users, balance, orders, positions, groups, symbols); all lists inside mask ok, outside empty/404 | |
| MGR-10 | any | admin removes a right while manager session live → next call 403; admin removes Manager right → 401 + refresh refused; admin changes manager IP allowlist → session revoked | session invalidation |
| MGR-11 | DeskManager | self-edit: PATCH own rights/group → 403; PATCH own name ok | |
| MGR-12 | ConfigManager | manager creating a manager with rights it lacks → 403; editing a manager that outranks it → 403 | scope |

### ROLE: TRADER — every order kind, every close — package `trader`
Personas: HedgeTrader (`e2e\real`, 10 000 USD), NetTrader (`e2e\net`), EurTrader (EUR account, USD symbol), Investor (investor password session), DisabledTrader (trade disabled), PoorTrader (balance 50).

| ID | Persona | Flow | Expected |
|---|---|---|---|
| TRD-1 login | all | Basic login → me, refresh, logout → old tokens dead; wrong password ×N → lockout; investor session flagged read-only | |
| TRD-2 market buy/sell | HedgeTrader | buy 0.10 instant at ask, sell 0.10 at bid → 2 positions; margin, equity, free margin numbers asserted; deals in/in | `position_create` ×2, `summary` |
| TRD-3 market on market-exec symbol | HedgeTrader | buy with client price ignored → filled at market; deviation irrelevant | |
| TRD-4 instant with deviation | HedgeTrader | price moves beyond deviation between request and fill → requote; accept requote → fill at new price; decline → no position | `requote` event |
| TRD-5 buy/sell limit | HedgeTrader | buy limit below ask ok (placed), above ask → rejected; tick crosses → filled; sell limit mirror | `order_create`, then `position_create` |
| TRD-6 buy/sell stop | HedgeTrader | buy stop above ask ok, below → rejected; tick crosses → filled; sell stop mirror | |
| TRD-7 stop-limit | HedgeTrader | buy stop-limit trigger above ask, limit below trigger → placed; trigger hit → becomes limit; bad sides → rejected | order type conversion row |
| TRD-8 SL/TP on orders | HedgeTrader | market with SL/TP inside stops level → rejected; valid → position carries SL/TP; tick hits SL → closed reason sl; TP → reason tp | `position_close`, deal out |
| TRD-9 expiry | HedgeTrader | pending GTC stays; `specified` expires at time; `day` expires at EOD; `specified_day` | `order_expire` |
| TRD-10 modify/cancel | HedgeTrader | modify pending price/SL/TP (volume unchanged in DB); modify inside freeze → rejected; cancel → canceled; cancel filled order → not found | |
| TRD-11 fill policies | HedgeTrader | FOK partial impossible → rejected; IOC partial → partial fill + remainder canceled (if symbol allows); volume min/max/step/limit gates | MT5 retcodes |
| TRD-12 close full | HedgeTrader | close buy at bid → profit = (bid−open)×lots×100k; balance += profit+swap+commission; position gone; `CheckBalances` ok | deal out |
| TRD-13 close partial | HedgeTrader | close 0.04 of 0.10 → position 0.06, deal out 0.04 with pro-rata storage; bad step 0.015 → rejected | |
| TRD-14 close-by | HedgeTrader | buy 0.10 + sell 0.10 → close-by → one deal out_by, both gone; unequal volumes → remainder stays | |
| TRD-15 close by opposite (netting) | NetTrader | buy 0.10 then sell 0.10 → position folds to 0 (closed); buy 0.10 then sell 0.04 → 0.06 long; sell 0.20 → flips to 0.10 short | one position row always |
| TRD-16 close by SL/TP/stop-out | PoorTrader | open max, move price → margin call event, then stop-out closes worst position first | `margin_call`, close reason so |
| TRD-17 close by dealer | HedgeTrader + Dealer | rule routes close to dealer; dealer confirms → closed; rejects → still open | |
| TRD-18 close by admin | HedgeTrader + admin | admin closes/deletes position → trader sees `position_close`; delete → no deal | |
| TRD-19 cross currency | EurTrader | buy USD symbol; P/L and margin converted at live rate; rate tick changes equity | |
| TRD-20 scope | Investor | any order/close/modify → 403; reads ok. DisabledTrader: order → rejected trade disabled. PoorTrader: not enough money retcode | |
| TRD-21 IDOR | HedgeTrader | act on another trader's order/position/request id → 404/refused; WS receives only own events | |
| TRD-22 session + market closed | HedgeTrader | symbol session closed / holiday → order rejected; quote stale (> quotes time) → rejected | |
| TRD-23 history | HedgeTrader | deals/orders/positions history paging and time filters after the above | counts match DB |

### SYSTEM flows (no role, whole backend) — package `system`
| ID | Flow | Expected |
|---|---|---|
| SYS-1 shard handoff | pod A alone → start B → leases split, trades land on owner only → 20 orders fired during join each answered once → kill B → A regains, ex-B trader trades, balances unchanged | redis `core:shard:*`, DB counts |
| SYS-2 dealing across pods | request queued on B, kill B → request recovered on A or swept with client told | no `request_*` orphan rows |
| SYS-3 restart durability | open positions + swap + pending orders → restart core → same state, no duplicate rows, volumes unchanged | |
| SYS-4 quote path | fake DDE/FIX source → hst-quote → `hstquote.tick.<sym>` (dot-symbol escaped) → core prices + trader WS market feed; source dies → runner restarts; feed priority arbiter switches source | |
| SYS-5 news path | local RSS → hst-news → `hstnews.item.*`, dedup, config edit applied live | |
| SYS-6 config propagation | every admin config change in ADM-2 → core reflects it in the next trade without restart | |
| SYS-7 shutdown | SIGTERM hst-server/core → readiness fails first, drains, exits < drain wait, no hang (workerstatus) | |
| SYS-8 health/503 | Influx down → history 503 + Retry-After; readyz reflects deps | |

## Order of writing (each file = one Go test file, templates first)
1. `harness/` + `admin/bootstrap_test.go` (ADM-1..3) + `trader/market_test.go` (TRD-2, TRD-12) — the template pair.
2. manager personas + MGR-2/4/5/9 (rights + masks).
3. trader order kinds TRD-5..11, closes TRD-13..18.
4. ADM-4..11, MGR-6..12, TRD-19..23.
5. system SYS-1..8 last (restarts pods/services).
Runs: `make test-e2e` in `tests/e2e` (stack up, migrate, start server + core A + quote + news + fake feed, `go test -tags e2e -count=1 -p 1 ./...`, stop). CI nightly and before merge to main.

