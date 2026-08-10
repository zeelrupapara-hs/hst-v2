import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
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
  moveRoutingDealer,
  removeRoutingDealer,
  updateRoutingRule,
} from "@/api/endpoints/routing.js";
import {
  ConditionRule_name,
  RouteCondition_date,
  RouteConditionGroups,
  routeConditionGroup,
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
          <span className="prop-select-btn-wrap">
            <button type="button" className="prop-select-btn" aria-label="Open list" tabIndex={-1} />
          </span>
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

  // the list order is the order a request is offered in, so it can be rearranged
  async function onMove(delta) {
    const row = rows?.[selected];
    const to = selected + delta;
    if (!row || to < 0 || to >= rows.length) return;
    const res = await moveRoutingDealer(ruleId, row.login, rows[to].dealer_index ?? to);
    if (!res.ok) {
      window.alert(res.message || "move failed");
      return;
    }
    setSelected(to);
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
        <div className="routing-grid-row routing-dealers-row">
          <div className="routing-btn-col routing-btn-col-split">
            <div className="routing-btn-group">
              <button type="button" disabled={selected == null || selected === 0} onClick={() => onMove(-1)}>
                Up
              </button>
              <button
                type="button"
                disabled={selected == null || selected >= (rows?.length ?? 0) - 1}
                onClick={() => onMove(1)}
              >
                Down
              </button>
            </div>
            <div className="routing-btn-group">
              <button type="button" onClick={() => setEditing({ login: 0 })}>Add</button>
              <button type="button" disabled={selected == null} onClick={() => setEditing(rows[selected])}>
                Edit
              </button>
              <button type="button" disabled={selected == null} onClick={onDelete}>Delete</button>
            </div>
          </div>
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
                  <td>
                    <span className="sym-symbol-cell">
                      <span className="routing-dealer-icon" aria-hidden="true">🖊</span>
                      {d.login}
                    </span>
                  </td>
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

function condGlyph(id) {
  if (RouteCondition_date.has(id)) return <span className="routing-glyph-dt">📅</span>;
  if (RouteCondition_text.has(id)) return <span className="routing-glyph-ab">ab</span>;
  return <span className="routing-glyph-01">01</span>;
}

/** The condition picker as the reference draws it: Request leaves on top, the other branches fold. */
function CondTypeSelect({ value, onChange }) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState(null);
  const btnRef = useRef(null);
  const [folded, setFolded] = useState({ Request: false, Account: true, Position: true, Order: true, Symbol: true });
  const byGroup = {};
  for (const [id, label] of Object.entries(RouteCondition_name)) {
    const g = routeConditionGroup(Number(id));
    (byGroup[g] ??= []).push({ id: Number(id), label });
  }

  function toggle() {
    if (!open && btnRef.current) {
      const r = btnRef.current.getBoundingClientRect();
      setPos({ top: r.bottom + 1, left: r.left, minWidth: r.width });
    }
    setOpen(!open);
  }

  return (
    <span className="routing-cond-select">
      <button ref={btnRef} type="button" className="routing-cond-current" onClick={toggle}>
        {condGlyph(value)}
        <span>{routeConditionGroup(value)}\{RouteCondition_name[value] ?? value}</span>
        <span className="routing-cond-arrow">▾</span>
      </button>
      {open &&
        pos &&
        createPortal(
        <span className="routing-cond-pop" style={{ position: "fixed", ...pos }}>
          <span className="routing-cond-root">
            <span className="routing-folder-icon">📁</span> Conditions
          </span>
          {RouteConditionGroups.map((g) => (
            <span key={g} className="routing-cond-branch">
              <button
                type="button"
                className="routing-cond-folder"
                onClick={() => setFolded({ ...folded, [g]: !folded[g] })}
              >
                {folded[g] ? "▸" : "▾"} <span className="routing-folder-icon">📁</span> {g}
              </button>
              {!folded[g] &&
                byGroup[g]?.map((c) => (
                  <button
                    key={c.id}
                    type="button"
                    className={`routing-cond-leaf${c.id === value ? " active" : ""}`}
                    onClick={() => {
                      onChange(c.id);
                      setOpen(false);
                    }}
                  >
                    {condGlyph(c.id)} {c.label}
                  </button>
                ))}
            </span>
          ))}
        </span>,
        document.body,
      )}
    </span>
  );
}

const newDraft = () => ({
  name: "",
  mode: 1,
  request: 0,
  type: 0,
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
  const close = useDialogStack(onClose);

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
      for (const key of ["name", "mode", "request", "type", "action", "action_value"]) {
        if (draft[key] !== original[key]) patch[key] = draft[key];
      }
      const res = await updateRoutingRule(ruleId, patch);
      if (!res.ok) {
        setError(res.message || "save failed");
        return;
      }
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
          width={620}
          height={500}
          onClose={close}
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
              <button type="button" onClick={close}>Cancel</button>
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
                  </div>
                  <div className="routing-conds-section">
                    <div className="routing-conds-left">
                      <span className="routing-conds-caption">Where conditions are:</span>
                      <div className="routing-conds-btns">
                        <button type="button" onClick={addCond}>Add</button>
                        <button type="button" disabled={selCond == null} onClick={() => setEditCond(selCond)}>
                          Edit
                        </button>
                        <button type="button" disabled={selCond == null} onClick={deleteCond}>Delete</button>
                      </div>
                    </div>
                    <div className="routing-conds-box">
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
                                    <CondTypeSelect value={c.condition} onChange={(v) => setCond(i, "condition", v)} />
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
                                      {condGlyph(c.condition)}
                                      {routeConditionGroup(c.condition)}\{RouteCondition_name[c.condition] ?? c.condition}
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
    </DialogOverlay>
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
