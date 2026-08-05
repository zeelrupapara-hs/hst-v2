import { useEffect, useState } from "react";
import {
  deleteGroupCommission,
  deleteGroupSymbol,
  fetchGroupCommissions,
  fetchGroupSymbols,
  updateGroupSymbol,
} from "@/api/endpoints/groups.js";
import { ContextMenu, listMenuHead } from "@/components/ui/ContextMenu.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { TradeMode_name } from "@/constants/symbols.js";
import { GroupTabIntro } from "./GroupTabs.jsx";
import { GroupSymbolDialog } from "./GroupSymbolDialog.jsx";
import { CommissionDialog } from "./CommissionDialog.jsx";

/** Symbols tab: the group's scope rows, applied top-down; `*` is the default row. */
export function GroupSymbolsTab({ groupId }) {
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(0);
  const [editing, setEditing] = useState(null);
  const [menu, setMenu] = useState(null);
  const isNew = groupId === "new";

  const load = () =>
    fetchGroupSymbols(groupId).then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    if (!isNew) load();
  }, [groupId]);

  // Order decides which rule wins, so a move swaps the two rows' config_index.
  async function move(delta) {
    const a = rows?.[selected];
    const b = rows?.[selected + delta];
    if (!a || !b) return;
    await updateGroupSymbol(groupId, a.symbol_id, { config_index: b.config_index });
    await updateGroupSymbol(groupId, b.symbol_id, { config_index: a.config_index });
    setSelected(selected + delta);
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
        Please set up individual parameters of symbols trade for the group.
      </GroupTabIntro>
      <div className="df-table-panel">
        <div className="df-table-toolbar grp-sym-buttons">
          <button type="button" disabled={selected <= 0} onClick={() => move(-1)}>Up</button>
          <button
            type="button"
            disabled={!rows?.length || selected >= rows.length - 1}
            onClick={() => move(1)}
          >
            Down
          </button>
          <span className="grp-sym-buttons-gap" />
          <button type="button" onClick={() => setEditing({ row: null })}>Add</button>
          <button type="button" disabled={!rows?.length} onClick={() => setEditing({ row: rows[selected] })}>
            Edit
          </button>
          <button type="button" disabled={!rows?.length} onClick={remove}>Delete</button>
        </div>
        <div
          className="df-table-main"
          onContextMenu={(e) => {
            e.preventDefault();
            setMenu({ x: e.clientX, y: e.clientY });
          }}
        >
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
                  onDoubleClick={() => setEditing({ row })}
                >
                  <td>
                    <span className="grp-sym-cell">
                      <Icon id="symbols-tree" /> {row.path}
                    </span>
                  </td>
                  <td>{row.spread_diff == null ? "Default" : `${row.spread_diff} pt`}</td>
                  <td>{row.trade_mode == null ? "Default" : TradeMode_name[row.trade_mode]}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            ...listMenuHead({
              onAdd: () => setEditing({ row: null }),
              onEdit: () => setEditing({ row: rows[selected] }),
              onDelete: remove,
              hasSelection: !!rows?.length,
            }),
            "sep",
            { label: "Up", disabled: selected <= 0, onClick: () => move(-1) },
            { label: "Down", disabled: !rows?.length || selected >= rows.length - 1, onClick: () => move(1) },
          ]}
        />
      )}
      {editing && (
        <GroupSymbolDialog
          groupId={groupId}
          row={editing.row}
          onClose={() => setEditing(null)}
          onSaved={load}
        />
      )}
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
  const [editing, setEditing] = useState(null);
  const isNew = groupId === "new";

  const load = () =>
    fetchGroupCommissions(groupId).then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    if (!isNew) load();
  }, [groupId]);

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
          <button type="button" onClick={() => setEditing({ commission: null })}>Add</button>
          <button
            type="button"
            disabled={!rows?.length}
            onClick={() => setEditing({ commission: rows[selected] })}
          >
            Edit
          </button>
          <button type="button" disabled={!rows?.length} onClick={remove}>Delete</button>
        </div>
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
                <tr
                  key={row.commission_id}
                  className={i === selected ? "selected" : ""}
                  onClick={() => setSelected(i)}
                  onDoubleClick={() => setEditing({ commission: row })}
                >
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
      {editing && (
        <CommissionDialog
          groupId={groupId}
          commission={editing.commission}
          onClose={() => setEditing(null)}
          onSaved={load}
        />
      )}
    </>
  );
}
