import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createLeverage,
  deleteLeverage,
  fetchLeverage,
  fetchLeverages,
  updateLeverage,
} from "@/api/endpoints/leverages.js";

const RANGE_MODE = {
  0: "Volume",
  1: "Volume per symbol",
  2: "Notional value",
  3: "Notional value per symbol",
};

const INFINITY_TO = 0;
const tierTo = (t) => (t.range_to === INFINITY_TO ? "infinity" : t.range_to);

const ruleSummary = (r) =>
  `${r.path}: ${(r.tiers || [])
    .map((t) => `${tierTo(t)} — ${t.margin_rate_initial}`)
    .join("; ")}`;

/** The rule editor: name, symbol mask, range mode and the tier ladder. */
function RuleEditor({ rule, onSave, onClose }) {
  const [draft, setDraft] = useState(
    rule ?? { name: "", description: "", path: "*", range_mode: 0, range_value_currency: "", tiers: [{ range_to: 0, margin_rate_initial: 1, margin_rate_maintenance: 1 }] },
  );
  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  const setTier = (i, key, value) =>
    set("tiers", draft.tiers.map((t, j) => (j === i ? { ...t, [key]: value } : t)));

  return (
    <div className="sym-session-dialog-overlay" onClick={onClose}>
      <div className="sym-session-dialog lev-rule-dialog" role="dialog" onClick={(e) => e.stopPropagation()}>
        <div className="sym-session-dialog-title">Leverage Rule</div>
        <div className="sym-session-dialog-body">
          <div className="form-grid">
            <label>Name</label>
            <input type="text" value={draft.name} onChange={(e) => set("name", e.target.value)} />
            <label>Description</label>
            <input type="text" value={draft.description} onChange={(e) => set("description", e.target.value)} />
            <label>Symbol</label>
            <input type="text" value={draft.path} onChange={(e) => set("path", e.target.value)} />
            <label>Range</label>
            <PropSelect
              fill
              value={draft.range_mode}
              options={Object.entries(RANGE_MODE).map(([v, l]) => ({ value: Number(v), label: l }))}
              onChange={(v) => set("range_mode", v)}
            />
            <label>Currency</label>
            <input
              type="text"
              value={draft.range_value_currency}
              onChange={(e) => set("range_value_currency", e.target.value.toUpperCase())}
            />
          </div>
          <table className="data-table data-table-grid df-sub-table">
            <thead>
              <tr>
                <th>To</th>
                <th>Initial margin rate</th>
                <th>Maintenance margin rate</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {draft.tiers.map((t, i) => (
                <tr key={i}>
                  {["range_to", "margin_rate_initial", "margin_rate_maintenance"].map((key) => (
                    <td key={key}>
                      <input
                        type="text"
                        className="df-cell-input"
                        value={t[key]}
                        onChange={(e) => setTier(i, key, Number(e.target.value) || 0)}
                      />
                    </td>
                  ))}
                  <td>
                    <button type="button" onClick={() => set("tiers", draft.tiers.filter((_, j) => j !== i))}>✕</button>
                  </td>
                </tr>
              ))}
              <tr>
                <td colSpan={4} className="df-cell-editable" onClick={() => set("tiers", [...draft.tiers, { range_to: 0, margin_rate_initial: 1, margin_rate_maintenance: 1 }])}>
                  + click to add…
                </td>
              </tr>
            </tbody>
          </table>
          <p className="module-note">Tier "To" 0 means infinity; the last tier should be 0.</p>
        </div>
        <div className="sym-session-dialog-footer">
          <span />
          <div className="sym-session-dialog-actions">
            <button type="button" className="sym-session-ok" onClick={() => draft.name.trim() && draft.path.trim() && onSave(draft)}>
              OK
            </button>
            <button type="button" className="sym-session-cancel" onClick={onClose}>Cancel</button>
          </div>
        </div>
      </div>
    </div>
  );
}

/** The profile dialog: a name over the rules grid; OK PUTs the whole profile. */
function LeverageDialog({ profileId, onClose, onSaved }) {
  const isNew = profileId === "new";
  const [draft, setDraft] = useState(isNew ? { name: "", flags: 0, rules: [] } : null);
  const [selected, setSelected] = useState(0);
  const [editor, setEditor] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(profileId);

  useEffect(() => {
    if (isNew) return;
    fetchLeverage(profileId).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load profile");
        return;
      }
      setDraft({ ...res.data, rules: res.data.rules || [] });
    });
  }, [profileId, isNew]);

  async function handleOk() {
    if (!draft.name?.trim()) {
      setError("Name is required");
      return;
    }
    const body = {
      name: draft.name.trim(),
      flags: draft.flags ?? 0,
      rules: draft.rules.map((r) => ({
        name: r.name,
        description: r.description || "",
        path: r.path,
        range_mode: r.range_mode ?? 0,
        range_value_currency: r.range_value_currency || "",
        range_value_currency_digits: r.range_value_currency_digits ?? 2,
        tiers: (r.tiers || []).map((t) => ({
          range_to: t.range_to ?? 0,
          margin_rate_initial: t.margin_rate_initial ?? 0,
          margin_rate_maintenance: t.margin_rate_maintenance ?? 0,
        })),
      })),
    };
    const res = isNew ? await createLeverage(body) : await updateLeverage(profileId, body);
    if (!res.ok) {
      setError(res.message || "save failed");
      return;
    }
    onSaved();
    onClose();
  }

  const saveRule = (rule) => {
    setDraft((prev) => ({
      ...prev,
      rules: editor.index == null
        ? [...prev.rules, rule]
        : prev.rules.map((r, i) => (i === editor.index ? rule : r)),
    }));
    setEditor(null);
  };

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
          title={isNew ? "Leverage: New" : `Leverage: ${draft?.name ?? "…"}`}
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
              <div className="sym-sessions-intro">
                <span className="sym-tab-intro-icon" aria-hidden="true">
                  <Icon id="leverages" size={48} />
                </span>
                <p>
                  Floating leverage profile: rules are checked top to bottom and the first matching
                  rule per symbol applies.
                </p>
              </div>
              <div className="form-grid">
                <label>Name</label>
                <input type="text" value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
              </div>
              <div className="df-table-panel">
                <div className="df-table-toolbar">
                  <button type="button" onClick={() => setEditor({ index: null })}>Add</button>
                  <button type="button" disabled={!draft.rules.length} onClick={() => setEditor({ index: selected })}>
                    Edit
                  </button>
                  <button
                    type="button"
                    disabled={!draft.rules.length}
                    onClick={() => setDraft({ ...draft, rules: draft.rules.filter((_, i) => i !== selected) })}
                  >
                    Delete
                  </button>
                </div>
                <div className="df-table-main">
                  <table className="data-table data-table-grid df-sub-table">
                    <thead>
                      <tr>
                        <th>Name</th>
                        <th>Symbols</th>
                        <th>Range</th>
                        <th>Tiers</th>
                      </tr>
                    </thead>
                    <tbody>
                      {draft.rules.map((r, i) => (
                        <tr
                          key={i}
                          className={i === selected ? "selected" : ""}
                          onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                          onDoubleClick={() => setEditor({ index: i })}
                        >
                          <td>{r.name}</td>
                          <td>{r.path}</td>
                          <td>{RANGE_MODE[r.range_mode]}</td>
                          <td>{(r.tiers || []).map((t) => `${tierTo(t)} - ${t.margin_rate_initial}`).join("; ")}</td>
                        </tr>
                      ))}
                      {!draft.rules.length && (
                        <tr>
                          <td colSpan={4} className="df-empty">No rules</td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}
        </SettingsDialog>
        {editor && (
          <RuleEditor
            rule={editor.index == null ? null : draft.rules[editor.index]}
            onSave={saveRule}
            onClose={() => setEditor(null)}
          />
        )}
      </div>
    </div>
  );
}

/** Leverage profiles list. */
export function LeveragesModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_groups !== false;

  const load = () => fetchLeverages().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
  }, []);

  async function onDelete(row) {
    if (!window.confirm(`Delete leverage profile '${row.name}'?`)) return;
    const res = await deleteLeverage(row.leverage_id);
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
              <th>Name</th>
              <th>Rules</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((row, i) => (
              <tr
                key={row.leverage_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ id: row.leverage_id })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="leverages" />
                    {row.name}
                  </span>
                </td>
                <td>{(row.rules || []).map(ruleSummary).join(" · ") || "—"}</td>
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
            { label: "Add", onClick: () => setDialog({ id: "new" }) },
            { label: "Edit", disabled: selected == null, onClick: () => setDialog({ id: rows[selected]?.leverage_id }) },
            "sep",
            { label: "Delete", disabled: selected == null, onClick: () => onDelete(rows[selected]) },
          ]}
        />
      )}
      {dialog && (
        <LeverageDialog profileId={dialog.id} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}
