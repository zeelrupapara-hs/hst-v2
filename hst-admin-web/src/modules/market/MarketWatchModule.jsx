import { useMemo, useState } from "react";
import { useMarketFeed } from "@/hooks/useMarketFeed.js";
import { useSymbols, useSymbolLiveness } from "@/hooks/useSymbols.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { price } from "@/lib/format.js";

// spread in points: the price difference scaled by the symbol's digits
const spreadPoints = (tick, digits) => Math.round((tick.ask - tick.bid) * 10 ** digits);

/** Live quotes for every symbol the feed carries, MT5's market watch. */
export function MarketWatchModule() {
  const ticks = useMarketFeed();
  const { symbols } = useSymbols();
  const { isLive } = useSymbolLiveness();
  const [filter, setFilter] = useState("");

  const digitsOf = useMemo(() => {
    const map = new Map();
    symbols.forEach((s) => map.set(s.symbol, s.digits ?? 5));
    return map;
  }, [symbols]);

  // computed every paint: the tick map mutates in place, so memoizing on it would freeze the grid
  const f = filter.trim().toUpperCase();
  const rows = symbols
    .filter((s) => !f || s.symbol.toUpperCase().includes(f))
    .map((s) => ({ symbol: s.symbol, digits: s.digits ?? 5, tick: ticks.get(s.symbol) }));

  return (
    <div className="module-root">
      <div className="table-wrap">
        <table className="data-table data-table-grid market-watch">
          <thead>
            <tr>
              <th>Symbol</th>
              <th>Bid</th>
              <th>Ask</th>
              <th>Spread</th>
              <th>High</th>
              <th>Low</th>
              <th>Change %</th>
            </tr>
          </thead>
          <tbody>
            {rows.map(({ symbol, digits, tick }) => (
              <tr key={symbol} className={isLive(symbol) ? "" : "mw-stale"}>
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="symbols" size={14} />
                    {symbol}
                  </span>
                </td>
                {tick ? (
                  <>
                    <td className={tick.dir > 0 ? "mw-up" : tick.dir < 0 ? "mw-down" : ""}>
                      {price(tick.bid, digits)}
                    </td>
                    <td className={tick.dir > 0 ? "mw-up" : tick.dir < 0 ? "mw-down" : ""}>
                      {price(tick.ask, digits)}
                    </td>
                    <td>{spreadPoints(tick, digits)}</td>
                    <td>{price(tick.high, digits)}</td>
                    <td>{price(tick.low, digits)}</td>
                    <td className={tick.changePercent < 0 ? "acc-loss" : "acc-profit"}>
                      {tick.changePercent.toFixed(2)}
                    </td>
                  </>
                ) : (
                  <><td>—</td><td>—</td><td>—</td><td>—</td><td>—</td><td>—</td></>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="filter-bar">
        <input
          type="text"
          placeholder="Filter symbols…"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
        <span className="grp-suffix">{rows.filter((r) => r.tick).length} / {rows.length} quoting</span>
      </div>
    </div>
  );
}
