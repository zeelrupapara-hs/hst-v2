import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { updateGroupSymbol } from "@/api/endpoints/groups.js";
import {
  CalcMode_name,
  ExecMode_name,
  SwapMode_name,
  TradeMode_name,
  toLots,
  fromLots,
} from "@/constants/symbols.js";
import { GroupTabIntro } from "./GroupTabs.jsx";

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

// Which override columns belong to which Use-default cluster.
const SECTIONS = {
  common: ["spread_diff", "spread_diff_balance", "stops_level", "freeze_level"],
  trade: ["trade_mode", "fill_flags", "expir_flags", "volume_min", "volume_max", "volume_step", "volume_limit"],
  execution: ["exec_mode", "ie_timeout", "ie_slip_profit", "ie_slip_losing"],
  margin: ["margin_initial", "margin_maintenance", "margin_hedged", "margin_liquidity"],
  margin_rate: [
    "margin_initial_buy", "margin_initial_sell", "margin_maintenance_buy", "margin_maintenance_sell",
  ],
  swaps: ["swap_mode", "swap_long", "swap_short", "swap_year_day"],
};

const TABS = [
  { id: "common", label: "Common" },
  { id: "trade", label: "Trade" },
  { id: "execution", label: "Execution" },
  { id: "margin", label: "Margin" },
  { id: "margin_rate", label: "Margin Rates" },
  { id: "swaps", label: "Swaps" },
];

function Num({ label, value, onChange, disabled, lots }) {
  const shown = value == null ? "" : lots ? toLots(value) : value;
  return (
    <>
      <label>{label}</label>
      <input
        type="text"
        disabled={disabled}
        value={shown}
        placeholder={disabled ? "Default" : ""}
        onChange={(e) => {
          const raw = e.target.value === "" ? 0 : Number(e.target.value) || 0;
          onChange(lots ? fromLots(raw) : raw);
        }}
      />
    </>
  );
}

function Sel({ label, value, names, onChange, disabled }) {
  return (
    <>
      <label>{label}</label>
      <PropSelect fill disabled={disabled} value={value ?? 0} options={enumOptions(names)} onChange={onChange} />
    </>
  );
}

/**
 * One group-symbol override row: every section carries a Use-default checkbox that greys
 * its fields; checked sections write back as NULL (inherit) through use_default_*.
 */
export function GroupSymbolDialog({ groupId, row, onClose, onSaved }) {
  const [activeTab, setActiveTab] = useState("common");
  const [draft, setDraft] = useState({ ...row });
  const [useDefault, setUseDefault] = useState(() =>
    Object.fromEntries(
      Object.entries(SECTIONS).map(([sec, keys]) => [sec, keys.every((k) => row[k] == null)]),
    ),
  );
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(row.symbol_id);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    const patch = {};
    for (const [sec, keys] of Object.entries(SECTIONS)) {
      if (useDefault[sec]) {
        patch[`use_default_${sec}`] = true;
      } else {
        for (const key of keys) if (draft[key] != null) patch[key] = draft[key];
      }
    }
    const res = await updateGroupSymbol(groupId, row.symbol_id, patch);
    if (!res.ok) {
      setError(res.message || "save failed");
      return;
    }
    onSaved();
    onClose();
  }

  function DefaultCheck({ sec, label }) {
    return (
      <label className="sym-check grp-use-default">
        <input
          type="checkbox"
          checked={useDefault[sec]}
          onChange={(e) => setUseDefault({ ...useDefault, [sec]: e.target.checked })}
        />{" "}
        {label}
      </label>
    );
  }

  function panel(tab) {
    const off = useDefault[tab];
    switch (tab) {
      case "common":
        return (
          <>
            <DefaultCheck sec="common" label="Use default common settings" />
            <div className="form-grid">
              <Num label="Spread difference" value={draft.spread_diff} disabled={off} onChange={(v) => set("spread_diff", v)} />
              <Num label="Spread balance" value={draft.spread_diff_balance} disabled={off} onChange={(v) => set("spread_diff_balance", v)} />
              <Num label="Limit & stop level" value={draft.stops_level} disabled={off} onChange={(v) => set("stops_level", v)} />
              <Num label="Freeze level" value={draft.freeze_level} disabled={off} onChange={(v) => set("freeze_level", v)} />
            </div>
          </>
        );
      case "trade":
        return (
          <>
            <DefaultCheck sec="trade" label="Use default trade settings" />
            <div className="form-grid">
              <Sel label="Trade" value={draft.trade_mode} names={TradeMode_name} disabled={off} onChange={(v) => set("trade_mode", v)} />
              <Num label="Minimum volume" value={draft.volume_min} lots disabled={off} onChange={(v) => set("volume_min", v)} />
              <Num label="Maximum volume" value={draft.volume_max} lots disabled={off} onChange={(v) => set("volume_max", v)} />
              <Num label="Volume step" value={draft.volume_step} lots disabled={off} onChange={(v) => set("volume_step", v)} />
              <Num label="Volume limit" value={draft.volume_limit} lots disabled={off} onChange={(v) => set("volume_limit", v)} />
            </div>
          </>
        );
      case "execution":
        return (
          <>
            <DefaultCheck sec="execution" label="Use default execution settings" />
            <div className="form-grid">
              <Sel label="Execution" value={draft.exec_mode} names={ExecMode_name} disabled={off} onChange={(v) => set("exec_mode", v)} />
              <Num label="Timeout" value={draft.ie_timeout} disabled={off} onChange={(v) => set("ie_timeout", v)} />
              <Num label="Profit slippage" value={draft.ie_slip_profit} disabled={off} onChange={(v) => set("ie_slip_profit", v)} />
              <Num label="Losing slippage" value={draft.ie_slip_losing} disabled={off} onChange={(v) => set("ie_slip_losing", v)} />
            </div>
          </>
        );
      case "margin":
        return (
          <>
            <DefaultCheck sec="margin" label="Use default margin settings" />
            <div className="form-grid">
              <Num label="Initial margin" value={draft.margin_initial} disabled={off} onChange={(v) => set("margin_initial", v)} />
              <Num label="Maintenance margin" value={draft.margin_maintenance} disabled={off} onChange={(v) => set("margin_maintenance", v)} />
              <Num label="Hedged margin" value={draft.margin_hedged} disabled={off} onChange={(v) => set("margin_hedged", v)} />
              <Num label="Liquidity margin rate" value={draft.margin_liquidity} disabled={off} onChange={(v) => set("margin_liquidity", v)} />
            </div>
          </>
        );
      case "margin_rate":
        return (
          <>
            <DefaultCheck sec="margin_rate" label="Use default margin rates" />
            <div className="form-grid">
              <Num label="Initial Buy" value={draft.margin_initial_buy} disabled={off} onChange={(v) => set("margin_initial_buy", v)} />
              <Num label="Initial Sell" value={draft.margin_initial_sell} disabled={off} onChange={(v) => set("margin_initial_sell", v)} />
              <Num label="Maintenance Buy" value={draft.margin_maintenance_buy} disabled={off} onChange={(v) => set("margin_maintenance_buy", v)} />
              <Num label="Maintenance Sell" value={draft.margin_maintenance_sell} disabled={off} onChange={(v) => set("margin_maintenance_sell", v)} />
            </div>
          </>
        );
      case "swaps":
        return (
          <>
            <DefaultCheck sec="swaps" label="Use default swaps" />
            <div className="form-grid">
              <Sel label="Type" value={draft.swap_mode} names={SwapMode_name} disabled={off} onChange={(v) => set("swap_mode", v)} />
              <Num label="Long positions" value={draft.swap_long} disabled={off} onChange={(v) => set("swap_long", v)} />
              <Num label="Short positions" value={draft.swap_short} disabled={off} onChange={(v) => set("swap_short", v)} />
              <Num label="Days in year" value={draft.swap_year_day} disabled={off} onChange={(v) => set("swap_year_day", v)} />
            </div>
          </>
        );
    }
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
          title={`Symbols: ${row.path}`}
          tabs={
            <div className="config-tabs">
              {TABS.map((tab) => (
                <button
                  key={tab.id}
                  type="button"
                  className={activeTab === tab.id ? "active" : ""}
                  onClick={() => setActiveTab(tab.id)}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          }
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={onClose}>Cancel</button>
            </div>
          }
        >
          <div className="config-panel active">
            <GroupTabIntro>
              Individual trading settings of the symbols for this group. Sections marked "Use
              default" inherit from the symbol master.
            </GroupTabIntro>
            {panel(activeTab)}
          </div>
        </SettingsDialog>
      </div>
    </div>
  );
}
