import { useEffect, useMemo, useState } from "react";
import {
  fetchAllDeals,
  fetchAllOrders,
  fetchAllPositions,
} from "@/api/endpoints/trades.js";
import { DealAction_name, DealEntry_name, OrderState_name, OrderType_name } from "@/constants/trades.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { formatNs } from "@/lib/time.js";

const px = (v, digits = 5) => (v ? v.toFixed(digits) : "");
const money = (v) => (v ?? 0).toFixed(2);
const lots = (v) => (v ?? 0).toFixed(2);

/** The blotter clock, with the optional millisecond part the reference menu toggles. */
const stamp = (ns, ms) =>
  !ns ? "" : ms ? `${formatNs(ns)}.${String(Math.floor(Number(ns) / 1e6) % 1000).padStart(3, "0")}` : formatNs(ns);

/** One request-bar blotter: fetch on mount, filter by login/symbol text client-side. */
function Blotter({ fetcher, columns, keyOf, journal = true }) {
  const [rows, setRows] = useState(null);
  const [filter, setFilter] = useState("");
  const [selected, setSelected] = useState(null);
  const [menu, setMenu] = useState(null);
  const [showMs, setShowMs] = useState(false);
  const [view, setView] = useState({ grid: true, autoArrange: true });

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
      <div
        className="table-wrap"
        onContextMenu={(e) => {
          e.preventDefault();
          setMenu({ x: e.clientX, y: e.clientY });
        }}
      >
        <table
          className={`data-table${view.grid ? " data-table-grid" : ""}${view.autoArrange ? " data-table-auto" : ""}`}
        >
          <thead>
            <tr>
              {columns.map((c, i) => (
                <th key={c.id ?? i}>{c.label}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {shown.map((r, i) => (
              <tr
                key={keyOf(r)}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
              >
                {columns.map((c, j) => (
                  <td key={c.id ?? j} className={c.className?.(r)}>{c.value(r, showMs)}</td>
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
      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            { label: "Edit", icon: "edit", shortcut: "Ctrl+U", disabled: true },
            { label: "Delete", icon: "delete", shortcut: "Ctrl+D", disabled: true },
            "sep",
            { label: "Request", disabled: true },
            { label: "Restore", disabled: true },
            "sep",
            {
              label: "Copy As",
              items: [
                { label: "Lines", disabled: true },
                { label: "List of Logins", disabled: true },
                { label: "List of Tickets", disabled: true },
              ],
            },
            { label: "Export", disabled: true },
            ...(journal ? [{ label: "Journal", disabled: true }] : []),
            "sep",
            { label: "Find", shortcut: "Ctrl+F", disabled: true },
            { label: "Show Milliseconds", checked: showMs, onClick: () => setShowMs(!showMs) },
            {
              label: "Auto Arrange",
              checked: view.autoArrange,
              onClick: () => setView((v) => ({ ...v, autoArrange: !v.autoArrange })),
            },
            { label: "Grid", checked: view.grid, onClick: () => setView((v) => ({ ...v, grid: !v.grid })) },
            { label: "Columns", items: columns.map((c) => ({ label: c.menuLabel ?? c.label, checked: true, disabled: true })) },
          ]}
        />
      )}
    </div>
  );
}

const plClass = (r) => (r.profit < 0 ? "acc-loss" : "acc-profit");

export function PositionsModule() {
  return (
    <Blotter
      fetcher={fetchAllPositions}
      journal={false}
      keyOf={(r) => r.position_id}
      columns={[
        { label: "Login", value: (r) => r.login },
        { label: "Symbol", value: (r) => r.symbol },
        { label: "Ticket", value: (r) => r.position_id },
        { label: "Time", value: (r, ms) => stamp(r.time_create, ms) },
        { label: "Type", value: (r) => (r.action === 0 ? "buy" : "sell") },
        { label: "Volume", value: (r) => lots(r.volume) },
        { label: "Price", value: (r) => px(r.price_open, r.digits) },
        { label: "S / L", value: (r) => px(r.price_sl, r.digits) },
        { label: "T / P", value: (r) => px(r.price_tp, r.digits) },
        { label: "Price", menuLabel: "Price (current)", value: (r) => px(r.price_current, r.digits) },
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
        { label: "Time", value: (r, ms) => stamp(r.time_setup, ms) },
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
        { label: "Time", value: (r, ms) => stamp(r.time, ms) },
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
