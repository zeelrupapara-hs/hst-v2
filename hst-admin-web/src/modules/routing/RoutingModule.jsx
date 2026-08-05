import { useEffect, useRef, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { fetchManagers } from "@/api/endpoints/managers.js";
import {
  addRoutingDealer,
  createRoutingRule,
  deleteRoutingRule,
  fetchRouting,
  fetchRoutingDealers,
  fetchRoutingRule,
  moveRoutingRule,
  removeRoutingDealer,
  updateRoutingRule,
} from "@/api/endpoints/routing.js";
import {
  ConditionRule_name,
  RouteAction_name,
  RouteActions_dealer,
  RouteActions_withValue,
  RouteCondition_name,
  RouteCondition_text,
  RouteFlags_labels,
  TypeFlags_labels,
} from "@/constants/routing.js";

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

const flagText = (labels, value) =>
  !value ? "All" : labels.filter(({ bit }) => (value & bit) !== 0).map((f) => f.label).join(", ") || "All";

const allBits = (labels) => labels.reduce((acc, f) => acc | f.bit, 0);

/** Multi-check combo: closed text is "All" for zero, right-click inside toggles the whole list. */
function FlagSelect({ labels, value, onChange }) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef(null);

  useEffect(() => {
    if (!open) return;
    const close = (e) => !rootRef.current?.contains(e.target) && setOpen(false);
    const onKey = (e) => e.key === "Escape" && setOpen(false);
    const t = setTimeout(() => document.addEventListener("mousedown", close), 0);
    document.addEventListener("keydown", onKey);
    return () => {
      clearTimeout(t);
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  const text = flagText(labels, value ?? 0);

  return (
    <div className="routing-multi" ref={rootRef}>
      <div className="prop-select prop-select-fill">
        <div className="prop-select-box" onClick={() => setOpen((o) => !o)}>
          <span className="prop-select-value" title={text}>{text}</span>
          <button type="button" className="prop-select-btn" aria-label="Open list" onClick={() => setOpen((o) => !o)} />
        </div>
      </div>
      {open && (
        <div
          className="routing-multi-list"
          onContextMenu={(e) => {
            e.preventDefault();
            onChange((value ?? 0) === 0 ? allBits(labels) : 0);
          }}
        >
          {labels.map(({ bit, label }) => (
            <label key={bit} className="sym-check">
              <input
                type="checkbox"
                checked={((value ?? 0) & bit) !== 0}
                onChange={() => onChange((value ?? 0) ^ bit)}
              />{" "}
              {label}
            </label>
          ))}
        </div>
      )}
    </div>
  );
}

/** Add / Edit / Delete stacked beside a grid, as the reference dialogs have them. */
function StackButtons({ onAdd, onEdit, onDelete, hasSelection }) {
  return (
    <div className="routing-btn-col">
      <button type="button" onClick={onAdd}>Add</button>
      <button type="button" disabled={!hasSelection} onClick={onEdit}>Edit</button>
      <button type="button" disabled={!hasSelection} onClick={onDelete}>Delete</button>
    </div>
  );
}

function DealersTab({ ruleId }) {
  const [rows, setRows] = useState(null);
  const [managers, setManagers] = useState([]);
  const [selected, setSelected] = useState(null);
  const [editing, setEditing] = useState(null);

  const load = () => fetchRoutingDealers(ruleId).then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    if (ruleId !== "new") load();
    fetchManagers().then((res) => res.ok && setManagers(res.data || []));
  }, [ruleId]);

  async function commit(login) {
    if (!login) return;
    if (editing?.login && editing.login !== login) await removeRoutingDealer(ruleId, editing.login);
    const res = await addRoutingDealer(ruleId, login);
    if (!res.ok) window.alert(res.message || "add failed");
    setEditing(null);
    load();
  }

  async function onDelete() {
    const row = rows?.[selected];
    if (!row) return;
    await removeRoutingDealer(ruleId, row.login);
    setSelected(null);
    load();
  }

  const options = managers.map((m) => ({ value: m.login, label: `${m.name} (${m.login})` }));

  return (
    <>
      <div className="sym-sessions-intro">
        <span className="sym-tab-intro-icon" aria-hidden="true">
          <Icon id="routing" size={48} />
        </span>
        <p>Please specify dealers who will process requests that meet the rule conditions.</p>
      </div>
      {ruleId === "new" ? (
        <p className="module-note">Save the rule first, then assign dealers here.</p>
      ) : (
        <div className="routing-grid-row">
          <StackButtons
            hasSelection={selected != null}
            onAdd={() => setEditing({ login: 0 })}
            onEdit={() => setEditing(rows[selected])}
            onDelete={onDelete}
          />
          <table
            className="data-table data-table-grid df-sub-table"
            onKeyDown={(e) => e.key === "Delete" && onDelete()}
            tabIndex={0}
          >
            <thead>
              <tr>
                <th>Login</th>
                <th>Name</th>
              </tr>
            </thead>
            <tbody>
              {(rows || []).map((d, i) => (
                <tr
                  key={d.login}
                  className={selected === i ? "selected" : ""}
                  onClick={() => setSelected(i)}
                  onDoubleClick={() => setEditing(d)}
                >
                  <td>{d.login}</td>
                  <td>{d.name || "—"}</td>
                </tr>
              ))}
              {editing && (
                <tr>
                  <td colSpan={2}>
                    <PropSelect
                      fill
                      value={editing.login || options[0]?.value}
                      options={options}
                      onChange={commit}
                    />
                  </td>
                </tr>
              )}
              {rows && !rows.length && !editing && (
                <tr>
                  <td colSpan={2} className="df-empty">No dealers assigned</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}

const newDraft = () => ({
  name: "",
  mode: 1,
  request: 0,
  type: 0,
  flags: 0,
  action: 1006,
  action_value: "",
  conditions: [],
});

function RuleDialog({ ruleId, onClose, onSaved }) {
  const isNew = ruleId === "new";
  const [tab, setTab] = useState("Common");
  const [draft, setDraft] = useState(isNew ? newDraft() : null);
  const [original, setOriginal] = useState(null);
  const [error, setError] = useState("");
  const [selCond, setSelCond] = useState(null);
  const [editCond, setEditCond] = useState(null);
  const { offset, onTitlePointerDown } = useDialogDrag(ruleId);

  useEffect(() => {
    if (isNew) return;
    fetchRoutingRule(ruleId).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load rule");
        return;
      }
      setDraft(res.data);
      setOriginal(res.data);
    });
  }, [ruleId, isNew]);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  const setCond = (i, key, value) =>
    set("conditions", draft.conditions.map((c, j) => (j === i ? { ...c, [key]: value } : c)));

  function addCond() {
    const next = [...(draft.conditions || []), { condition: 1, rule: 0, value: "" }];
    set("conditions", next);
    setSelCond(next.length - 1);
    setEditCond(next.length - 1);
  }

  function deleteCond() {
    if (selCond == null) return;
    set("conditions", draft.conditions.filter((_, j) => j !== selCond));
    setSelCond(null);
    setEditCond(null);
  }

  async function handleOk() {
    if (!draft.name?.trim()) {
      setError("Name is required");
      return;
    }
    const conditions = (draft.conditions || []).map((c) => ({
      condition: c.condition ?? 1,
      rule: c.rule ?? 0,
      value: String(c.value ?? ""),
    }));
    if (isNew) {
      const res = await createRoutingRule({ ...draft, conditions });
      if (!res.ok) {
        setError(res.message || "create failed");
        return;
      }
    } else {
      const patch = { conditions };
      for (const key of ["name", "mode", "request", "type", "flags", "action", "action_value"]) {
        if (draft[key] !== original[key]) patch[key] = draft[key];
      }
      const res = await updateRoutingRule(ruleId, patch);
      if (!res.ok) {
        setError(res.message || "save failed");
        return;
      }
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
          width={640}
          onClose={onClose}
          onTitlePointerDown={onTitlePointerDown}
          title={isNew ? "Routing: New" : `Routing: ${draft?.name ?? "…"}`}
          tabs={
            <div className="config-tabs">
              {["Common", "Dealers"].map((t) => (
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
          {draft && (
            <div className="config-panel active">
              {tab === "Common" ? (
                <>
                  <div className="sym-sessions-intro">
                    <span className="sym-tab-intro-icon" aria-hidden="true">
                      <Icon id="routing" size={48} />
                    </span>
                    <p>
                      Using routing rules, one can adjust processing of trade requests by
                      different conditions. Please specify conditions, for which the rule will be
                      applied.
                    </p>
                  </div>
                  <div className="form-grid">
                    <span />
                    <label className="sym-check routing-enable">
                      <input
                        type="checkbox"
                        checked={draft.mode === 1}
                        onChange={(e) => set("mode", e.target.checked ? 1 : 0)}
                      />{" "}
                      Enable this rule
                    </label>
                    <label>Name</label>
                    <input type="text" value={draft.name} onChange={(e) => set("name", e.target.value)} />
                    <label>Perform action</label>
                    <div className="routing-action-row">
                      <PropSelect fill value={draft.action} options={enumOptions(RouteAction_name)} onChange={(v) => set("action", v)} />
                      {RouteActions_withValue.has(draft.action) && (
                        <input
                          type="text"
                          value={draft.action_value}
                          onChange={(e) => set("action_value", e.target.value)}
                        />
                      )}
                      {RouteActions_dealer.has(draft.action) && (
                        <label className="sym-check">
                          <input
                            type="checkbox"
                            checked={String(draft.action_value) === "1"}
                            onChange={(e) => set("action_value", e.target.checked ? "1" : "0")}
                          />{" "}
                          skip this rule if no dealers online
                        </label>
                      )}
                    </div>
                    <label>Where request is</label>
                    <FlagSelect labels={RouteFlags_labels} value={draft.request} onChange={(v) => set("request", v)} />
                    <label>Where order is</label>
                    <FlagSelect labels={TypeFlags_labels} value={draft.type} onChange={(v) => set("type", v)} />
                    <label>Where conditions are</label>
                    <div className="routing-grid-row">
                      <StackButtons
                        hasSelection={selCond != null}
                        onAdd={addCond}
                        onEdit={() => setEditCond(selCond)}
                        onDelete={deleteCond}
                      />
                      <table
                        className="data-table data-table-grid df-sub-table"
                        tabIndex={0}
                        onKeyDown={(e) => e.key === "Delete" && deleteCond()}
                      >
                        <thead>
                          <tr>
                            <th>Type</th>
                            <th>Condition</th>
                            <th className="num">Value</th>
                          </tr>
                        </thead>
                        <tbody>
                          {(draft.conditions || []).map((c, i) => (
                            <tr
                              key={i}
                              className={selCond === i ? "selected" : ""}
                              onClick={() => setSelCond(i)}
                              onDoubleClick={() => setEditCond(i)}
                            >
                              {editCond === i ? (
                                <>
                                  <td>
                                    <PropSelect fill value={c.condition} options={enumOptions(RouteCondition_name)} onChange={(v) => setCond(i, "condition", v)} />
                                  </td>
                                  <td>
                                    <PropSelect fill value={c.rule} options={enumOptions(ConditionRule_name)} onChange={(v) => setCond(i, "rule", v)} />
                                  </td>
                                  <td>
                                    <input
                                      type="text"
                                      className="df-cell-input"
                                      value={c.value ?? ""}
                                      onChange={(e) => setCond(i, "value", e.target.value)}
                                      onBlur={() => setEditCond(null)}
                                      onKeyDown={(e) => e.key === "Enter" && setEditCond(null)}
                                    />
                                  </td>
                                </>
                              ) : (
                                <>
                                  <td>
                                    <span className="sym-symbol-cell">
                                      <span className={RouteCondition_text.has(c.condition) ? "routing-glyph-ab" : "routing-glyph-01"}>
                                        {RouteCondition_text.has(c.condition) ? "ab" : "01"}
                                      </span>
                                      {RouteCondition_name[c.condition] ?? c.condition}
                                    </span>
                                  </td>
                                  <td>{ConditionRule_name[c.rule] ?? c.rule}</td>
                                  <td className="num">{c.value}</td>
                                </>
                              )}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </>
              ) : (
                <DealersTab ruleId={ruleId} />
              )}
            </div>
          )}
        </SettingsDialog>
      </div>
    </div>
  );
}

/** Routing rules list: list position IS priority; Up/Down reorder server-side. */
export function RoutingModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [dealers, setDealers] = useState({});
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_requests !== false;

  const load = () =>
    fetchRouting().then(async (res) => {
      if (!res.ok) return;
      const list = (res.data || []).sort((a, b) => a.routing_index - b.routing_index);
      setRows(list);
      const found = await Promise.all(list.map((r) => fetchRoutingDealers(r.routing_id)));
      setDealers(Object.fromEntries(list.map((r, i) => [r.routing_id, found[i].ok ? found[i].data || [] : []])));
    });

  useEffect(() => {
    load();
  }, []);

  async function onDelete(row) {
    if (!window.confirm(`Delete routing rule '${row.name}'?`)) return;
    const res = await deleteRoutingRule(row.routing_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    load();
  }

  async function move(dir) {
    const row = rows?.[selected];
    if (!row) return;
    await moveRoutingRule(row.routing_id, dir);
    setSelected(dir === "up" ? selected - 1 : selected + 1);
    load();
  }

  // Enable/Disable have no endpoint of their own; they are a patch of the rule's mode
  async function setMode(mode) {
    const row = rows?.[selected];
    if (!row) return;
    const res = await updateRoutingRule(row.routing_id, { mode });
    if (!res.ok) window.alert(res.message || "save failed");
    load();
  }

  const row = selected == null ? null : rows?.[selected];

  return (
    <div className="module-root">
      <p className="module-note">
        Any change in the routing table leads to that all requests currently processed according
        to these rules are moved to the beginning of the table and go through all rules again.
      </p>
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Name</th>
              <th>Action</th>
              <th>Dealers</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((r, i) => (
              <tr
                key={r.routing_id}
                className={`${selected === i ? "selected" : ""}${r.mode === 1 ? "" : " nav-feed-disabled"}`}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ id: r.routing_id })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="routing" />
                    {r.name}
                  </span>
                </td>
                <td>{RouteAction_name[r.action] ?? r.action}</td>
                <td>
                  {(dealers[r.routing_id] || []).map((d) => `${d.name} (${d.login})`).join(", ")}
                </td>
              </tr>
            ))}
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
              onAdd: () => setDialog({ id: "new" }),
              onEdit: () => setDialog({ id: row.routing_id }),
              onDelete: () => onDelete(row),
              hasSelection: !!row,
            }),
            ...listMenuTail({
              on: {
                moveUp: row ? () => move("up") : undefined,
                moveDown: row ? () => move("down") : undefined,
              },
              extras: [
                "sep",
                { label: "Enable", disabled: !row || row.mode === 1, onClick: () => setMode(1) },
                { label: "Disable", disabled: !row || row.mode === 0, onClick: () => setMode(0) },
              ],
            }),
          ]}
        />
      )}
      {dialog && <RuleDialog ruleId={dialog.id} onClose={() => setDialog(null)} onSaved={load} />}
    </div>
  );
}
