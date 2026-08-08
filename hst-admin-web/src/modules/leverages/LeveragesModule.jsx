import { useEffect, useRef, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu, gridMenuHead, gridMenuTail, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { SymbolScopeSelectField } from "@/components/ui/SymbolScopeSelectField.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
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

const emptyTier = () => ({ range_to: "∞", margin_rate_initial: "1.00", margin_rate_maintenance: "1.00" });

// the editor keeps every numeric cell as typed text, so "1." and "" survive keystrokes
const toDraftRule = (rule) =>
  rule
    ? {
        ...rule,
        tiers: (rule.tiers || []).map((t, i, arr) => ({
          range_to: i === arr.length - 1 ? "∞" : String(Number(t.range_to) === INFINITY_TO ? 0 : t.range_to),
          margin_rate_initial: String(t.margin_rate_initial ?? 0),
          margin_rate_maintenance: String(t.margin_rate_maintenance ?? 0),
        })),
      }
    : { name: "", description: "", path: "*", range_mode: 0, range_value_currency: "", tiers: [emptyTier()] };

const numOf = (v) => (v === "" || v === "∞" ? 0 : Number(v) || 0);

/** The rule editor: name, symbol mask, range mode and the tier ladder. */
function RuleEditor({ rule, onSave, onClose }) {
  const [draft, setDraft] = useState(() => toDraftRule(rule));
  const [selectedTier, setSelectedTier] = useState(0);
  const [selectAllTiers, setSelectAllTiers] = useState(false);
  const [menu, setMenu] = useState(null);
  const tierInputRefs = useRef([]);
  const close = useDialogStack(onClose);
  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  const setTier = (i, key, value) =>
    set("tiers", draft.tiers.map((t, j) => (j === i ? { ...t, [key]: value } : t)));

  const addTier = () => {
    setDraft((prev) => {
      const tiers = [...prev.tiers];
      const insertAt = tiers.length > 0 ? tiers.length - 1 : 0;
      tiers.splice(insertAt, 0, { range_to: "10", margin_rate_initial: "1.00", margin_rate_maintenance: "1.00" });
      setSelectedTier(insertAt);
      setSelectAllTiers(false);
      return { ...prev, tiers };
    });
  };

  const deleteTier = (index) => {
    if (draft.tiers.length < 2 || index >= draft.tiers.length - 1) return;
    setDraft((prev) => ({ ...prev, tiers: prev.tiers.filter((_, j) => j !== index) }));
    setSelectedTier((i) => Math.max(0, Math.min(i, draft.tiers.length - 2)));
    setSelectAllTiers(false);
  };

  const openTierMenu = (e, index) => {
    e.preventDefault();
    e.stopPropagation();
    setSelectedTier(index);
    setSelectAllTiers(false);
    setMenu({ x: e.clientX, y: e.clientY, index });
  };

  const editTier = () => {
    const el = tierInputRefs.current[menu?.index ?? selectedTier]?.[0];
    el?.focus();
    el?.select();
  };

  const copyTier = () => {
    const i = menu?.index ?? selectedTier;
    const t = draft.tiers[i];
    if (!t) return;
    const text = [t.range_to, t.margin_rate_initial, t.margin_rate_maintenance].join("\t");
    navigator.clipboard?.writeText(text).catch(() => {});
  };

  const currencyRequired = needsCurrency(draft.range_mode);
  const canSave =
    draft.name.trim() && draft.path.trim() && (!currencyRequired || draft.range_value_currency.trim());

  const submit = () =>
    canSave &&
    onSave({
      ...draft,
      range_mode: Number(draft.range_mode),
      tiers: draft.tiers.map((t, i) => ({
        range_to: i === draft.tiers.length - 1 ? 0 : numOf(t.range_to),
        margin_rate_initial: numOf(t.margin_rate_initial),
        margin_rate_maintenance: numOf(t.margin_rate_maintenance),
      })),
    });

  return (
    <DialogOverlay className="sym-session-dialog-overlay" onClose={onClose}>
      <div className="sym-session-dialog lev-rule-dialog" role="dialog">
        <div className="sym-session-dialog-title">
          <span>Leverage Rule</span>
          <span className="lev-title-glyphs">
            <button type="button" className="lev-title-btn" aria-label="Help" disabled>?</button>
            <button type="button" className="lev-title-btn" aria-label="Close" onClick={close}>✕</button>
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
            <SymbolScopeSelectField
              label="Symbol"
              value={draft.path}
              onChange={(v) => set("path", v)}
            />
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
          <table
            className="data-table data-table-grid df-sub-table lev-tier-table"
            onContextMenu={(e) => {
              const row = e.target.closest("tbody tr[data-tier-index]");
              if (!row) return;
              openTierMenu(e, Number(row.dataset.tierIndex));
            }}
          >
            <thead>
              <tr>
                <th className="lev-icon-col" />
                <th className="num">To</th>
                <th className="num">Initial margin rate</th>
                <th className="num">Maintenance margin rate</th>
              </tr>
            </thead>
            <tbody>
              {draft.tiers.map((t, i) => {
                const isLast = i === draft.tiers.length - 1;
                if (!tierInputRefs.current[i]) tierInputRefs.current[i] = [];
                return (
                  <tr
                    key={i}
                    data-tier-index={i}
                    className={selectAllTiers || i === selectedTier ? "selected" : ""}
                    onClick={() => {
                      setSelectedTier(i);
                      setSelectAllTiers(false);
                    }}
                    onContextMenu={(e) => openTierMenu(e, i)}
                  >
                    <td className="lev-icon-col">
                      <Icon id="leverages" />
                    </td>
                    <td className="num">
                      {isLast ? (
                        <span className="lev-tier-infinity" title="No upper limit">∞</span>
                      ) : (
                        <input
                          ref={(el) => {
                            tierInputRefs.current[i][0] = el;
                          }}
                          type="text"
                          className="df-cell-input lev-num-input"
                          value={t.range_to}
                          onContextMenu={(e) => openTierMenu(e, i)}
                          onChange={(e) => setTier(i, "range_to", e.target.value)}
                        />
                      )}
                    </td>
                    {["margin_rate_initial", "margin_rate_maintenance"].map((key, col) => (
                      <td key={key} className="num">
                        <input
                          ref={(el) => {
                            tierInputRefs.current[i][col + 1] = el;
                          }}
                          type="text"
                          className="df-cell-input lev-num-input"
                          value={t[key]}
                          onContextMenu={(e) => openTierMenu(e, i)}
                          onChange={(e) => setTier(i, key, e.target.value)}
                        />
                      </td>
                    ))}
                  </tr>
                );
              })}
              <tr>
                <td colSpan={4} className="df-cell-editable" onClick={addTier}>
                  + click to add…
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div className="sym-session-dialog-footer">
          <div className="sym-session-dialog-actions">
            <button type="button" className="sym-session-ok" disabled={!canSave} onClick={submit}>
              OK
            </button>
            <button type="button" className="sym-session-cancel" onClick={close}>
              Cancel
            </button>
          </div>
        </div>
        {menu && (
          <ContextMenu
            x={menu.x}
            y={menu.y}
            onClose={() => setMenu(null)}
            items={[
              ...gridMenuHead({
                onAdd: addTier,
                onEdit: editTier,
                onDelete: () => deleteTier(menu.index),
                hasSelection: draft.tiers.length > 0,
              }).map((item) =>
                item.label === "Delete"
                  ? {
                      ...item,
                      disabled: draft.tiers.length < 2 || menu.index >= draft.tiers.length - 1,
                    }
                  : item,
              ),
              ...gridMenuTail({
                onSelectAll: () => setSelectAllTiers(true),
                onCopy: copyTier,
              }),
            ]}
          />
        )}
      </div>
    </DialogOverlay>
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
  const close = useDialogStack(onClose);

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
    close();
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
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          height={520}
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          title="Leverage"
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
    </DialogOverlay>
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
    const hasRow = selected != null && selected !== "add";
    const head = listMenuHead({
      onAdd: () => setDialog({ id: "new" }),
      onEdit: () => setDialog({ id: rows[selected]?.leverage_id }),
      onDelete: () => onDelete(rows[selected]),
      hasSelection: hasRow,
    });
    head.splice(2, 0, { label: "Edit Groups", disabled: true });
    return [
      ...head,
      "sep",
      { label: "Groups", disabled: true },
      { label: "Assign", items: maskItems("To") },
      { label: "Remove", items: maskItems("From") },
      ...listMenuTail({
        on: {
          sort: () => setRows([...(rows || [])].sort((a, b) => a.name.localeCompare(b.name))),
        },
        extras: [
          "sep",
          { label: "Enable", disabled: true },
          { label: "Disable", disabled: true },
        ],
      }),
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
                onContextMenu={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  setSelected(i);
                  setMenu({ x: e.clientX, y: e.clientY });
                }}
                onDoubleClick={() => canEdit && setDialog({ id: row.leverage_id })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="leverages" />
                    {row.name}
                  </span>
                </td>
                <td className="lev-rules-cell">{(row.rules || []).map(ruleSummary).join(", ") || "—"}</td>
              </tr>
            ))}
            {rows && rows.length === 0 && (
              <tr>
                <td colSpan={2} className="df-empty">No leverage profiles configured</td>
              </tr>
            )}
            {canEdit && (
              <tr
                className={`df-add-row${selected === "add" ? " selected" : ""}`}
                onClick={() => setSelected("add")}
                onDoubleClick={() => setDialog({ id: "new" })}
              >
                <td colSpan={2}>
                  <span className="df-add-plus" aria-hidden="true">+</span>
                  click to add…
                </td>
              </tr>
            )}
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
