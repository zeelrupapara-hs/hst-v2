import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
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
  RouteActions_withValue,
  RouteCondition_name,
  RouteFlags_labels,
  TypeFlags_labels,
} from "@/constants/routing.js";

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

function FlagChecks({ labels, value, onChange }) {
  return (
    <div className="routing-flag-grid">
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
  );
}

function DealersTab({ ruleId }) {
  const [rows, setRows] = useState(null);
  const [adding, setAdding] = useState("");

  const load = () => fetchRoutingDealers(ruleId).then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    if (ruleId !== "new") load();
  }, [ruleId]);

  if (ruleId === "new") {
    return <p className="module-note">Save the rule first, then assign dealers here.</p>;
  }

  async function add() {
    const login = Number(adding);
    if (!login) return;
    const res = await addRoutingDealer(ruleId, login);
    if (!res.ok) window.alert(res.message || "add failed");
    setAdding("");
    load();
  }

  return (
    <div className="df-table-panel">
      <div className="df-table-toolbar">
        <input
          type="text"
          className="df-cell-input"
          placeholder="Dealer login"
          value={adding}
          onChange={(e) => setAdding(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && add()}
        />
        <button type="button" onClick={add}>Add</button>
      </div>
      <div className="df-table-main">
        <table className="data-table data-table-grid df-sub-table">
          <thead>
            <tr>
              <th>Login</th>
              <th>Name</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((d) => (
              <tr key={d.login}>
                <td>{d.login}</td>
                <td>{d.name || "—"}</td>
                <td>
                  <button type="button" onClick={() => removeRoutingDealer(ruleId, d.login).then(load)}>✕</button>
                </td>
              </tr>
            ))}
            {rows && !rows.length && (
              <tr>
                <td colSpan={3} className="df-empty">No dealers assigned</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
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
          onTitlePointerDown={onTitlePointerDown}
          title={isNew ? "Routing Rule: New" : `Routing Rule: ${draft?.name ?? "…"}`}
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
                      Rules run top to bottom; the first matching rule decides what happens to a
                      trade request.
                    </p>
                  </div>
                  <div className="form-grid">
                    <label>Enable this rule</label>
                    <input
                      type="checkbox"
                      className="df-checkbox"
                      checked={draft.mode === 1}
                      onChange={(e) => set("mode", e.target.checked ? 1 : 0)}
                    />
                    <label>Name</label>
                    <input type="text" value={draft.name} onChange={(e) => set("name", e.target.value)} />
                    <label>Perform action</label>
                    <PropSelect fill value={draft.action} options={enumOptions(RouteAction_name)} onChange={(v) => set("action", v)} />
                    {RouteActions_withValue.has(draft.action) && (
                      <>
                        <label>Value</label>
                        <input type="text" value={draft.action_value} onChange={(e) => set("action_value", e.target.value)} />
                      </>
                    )}
                  </div>
                  <fieldset className="fieldset">
                    <legend>Where request is (none checked = all)</legend>
                    <FlagChecks labels={RouteFlags_labels} value={draft.request} onChange={(v) => set("request", v)} />
                  </fieldset>
                  <fieldset className="fieldset">
                    <legend>Where order is (none checked = all)</legend>
                    <FlagChecks labels={TypeFlags_labels} value={draft.type} onChange={(v) => set("type", v)} />
                  </fieldset>
                  <fieldset className="fieldset">
                    <legend>Where conditions are</legend>
                    <div className="df-table-toolbar">
                      <button
                        type="button"
                        onClick={() => set("conditions", [...(draft.conditions || []), { condition: 1001, rule: 0, value: "" }])}
                      >
                        Add
                      </button>
                    </div>
                    <table className="data-table data-table-grid df-sub-table">
                      <tbody>
                        {(draft.conditions || []).map((c, i) => (
                          <tr key={i}>
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
                              />
                            </td>
                            <td>
                              <button type="button" onClick={() => set("conditions", draft.conditions.filter((_, j) => j !== i))}>✕</button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </fieldset>
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
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_requests !== false;

  const load = () =>
    fetchRouting().then((res) => {
      if (res.ok) setRows((res.data || []).sort((a, b) => a.routing_index - b.routing_index));
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

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Name</th>
              <th>Action</th>
              <th>State</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((row, i) => (
              <tr
                key={row.routing_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ id: row.routing_id })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="routing" />
                    {row.name}
                  </span>
                </td>
                <td>{RouteAction_name[row.action] ?? row.action}</td>
                <td className={row.mode === 1 ? "" : "nav-feed-disabled"}>
                  {row.mode === 1 ? "Enabled" : "Disabled"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {dialog && <RuleDialog ruleId={dialog.id} onClose={() => setDialog(null)} onSaved={load} />}
    </div>
  );
}
