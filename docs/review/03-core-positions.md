# core-positions review

Scope: hst-core handler/positions.go, positions_hedge.go, floating_margin.go, stopout.go, commissions.go, commissions_accrual.go, end_of_day.go, deals.go, accounts.go, query.go, ticks.go, money.go, util.go, store.go, parts of orders.go/pending.go/validate.go/load.go/publish.go, internal/quote, internal/book, internal/settings (calendar.go, settings.go), model/trade.go, balance.go, tick.go, leverage.go, symbol.go, group.go; hst-server internal/server/v1/balance.go, positions.go, positions_http.go for the api side. 30 files, ~6,900 lines. Flows: open/close/close-by/partial close, SL/TP on tick, margin + floating P/L, margin call / stop out, swaps + EOD, commission accrual, balance ops from api.

## Summary
The engine's fill/close/margin shape is sound: one lock per account, money settled through `CalculateAccountMargins`, DB writes in one tx per trade, publishes after commit. The money is not: swaps accrued nightly are never written to the `positions` row (lost on restart or shard move), period commissions read the legacy `volume` column with the extended `Lots()` divisor (10,000x under-charged), the hedged-margin leg ignores leverage and calc mode, commission and swap currencies are converted wrongly or not at all, and floating P/L uses the conversion rate frozen at open. Closes skip the account-rights check and any volume step check. Several paths mutate the book before the DB write and keep the change when the write fails. Readability is good for a junior: names are plain, comments explain intent; positions.go (941 lines) and the lock-passing conventions (`removeOrder` unlocks what the caller locked) are the two things to simplify first.

## Findings
| # | Sev | Type | Where | What | Why it matters | Fix |
|---|---|---|---|---|---|---|
| 1 | P0 | bug | hst-core/handler/end_of_day.go:140,156 → positions.go:635-640 | `AccrueSwap` does `p.Storage += swap` then `SavePositionAndPublishAsync`, whose SQL updates only `price_sl, price_tp, activation_flags, time_update`; `flushAccount` (shards.go:184-189) saves the account only | Every swap ever accrued on an open position is lost on pod restart or shard move; the client's balance on close then differs from what they were shown | Add `storage = $n` to that UPDATE (or a dedicated `UPDATE hst.positions SET storage`), and write it synchronously in the EOD job |
| 2 | P0 | bug | hst-core/handler/commissions_accrual.go:166,193 | `SELECT login, symbol, SUM(volume) …` reads the legacy `volume` column (written as `model.Legacy(d.Volume)` in store.go:135) but converts with `model.Lots(volume)` which divides by `VolumeUnitExt` (1e8) | Daily/monthly per-volume commissions and tier selection are 10,000x too small; `Money` is right (divided by `VolumeUnit`) so the two measures disagree | `SUM(volume_ext)` with `Lots`, or keep `volume` and divide by `VolumeUnit` |
| 3 | P1 | bug | hst-core/handler/positions.go:80,757; 530,547,560,606 | `RateProfit` is computed once at open (`h.RateProfit(...)`) and reused for floating P/L (`CalcPosition`) and for every close | On a cross-currency account the P/L in deposit currency is valued at the open-time rate for the position's whole life; equity, margin level and stop out drift from the real value | Refresh `p.RateProfit` in `CalcPosition` (and before closing) from `h.RateProfit`; keep the open-time rate only on the deal row |
| 4 | P1 | bug | hst-core/handler/positions_hedge.go:88-90 | Hedged part: `margin += model.Lots(smaller.volume) * r.MarginHedged * rate` — no leverage, no price, no calc mode, no `MarginRate` multiplier; the uncovered part next to it goes through `MarginForType` | Under forex the covered leg is charged `leverage` times too much (e.g. 1:100 → 100x); under CFD it ignores price | Run the hedged lots through `marginBasePlain`-style formula with `MarginHedged` as the contract size, then apply the multiplier |
| 5 | P1 | bug | hst-core/handler/commissions.go:80-97,148-171 | `CommissionFor`/`charge` return `tier.Value` (or `Value*lots`) straight onto `e.Account.Balance` (money.go:262-263); `Commission.Currency` / `CommissionTier.Currency` are loaded (line 241,270) but never used | A commission defined in USD is charged 1:1 on a EUR/JPY account | Convert with `h.crossRate(group, tier.Currency, a.Currency, …)` before charging; warn when no rate |
| 6 | P1 | bug | hst-core/handler/end_of_day.go:257-259 | Swap converted with `h.crossRate(a.Group, r.CurrencyProfit, a.Currency, …)` for every mode; `SwapMarginCurrency` is in `CurrencyMargin`, `SwapDepositRate` is already in deposit currency | Wrong direction/currency for two of the modes; mis-charged swap | Pick the source currency per `SwapMode` |
| 7 | P1 | validation | hst-core/handler/positions.go:162-223, 257-331 | `ClosePosition` and `CloseByPosition` never call `checkAccount` (validate.go) and never check `req.Volume` against `VolumeStep`/`VolumeMin`; `closingOrder` (701-703) only clamps to `p.Volume` | A trade-disabled / investor / read-only login can close positions; a partial close of 0.001 on a 0.01-step symbol is accepted, leaving an unsteppable remainder | Call `checkAccount` first; validate `req.Volume` with the same step/min check orders use |
| 8 | P1 | bug | hst-core/handler/positions.go:814-815 ← hst-server internal/server/v1/positions_http.go:511 | `FixPosition` sets `p.Volume = req.Volume` and `p.VolumeExt = model.FromLegacy(req.Volume)`; the server sends `validVolumeExt / ExtPerUnit` (legacy units), while everywhere else `p.Volume` is extended (load.go:532 scans `volume_ext` into `p.Volume`) | After a fix the in-memory volume is 10,000x too small until the next reload: P/L, margin and any close on it are wrong | Send extended volume from the server and set `p.Volume = req.Volume` only (drop `VolumeExt`, it is unused in the engine) |
| 9 | P1 | bug | hst-core/handler/commissions_accrual.go:22-26, 66, 203-215, 231 | Daily window is `[midnight, at)` where `at` = EOD start; the charge deal is stamped `Now()` later. Deals between EOD and midnight never fall in any window; if the job crosses midnight the deal lands in tomorrow's window and `chargedBetween` skips tomorrow's charge | Silent under- or non-charging depending on EOD time and job duration | Window = `[previous EOD, this EOD)`; mark the period on the deal (comment or a `period_from` column) rather than by timestamp |
| 10 | P1 | bug | hst-core/handler/orders.go:476-497 (bookFill), store.go:117-141, deals.go:20-31 | `SettleSwap` moves `p.Storage` onto the balance, `ChargeCommission` takes commission off it, but the closing deal row carries `Storage = 0`; hst-server `CheckBalances` (balance.go:136) sums `profit + storage + fee` from deals | Every account that ever paid swap or commission fails the balance check; history does not show the swap | Set `d.Storage = p.Storage` (pro rata on partial) on the out-deal; include `commission` in the server's sum |
| 11 | P1 | bug | hst-core/handler/positions.go:140-150, 225-234, 339-348, 410-419; pending.go:99-115 | The book is mutated (levels changed, position deleted, order removed) before the DB write; on error the code logs and returns `RetError` but the memory change stays | Client is told "error" while the engine has closed the position; DB and memory diverge until the next reload, a later reload resurrects the position | Snapshot before, restore on write failure (or treat the write error as fatal for the entry and reload it) |
| 12 | P2 | concurrency | hst-core/handler/store.go:170-188 vs internal/book/book.go:209-235 | `unwatchIfLast` holds `e.mu` then takes `Book.mu.Lock` in `Unwatch`; `Book.Positions()/Orders()` hold `Book.mu.RLock` then `e.Lock()` | Lock-order inversion → deadlock; today only reachable from the startup log (load.go:65) | Release `e.mu` before `Unwatch`, or have `Positions()` copy entries first like `Each` does |
| 13 | P2 | concurrency | hst-core/handler/commissions.go:83-84 vs 283-285 | `CommissionFor` reads `h.commissions[...]` without `h.mu`; `LoadCommission` swaps the map under `h.mu.Lock`; `periodCommission` (accrual.go:96) does lock | Data race on reload under load | `h.mu.RLock()` around the slice fetch |
| 14 | P2 | concurrency | hst-core/handler/end_of_day.go:155-157 → positions.go:650-657 | `SavePositionAndPublishAsync(account.Group, p)` hands the live `*Position` to a worker that reads it without the entry lock while ticks keep writing it | Torn reads of SL/TP/time published to WS | Pass a copy (`saved := *p`) as `UpdatePosition` does |
| 15 | P2 | bug | hst-core/handler/store.go:66-93 | `writeFill` swaps the temp id for the real one and calls `Accounts.Watch` before `tx.Commit` (147) | If commit fails the book holds a real position id whose row does not exist | Move the id swap and `Watch` after commit |
| 16 | P2 | bug | hst-core/handler/floating_margin.go:79-100 | Exposures whose rule is not found are appended to `plain` after the `plain` loop already ran | Those symbols contribute zero margin, silently; dead today because `findLeverageRule` mirrors `MatchLeverageRule`, but one edit away | Handle the fallback inline or run the `plain` loop after |
| 17 | P2 | bug | hst-core/handler/stopout.go:314-323 | `CompensateNegativeBalance` ignores the `NewBalance` result and logs "compensated" | A refused compensation is reported as done | Check `res.RetCode` |
| 18 | P2 | bug | hst-core/handler/end_of_day.go:183-197 | `ReleaseAccumulatedProfit` moves `BlockedProfit` into `Balance` with no deal row | Balance no longer equals the deal sum; the client sees the balance jump with no history line | Write a `DealAction_balance`-class deal (or record it on the account and exclude in the check) |
| 19 | P2 | bug | hst-core/handler/util.go:27, end_of_day.go:218-220,264-268, calendar.go:96 | Sessions/holidays use `time.Now().UTC()`; swap weekday and EOD use local `time.Now()` | A pod in a non-UTC zone charges the triple swap on a different day and runs EOD at a different instant than the session tables expect | One clock for the engine (UTC, or a configured server zone) |
| 20 | P2 | resilience | hst-core/handler/accounts.go:25 | `NewBalance(context.Background(), …)` — no timeout on the DB tx | A hung DB hangs the NATS handler | Use a bounded context as the trade handlers do |
| 21 | P2 | bug | hst-core/handler/floating_margin.go:187-205 | `orderMaintenanceMargin` finds "the" order by matching `Lots(o.VolumeCurrent) == exp.lots` | Two pending orders of the same size on one symbol share the first one's maintenance; fragile heuristic | Carry the order id in `marginExposure` |
| 22 | P2 | bug | hst-core/handler/money.go:206-208 | `NormalisePrice` does `int64(price*p+0.5)` | Overflow / undefined on Inf or very large values (swap with zero `SwapYearDay` is guarded by settings.go:283, but the helper itself is not) | `math.Round(price*p)/p` |
| 23 | P3 | readability | hst-core/handler/positions.go:669-670 | Doc comment for `positionById` sits above `firstInLine` | Misleads a reader | Move it |
| 24 | P3 | readability | hst-core/handler/pending.go:97-124, stopout.go:187-194 | `removeOrder` unlocks an entry the caller locked; callers re-lock right after | Asymmetric locking is the kind of thing a junior breaks | Have callers hold nothing; `removeOrder` locks and unlocks itself |
| 25 | P3 | readability | hst-core/handler/positions_hedge.go:122-129 | `seen` map over a map iteration is redundant | Noise | Drop it |
| 26 | P3 | readability | hst-core/handler/stopout.go:86 | `call*1.05` hysteresis is a magic number | Unexplained | Named constant with a one-line comment |
| 27 | P3 | readability | hst-core/handler/positions.go:463-467 | FIFO skip logs a warning on every tick while an older position exists | Log spam | Log once per position (latch) or at debug |
| 28 | P3 | readability | hst-core/handler/positions.go (941 lines), `CloseByPosition` 108 lines | One file holds verbs, fill arithmetic, fix/delete and helpers | Onboarding cost | Split: verbs / fill math / corrections |
| 29 | P3 | bug | hst-core/handler/stopout.go:175, pending.go:149 | `e.Account.Group` read outside the lock | Benign today (group rarely changes) | Capture under the lock |
| 30 | P3 | structure | hst-core/handler/commissions_accrual.go:164-170 | `turnoverBetween` reads every pod's deals on every pod | N pods × full scan at EOD | Filter by the pod's logins (`= ANY($n)`) |

## Flow traces

**Open (market / pending fill)**
1. orders.go → `Execute` positions.go:31 → `NewPosition` :56 (margin at open price :86, `RateProfit` frozen :80 — finding 3) or `netInto` :486 → grow/reduce/close/reverse :512-601.
2. `bookFill` orders.go:476: commission → swap settle → realised profit → temp id → `CalculateAccountMargins`.
3. `SaveOrderAndPublish` orders.go:403 → `save`/`saveFill` store.go:15/52 → `writeFill` :64 (id swap before commit — finding 15) → `PublishTrade` publish.go:111.
Verdict: OK apart from findings 3, 15.

**Close / partial close (client)**
1. hst-server positions.go:124-160 (`LotsToVolume` → ext units) → NATS → `ClosePosition` positions.go:162.
2. No `checkAccount`, no step check (finding 7) → `firstInLine` :190 → `closingOrder` :195 → `checkExecution` validate.go:198 → `Route` :206 → price :216-220 → `Execute` :223 (`reducePosition` :527 or `closeInto` :545) → `bookFill` → unlock → save (finding 11) → reply.
Verdict: gap at step 2 (validation) and step 2 save failure.

**Close-by**
1. `CloseByPosition` positions.go:257: hedging only :279, same symbol/opposite side :283-290, volume = min :298, order type close_by :305, route :323, `closeAgainst` :604 (profit on `p` at `by.PriceOpen`, two out_by deals, second with zero profit — MT5-consistent), `takeOff` :619. Save as above.
Verdict: OK (same rights gap as close, finding 7).

**SL/TP on tick**
1. hst-quote → `NotifyAll` ticks.go:11 → `Watching` → worker → `CalculateAccountProfits` :31 (lock) → group spread :43 → `CalcPosition` positions.go:745 → settle → summary → `CookPosition` :427 (SL/TP at close price, flags, FIFO, MarginFlagCheckSLTP → `coversAfterClose` :767) → unlock.
2. `CloseAtMarket` :367 → re-lock, re-check still open :379, route :394, `Execute`, save.
3. `checkStopOut` last ticks.go:94.
Verdict: OK.

**Margin + floating P/L**
1. `CalculateAccountMargins` positions_hedge.go:175 → `RemargeAccount` :119 → `accountMarginBreakdown` floating_margin.go:48 (plain vs leverage-rule tiers, pending reservations) → `MarginForSymbol` positions_hedge.go:35 (netting / one leg / larger leg / uncovered + hedged — finding 4) → `SpreadMargin` :97 → maintenance → `Settle` money.go:137 (equity, free margin per `FreeMarginMode`, level against maintenance).
Verdict: gap at `MarginForSymbol` hedged branch (4), stale `RateProfit` (3), dead fallback (16).

**Margin call / stop out**
1. `checkStopOut` stopout.go:29: busy latch :32 → settle :37 → fully-hedged case :41-43 → level :53 → margin call once :60-82 with 5% hysteresis :86 → stop out :94.
2. `stopOutPendings` :136 (largest reservation first, routable, `removeOrder` with caller's lock) → `stopOutPositions` :208 (`stopOutOrder` :262 biggest loser / FIFO oldest per symbol, `logStopOut` :366, `closeAtMarket` with "[so at …]") → `CompensateNegativeBalance` :300 (finding 17).
Verdict: OK apart from 17, 24.

**Swaps + EOD**
1. `RunEndOfDay` end_of_day.go:54 (local clock — finding 19) → `EndOfDayProcess` :71: expiry → `SwapsJob` :102 → `AccrueSwap` :116 (`CalculateSwaps` :205, currency — finding 6) → async position save that drops storage (finding 1) → `ReleaseAccumulatedProfit` :166 (no deal — 18) → daily/monthly commissions.
Verdict: gap at step 1 (finding 1).

**Commission accrual**
1. Instant: `bookFill` → `ChargeCommission` money.go:260 → `CommissionFor` commissions.go:80 (covers/tier/charge; no currency — finding 5; no lock — 13).
2. Period: `chargePeriod` accrual.go:43 → `turnoverBetween` :164 (legacy volume — finding 2) → `chargedBetween` :203 (window — 9) → `periodCommission` :90 → `applyPeriodCharge` :218 → `SaveBalanceAndPublish` accounts.go:160.
Verdict: gaps 2, 5, 9.

**Balance ops from api**
1. hst-server balance.go:45 `makeBalance` (struct validate, action, amount≠0, `AllowNegative` only for correction) → `request(SubjectSystemBalance)` → hst-core `BalanceSystemEventHandler` accounts.go:14 → `forwarded` :21 → `NewBalance` :37 (action, amount/fix, account, negative guard :81, credit vs balance, deal :98, settle) → `SaveBalanceAndPublish` :160 (deal + account in one tx, publish after commit) → reply.
2. Fix path :57 writes value without a deal (by design, server balance.go:97-99).
Verdict: OK apart from finding 20.

## Readability notes for onboarding
- positions.go — every position verb plus the fill arithmetic and manager corrections; clear names, too long; split first, fix the misplaced comment.
- positions_hedge.go — margin per symbol and account remarge; clear, the hedged-leg formula needs a reference comment once fixed.
- floating_margin.go — leverage tiers; the `plain`/`byRule` two-pass is easy to misread (finding 16); add a 3-line overview.
- stopout.go — readable top-down; the lock passing into `removeOrder` is the trap.
- commissions.go / commissions_accrual.go — clear; constants are unnamed ints in the DB sense, fine; say where currency conversion happens once it exists.
- end_of_day.go — short; swap modes would benefit from one comment per case naming the MT5 mode.
- deals.go — clear; note that `Storage` on out-deals must be filled.
- accounts.go, query.go, ticks.go — clear, short.
- money.go, util.go — pure helpers, good home for a junior to start.
- store.go — one tx per trade; note the pre-commit id swap.
- internal/quote, internal/book, calendar.go — small, clear.

## Open questions
- `MarginRates.For` returns the raw array value; a symbol row with all-zero `margin_rate_initial` yields zero margin (money.go:34-42). Is a zero rate defaulted to 1 on the server side, or should the engine guard it?
- hst-server `CheckBalances` sums `profit + storage + fee` but not `commission`; is commission meant to be folded into `profit` on deal rows, or is the server query incomplete (finding 10)?
- Should period (daily/monthly) per-trade tiers charge `Value` once per period or once per deal (accrual.go:151)? MT5 charges per deal.
- Is `p.VolumeExt` on `model.Position` used by anything? The engine reads/writes only `p.Volume` except in `FixPosition`.
- Server time zone: is the engine expected to run in UTC only? Sessions assume it, EOD and swaps do not.
