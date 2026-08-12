import { useEffect, useState } from "react";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { request } from "@/api/client.js";
import { DealAction_name, DealEntry_name } from "@/constants/trades.js";
import { formatNs } from "@/lib/time.js";
import { money } from "@/lib/format.js";

const DAY = 86400;

const PERIODS = [
  { value: "today", label: "Today" },
  { value: "3days", label: "Last 3 days" },
  { value: "week", label: "Last week" },
  { value: "month", label: "Last month" },
  { value: "3months", label: "Last 3 months" },
  { value: "6months", label: "Last 6 months" },
  { value: "all", label: "All history" },
];

const since = (period) => {
  const now = Math.floor(Date.now() / 1000);
  switch (period) {
    case "today": return now - DAY;
    case "3days": return now - 3 * DAY;
    case "week": return now - 7 * DAY;
    case "month": return now - 30 * DAY;
    case "3months": return now - 90 * DAY;
    case "6months": return now - 180 * DAY;
    default: return 0;
  }
};

const toInput = (unix) => new Date(unix * 1000).toISOString().slice(0, 10);

/** The account's deal history, MT5's History tab: a period, a range, a Request button. */
export function AccountHistoryTab({ login }) {
  const [period, setPeriod] = useState("month");
  const [from, setFrom] = useState(toInput(since("month")));
  const [to, setTo] = useState(toInput(Math.floor(Date.now() / 1000)));
  const [rows, setRows] = useState(null);

  function load(fromSec, toSec) {
    const range = `from=${fromSec}&to=${toSec + DAY}`;
    request(`/api/v1/deals/accounts/${login}?limit=500&${range}`).then(
      (res) => setRows(res.ok ? res.data || [] : []),
    );
  }

  useEffect(() => {
    load(since(period), Math.floor(Date.now() / 1000));
  }, [login]);

  function pick(p) {
    setPeriod(p);
    const f = since(p);
    setFrom(toInput(f));
    setTo(toInput(Math.floor(Date.now() / 1000)));
    load(f, Math.floor(Date.now() / 1000));
  }

  const requestRange = () =>
    load(Math.floor(new Date(from).getTime() / 1000), Math.floor(new Date(to).getTime() / 1000));

  return (
    <div className="history-tab">
      <div className="table-wrap history-table">
        <table className="data-table data-table-grid">
          <thead>
            <tr>
              <th>Time</th><th>Deal</th><th>Order</th><th>Symbol</th><th>Type</th>
              <th>Entry</th><th>Volume</th><th>Price</th><th>Profit</th><th>Comment</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((r) => (
              <tr key={r.deal_id}>
                <td>{formatNs(r.time)}</td>
                <td>{r.deal_id}</td>
                <td>{r.order_id || ""}</td>
                <td>{r.symbol}</td>
                <td>{DealAction_name[r.action] ?? r.action}</td>
                <td>{DealEntry_name[r.entry] ?? r.entry}</td>
                <td>{(r.volume ?? 0).toFixed(2)}</td>
                <td>{r.price ? r.price.toFixed(r.digits ?? 5) : ""}</td>
                <td className={r.profit < 0 ? "acc-loss" : "acc-profit"}>{money(r.profit)}</td>
                <td>{r.comment}</td>
              </tr>
            ))}
            {rows !== null && !rows.length && (
              <tr><td colSpan={10} className="df-empty">No history for the period</td></tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="filter-bar history-bar">
        <PropSelect value={period} options={PERIODS} onChange={pick} />
        <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
        <span>-</span>
        <input type="date" value={to} onChange={(e) => setTo(e.target.value)} />
        <button type="button" onClick={requestRange}>Request</button>
        <span className="grp-suffix">{rows ? `${rows.length} deals` : "Loading…"}</span>
      </div>
    </div>
  );
}
