import { useEffect, useMemo, useState } from "react";
import {
  checkPositions,
  deleteDeal,
  deletePosition,
  fetchAllDeals,
  fetchAllOrders,
  fetchAllPositions,
  fixPosition,
  updateDeal,
} from "@/api/endpoints/trades.js";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { DealAction_name, DealEntry_name, OrderState_name, OrderType_name } from "@/constants/trades.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { formatNs } from "@/lib/time.js";
import { useLiveAccounts } from "@/hooks/useLiveAccounts.js";
import { useMarketFeed } from "@/hooks/useMarketFeed.js";
import { BulkCloseDialog } from "./BulkCloseDialog.jsx";

const px = (v, digits = 5) => (v ? v.toFixed(digits) : "");
const money = (v) => (v ?? 0).toFixed(2);
const lots = (v) => (v ?? 0).toFixed(2);

/** The blotter clock, with the optional millisecond part the reference menu toggles. */
const stamp = (ns, ms) =>
  !ns ? "" : ms ? `${formatNs(ns)}.${String(Math.floor(Number(ns) / 1e6) % 1000).padStart(3, "0")}` : formatNs(ns);

/** One request-bar blotter: fetch on mount, filter by login/symbol text client-side.
 * liveRow, when given, patches each row with live values just before it is painted. */
function Blotter({ fetcher, columns, keyOf, journal = true, liveRow, rowClass, onEdit, onDelete, extraItems }) {
  const [rows, setRows] = useState(null);
  const [filter, setFilter] = useState("");
  const [selected, setSelected] = useState(null);
  const [menu, setMenu] = useState(null);
  const [showMs, setShowMs] = useState(false);
  const [view, setView] = useState({ grid: true, autoArrange: true });

  const reload = () => fetcher().then((res) => setRows(res.ok ? res.data || [] : []));

  useEffect(() => {
    reload();
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
            {shown.map((raw, i) => {
              const r = liveRow ? liveRow(raw) : raw;
              return (
                <tr
                  key={keyOf(r)}
                  className={`${selected === i ? "selected" : ""}${rowClass ? rowClass(r) : ""}`.trim()}
                  onClick={() => setSelected(i)}
                  onContextMenu={() => setSelected(i)}
                >
                  {columns.map((c, j) => (
                    <td key={c.id ?? j} className={c.className?.(r)}>{c.value(r, showMs)}</td>
                  ))}
                </tr>
              );
            })}
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
            { label: "Edit", icon: "edit", shortcut: "Ctrl+U", disabled: !onEdit || selected == null, onClick: () => onEdit(shown[selected], reload) },
            { label: "Delete", icon: "delete", shortcut: "Ctrl+D", disabled: !onDelete || selected == null, onClick: () => onDelete(shown[selected], reload) },
            ...(extraItems ? extraItems(selected != null ? shown[selected] : null, reload) : []),
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

/** The MT5 deal editor: rewrite the ledger row, then check positions and balance. */
function DealDialog({ deal, onClose, onSaved }) {
  const [form, setForm] = useState({
    volume: String(deal.volume ?? 0),
    price: String(deal.price ?? 0),
    profit: String(deal.profit ?? 0),
    storage: String(deal.storage ?? 0),
    commission: String(deal.commission ?? 0),
    fee: String(deal.fee ?? 0),
    comment: deal.comment ?? "",
  });
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(deal.deal_id);
  const close = useDialogStack(onClose);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));

  async function handleOk() {
    const res = await updateDeal(deal.deal_id, {
      volume: Number(form.volume) || 0,
      price: Number(form.price) || 0,
      profit: Number(form.profit) || 0,
      storage: Number(form.storage) || 0,
      commission: Number(form.commission) || 0,
      fee: Number(form.fee) || 0,
      comment: form.comment,
    });
    if (!res.ok) {
      setError(res.message || "update failed");
      return;
    }
    onSaved();
    close();
  }

  return (
    <DialogOverlay>
      <div className="dialog-positioner" style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}>
        <SettingsDialog
          draggable
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          width={420}
          height={430}
          className="balance-dialog"
          title={`Deal: #${deal.deal_id} — ${deal.login} ${deal.symbol || ""}`}
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="balance" size={40} />
              </span>
              <p>
                Deals are the ledger. After an edit, run Check on positions and
                Check Balance on the account to fix what drifted.
              </p>
            </div>
            <div className="form-grid">
              <label>Volume</label>
              <input type="text" value={form.volume} onChange={set("volume")} autoFocus />
              <label>Price</label>
              <input type="text" value={form.price} onChange={set("price")} />
              <label>Profit</label>
              <input type="text" value={form.profit} onChange={set("profit")} />
              <label>Swap</label>
              <input type="text" value={form.storage} onChange={set("storage")} />
              <label>Commission</label>
              <input type="text" value={form.commission} onChange={set("commission")} />
              <label>Fee</label>
              <input type="text" value={form.fee} onChange={set("fee")} />
              <label>Comment</label>
              <input type="text" maxLength={64} value={form.comment} onChange={set("comment")} />
            </div>
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

export function PositionsModule() {
  const live = useLiveAccounts();
  const ticks = useMarketFeed();
  const [checks, setChecks] = useState(null);
  const [bulk, setBulk] = useState(null);

  // the engine's summary pairs carry each position's live profit; the feed carries the price
  const liveRow = (r) => {
    const profit = live.get(r.login)?.positions?.[r.position_id];
    const tick = ticks.get(r.symbol);
    const current = tick ? (r.action === 0 ? tick.bid : tick.ask) : r.price_current;
    return { ...r, profit: profit ?? r.profit, price_current: current };
  };

  // the audit: every open position recomputed from its deals, mismatches red
  async function runCheck() {
    const res = await checkPositions();
    if (!res.ok) {
      window.alert(res.message || "check failed");
      return;
    }
    setChecks(new Map(res.data.map((r) => [r.position_id, r])));
    const invalid = res.data.filter((r) => !r.ok).length;
    window.alert(invalid
      ? `${invalid} of ${res.data.length} positions do not match their deals`
      : `All ${res.data.length} positions match their deals`);
  }

  async function runFix(row, reload) {
    const res = await fixPosition(row.position_id);
    if (!res.ok) {
      window.alert(res.message || "fix failed");
      return;
    }
    setChecks((prev) => {
      const next = new Map(prev ?? []);
      next.set(row.position_id, res.data);
      return next;
    });
    reload();
  }

  async function runDelete(row, reload) {
    if (!window.confirm(`Delete position #${row.position_id} of ${row.login} without a deal?`)) return;
    const res = await deletePosition(row.position_id);
    if (!res.ok) {
      window.alert(res.message || "delete failed");
      return;
    }
    reload();
  }

  const badRow = (r) => checks?.get(r.position_id)?.ok === false;

  return (
    <>
    <Blotter
      fetcher={fetchAllPositions}
      journal={false}
      liveRow={liveRow}
      keyOf={(r) => r.position_id}
      rowClass={(r) => (badRow(r) ? " acc-invalid" : "")}
      extraItems={(row, reload) => [
        "sep",
        { label: "Check", onClick: runCheck },
        { label: "Fix Position", disabled: !row || !badRow(row), onClick: () => runFix(row, reload) },
        { label: "Delete Position", disabled: !row, onClick: () => runDelete(row, reload) },
        { label: "Bulk Close…", onClick: () => setBulk({ reload, symbol: row?.symbol }) },
      ]}
      columns={[
        { label: "Login", value: (r) => r.login },
        { label: "Symbol", value: (r) => r.symbol },
        { label: "Ticket", value: (r) => r.position_id },
        { label: "Time", value: (r, ms) => stamp(r.time_create, ms) },
        { label: "Type", value: (r) => (r.action === 0 ? "buy" : "sell") },
        {
          label: "Volume",
          value: (r) => (badRow(r)
            ? `${lots(r.volume)} / ${lots(checks.get(r.position_id).valid_volume)}`
            : lots(r.volume)),
        },
        {
          label: "Price",
          value: (r) => (badRow(r)
            ? `${px(r.price_open, r.digits)} / ${px(checks.get(r.position_id).valid_price, r.digits)}`
            : px(r.price_open, r.digits)),
        },
        { label: "S / L", value: (r) => px(r.price_sl, r.digits) },
        { label: "T / P", value: (r) => px(r.price_tp, r.digits) },
        { label: "Price", menuLabel: "Price (current)", value: (r) => px(r.price_current, r.digits) },
        { label: "Profit", value: (r) => money(r.profit), className: plClass },
      ]}
    />
    {bulk && <BulkCloseDialog initialSymbol={bulk.symbol} onClose={() => setBulk(null)} onDone={() => bulk.reload()} />}
    </>
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
  const [editing, setEditing] = useState(null);

  async function runDelete(row, reload) {
    if (!window.confirm(`Delete deal #${row.deal_id} of ${row.login}? Positions and balance must be checked after.`)) return;
    const res = await deleteDeal(row.deal_id);
    if (!res.ok) {
      window.alert(res.message || "delete failed");
      return;
    }
    reload();
  }

  return (
    <>
    <Blotter
      fetcher={fetchAllDeals}
      keyOf={(r) => r.deal_id}
      onEdit={(row, reload) => setEditing({ row, reload })}
      onDelete={runDelete}
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
    {editing && (
      <DealDialog
        deal={editing.row}
        onClose={() => setEditing(null)}
        onSaved={() => editing.reload()}
      />
    )}
    </>
  );
}
