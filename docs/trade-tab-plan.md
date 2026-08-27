# Symbol "Trade" tab — findings and plan

Sources: MT5 docs `symbol_settings_trade.htm`, `profit_calculation.htm`, `group_symbols_*.htm`, IMTConSymbol / IMTConGroupSymbol / IMTPosition / IMTDeal PDFs; current code in hst-core, hst-server, hst-admin-web, hst-web.

## What MT5 does (per field) vs what we do

| Field | MT5 rule | Ours today | Verdict |
|---|---|---|---|
| Contract size | `double`; positions, orders **and deals each store their own copy at open** (`IMTPosition::ContractSize`, `IMTDeal::ContractSize`); a later symbol change never re-values open positions | DB `NUMERIC(20,8)` fine, positions/orders/deals already **store** it — but `ProfitFor`/`marginBase` read the **live** `Rules.ContractSize`, so an admin edit instantly re-values every open position (this is the "realtime change applied to client" you saw). Admin input can't take `0.01`: plain `Field` runs `Number(v)` on every keystroke, so `"0."` collapses to `0` | **Bug ×2** |
| Limit & stop level | channel in points, SL/TP/pending refused inside it | `checkStops` | OK |
| Calculation | 15 calc modes; Forex/CFD/Futures formulas; exchange modes trigger on Last | `money.go` covers Forex, Forex-no-leverage, CFD, CFD-leverage, CFD-index, Futures (tick value/size), exchange stocks; formula per doc | OK — re-verified below |
| Freeze level | no SL/TP/pending edit, no close/volume change of a position inside it | `checkFreeze` on edits; **close is not frozen** | minor gap |
| Trade mode | Disabled = trade requests refused only; quotes/charts are governed by *quoting sessions*, not trade mode. Close-only: only closing. Long/short only | core correct (close allowed when disabled). hst-web greys **chart** and kills ticks for Disabled/Close-only; long/short-only not reflected (both buttons enabled) | **UI bug** |
| Max quote delay | trading auto-disabled after N s without ticks | `checkQuote` | OK |
| GTC mode | 0 GTC; **1 = at day change delete all pendings and clear SL/TP of positions; 2 = delete pendings, keep SL/TP** | stored + shown, **never enforced** — EOD only expires orders whose own type is "Day" | **Not implemented** |
| Convert profit | Forex-only bit in `trade_flags`. *by deal* (default): profit→deposit at the pair rate in the closing direction (Ask when closing a buy, Bid when closing a sell), profitability ignored. *by market*: Bid if the deal is profitable, Ask if losing. Cross via USD when no direct pair | `trade_flags` **never loaded by core**; conversion always direction-based = "by deal" | explain + implement |
| Filling | bitmask FOK=1 IOC=2 BOC=4; Return implicit; terminal offers the flag only for **market orders**; instant/request execution forces FOK; BOC only for limit/stop-limit | core validates; hst-web has **no filling dropdown**, hardcodes 0 | **Missing in UI** |
| Trading signals | signals-copying gate | no signals product | skip |
| Expiration | bitmask GTC=1 Day=2 Specified=4 SpecifiedDay=8; terminal shows only ticked | core validates; hst-web dropdown **static** | **UI bug** |
| Orders | bitmask Market=1 Limit=2 Stop=4 StopLimit=8 SL=16 TP=32 CloseBy=64; **0 = nothing allowed** | core treats `0` as "no restriction" (same for fill/expir flags); hst-web dropdown static, no stop-limit, SL/TP always shown | **Core + UI bug** |
| Tick size / value | profit/margin/swaps for Futures & CFD-index (`(Close−Open)·Lots·TickValue/TickSize`); tick size also rounds feed prices when no DOM; deals snapshot both | used for futures profit and points-swap only; deals store them; **positions don't**; no feed rounding | partial |
| Volume min/step/max/limit | integer units (1/10000 lot; Ext 1/1e8); min not applied on close; max also caps stop-out closes; limit = one-direction cap | all enforced; Ext canonical. **Terminal never refreshes** — `symbol_updated` WS event is published and the trader socket is subscribed, but hst-web has no handler and fetches `/symbols` once at boot | **UI bug** |
| Group overrides | per-group record per symbol path, `X`/`XDefault` pairs, "Use default …" = inherit from symbol. Overridable: trade mode, fill/expir/order flags, stops/freeze, volumes, exec, spread, margin, swaps, IE/RE. **Not** overridable: contract size, tick size/value, calc mode, GTC mode, quote delay, trade flags | `NULL` = inherit, first matching `config_index` row wins, resolved values are what the terminal receives via core | OK as designed |

Re-verified calculation formulas (`money.go`): Forex margin `lots·CS/leverage`, CFD `lots·CS·price`, CFD-leverage `/leverage`, CFD-index/futures `lots·MarginInitial` (fallback CS·price), profit `diff·CS·lots` (futures `diff/TickSize·TickValue·lots`), × RateProfit. Matches the doc. The only correctness hole is *which* contract size is used (live vs stored) — fixed by item 1.

## Plan (ordered by risk to the core, lowest first)

Every core change is done by substituting values **into the existing `Rules` object at the call site**, not by changing formulas, signatures or the settings resolution. The symbol → group-override → Rules flow stays exactly as it is.

1. **Contract size is fixed at open (core, small).**
   Positions/orders/deals already carry `contract_size`. Add one helper `rulesForPosition(r, p)` returning a shallow copy of `Rules` with `ContractSize` (and `TickSize`/`TickValue`) taken from the position when set; use it at the ~12 profit/margin call sites (`positions.go`, `floating_margin.go`, `positions_hedge.go`, `stopout.go`, `CalcPosition`). New positions keep using live rules, so a symbol edit only affects trades opened after it — exactly MT5.
   Migration: add `tick_size`, `tick_value` to `hst.positions` (deals already have them), copied at open. Existing positions with `contract_size = 0` (older rows) fall back to live rules.
   Admin: contract-size input keeps the raw text while typing (same buffering `FmtInput` already does for volumes) so `0.01` can be entered; API validation `gt=0` on patch so `0` cannot be written by mistake.

2. **Terminal refreshes symbol settings live (hst-web, small).**
   Handle the existing `symbol_updated` WS event → debounce 500 ms → refetch `/symbols` and merge into `useSymbolStore` (prices untouched). Min/step/max, trade mode, flags, digits… all update without reload.

3. **Order/filling/expiration flags drive the order form (hst-web + tiny core fix).**
   - Core: treat `order_flags = 0`, `expir_flags = 0`, `fill_flags = 0` as "nothing allowed" (MT5 `*_FLAGS_NONE`). Migration sets the MT5 defaults (127 / 15 / 3) on existing rows that are 0 so nothing breaks; the create form already uses those defaults.
   - hst-web order form: order-type list built from `order_flags` (market, limit, stop, stop-limit); expiration list from `expir_flags`; SL/TP inputs hidden when their bits are off; new **Filling** dropdown from `fill_flags`, shown for market orders only, forced to FOK for instant/request execution, BOC offered only on limit/stop-limit (per doc); one-click sends the first allowed policy. Long-only/short-only disable the Sell/Buy button respectively.

4. **Trade mode Disabled / Close-only keep the chart alive (hst-web, small).**
   `isSymbolLive` splits into `canTrade` (buttons, one-click) and `hasQuotes` (chart, dot). Chart stays live; Buy/Sell disabled; Close button still allowed (core already permits it).

5. **GTC modes 1 and 2 (core EOD, contained).**
   In `EndOfDayProcess`, per symbol with `gtc_mode = 1`: cancel every pending order and clear SL/TP on open positions (journal + WS events through the existing cancel/modify paths); `gtc_mode = 2`: cancel pendings only. Uses the existing `ExpireOrders` and SL/TP modify functions — no new state.

6. **Convert profit (core, contained).**
   Load `trade_flags` into `model.Symbol`/`Rules`. In `RateProfit` at **close**, Forex only: bit clear → keep today's direction-based rate ("by deal"); bit set → Bid of the conversion pair if the deal is profitable, Ask if losing ("by market"). Floating profit unchanged. Explanation for you: it only changes *which side of the conversion pair* is used when a Forex deal's profit is turned into the deposit currency; with USD deposits and USD-quoted pairs it makes no difference at all.

7. **Freeze level on close (core, one guard).** `ClosePosition` refuses with `RetTradeFrozen` when market is within `freeze_level` points of the position's SL/TP, per doc.

8. **Tick size rounding of feed prices (hst-quote, optional).** Round bid/ask to `tick_size` when the symbol has no market depth. Low priority; only matters for exchange-style symbols.

9. **Housekeeping found on the way:** `settings.go:279` defaults IE slippage to `spread_diff` (wrong base field); trading-signals flags stay dead (not offered).

## Test plan (Go e2e harness, per role)
- manager: set contract size 1 → trader opens 1 lot → manager sets contract size 200 → trader's position profit/margin unchanged, a **new** position uses 200.
- trader: min/step/max changed by manager → order form limits update within 1 s without reload; volume outside → rejected.
- manager: order flags = Limit only → terminal shows only Buy/Sell Limit, no SL/TP fields; core rejects a market order.
- manager: trade mode Disabled → chart still ticks, Buy/Sell disabled, close works; Long only → Sell disabled.
- EOD with gtc 1 / 2 on a symbol with pendings + SL/TP.
- Forex close with convert-profit by deal vs by market on a non-USD deposit account.

## Not changing
- Settings resolution (`Store.For`, `pick`, NULL-inherit, first-match) — correct per MT5.
- Profit/margin formulas — verified.
- Anything under Calculation, Limit&stop, Max quote delay — working.
