import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { SymbolTreeSelect } from "@/components/ui/SymbolTreeSelect.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createHoliday,
  deleteHoliday,
  fetchHolidays,
  reorderHolidays,
  updateHoliday,
} from "@/api/endpoints/holidays.js";
import { minutesToTime } from "@/lib/symbolSessions.js";

const pad = (n) => String(n).padStart(2, "0");
const dayLabel = (h) => `${h.year ? h.year : "****"}.${pad(h.month)}.${pad(h.day)}`;
const parseTime = (s) => {
  const m = String(s).match(/^(\d{1,2}):(\d{2})$/);
  return m ? Math.min(1439, Number(m[1]) * 60 + Number(m[2])) : 0;
};

function TimeSpin({ value, onChange }) {
  const step = (d) => onChange(Math.min(1439, Math.max(0, value + d)));
  return (
    <span className="hol-spin">
      <input type="text" value={minutesToTime(value)} onChange={(e) => onChange(parseTime(e.target.value))} />
      <span className="hol-spin-btns">
        <button type="button" className="hol-spin-up" onClick={() => step(1)} tabIndex={-1} aria-label="up" />
        <button type="button" className="hol-spin-down" onClick={() => step(-1)} tabIndex={-1} aria-label="down" />
      </span>
    </span>
  );
}

function HolidayDialog({ holiday, onClose, onSaved }) {
  const isNew = !holiday;
  const [tab, setTab] = useState("Common");
  const [draft, setDraft] = useState(
    holiday
      ? { ...holiday, symbols: [...(holiday.symbols || [])] }
      : { mode: 1, year: 0, month: 1, day: 1, from: 0, to: 0, description: "", symbols: ["*"] },
  );
  const [picking, setPicking] = useState(false);
  const [pickedSymbol, setPickedSymbol] = useState(null);
  // which row the inline tree editor is open on; "add" is the click-to-add row
  const [editingMask, setEditingMask] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(holiday?.holiday_id ?? "new");
  const close = useDialogStack(onClose);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));
  const everyYear = !draft.year;
  const addingMask = editingMask === "add";

  const dateText = everyYear
    ? `${pad(draft.month)}.${pad(draft.day)}`
    : `${draft.year}.${pad(draft.month)}.${pad(draft.day)}`;

  function onDateText(text) {
    const m = text.match(/^(?:(\d{4})\.)?(\d{1,2})\.(\d{1,2})$/);
    if (!m) return;
    setDraft((prev) => ({
      ...prev,
      year: m[1] ? Number(m[1]) : prev.year,
      month: Math.min(12, Math.max(1, Number(m[2]))),
      day: Math.min(31, Math.max(1, Number(m[3]))),
    }));
  }

  /** Commits what the inline tree editor returned; index null appends. */
  function commitMask(value, index) {
    setEditingMask(null);

    const mask = (value || "").trim();
    if (!mask) return;

    // the same mask twice changes nothing and only makes the list harder to read
    const clash = draft.symbols.findIndex((m) => m === mask);
    if (clash !== -1 && clash !== index) {
      setPickedSymbol(clash);
      return;
    }

    const next = [...draft.symbols];
    if (index == null) next.push(mask);
    else next[index] = mask;

    set("symbols", next);
    setPickedSymbol(index == null ? next.length - 1 : index);
  }

  function moveMask(step) {
    const to = pickedSymbol + step;
    if (pickedSymbol == null || to < 0 || to >= draft.symbols.length) return;

    const next = [...draft.symbols];
    [next[pickedSymbol], next[to]] = [next[to], next[pickedSymbol]];

    set("symbols", next);
    setPickedSymbol(to);
  }

  async function handleOk() {
    const symbols = draft.symbols.map((s) => s.trim()).filter(Boolean);
    if (!symbols.length) {
      setError("At least one symbol mask is required");
      return;
    }
    const body = {
      year: draft.year || 0,
      month: draft.month,
      day: draft.day,
      from: draft.from,
      to: draft.to,
      description: draft.description || "",
      mode: draft.mode ?? 1,
      symbols,
    };
    const res = isNew ? await createHoliday(body) : await updateHoliday(holiday.holiday_id, body);
    if (!res.ok) {
      setError(res.message || "save failed");
      return;
    }
    onSaved();
    close();
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          width={531}
          height={360}
          title="Holiday"
          tabs={
            <div className="config-tabs">
              {["Common", "Symbols"].map((t) => (
                <button key={t} type="button" className={tab === t ? "active" : ""} onClick={() => setTab(t)}>
                  {t}
                </button>
              ))}
            </div>
          }
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className={`config-panel active${tab === "Symbols" ? " hol-sym-tab" : ""}`}>
            {tab === "Common" ? (
              <>
                <div className="sym-sessions-intro">
                  <span className="sym-tab-intro-icon" aria-hidden="true">
                    <Icon id="holidays" size={48} />
                  </span>
                  <p>
                    Holidays are used for limitation of trading time at the server. Please specify
                    the settings and description of day.
                  </p>
                </div>
                <div className="form-grid grp-check-stack hol-check-stack">
                  <label className="sym-check">
                    <input
                      type="checkbox"
                      checked={draft.mode === 1}
                      onChange={(e) => set("mode", e.target.checked ? 1 : 0)}
                    />{" "}
                    Enable
                  </label>
                  <label className="sym-check">
                    <input
                      type="checkbox"
                      checked={everyYear}
                      onChange={(e) => set("year", e.target.checked ? 0 : new Date().getFullYear())}
                    />{" "}
                    Every year
                  </label>
                </div>
                <div className="form-grid hol-common">
                  <label>Date</label>
                  <span className="hol-date-row">
                    <span className="hol-date">
                      <input type="text" value={dateText} onChange={(e) => onDateText(e.target.value)} />
                      <button
                        type="button"
                        className="hol-date-btn"
                        title="Calendar"
                        onClick={() => setPicking((p) => !p)}
                      >
                        <Icon id="holidays" /> ▼
                      </button>
                      {picking && (
                        <input
                          type="date"
                          className="hol-date-picker"
                          value={`${draft.year || new Date().getFullYear()}-${pad(draft.month)}-${pad(draft.day)}`}
                          onChange={(e) => {
                            const [y, m, d] = e.target.value.split("-").map(Number);
                            setDraft((prev) => ({
                              ...prev,
                              year: prev.year ? y : 0,
                              month: m,
                              day: d,
                            }));
                            setPicking(false);
                          }}
                        />
                      )}
                    </span>
                    <span className="hol-worktime">
                      <label className="hol-inline-label">Work time</label>
                      <TimeSpin value={draft.from} onChange={(v) => set("from", v)} />
                      <span className="hol-dash">-</span>
                      <TimeSpin value={draft.to} onChange={(v) => set("to", v)} />
                    </span>
                  </span>
                  <label>Description</label>
                  <input
                    type="text"
                    className="wide"
                    value={draft.description}
                    onChange={(e) => set("description", e.target.value)}
                  />
                </div>
              </>
            ) : (
              <>
                <div className="sym-sessions-intro">
                  <span className="sym-tab-intro-icon" aria-hidden="true">
                    <Icon id="holidays" size={48} />
                  </span>
                  <p>Please specify symbols which will be affected by the holiday.</p>
                </div>
                <div className="df-table-panel hol-sym-panel">
                  <div className="df-table-toolbar grp-sym-buttons">
                    <div className="grp-sym-buttons-move">
                      <button
                        type="button"
                        disabled={addingMask || pickedSymbol == null || pickedSymbol <= 0}
                        onClick={() => moveMask(-1)}
                      >
                        Up
                      </button>
                      <button
                        type="button"
                        disabled={
                          addingMask || pickedSymbol == null || pickedSymbol >= draft.symbols.length - 1
                        }
                        onClick={() => moveMask(1)}
                      >
                        Down
                      </button>
                    </div>
                    <div className="grp-sym-buttons-actions">
                      <button type="button" onClick={() => setEditingMask("add")}>
                        Add
                      </button>
                      <button
                        type="button"
                        disabled={addingMask || pickedSymbol == null}
                        onClick={() => setEditingMask(pickedSymbol)}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        disabled={addingMask || pickedSymbol == null}
                        onClick={() => {
                          set("symbols", draft.symbols.filter((_, i) => i !== pickedSymbol));
                          setPickedSymbol(null);
                        }}
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                  <div className="df-table-main hol-sym-table">
                    <table className="data-table data-table-grid df-sub-table">
                      <thead>
                        <tr>
                          <th>Symbol</th>
                        </tr>
                      </thead>
                      <tbody>
                        {draft.symbols.map((mask, i) => (
                          <tr
                            key={`${mask}-${i}`}
                            className={
                              editingMask === i
                                ? "hol-sym-edit-row"
                                : !addingMask && pickedSymbol === i
                                  ? "selected"
                                  : ""
                            }
                            onClick={() => setPickedSymbol(i)}
                            onDoubleClick={() => setEditingMask(i)}
                          >
                            <td>
                              {editingMask === i ? (
                                <SymbolTreeSelect
                                  value={mask}
                                  onCommit={(v) => commitMask(v, i)}
                                  onCancel={() => setEditingMask(null)}
                                />
                              ) : (
                                <span className="grp-sym-cell">
                                  <Icon id="symbols-tree" /> {mask}
                                </span>
                              )}
                            </td>
                          </tr>
                        ))}
                        {addingMask ? (
                          <tr className="hol-sym-edit-row">
                            <td>
                              <SymbolTreeSelect
                                value="*"
                                onCommit={(v) => commitMask(v, null)}
                                onCancel={() => setEditingMask(null)}
                              />
                            </td>
                          </tr>
                        ) : (
                          <tr className="df-add-row" onClick={() => setEditingMask("add")}>
                            <td>
                              <span className="df-add-plus" aria-hidden="true">
                                +
                              </span>
                              click to add...
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              </>
            )}
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

/** Trading holidays: **** in Day means every year; From = To = 00:00 closes the whole day. */
export function HolidaysModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_holidays !== false;

  const load = () => fetchHolidays().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
  }, []);

  async function onDelete(row) {
    if (!window.confirm(`Delete holiday '${row.description || dayLabel(row)}'?`)) return;
    const res = await deleteHoliday(row.holiday_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  function saved() {
    load();
    session.refreshNav?.();
  }

  // holiday rules are evaluated top-down, so list order is a setting, not a view preference
  async function applyOrder(next, keepIndex) {
    const res = await reorderHolidays(next.map((r) => r.holiday_id));
    if (!res.ok) {
      window.alert(res.message || "reorder failed");
      return;
    }
    setRows(next);
    setSelected(keepIndex);
  }

  function move(delta) {
    const i = selected;
    const j = i + delta;
    if (rows == null || i == null || j < 0 || j >= rows.length) return;
    const next = [...rows];
    [next[i], next[j]] = [next[j], next[i]];
    applyOrder(next, j);
  }

  function sortAlphabetically() {
    if (!rows?.length) return;
    const next = [...rows].sort((a, b) => dayLabel(a).localeCompare(dayLabel(b)));
    applyOrder(next, null);
  }

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Day</th>
              <th>From</th>
              <th>To</th>
              <th>Description</th>
              <th>Symbols</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((row, i) => (
              <tr
                key={row.holiday_id}
                className={`${selected === i ? "selected" : ""}${row.mode === 1 ? "" : " nav-feed-disabled"}`.trim()}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ holiday: row })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="holidays" />
                    {dayLabel(row)}
                  </span>
                </td>
                <td>{minutesToTime(row.from)}</td>
                <td>{minutesToTime(row.to)}</td>
                <td>{row.description}</td>
                <td>{(row.symbols || []).join("; ")}</td>
              </tr>
            ))}
            {rows?.length === 0 && (
              <tr>
                <td colSpan={5} className="df-empty">No holidays configured</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      {menu && canEdit && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            ...listMenuHead({
              onAdd: () => setDialog({ holiday: null }),
              onEdit: () => setDialog({ holiday: rows[selected] }),
              onDelete: () => onDelete(rows[selected]),
              hasSelection: selected != null,
            }),
            ...listMenuTail({
              on: {
                moveUp: selected != null ? () => move(-1) : undefined,
                moveDown: selected != null ? () => move(1) : undefined,
                sort: sortAlphabetically,
              },
            }),
          ]}
        />
      )}
      {dialog && (
        <HolidayDialog holiday={dialog.holiday} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}
