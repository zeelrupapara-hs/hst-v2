import { useEffect, useState } from "react";
import {
  createGroupCommission,
  createGroupSymbol,
  deleteGroupCommission,
  deleteGroupSymbol,
  fetchGroupCommissions,
  fetchGroupSymbols,
} from "@/api/endpoints/groups.js";
import { TradeMode_name } from "@/constants/symbols.js";
import { GroupTabIntro } from "./GroupTabs.jsx";

/** Symbols tab: the group's scope rows, applied top-down; `*` is the default row. */
export function GroupSymbolsTab({ groupId, onOpenRow }) {
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(0);
  const [adding, setAdding] = useState("");
  const isNew = groupId === "new";

  const load = () =>
    fetchGroupSymbols(groupId).then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    if (!isNew) load();
  }, [groupId]);

  async function add() {
    const path = adding.trim();
    if (!path) return;
    const res = await createGroupSymbol(groupId, { path });
    if (!res.ok) {
      window.alert(res.message || "add failed");
      return;
    }
    setAdding("");
    load();
  }

  async function remove() {
    const row = rows?.[selected];
    if (!row) return;
    await deleteGroupSymbol(groupId, row.symbol_id);
    load();
  }

  if (isNew) {
    return (
      <GroupTabIntro>Save the group first, then configure its symbol rules here.</GroupTabIntro>
    );
  }

  return (
    <>
      <GroupTabIntro>
        Please specify the list of symbol groups with the individual trading settings available for
        the clients of this group. Rules apply top-down; the first match wins.
      </GroupTabIntro>
      <div className="df-table-panel">
        <div className="df-table-toolbar">
          <input
            type="text"
            className="df-cell-input grp-symbol-add"
            placeholder={String.raw`Forex\* or !EURUSD`}
            value={adding}
            onChange={(e) => setAdding(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && add()}
          />
          <button type="button" onClick={add}>Add</button>
          <button type="button" disabled={!rows?.length} onClick={() => onOpenRow?.(rows[selected])}>
            Edit
          </button>
          <button type="button" disabled={!rows?.length} onClick={remove}>Delete</button>
        </div>
        <div className="df-table-main">
          <table className="data-table data-table-grid df-sub-table">
            <thead>
              <tr>
                <th>Symbol</th>
                <th>Spread</th>
                <th>Trade</th>
              </tr>
            </thead>
            <tbody>
              {(rows || []).map((row, i) => (
                <tr
                  key={row.symbol_id}
                  className={i === selected ? "selected" : ""}
                  onClick={() => setSelected(i)}
                  onDoubleClick={() => onOpenRow?.(row)}
                >
                  <td>{row.path}</td>
                  <td>{row.spread_diff == null ? "Default" : row.spread_diff}</td>
                  <td>{row.trade_mode == null ? "Default" : TradeMode_name[row.trade_mode]}</td>
                </tr>
              ))}
              {rows && !rows.length && (
                <tr>
                  <td colSpan={3} className="df-empty">No symbol rules — everything inherits defaults</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

const COMMISSION_MODE = { 0: "Standard", 1: "Agent", 2: "Fee" };
const RANGE_MODE = { 0: "Volume", 1: "Turnover (money)", 2: "Turnover (volume)", 3: "Notional", 4: "Profit" };
const CHARGE_MODE = { 0: "Instant", 1: "Daily", 2: "Monthly" };

/** Commissions tab: header rows with tiers behind them. */
export function GroupCommissionsTab({ groupId }) {
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(0);
  const [adding, setAdding] = useState(null);
  const isNew = groupId === "new";

  const load = () =>
    fetchGroupCommissions(groupId).then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    if (!isNew) load();
  }, [groupId]);

  async function add() {
    if (!adding?.name?.trim()) return;
    const res = await createGroupCommission(groupId, {
      name: adding.name.trim(),
      path: adding.path || "*",
      tiers: [],
    });
    if (!res.ok) {
      window.alert(res.message || "add failed");
      return;
    }
    setAdding(null);
    load();
  }

  async function remove() {
    const row = rows?.[selected];
    if (!row) return;
    await deleteGroupCommission(groupId, row.commission_id);
    load();
  }

  if (isNew) {
    return <GroupTabIntro>Save the group first, then configure commissions here.</GroupTabIntro>;
  }

  return (
    <>
      <GroupTabIntro>
        Please specify the commissions charged for the clients of this group.
      </GroupTabIntro>
      <div className="df-table-panel">
        <div className="df-table-toolbar">
          <button type="button" onClick={() => setAdding({ name: "", path: "*" })}>Add</button>
          <button type="button" disabled={!rows?.length} onClick={remove}>Delete</button>
        </div>
        {adding && (
          <div className="df-table-toolbar">
            <input
              type="text"
              className="df-cell-input"
              placeholder="Name"
              value={adding.name}
              onChange={(e) => setAdding({ ...adding, name: e.target.value })}
            />
            <input
              type="text"
              className="df-cell-input"
              placeholder="Symbol mask"
              value={adding.path}
              onChange={(e) => setAdding({ ...adding, path: e.target.value })}
            />
            <button type="button" onClick={add}>OK</button>
            <button type="button" onClick={() => setAdding(null)}>Cancel</button>
          </div>
        )}
        <div className="df-table-main">
          <table className="data-table data-table-grid df-sub-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Symbol</th>
                <th>Type</th>
                <th>Range</th>
                <th>Charge</th>
              </tr>
            </thead>
            <tbody>
              {(rows || []).map((row, i) => (
                <tr key={row.commission_id} className={i === selected ? "selected" : ""} onClick={() => setSelected(i)}>
                  <td>{row.name}</td>
                  <td>{row.path}</td>
                  <td>{COMMISSION_MODE[row.mode] ?? row.mode}</td>
                  <td>{RANGE_MODE[row.mode_range] ?? row.mode_range}</td>
                  <td>{CHARGE_MODE[row.mode_charge] ?? row.mode_charge}</td>
                </tr>
              ))}
              {rows && !rows.length && (
                <tr>
                  <td colSpan={5} className="df-empty">No commissions</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
