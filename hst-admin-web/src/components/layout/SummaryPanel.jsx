import { useEffect, useState } from "react";
import { fetchAllPositions } from "@/api/endpoints/trades.js";
import { onEvent } from "@/api/socket.js";
import { useLiveAccounts } from "@/hooks/useLiveAccounts.js";
import { money } from "@/lib/format.js";

const REFRESH_EVERY_MS = 15000;

/** Per-symbol totals over every open position the manager's masks reach, profit live. */
export function SummaryPanel() {
  const [positions, setPositions] = useState([]);
  const live = useLiveAccounts();

  useEffect(() => {
    const load = () => fetchAllPositions().then((res) => res.ok && setPositions(res.data || []));
    load();
    const timer = setInterval(load, REFRESH_EVERY_MS);
    // a position opening or closing reshapes the totals right away
    const off = onEvent("*", (event) => /^position_/.test(event.type) && load());
    return () => {
      clearInterval(timer);
      off();
    };
  }, []);

  // group by symbol; profit prefers the engine's live number for each position
  const bySymbol = new Map();
  for (const p of positions) {
    const row = bySymbol.get(p.symbol) ?? {
      symbol: p.symbol, count: 0, buyVolume: 0, buyWeighted: 0, sellVolume: 0, sellWeighted: 0, profit: 0,
    };
    const profit = live.get(p.login)?.positions?.[p.position_id] ?? p.profit ?? 0;
    row.count += 1;
    row.profit += profit;
    if (p.action === 0) {
      row.buyVolume += p.volume;
      row.buyWeighted += p.price_open * p.volume;
    } else {
      row.sellVolume += p.volume;
      row.sellWeighted += p.price_open * p.volume;
    }
    bySymbol.set(p.symbol, row);
  }

  const rows = [...bySymbol.values()].sort((a, b) => a.symbol.localeCompare(b.symbol));
  const total = rows.reduce((sum, r) => sum + r.profit, 0);

  return (
    <table className="journal-table summary-table">
      <thead>
        <tr>
          <th>Symbol</th>
          <th>Positions</th>
          <th>Buy Volume</th>
          <th>Buy Price</th>
          <th>Sell Volume</th>
          <th>Sell Price</th>
          <th>Net Volume</th>
          <th>Profit</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r) => (
          <tr key={r.symbol}>
            <td>{r.symbol}</td>
            <td>{r.count}</td>
            <td>{r.buyVolume.toFixed(2)}</td>
            <td>{r.buyVolume ? (r.buyWeighted / r.buyVolume).toFixed(5) : ""}</td>
            <td>{r.sellVolume.toFixed(2)}</td>
            <td>{r.sellVolume ? (r.sellWeighted / r.sellVolume).toFixed(5) : ""}</td>
            <td>{(r.buyVolume - r.sellVolume).toFixed(2)}</td>
            <td className={r.profit < 0 ? "acc-loss" : "acc-profit"}>{money(r.profit)}</td>
          </tr>
        ))}
        {rows.length > 0 && (
          <tr className="summary-total">
            <td>Total</td>
            <td>{positions.length}</td>
            <td colSpan={5} />
            <td className={total < 0 ? "acc-loss" : "acc-profit"}>{money(total)}</td>
          </tr>
        )}
        {!rows.length && (
          <tr><td colSpan={8}>No open positions.</td></tr>
        )}
      </tbody>
    </table>
  );
}
