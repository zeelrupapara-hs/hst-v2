import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { SymbolScopeSelectField } from "@/components/ui/SymbolScopeSelectField.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { createGroupCommission, updateGroupCommission } from "@/api/endpoints/groups.js";
import { GroupTabIntro } from "./GroupTabs.jsx";
import { FlagCombo } from "@/modules/symbols/SymbolTabs.jsx";
import {
  COMMISSION_ACTION,
  COMMISSION_CHARGE,
  COMMISSION_CHARGE_INSTANT,
  COMMISSION_ENTRY,
  COMMISSION_MODE,
  COMMISSION_MODE_FEE,
  COMMISSION_PROFIT,
  COMMISSION_RANGE,
  COMMISSION_REASON_FLAGS,
  TIER_MODE,
  TIER_TYPE,
  commissionFromRow,
  commissionToBody,
  enumOptions,
  newCommissionTier,
  reasonFlagsForDisplay,
  reasonFlagsForSave,
} from "@/lib/commissionConfig.js";

/** Group commission editor — MT5 Manager layout (header, category radios, deal filters, tier ladder). */
export function CommissionDialog({ groupId, commission, onClose, onSaved }) {
  const isNew = !commission;
  const [draft, setDraft] = useState(() => commissionFromRow(commission));
  const [selected, setSelected] = useState(0);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(commission?.commission_id ?? "new");
  const close = useDialogStack(onClose);
  const isFee = draft.mode === COMMISSION_MODE_FEE;

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));
  const setTier = (i, key, value) =>
    setDraft((prev) => ({
      ...prev,
      tiers: prev.tiers.map((t, j) => (j === i ? { ...t, [key]: value } : t)),
    }));

  function setMode(mode) {
    setDraft((prev) => ({
      ...prev,
      mode,
      mode_charge: mode === COMMISSION_MODE_FEE ? COMMISSION_CHARGE_INSTANT : prev.mode_charge,
      mode_entry: mode === COMMISSION_MODE_FEE ? (prev.mode_entry ?? 0) : prev.mode_entry,
      mode_action: mode === COMMISSION_MODE_FEE ? (prev.mode_action ?? 0) : prev.mode_action,
      mode_profit: mode === COMMISSION_MODE_FEE ? (prev.mode_profit ?? 0) : prev.mode_profit,
      mode_reason:
        mode === COMMISSION_MODE_FEE
          ? prev.mode_reason == null || prev.mode_reason === 0
            ? 0
            : prev.mode_reason
          : prev.mode_reason,
    }));
  }

  async function handleOk() {
    if (!draft.name?.trim()) {
      setError("Name is required");
      return;
    }
    const body = commissionToBody(draft);
    const res = isNew
      ? await createGroupCommission(groupId, body)
      : await updateGroupCommission(groupId, commission.commission_id, body);
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
          className="sym-config-window grp-comm-dialog"
          width={787}
          height={620}
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          title="Commissions"
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active grp-comm-panel">
            <GroupTabIntro>Adjustment of commissions parameters of the symbol.</GroupTabIntro>

            <div className="grp-comm-form">
              <label className="grp-comm-label">Name</label>
              <input
                type="text"
                className="grp-comm-field-wide"
                value={draft.name}
                onChange={(e) => set("name", e.target.value)}
              />

              <label className="grp-comm-label">Description</label>
              <input
                type="text"
                className="grp-comm-field-wide"
                value={draft.description ?? ""}
                onChange={(e) => set("description", e.target.value)}
              />

              <div className="grp-comm-body">
                <div className="grp-comm-left">
                  <SymbolScopeSelectField
                    label="Symbol"
                    value={draft.path ?? "*"}
                    onChange={(v) => set("path", v || "*")}
                  />
                  <label className="grp-comm-label">Range</label>
                  <PropSelect
                    fill
                    className="grp-comm-select"
                    value={draft.mode_range}
                    options={enumOptions(COMMISSION_RANGE)}
                    onChange={(v) => set("mode_range", v)}
                  />
                  <label className="grp-comm-label">Charge</label>
                  <PropSelect
                    fill
                    className="grp-comm-select"
                    value={draft.mode_charge}
                    options={enumOptions(COMMISSION_CHARGE)}
                    disabled={isFee}
                    onChange={(v) => set("mode_charge", v)}
                  />
                  <label className="grp-comm-label">Turnover currency</label>
                  <input
                    type="text"
                    className="grp-comm-select"
                    value={draft.turnover_currency ?? ""}
                    onChange={(e) => set("turnover_currency", e.target.value.toUpperCase())}
                  />
                </div>

                <fieldset className="grp-comm-types">
                  {Object.entries(COMMISSION_MODE).map(([value, label]) => (
                    <label key={value} className="grp-comm-type-opt">
                      <input
                        type="radio"
                        name="commission-mode"
                        checked={draft.mode === Number(value)}
                        onChange={() => setMode(Number(value))}
                      />
                      <span>{label}</span>
                    </label>
                  ))}
                </fieldset>
              </div>

              {isFee && (
                <div className="grp-comm-filters">
                  <label className="grp-comm-label">Deal entry</label>
                  <PropSelect
                    fill
                    className="grp-comm-select"
                    value={draft.mode_entry ?? 0}
                    options={enumOptions(COMMISSION_ENTRY)}
                    onChange={(v) => set("mode_entry", v)}
                  />
                  <label className="grp-comm-label">Deal action</label>
                  <PropSelect
                    fill
                    className="grp-comm-select"
                    value={draft.mode_action ?? 0}
                    options={enumOptions(COMMISSION_ACTION)}
                    onChange={(v) => set("mode_action", v)}
                  />
                  <label className="grp-comm-label">Deal profit</label>
                  <PropSelect
                    fill
                    className="grp-comm-select"
                    value={draft.mode_profit ?? 0}
                    options={enumOptions(COMMISSION_PROFIT)}
                    onChange={(v) => set("mode_profit", v)}
                  />
                  <label className="grp-comm-label">Deal reason</label>
                  <FlagCombo
                    hideLabel
                    labels={COMMISSION_REASON_FLAGS}
                    value={reasonFlagsForDisplay(draft.mode_reason)}
                    emptyLabel="None"
                    onChange={(v) => set("mode_reason", reasonFlagsForSave(v))}
                  />
                </div>
              )}
            </div>

            <div className="df-table-panel grp-comm-tiers">
              <div className="df-table-toolbar">
                <button
                  type="button"
                  onClick={() =>
                    setDraft((prev) => ({
                      ...prev,
                      tiers: [...prev.tiers, newCommissionTier()],
                    }))
                  }
                >
                  Add
                </button>
                <button type="button" disabled={!draft.tiers.length}>
                  Edit
                </button>
                <button
                  type="button"
                  disabled={!draft.tiers.length}
                  onClick={() =>
                    setDraft((prev) => ({
                      ...prev,
                      tiers: prev.tiers.filter((_, i) => i !== selected),
                    }))
                  }
                >
                  Delete
                </button>
              </div>
              <div className="df-table-main">
                <table className="data-table data-table-grid df-sub-table grp-comm-tier-table">
                  <thead>
                    <tr>
                      <th>From</th>
                      <th>To</th>
                      <th>Commission</th>
                      <th>Minimal</th>
                      <th>Maximal</th>
                      <th>Mode</th>
                      <th>Currency</th>
                      <th>Type</th>
                    </tr>
                  </thead>
                  <tbody>
                    {draft.tiers.map((t, i) => (
                      <tr
                        key={i}
                        className={i === selected ? "selected" : ""}
                        onClick={() => setSelected(i)}
                      >
                        {["range_from", "range_to", "value", "minimal", "maximal"].map((key) => (
                          <td key={key}>
                            <input
                              type="text"
                              className="df-cell-input"
                              value={t[key] ?? 0}
                              disabled={key === "maximal"}
                              title={key === "maximal" ? "Not stored yet" : undefined}
                              onChange={(e) => setTier(i, key, Number(e.target.value) || 0)}
                            />
                          </td>
                        ))}
                        <td>
                          <PropSelect
                            fill
                            value={t.mode}
                            options={enumOptions(TIER_MODE)}
                            onChange={(v) => setTier(i, "mode", v)}
                          />
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
                          <PropSelect
                            fill
                            value={t.type}
                            options={enumOptions(TIER_TYPE)}
                            onChange={(v) => setTier(i, "type", v)}
                          />
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
    </DialogOverlay>
  );
}
