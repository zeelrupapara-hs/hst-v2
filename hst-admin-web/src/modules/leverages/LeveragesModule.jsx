import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu, listMenuHead } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { fetchSymbols } from "@/api/endpoints/symbols.js";
import {
  createLeverage,
  deleteLeverage,
  fetchLeverage,
  fetchLeverages,
  updateLeverage,
} from "@/api/endpoints/leverages.js";

// mirrors model.RangeMode_name
const RANGE_MODE_NAME = {
  0: "volume",
  1: "volume_per_symbol",
  2: "notional_value",
  3: "notional_value_per_symbol",
};

const rangeWord = (mode) => (RANGE_MODE_NAME[mode] || "").replace(/_/g, " ");
const rangeLabel = (mode) => rangeWord(mode).replace(/^./, (c) => c.toUpperCase());
const needsCurrency = (mode) => Number(mode) === 2 || Number(mode) === 3;

const INFINITY_TO = 0;
const fmtNum = (n) => Number(n || 0).toLocaleString("en-US").replace(/,/g, " ");
const fmtRate = (n) => Number(n || 0).toFixed(2);
const tierTo = (t) => (Number(t.range_to) === INFINITY_TO ? "∞" : fmtNum(t.range_to));

// the platform prints each level as "<from>-<to> - <initial> - <maintenance>"
const tierSummary = (tiers = []) => {
  let from = 0;
  return tiers
    .map((t) => {
      const text = `${fmtNum(from)}-${tierTo(t)} - ${fmtRate(t.margin_rate_initial)} - ${fmtRate(t.margin_rate_maintenance)}`;
      from = Number(t.range_to) || 0;
      return text;
    })
    .join("; ");
};

const ruleSummary = (r) => `${r.name}: ${r.path}`;

const emptyTier = () => ({ range_to: "0", margin_rate_initial: "1.00", margin_rate_maintenance: "1.00" });

// the editor keeps every numeric cell as typed text, so "1." and "" survive keystrokes
const toDraftRule = (rule) =>
  rule
    ? { ...rule, tiers: (rule.tiers || []).map((t) => ({
        range_to: Number(t.range_to) === INFINITY_TO ? "∞" : String(t.range_to),
        margin_rate_initial: String(t.margin_rate_initial ?? 0),
        margin_rate_maintenance: String(t.margin_rate_maintenance ?? 0),
      })) }
    : { name: "", description: "", path: "*", range_mode: 0, range_value_currency: "", tiers: [emptyTier()] };

const numOf = (v) => (v === "" || v === "∞" ? 0 : Number(v) || 0);

/** The rule editor: name, symbol mask, range mode and the tier ladder. */
function RuleEditor({ rule, onSave, onClose }) {
  const [draft, setDraft] = useState(() => toDraftRule(rule));
  const [paths, setPaths] = useState([]);
  const [menu, setMenu] = useState(null);
  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  useEffect(() => {
    fetchSymbols().then((res) => {
      if (!res.ok) return;
      setPaths([...new Set((res.data || []).map((s) => s.path || s.symbol).filter(Boolean))]);
    });
  }, []);

  const setTier = (i, key, value) =>
    set("tiers", draft.tiers.map((t, j) => (j === i ? { ...t, [key]: value } : t)));

  const addTier = () => set("tiers", [...draft.tiers, emptyTier()]);

  const currencyRequired = needsCurrency(draft.range_mode);
  const canSave =
    draft.name.trim() && draft.path.trim() && (!currencyRequired || draft.range_value_currency.trim());

  const submit = () =>
    canSave &&
    onSave({
      ...draft,
      range_mode: Number(draft.range_mode),
      tiers: draft.tiers.map((t) => ({
        range_to: numOf(t.range_to),
        margin_rate_initial: numOf(t.margin_rate_initial),
        margin_rate_maintenance: numOf(t.margin_rate_maintenance),
      })),
    });

  return (
    <div className="sym-session-dialog-overlay" onClick={onClose}>
      <div className="sym-session-dialog lev-rule-dialog" role="dialog" onClick={(e) => e.stopPropagation()}>
        <div className="sym-session-dialog-title">
          <span>Leverage Rule</span>
          <span className="lev-title-glyphs">
            <button type="button" className="lev-title-btn" aria-label="Help" disabled>?</button>
            <button type="button" className="lev-title-btn" aria-label="Close" onClick={onClose}>✕</button>
          </span>
        </div>
        <div className="sym-session-dialog-body">
          <div className="sym-sessions-intro">
            <span className="sym-tab-intro-icon" aria-hidden="true">
              <Icon id="leverages" size={48} />
            </span>
            <p>
              Set up a <b>leverage rule</b>. Select a symbol and specify leverage levels based on
              position size or value. You can check the total volume/value for the selected group of
              symbols or a separate symbol by selecting the corresponding range option.
            </p>
          </div>
          <div className="form-grid">
            <label>Name</label>
            <input type="text" className="wide" value={draft.name} onChange={(e) => set("name", e.target.value)} />
            <label>Description</label>
            <input type="text" className="wide" value={draft.description} onChange={(e) => set("description", e.target.value)} />
            <label>Symbol</label>
            <input
              type="text"
              className="wide"
              list="lev-symbol-paths"
              value={draft.path}
              onChange={(e) => set("path", e.target.value)}
            />
            <datalist id="lev-symbol-paths">
              {paths.map((p) => (
                <option key={p} value={p} />
              ))}
            </datalist>
            <label>Range</label>
            <PropSelect
              value={Number(draft.range_mode)}
              options={Object.keys(RANGE_MODE_NAME).map((v) => ({ value: Number(v), label: rangeLabel(v) }))}
              onChange={(v) => set("range_mode", v)}
            />
            <label>Currency</label>
            <input
              type="text"
              disabled={!currencyRequired}
              value={draft.range_value_currency}
              onChange={(e) => set("range_value_currency", e.target.value.toUpperCase())}
            />
          </div>
          <table className="data-table data-table-grid df-sub-table lev-tier-table">
            <thead>
              <tr>
                <th className="lev-icon-col" />
                <th className="num">To</th>
                <th className="num">Initial margin rate</th>
                <th className="num">Maintenance margin rate</th>
              </tr>
            </thead>
            <tbody>
              {draft.tiers.map((t, i) => (
                <tr
                  key={i}
                  onContextMenu={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setMenu({ x: e.clientX, y: e.clientY, index: i });
                  }}
                >
                  <td className="lev-icon-col">
                    <Icon id="leverages" />
                  </td>
                  {["range_to", "margin_rate_initial", "margin_rate_maintenance"].map((key) => (
                    <td key={key} className="num">
                      <input
                        type="text"
                        className="df-cell-input lev-num-input"
                        value={t[key]}
                        onChange={(e) => setTier(i, key, e.target.value)}
                      />
                    </td>
                  ))}
                </tr>
              ))}
              <tr>
                <td colSpan={4} className="df-cell-editable" onClick={addTier}>
                  + click to add…
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div className="sym-session-dialog-footer lev-footer-centered">
          <button type="button" className="sym-session-ok" disabled={!canSave} onClick={submit}>
            OK
          </button>
          <button type="button" className="sym-session-cancel lev-default-btn" onClick={onClose}>Cancel</button>
        </div>
        {menu && (
          <ContextMenu
            x={menu.x}
            y={menu.y}
            onClose={() => setMenu(null)}
            items={[
              { label: "Add", shortcut: "Ctrl+N", onClick: addTier },
              {
                label: "Delete",
                shortcut: "Ctrl+D",
                disabled: draft.tiers.length < 2,
                onClick: () => set("tiers", draft.tiers.filter((_, j) => j !== menu.index)),
              },
            ]}
          />
        )}
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
  const [menu, setMenu] = useState(null);
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

  const moveRule = (from, to) =>
    setDraft((prev) => {
      const rules = [...prev.rules];
      const [moved] = rules.splice(from, 1);
      rules.splice(to, 0, moved);
      return { ...prev, rules };
    });

  return (
    <div className="dialog-overlay" onClick={onClose} role="presentation">
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          draggable
          onClose={onClose}
          onTitlePointerDown={onTitlePointerDown}
          title="Leverage"
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
                  Create a list of rules to implement floating leverage for a group of financial
                  instruments. To apply the rules, choose the leverage configuration in the group&apos;s
                  margin settings. Use Automations for scheduled leverage adjustments.
                </p>
              </div>
              <div className="form-grid">
                <label>Name</label>
                <input type="text" className="wide" value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
              </div>
              <div className="df-table-panel">
                <div className="df-table-main">
                  <table className="data-table data-table-grid df-sub-table">
                    <thead>
                      <tr>
                        <th>Name</th>
                        <th>Symbols</th>
                        <th>Range</th>
                        <th>Currency</th>
                        <th>Tiers</th>
                      </tr>
                    </thead>
                    <tbody>
                      {draft.rules.map((r, i) => (
                        <tr
                          key={i}
                          className={i === selected ? "selected" : ""}
                          onClick={() => setSelected(i)}
                          onContextMenu={(e) => {
                            e.preventDefault();
                            setSelected(i);
                            setMenu({ x: e.clientX, y: e.clientY, index: i });
                          }}
                          onDoubleClick={() => setEditor({ index: i })}
                        >
                          <td>
                            <span className="sym-symbol-cell">
                              <Icon id="leverages" />
                              {r.name}
                            </span>
                          </td>
                          <td>{r.path}</td>
                          <td>{rangeWord(r.range_mode)}</td>
                          <td>{r.range_value_currency}</td>
                          <td>{tierSummary(r.tiers)}</td>
                        </tr>
                      ))}
                      <tr>
                        <td colSpan={5} className="df-cell-editable" onClick={() => setEditor({ index: null })}>
                          + click to add…
                        </td>
                      </tr>
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
                      onAdd: () => setEditor({ index: null }),
                      onEdit: () => setEditor({ index: menu.index }),
                      onDelete: () =>
                        setDraft({ ...draft, rules: draft.rules.filter((_, i) => i !== menu.index) }),
                      hasSelection: draft.rules.length > 0,
                    }),
                    "sep",
                    {
                      label: "Move Up",
                      disabled: menu.index < 1,
                      onClick: () => moveRule(menu.index, menu.index - 1),
                    },
                    {
                      label: "Move Down",
                      disabled: menu.index >= draft.rules.length - 1,
                      onClick: () => moveRule(menu.index, menu.index + 1),
                    },
                  ]}
                />
              )}
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

const GROUP_MASKS = ["All", "real*", "demo*", "managers*", "contest*", "Risk*", "coverage*"];
const maskItems = (verb) =>
  GROUP_MASKS.map((m) => ({ label: `${verb} ${m} Groups`, disabled: true }));

/** Leverage profiles list. */
export function LeveragesModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_groups !== false;

  // the list endpoint answers without rules, so the Rules column needs each profile's tree
  const load = async () => {
    const res = await fetchLeverages();
    if (!res.ok) return;
    const list = res.data || [];
    const details = await Promise.all(list.map((r) => fetchLeverage(r.leverage_id)));
    setRows(list.map((r, i) => (details[i].ok ? { ...r, rules: details[i].data.rules || [] } : r)));
  };

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

  const menuItems = () => {
    const head = listMenuHead({
      onAdd: () => setDialog({ id: "new" }),
      onEdit: () => setDialog({ id: rows[selected]?.leverage_id }),
      onDelete: () => onDelete(rows[selected]),
      hasSelection: selected != null,
    });
    head.splice(2, 0, { label: "Edit Groups", disabled: true });
    return [
      ...head,
      "sep",
      { label: "Groups", disabled: true },
      { label: "Assign", items: maskItems("To") },
      { label: "Remove", items: maskItems("From") },
      "sep",
      { label: "Move Up", disabled: true },
      { label: "Move Down", disabled: true },
      {
        label: "Sort Alphabetically",
        onClick: () => setRows([...(rows || [])].sort((a, b) => a.name.localeCompare(b.name))),
      },
      "sep",
      { label: "Enable", disabled: true },
      { label: "Disable", disabled: true },
    ];
  };

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
                <td className="lev-rules-cell">{(row.rules || []).map(ruleSummary).join(", ")}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {menu && canEdit && (
        <ContextMenu x={menu.x} y={menu.y} onClose={() => setMenu(null)} items={menuItems()} />
      )}
      {dialog && (
        <LeverageDialog profileId={dialog.id} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}
