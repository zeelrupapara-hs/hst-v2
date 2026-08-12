import { useEffect, useState } from "react";
import { fetchExposure } from "@/api/endpoints/exposure.js";
import { onEvent } from "@/api/socket.js";
import { money } from "@/lib/format.js";

const REFRESH_EVERY_MS = 15000;

/** Per-currency exposure the risk manager watches, long/short bar scaled to the biggest book. */
export function ExposureTab() {
  const [data, setData] = useState({ rows: [], total_usd: 0 });

  useEffect(() => {
    const load = () =>
      fetchExposure().then((res) => {
        if (!res.ok) return;
        let d = res.data;
        if (d?.currencies) d = { rows: d.currencies, total_usd: d.totals?.net_usd ?? 0 };
        // backend may hand back a bare array or {rows, total_usd}
        if (Array.isArray(d)) {
          setData({ rows: d, total_usd: d.reduce((s, r) => s + (r.net_usd ?? 0), 0) });
        } else {
          setData({ rows: d?.rows ?? [], total_usd: d?.total_usd ?? 0 });
        }
      });
    load();
    const timer = setInterval(load, REFRESH_EVERY_MS);
    const off = onEvent("*", (event) => /^position_/.test(event.type) && load());
    return () => {
      clearInterval(timer);
      off();
    };
  }, []);

  const rows = data.rows;
  const maxAbs = Math.max(1e-9, ...rows.map((r) => Math.abs(r.net_usd ?? 0)));

  return (
    <table className="journal-table summary-table">
      <thead>
        <tr>
          <th>Currency</th>
          <th>Clients</th>
          <th>Coverage</th>
          <th>Net</th>
          <th>Rate</th>
          <th>Net USD</th>
          <th>Long / Short</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r) => {
          const usd = r.net_usd ?? 0;
          const pct = Math.min(100, (Math.abs(usd) / maxAbs) * 100);
          return (
            <tr key={r.currency}>
              <td>{r.currency}</td>
              <td>{money(r.clients)}</td>
              <td>{money(r.coverage)}</td>
              <td>{money(r.net)}</td>
              <td>{r.rate ? Number(r.rate.toPrecision(6)) : "—"}</td>
              <td className={usd < 0 ? "acc-loss" : "acc-profit"}>{money(usd)}</td>
              <td>
                <div className="exp-bar">
                  <div className="exp-bar-left">
                    {usd < 0 && <div className="exp-bar-short" style={{ width: `${pct}%` }} />}
                  </div>
                  <div className="exp-bar-right">
                    {usd >= 0 && <div className="exp-bar-long" style={{ width: `${pct}%` }} />}
                  </div>
                </div>
              </td>
            </tr>
          );
        })}
        {rows.length > 0 && (
          <tr className="summary-total">
            <td>Total</td>
            <td colSpan={4} />
            <td className={data.total_usd < 0 ? "acc-loss" : "acc-profit"}>{money(data.total_usd)}</td>
            <td />
          </tr>
        )}
        {!rows.length && (
          <tr><td colSpan={7}>No exposure.</td></tr>
        )}
      </tbody>
    </table>
  );
}
