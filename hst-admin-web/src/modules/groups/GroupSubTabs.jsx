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
    if (!a || !b || selected === "add") return;
    await updateGroupSymbol(groupId, a.symbol_id, { config_index: b.config_index });
    await updateGroupSymbol(groupId, b.symbol_id, { config_index: a.config_index });
    setSelected(selected + delta);
    load();
  }

  async function remove() {
    if (selected === "add") return;
    const row = rows?.[selected];
    if (!row) return;
    await deleteGroupSymbol(groupId, row.symbol_id);
    setSelected(Math.max(0, selected - 1));
    load();
  }

  if (isNew) {
    return (
      <GroupTabIntro>Save the group first, then configure its symbol rules here.</GroupTabIntro>
    );
  }

  const hasRows = (rows?.length ?? 0) > 0;
  const addRowActive = selected === "add";
  const rowIndex = addRowActive ? -1 : selected;
  const selectedRow = hasRows && rowIndex >= 0 ? rows[rowIndex] : null;

  return (
    <>
      <div className="grp-sym-tab">
        <GroupTabIntro>
          Please set up individual parameters of symbols trade for the group.
        </GroupTabIntro>
        <div className="df-table-panel">
          <div className="df-table-toolbar grp-sym-buttons">
            <div className="grp-sym-buttons-move">
              <button type="button" disabled={addRowActive || rowIndex <= 0} onClick={() => move(-1)}>
                Up
              </button>
              <button
                type="button"
                disabled={addRowActive || !hasRows || rowIndex >= rows.length - 1}
                onClick={() => move(1)}
              >
                Down
              </button>
            </div>
            <div className="grp-sym-buttons-actions">
              <button type="button" onClick={() => setEditing({ row: null })}>Add</button>
              <button
                type="button"
                disabled={addRowActive || !selectedRow}
                onClick={() => setEditing({ row: selectedRow })}
              >
                Edit
              </button>
              <button type="button" disabled={addRowActive || !selectedRow} onClick={remove}>
                Delete
              </button>
            </div>
          </div>
          <div
            className="df-table-main grp-sym-table"
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
                    className={!addRowActive && i === rowIndex ? "selected" : ""}
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
                <tr
                  className={`df-add-row${addRowActive ? " selected" : ""}`}
                  onClick={() => setSelected("add")}
                  onDoubleClick={() => setEditing({ row: null })}
                >
                  <td colSpan={3}>
                    <span className="df-add-plus" aria-hidden="true">
                      +
                    </span>
                    click to add...
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
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
              onEdit: () => selectedRow && setEditing({ row: selectedRow }),
              onDelete: remove,
              hasSelection: !!selectedRow,
            }),
            "sep",
            { label: "Up", disabled: addRowActive || rowIndex <= 0, onClick: () => move(-1) },
            {
              label: "Down",
              disabled: addRowActive || !hasRows || rowIndex >= rows.length - 1,
              onClick: () => move(1),
            },
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

import {
  COMMISSION_CHARGE,
  COMMISSION_MODE_SHORT,
  COMMISSION_RANGE,
} from "@/lib/commissionConfig.js";

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
                  <td>{COMMISSION_MODE_SHORT[row.mode] ?? row.mode}</td>
                  <td>{COMMISSION_RANGE[row.mode_range] ?? row.mode_range}</td>
                  <td>{COMMISSION_CHARGE[row.mode_charge] ?? row.mode_charge}</td>
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
