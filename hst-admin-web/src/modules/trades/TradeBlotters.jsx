import { useEffect, useMemo, useState } from "react";
import {
  fetchAllDeals,
  fetchAllOrders,
  fetchAllPositions,
} from "@/api/endpoints/trades.js";
import { DealAction_name, DealEntry_name, OrderState_name, OrderType_name } from "@/constants/trades.js";
import { formatNs } from "@/lib/time.js";

const px = (v, digits = 5) => (v ? v.toFixed(digits) : "");
const money = (v) => (v ?? 0).toFixed(2);
const lots = (v) => (v ?? 0).toFixed(2);

/** One request-bar blotter: fetch on mount, filter by login/symbol text client-side. */
function Blotter({ fetcher, columns, keyOf }) {
  const [rows, setRows] = useState(null);
  const [filter, setFilter] = useState("");

  useEffect(() => {
    fetcher().then((res) => setRows(res.ok ? res.data || [] : []));
  }, [fetcher]);

  const shown = useMemo(() => {
    if (!filter.trim()) return rows || [];
    const f = filter.trim().toLowerCase();
    return (rows || []).filter(
      (r) => String(r.login).includes(f) || (r.symbol || "").toLowerCase().includes(f),
    );
  }, [rows, filter]);

  return (
    <div className="module-root">
      <div className="table-wrap">
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              {columns.map((c, i) => (
                <th key={c.id ?? i}>{c.label}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {shown.map((r) => (
              <tr key={keyOf(r)}>
                {columns.map((c, i) => (
                  <td key={c.id ?? i} className={c.className?.(r)}>{c.value(r)}</td>
                ))}
              </tr>
            ))}
            {rows !== null && !shown.length && (
              <tr>
                <td colSpan={columns.length} className="df-empty">No records</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="filter-bar">
        <span>Request:</span>
        <input
          type="text"
          placeholder="login or symbol, empty = all"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
        <span className="grp-suffix">{rows ? `${shown.length} of ${rows.length}` : "Loading…"}</span>
      </div>
    </div>
  );
}

const plClass = (r) => (r.profit < 0 ? "acc-loss" : "acc-profit");

export function PositionsModule() {
  return (
    <Blotter
      fetcher={fetchAllPositions}
      keyOf={(r) => r.position_id}
      columns={[
        { label: "Login", value: (r) => r.login },
        { label: "Symbol", value: (r) => r.symbol },
        { label: "Ticket", value: (r) => r.position_id },
        { label: "Time", value: (r) => formatNs(r.time_create) },
        { label: "Type", value: (r) => (r.action === 0 ? "buy" : "sell") },
        { label: "Volume", value: (r) => lots(r.volume) },
        { label: "Price", value: (r) => px(r.price_open, r.digits) },
        { label: "S / L", value: (r) => px(r.price_sl, r.digits) },
        { label: "T / P", value: (r) => px(r.price_tp, r.digits) },
        { label: "Price", value: (r) => px(r.price_current, r.digits) },
        { label: "Profit", value: (r) => money(r.profit), className: plClass },
      ]}
    />
  );
}

export function OrdersModule() {
  return (
    <Blotter
      fetcher={fetchAllOrders}
      keyOf={(r) => r.order_id}
      columns={[
        { label: "Login", value: (r) => r.login },
        { label: "Symbol", value: (r) => r.symbol },
        { label: "Ticket", value: (r) => r.order_id },
        { label: "Time", value: (r) => formatNs(r.time_setup) },
        { label: "Type", value: (r) => OrderType_name[r.type] ?? r.type },
        { label: "Volume", value: (r) => lots(r.volume_initial ?? r.volume) },
        { label: "Price", value: (r) => px(r.price_order, r.digits) },
        { label: "S / L", value: (r) => px(r.price_sl, r.digits) },
        { label: "T / P", value: (r) => px(r.price_tp, r.digits) },
        { label: "State", value: (r) => OrderState_name[r.state] ?? r.state },
      ]}
    />
  );
}

export function DealsModule() {
  return (
    <Blotter
      fetcher={fetchAllDeals}
      keyOf={(r) => r.deal_id}
      columns={[
        { label: "Login", value: (r) => r.login },
        { label: "Deal", value: (r) => r.deal_id },
        { label: "Order", value: (r) => r.order_id || "" },
        { label: "Time", value: (r) => formatNs(r.time) },
        { label: "Symbol", value: (r) => r.symbol },
        { label: "Action", value: (r) => DealAction_name[r.action] ?? r.action },
        { label: "Entry", value: (r) => DealEntry_name[r.entry] ?? r.entry },
        { label: "Volume", value: (r) => lots(r.volume) },
        { label: "Price", value: (r) => px(r.price, r.digits) },
        { label: "Profit", value: (r) => money(r.profit), className: plClass },
      ]}
    />
  );
}
