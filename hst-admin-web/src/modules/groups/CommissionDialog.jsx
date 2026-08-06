import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { createGroupCommission, updateGroupCommission } from "@/api/endpoints/groups.js";
import { GroupTabIntro } from "./GroupTabs.jsx";

// Mirrors hst-server/model/commission.go.
const MODE = { 0: "Standard", 1: "Agent" };
const RANGE = { 0: "Volume", 1: "Turnover (money)", 2: "Turnover (volume)" };
const CHARGE = { 0: "Daily", 1: "Monthly", 2: "Instant" };
const TIER_MODE = {
  0: "Deposit currency",
  1: "Base currency",
  2: "Profit currency",
  3: "Margin currency",
  4: "Points",
  5: "Percent",
  6: "Specified currency",
};
const TIER_TYPE = { 0: "Per trade", 1: "Per volume" };

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

const newTier = () => ({
  mode: 0,
  type: 0,
  value: 0,
  range_from: 0,
  range_to: 0,
  minimal: 0,
  currency: "",
});

/** The commission entry: header fields over the tier ladder; ranges apply first-match. */
export function CommissionDialog({ groupId, commission, onClose, onSaved }) {
  const isNew = !commission;
  const [draft, setDraft] = useState(
    commission
      ? { ...commission, tiers: (commission.tiers || []).map((t) => ({ ...t })) }
      : { name: "", description: "", path: "*", mode: 0, mode_range: 0, mode_charge: 2, turnover_currency: "", tiers: [newTier()] },
  );
  const [selected, setSelected] = useState(0);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(commission?.commission_id ?? "new");

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));
  const setTier = (i, key, value) =>
    set("tiers", draft.tiers.map((t, j) => (j === i ? { ...t, [key]: value } : t)));

  async function handleOk() {
    if (!draft.name?.trim()) {
      setError("Name is required");
      return;
    }
    const body = {
      name: draft.name.trim(),
      description: draft.description || "",
      path: draft.path || "*",
      mode: draft.mode ?? 0,
      mode_range: draft.mode_range ?? 0,
      mode_charge: draft.mode_charge ?? 2,
      turnover_currency: draft.turnover_currency || "",
      tiers: draft.tiers.map((t) => ({
        mode: t.mode ?? 0,
        type: t.type ?? 0,
        value: t.value ?? 0,
        range_from: t.range_from ?? 0,
        range_to: t.range_to ?? 0,
        minimal: t.minimal ?? 0,
        currency: t.currency || "",
      })),
    };
    const res = isNew
      ? await createGroupCommission(groupId, body)
      : await updateGroupCommission(groupId, commission.commission_id, body);
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
          height={540}
          onClose={onClose}
          onTitlePointerDown={onTitlePointerDown}
          title={`Commission: ${draft.name || "New"}`}
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
            <GroupTabIntro>
              Commission charged for the clients of this group. Levels apply to the first matching
              range; ranges must not overlap.
            </GroupTabIntro>
            <div className="form-grid sym-form-two-col">
              <label>Name</label>
              <input type="text" value={draft.name} onChange={(e) => set("name", e.target.value)} />
              <label>Description</label>
              <input type="text" value={draft.description ?? ""} onChange={(e) => set("description", e.target.value)} />
              <label>Symbol</label>
              <input type="text" value={draft.path ?? "*"} onChange={(e) => set("path", e.target.value)} />
              <label>Type</label>
              <PropSelect fill value={draft.mode} options={enumOptions(MODE)} onChange={(v) => set("mode", v)} />
              <label>Range</label>
              <PropSelect fill value={draft.mode_range} options={enumOptions(RANGE)} onChange={(v) => set("mode_range", v)} />
              <label>Charge</label>
              <PropSelect fill value={draft.mode_charge} options={enumOptions(CHARGE)} onChange={(v) => set("mode_charge", v)} />
              <label>Turnover currency</label>
              <input
                type="text"
                value={draft.turnover_currency ?? ""}
                onChange={(e) => set("turnover_currency", e.target.value.toUpperCase())}
              />
            </div>
            <div className="df-table-panel">
              <div className="df-table-toolbar">
                <button type="button" onClick={() => set("tiers", [...draft.tiers, newTier()])}>Add</button>
                <button
                  type="button"
                  disabled={!draft.tiers.length}
                  onClick={() => set("tiers", draft.tiers.filter((_, i) => i !== selected))}
                >
                  Delete
                </button>
              </div>
              <div className="df-table-main">
                <table className="data-table data-table-grid df-sub-table">
                  <thead>
                    <tr>
                      <th>From</th>
                      <th>To</th>
                      <th>Commission</th>
                      <th>Minimal</th>
                      <th>Mode</th>
                      <th>Currency</th>
                      <th>Type</th>
                    </tr>
                  </thead>
                  <tbody>
                    {draft.tiers.map((t, i) => (
                      <tr key={i} className={i === selected ? "selected" : ""} onClick={() => setSelected(i)}>
                        {["range_from", "range_to", "value", "minimal"].map((key) => (
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
                          <PropSelect fill value={t.mode} options={enumOptions(TIER_MODE)} onChange={(v) => setTier(i, "mode", v)} />
                        </td>
                        <td>
                          <input
                            type="text"
                            className="df-cell-input"
                            value={t.currency ?? ""}
                            onChange={(e) => setTier(i, "currency", e.target.value.toUpperCase())}
                          />
                        </td>
                        <td>
                          <PropSelect fill value={t.type} options={enumOptions(TIER_TYPE)} onChange={(v) => setTier(i, "type", v)} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </SettingsDialog>
      </div>
    </div>
  );
}
