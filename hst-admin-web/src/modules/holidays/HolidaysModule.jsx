import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createHoliday,
  deleteHoliday,
  fetchHolidays,
  updateHoliday,
} from "@/api/endpoints/holidays.js";
import { minutesToTime } from "@/lib/symbolSessions.js";

const pad = (n) => String(n).padStart(2, "0");
const dayLabel = (h) => `${h.year ? h.year : "****"}.${pad(h.month)}.${pad(h.day)}`;
const parseTime = (s) => {
  const m = String(s).match(/^(\d{1,2}):(\d{2})$/);
  return m ? Math.min(1439, Number(m[1]) * 60 + Number(m[2])) : 0;
};

function HolidayDialog({ holiday, onClose, onSaved }) {
  const isNew = !holiday;
  const [tab, setTab] = useState("Common");
  const [draft, setDraft] = useState(
    holiday
      ? { ...holiday, symbols: (holiday.symbols || []).join("; ") }
      : { mode: 1, year: 0, month: 1, day: 1, from: 0, to: 0, description: "", symbols: "*" },
  );
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(holiday?.holiday_id ?? "new");

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    const symbols = draft.symbols.split(";").map((s) => s.trim()).filter(Boolean);
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
    onClose();
  }

  return (
    <div className="dialog-overlay" onClick={onClose} role="presentation">
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          draggable
          onTitlePointerDown={onTitlePointerDown}
          width={480}
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
              <button type="button" onClick={onClose}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="holidays" size={48} />
              </span>
              <p>
                Holidays are used for limitation of trading time at the server. Please specify the
                settings and description of day.
              </p>
            </div>
            {tab === "Common" ? (
              <>
                <div className="form-grid grp-check-stack">
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
                      checked={!draft.year}
                      onChange={(e) => set("year", e.target.checked ? 0 : new Date().getFullYear())}
                    />{" "}
                    Every year
                  </label>
                </div>
                <div className="form-grid">
                  {draft.year ? (
                    <>
                      <label>Year</label>
                      <input type="text" value={draft.year} onChange={(e) => set("year", Number(e.target.value) || 0)} />
                    </>
                  ) : null}
                  <label>Date</label>
                  <span className="grp-suffixed">
                    <input
                      type="text"
                      value={`${pad(draft.month)}.${pad(draft.day)}`}
                      onChange={(e) => {
                        const m = e.target.value.match(/^(\d{1,2})\.(\d{1,2})$/);
                        if (m) {
                          set("month", Math.min(12, Math.max(1, Number(m[1]))));
                          set("day", Math.min(31, Math.max(1, Number(m[2]))));
                        }
                      }}
                    />
                    <span className="grp-suffix">MM.DD</span>
                  </span>
                  <label>Work time</label>
                  <span className="grp-suffixed">
                    <input
                      type="text"
                      value={minutesToTime(draft.from)}
                      onChange={(e) => set("from", parseTime(e.target.value))}
                    />
                    <span className="grp-suffix">—</span>
                    <input
                      type="text"
                      value={minutesToTime(draft.to)}
                      onChange={(e) => set("to", parseTime(e.target.value))}
                    />
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
              <div className="form-grid">
                <label>Symbols</label>
                <input
                  type="text"
                  className="wide"
                  placeholder={String.raw`Forex\*; CFD\*; !EURUSD`}
                  value={draft.symbols}
                  onChange={(e) => set("symbols", e.target.value)}
                />
              </div>
            )}
          </div>
        </SettingsDialog>
      </div>
    </div>
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
                className={selected === i ? "selected" : ""}
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
                <td className={row.mode === 1 ? "" : "nav-feed-disabled"}>{row.description}</td>
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
            { label: "Add", onClick: () => setDialog({ holiday: null }) },
            { label: "Edit", disabled: selected == null, onClick: () => setDialog({ holiday: rows[selected] }) },
            "sep",
            { label: "Delete", disabled: selected == null, onClick: () => onDelete(rows[selected]) },
          ]}
        />
      )}
      {dialog && (
        <HolidayDialog holiday={dialog.holiday} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}
