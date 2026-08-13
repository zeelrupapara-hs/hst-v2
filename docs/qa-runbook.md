# QA runbook — platform bring-up, phase by phase

A repeatable drill from an empty-ish server to a verified buy/sell, in the strict order
the platform depends on: **feed → symbol → group → account → trade → money**. Each phase
lists the steps, the checks, and the expected result; anything that fails goes into the
findings report at the bottom, in the exact format defined there.

Run it against local (`http://localhost:5175` admin/manager, `:5173` trader) or staging
(`:8081` / `:80`). Use Playwright for every UI step and record evidence (screenshot or
DB row) for every ✗.

> **Prompt to run this:** "Run docs/qa-runbook.md phase by phase with Playwright against
> <environment>. Do not skip a phase gate. Produce the findings report in the format at
> the bottom of the file."

---

## Phase 0 — preflight

| Step | Check | Expected |
|---|---|---|
| services | hst-server :8080, hst-core, hst-quote, sim/feed, nats, redis, postgres, influx | all up |
| tick bus | `nats sub "hstquote.tick.>" --count 1` | a tick within 5 s |
| panels | admin login (Administrator terminal), manager login (Manager terminal), trader login | all three sign in |

**Gate:** no ticks on the bus ⇒ everything downstream will "work but stay grey" — stop
and fix the feed first (Phase 1).

## Phase 1 — datafeed

1. Admin panel → Datafeeds. Is there a feed row? Is its status **connected**?
2. If none / disconnected: create the feed (FIX host, port, credentials from the sim),
   enable it, watch the status flip to connected and the journal log the connection.
3. Check the feed's **symbol subscriptions** — every symbol you plan to trade must be
   requested by the feed (the sim logs `requested unknown symbol X` for bad names).

| Check | Expected | If wrong |
|---|---|---|
| feed status | connected, ticks/s > 0 | connection settings; sim running; journal shows "datafeed lost connection" |
| per-symbol flow | `redis-cli get hstquote:last:<SYM>` fresh | symbol not in feed scope → add it to the feed's symbol list |

## Phase 2 — symbols

1. Admin → Symbols. For each traded symbol verify the master record: **digits, contract
   size, currencies (base/profit/margin), volume min/max/step, stops level, exec mode,
   fill flags, trade mode Full, sessions**.
2. Market Watch (manager): the symbol must tick live. Dead symbol + working feed ⇒ the
   symbol name doesn't match the feed's source; fix the mapping.
3. Emergency path: throw a manual quote (Market Watch → double-click → Quotes → Send)
   — the symbol must turn live everywhere from that single quote.

**Expected-if-changed examples to spot-check:**
- set `trade mode = disabled` → new orders refuse `10001`, closes still pass
- set `stops level = 100` → SL closer than 100 points refuses `10006`
- set `volume max = 0.5` → buy 1.00 refuses `10008`

## Phase 3 — group + group symbols

1. Admin → Groups → create the QA group (e.g. `qa\real`): **currency, leverage default,
   margin mode = Hedging, margin call 100 / stop out 50 (percent), limits** (max
   orders/positions), company/server fields.
2. Add **group symbols**: the group trades only what its group-symbol records reach
   (path masks — `Forex\*` or one symbol at a time). Set the spread override, swaps,
   commission here if the group differs from the master.
3. Verify in the manager panel: Groups grid shows the group with the right margin column
   (percent shows `%`, money mode shows the currency).

| Check | Expected | If wrong |
|---|---|---|
| group symbol reach | trader sees exactly these symbols in Market Watch | order on an unlisted symbol refuses `10010 unknown symbol` |
| spread override +N points | trader's bid/ask N points wider than the raw feed | override not applied → group-symbol record missing |
| leverage | margin = vol × contract × price / leverage in the group currency | wrong margin → leverage or margin-currency conversion |

## Phase 4 — account

1. Manager → Trading Accounts → New: create the trader **in the QA group**, note the
   login and master password. Rights: enabled + trading allowed (not investor).
2. Fund it: account dialog → Balance tab → Deposit (a real deal — never SQL).
3. Verify: balance shows in the grid, the deposit deal exists, **Check Balance** is
   green (balance ≡ Σ deals — the ledger law from day one).

## Phase 5 — trade, one by one (trader terminal)

Log in as the QA trader at the trader terminal. For every step capture: the on-screen
bid/ask at the click, the result event, the DB row.

| # | Action | Expected |
|---|---|---|
| 5.1 | market **buy** 0.01 | fills at the shown **ask** (± the tick that moved), position appears live with floating P/L on **bid** |
| 5.2 | market **sell** 0.01 | fills at the shown **bid**; hedging: a second, separate position — no netting |
| 5.3 | rapid fire: 2 buys in 1 s, then 3 | each fills at its own fresh tick; refusals only where margin runs out (`10003`) |
| 5.4 | buy **limit** below market | sits as `placed`; fills at its own price when Ask ≤ price |
| 5.5 | sell **limit** above market | fills when Bid ≥ price |
| 5.6 | buy/sell **stop** | fills at market when Ask/Bid crosses the price |
| 5.7 | **stop-limit** | converts to a limit at the trigger, then fills as a limit |
| 5.8 | pending inside the stops-level distance | refused `10006` |
| 5.9 | set / modify **SL & TP** on a position | accepted beyond stops level; SL hit closes on bid (long) with comment `[sl]` |
| 5.10 | **partial close** half the volume | out-deal for the part, position volume halves, partial profit into balance |
| 5.11 | **full close** | position gone, profit = (close − open) × vol × contract to the cent |
| 5.12 | **close-by** the buy against the sell | two `out_by` deals at the older leg's open price |
| 5.13 | oversize order | refused `10008` (above max) or queued to dealer if a rule routes it |
| 5.14 | drain free margin, then order | refused `10003 not enough money` |
| 5.15 | margin call / stop out drill | warn at 100 %, forced closes at 50 %, biggest loser first, `[so]` comments, stopout_log rows |

## Phase 6 — money truth

1. **Check Balance** (manager, accounts grid) → every account green; any red means a
   money path skipped the ledger — find the deal that's missing.
2. **Check Positions** → every open position's volume/price matches its deals.
3. Hand-verify one P/L and one margin figure against the formulas
   (`docs/order-lifecycle.md §4`).
4. Journal: every action above must have its line (order requested, trade done, deal,
   balance op) with the right actor login.

---

## The findings report (the only output format)

```markdown
# QA findings — <env> — <date>

## Score
- Phases passed: X / 6
- Checks passed: NN / MM

## Missing (does not exist yet)
| # | Area | What is missing | Blocks |
|---|---|---|---|
| 1 | datafeed | no connection configured for <feed> | everything |

## Broken (exists but wrong)
| # | Area | Setting / feature | Observed | Expected | Evidence |
|---|---|---|---|---|---|
| 1 | group qa\real | spread override | bid/ask = raw feed | +2 points wider | screenshot / redis vs terminal |

## Setting-change predictions (verify after fixing)
| # | If I change | Expected result |
|---|---|---|
| 1 | group-symbol spread 0 → 2 | trader spread widens 2 points on the next tick |
| 2 | stops level 10 → 100 | SL within 100 points refuses 10006 |
| 3 | leverage 100 → 50 | margin per lot doubles; free margin halves headroom |

## Working end to end
- <one line per verified flow, with the number that proves it>
```

Rules for the report: every ✗ has evidence; every "broken" row has the exact setting to
change and the expected result after changing it; nothing marked working without a
number (a price, a deal id, a balance) behind it.
