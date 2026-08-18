# QA — trader-terminal effects of group & group-symbol settings

One kind of test only: **change one setting → do the named trade in the trader terminal
→ compare against the expected output.** Everything is observed from the terminal
(prices, refusal messages, fills, P/L); the settings are changed in the admin/manager
panel between attempts. Change **one setting at a time**, test, then put it back.

Baseline for every test (set once):
- group `qa\real` — USD, leverage 1:100, **hedging**, margin call 100 % / stop out 50 %
- group symbol `EURUSD` — digits 5, contract 100 000, vol 0.01–100 step 0.01,
  stops level 10, freeze 5, exec Instant, trade mode Full, fill FOK+IOC, spread 0
- trader account in `qa\real`, balance 10 000, no positions
- a routing rule that confirms at market (otherwise every order refuses
  `no routing rule admitted this request`)

Result of the run = the table itself with a ✓/✗ per row + the findings block at the end.

---

## A. Group-symbol settings → terminal effect

| # | Setting | Change to | Do in the terminal | Expected output |
|---|---|---|---|---|
| A1 | spread | 0 → 20 points | just watch the quote panel | bid/ask widen by 20 points vs the raw feed on the next tick; open positions' floating P/L drops by the extra spread |
| A2 | spread balance | 20 spread, balance −10 bid / +10 ask | watch the quote | bid = mid − 10 pts, ask = mid + 10 pts (skew follows the balance) |
| A3 | trade mode | Full → Disabled | buy 0.01 / close an open position | buy refused "trading is disabled"; **close still works** |
| A4 | trade mode | Full → Close-only | buy 0.01 / close | buy refused "closing only"; close fills |
| A5 | trade mode | Full → Long-only | buy / sell | buy fills; sell refused "trading is disabled" |
| A6 | volume min | 0.01 → 0.10 | buy 0.05 | refused "invalid volume" |
| A7 | volume max | 100 → 0.50 | buy 1.00 | refused "invalid volume" |
| A8 | volume step | 0.01 → 0.10 | buy 0.15 | refused "invalid volume"; 0.20 fills |
| A9 | volume limit (per symbol) | ∞ → 0.10 | buy 0.05, then buy 0.10 more | first fills; second refused "position volume limit reached" (open positions **plus pendings** count) |
| A10 | stops level | 10 → 100 points | buy with SL 50 points away | refused "stops are too close"; SL 150 points away accepted |
| A11 | stops level (pending) | 100 | place buy limit 50 points under market | refused "stops are too close"; 150 points under → placed |
| A12 | freeze level | 5 → 50 points | modify SL of an open position to within 50 points of market | refused "order is frozen"; **new** orders are exempt from freeze |
| A13 | exec mode | Instant → Request | buy at a stale/slipped price | instant refuses with a requote (bid/ask echoed); request mode walks routing as a "request" kind — with a dealer rule it queues |
| A14 | max instant volume | ∞ → 0.05 | buy 0.10 (instant mode) | not refused — silently reclassified as *request* kind; with only a confirm-market rule it still fills; with a dealer rule on requests it queues |
| A15 | deviation (profit/loss) | tighten to 1 point | buy while price is moving | requote when the fill would slip beyond 1 point; the requote carries the new bid/ask |
| A16 | fill flags | remove IOC | buy with fill policy IOC | refused "fill policy not allowed" |
| A17 | order flags | remove SL/TP flag | buy with SL set | refused "stops are too close" (SL not allowed at all) |
| A18 | order flags | remove buy-limit type | place buy limit | refused "trading is disabled" |
| A19 | expiry flags | remove "specified" | place pending with a date | refused "expiry not allowed"; GTC still placed |
| A20 | quotes time | 60 → 1 s, stop the feed | buy | refused "no price for this symbol" once the quote is older than 1 s |
| A21 | swap long/short | 0 → −5 / +1 | hold a buy and a sell overnight (or run end-of-day manually) | buy's storage −5 pts worth, sell's +1; storage lands in the position row and, on close, in the deal |
| A22 | commission | 0 → 5 USD/lot | buy 0.10, close it | 0.50 commission on the deal; balance = profit − 0.50 |
| A23 | request timeout | 30 → 5 s (with a dealer rule) | buy 2.00 (routes to desk), dealer does nothing | "timed out" after ~5 s instead of 30 |
| A24 | sessions | close today's session | buy / close | buy refused "market is closed"; **close still fills** (out-direction bypasses the clock) |

## B. Group settings → terminal effect

| # | Setting | Change to | Do in the terminal | Expected output |
|---|---|---|---|---|
| B1 | leverage | 100 → 50 | buy 0.10 EURUSD @ ~1.08 | margin doubles: ~216.70 instead of ~108.35 (margin = vol × contract × price ÷ leverage) |
| B2 | currency | USD → EUR (new group) | buy 0.10, watch P/L | all money in EUR; P/L converted profit-currency → EUR by live rate (missing cross triangulates through USD) |
| B3 | margin call level | 100 → 500 % | open until level ≈ 400 % | margin-call warning fires (once) although the account is far from danger; clears only above 525 % (call × 1.05) |
| B4 | stop out level | 50 → 90 % | withdraw until level < 90 % | forced closing starts at 90 %: margined pendings deleted first, then biggest loser, one at a time, `[so at X%]` comments |
| B5 | stop-out mode | percent → money | set stop out = 9 000 | forced closing when **equity** < 9 000 USD (currency of the group); comment `[so at 8999.xx]` without % |
| B6 | limit: max orders | 200 → 2 | place 3 pendings | third refused "too many orders" |
| B7 | limit: max positions | 100 → 2 | open 3 positions | third refused "too many orders" |
| B8 | limit: positions value | ∞ → 0.5 lots | open 0.3 + 0.3 | second refused "position volume limit reached" (whole book, all symbols) |
| B9 | limit: symbols | ∞ → 1 | open EURUSD, then GBPUSD | GBPUSD refused "position volume limit reached" |
| B10 | flag: hedge prohibit | off → on | buy 0.01, then sell 0.01 same symbol | sell refused "hedging is not allowed" |
| B11 | flag: expiration | off | pending with any non-GTC expiry | refused "expiry not allowed" |
| B12 | flag: expert trading | off | order sent with an expert id | refused "trading is disabled" |
| B13 | flag: trailing | off | expert moves an existing SL | refused "stops are too close" |
| B14 | flag: SO compensation | on | drive the account to negative balance via stop out | once flat, a `so_compensation` deal tops balance back to exactly 0 |
| B15 | free-margin mode: day profit/loss | on | close a winning position | profit goes to **blocked profit**, not balance, until end-of-day; a loss hits balance immediately |
| B16 | margin flag: check SL/TP | on | SL that would leave the rest of the book under-margined | the SL simply does not fire while closing would break coverage |
| B17 | margin flag: check process | on | order confirmed by a dealer after margin ran out | second money check at confirm time → "not enough money" |
| B18 | history limit | ∞ → 1 month | trader opens History → All history | only the last month of deals returned |
| B19 | group symbol removed | delete the EURUSD group-symbol row | buy EURUSD | refused "unknown symbol" — the group only trades what its group-symbol records reach |

## C. Account-state effects (no setting — the state itself)

| # | State | Do | Expected |
|---|---|---|---|
| C1 | free margin < need | buy beyond the free margin | refused "not enough money"; need shown in server log |
| C2 | rights: investor / read-only | any order | refused "trading is disabled"; closes too |
| C3 | account disabled | login / order | login refused or order "account disabled" |
| C4 | FIFO (older same-side position) | close the newer of two same-side positions | refused "an older position on this symbol must be closed first" |

---

## How to run one row

1. Note the baseline behaviour (do the terminal action once **before** the change).
2. Change exactly that one setting in the admin/manager panel. Wait one tick
   (settings reload over NATS — no restart).
3. Repeat the same terminal action. Record: the on-screen quote, the exact message or
   fill price, the deal row if any.
4. Mark the row ✓ (matched expected) or ✗ (evidence attached), revert the setting,
   confirm the baseline behaviour is back.

## Findings block (append after a run)

```markdown
# Run <env> <date> — A: n/24 ✓  B: n/19 ✓  C: n/4 ✓

## ✗ rows (setting did not behave as expected)
| Row | Setting changed | Observed | Expected | Evidence |
|---|---|---|---|---|
| A10 | stops level 10 → 100 | SL at 50 pts accepted, deal #123 | refuse "stops are too close" | screenshot + deal id |

## Settings that only took effect after a restart (should apply live)
| Row | Setting | How it was proven |
|---|---|---|
| A21 | swap long −5 | no storage after end-of-day; appeared only after core restart |

## Settings with no visible effect at all (dead setting?)
| Row | Setting | What was tried |
|---|---|---|
| B18 | history limit 1 month | History tab still returns all deals on "All history" |

## ✓ summary
- A: rows … verified, each with the message/price/deal id recorded
- B: rows … verified
- C: rows … verified
- Baseline restored after every row: yes/no (list any row left changed)
```
