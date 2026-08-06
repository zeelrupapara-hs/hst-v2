import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { createGroupSymbol, updateGroupSymbol } from "@/api/endpoints/groups.js";
import { fetchSymbols } from "@/api/endpoints/symbols.js";
import {
  ExecMode_name,
  SwapYearDays_options,
  SwapMode_name,
  TradeMode_name,
  toLots,
  fromLots,
} from "@/constants/symbols.js";

const DEFAULT = "";

// Every override control offers the literal Default entry; it writes NULL (inherit).
const withDefault = (options) => [{ value: DEFAULT, label: "Default" }, ...options];

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

// A bit-mask combo: Default, each single flag, and All when every bit is set.
const flagOptions = (labels) => {
  const all = labels.reduce((acc, f) => acc | f.bit, 0);
  return withDefault([
    { value: 0, label: "None" },
    ...labels.map((f) => ({ value: f.bit, label: f.label })),
    { value: all, label: "All" },
  ]);
};

const FILLING_LABELS = [
  { bit: 1, label: "Fill or Kill" },
  { bit: 2, label: "Immediate or Cancel" },
  { bit: 4, label: "Book or Cancel" },
];

const EXPIRATION_LABELS = [
  { bit: 1, label: "Good till canceled" },
  { bit: 2, label: "Day" },
  { bit: 4, label: "Specified time" },
  { bit: 8, label: "Specified day" },
];

const ORDER_LABELS = [
  { bit: 1, label: "Market orders" },
  { bit: 2, label: "Limit orders" },
  { bit: 4, label: "Stop orders" },
  { bit: 8, label: "Stop limit orders" },
  { bit: 16, label: "Stop loss" },
  { bit: 32, label: "Take profit" },
  { bit: 64, label: "Close by" },
];

// model.SymbolMarginFlags: bits 1|2 are the additional checks, 4|8|16 the switches below.
const MARGIN_CHECK_BITS = 3;
const MARGIN_HEDGE_LARGE_LEG = 4;
const MARGIN_EXCLUDE_PL = 8;
const MARGIN_RECALC_RATES = 16;
const MARGIN_CHECK_OPTIONS = withDefault([
  { value: 0, label: "None" },
  { value: 1, label: "Check on trade process" },
  { value: 2, label: "Check stop loss and take profit" },
  { value: 3, label: "All" },
]);

const IE_FAST_CONFIRMATION = 1;
const RE_CONFIRM_ORDERS = 1;
const SWAP_CONSIDER_HOLIDAYS = 1;

const YES_NO = withDefault([
  { value: 0, label: "No" },
  { value: 1, label: "Yes" },
]);

// permissions_book_depth is a plain count; 0 means no limit.
const BOOK_DEPTH_OPTIONS = withDefault(
  Array.from({ length: 33 }, (_, i) => ({ value: i, label: i === 0 ? "unlimited" : String(i) })),
);

const SWAP_DAYS = [
  ["swap_rate_sunday", "Sunday"],
  ["swap_rate_monday", "Monday"],
  ["swap_rate_tuesday", "Tuesday"],
  ["swap_rate_wednesday", "Wednesday"],
  ["swap_rate_thursday", "Thursday"],
  ["swap_rate_friday", "Friday"],
  ["swap_rate_saturday", "Saturday"],
];

const RATE_COLUMNS = [
  ["", "Market Order"],
  ["_limit", "Limit Order"],
  ["_stop", "Stop Order"],
  ["_stop_limit", "Stop Limit Order"],
];

const rateKeys = (base) =>
  ["buy", "sell"].flatMap((side) => RATE_COLUMNS.map(([sfx]) => `${base}_${side}${sfx}`));

// One Use-default control per reference cluster; the wire clears whole sections only.
const CLUSTERS = {
  spread: { section: "common", keys: ["permissions_flags", "permissions_book_depth", "spread_diff", "spread_diff_balance"] },
  volumes: { section: "common", keys: ["volume_min", "volume_max", "volume_step"] },
  limit: { section: "common", keys: ["volume_limit"] },
  trade: { section: "trade", keys: ["trade_mode", "fill_flags", "expir_flags", "order_flags"] },
  trade_level: { section: "trade", keys: ["stops_level", "freeze_level"] },
  execution: {
    section: "execution",
    keys: ["exec_mode", "ie_timeout", "ie_slip_profit", "ie_slip_losing", "ie_volume_max", "ie_flags", "re_timeout", "re_flags"],
  },
  margin_values: { section: "margin", keys: ["margin_initial", "margin_hedged", "margin_maintenance"] },
  margin_settings: { section: "margin", keys: ["margin_flags"] },
  margin_rate: {
    section: "margin_rate",
    keys: ["margin_liquidity", "margin_currency", ...rateKeys("margin_initial"), ...rateKeys("margin_maintenance")],
  },
  swaps: {
    section: "swaps",
    keys: [
      "swap_mode", "swap_long", "swap_short", "swap_year_day", "swap_flags",
      ...SWAP_DAYS.map(([key]) => key),
    ],
  },
};

const TABS = [
  { id: "common", label: "Common", intro: "Please set up main parameters of symbols for the group." },
  {
    id: "trade",
    label: "Trade",
    intro: "The setting of symbol trading for the group. Please specify trade mode, filling and expiration modes, etc.",
  },
  { id: "execution", label: "Execution", intro: "Please set up parameters of orders execution by symbols for the group." },
  { id: "margin", label: "Margin", intro: "Please specify margin requirements for the symbols group." },
  { id: "margin_rate", label: "Margin Rates", intro: "Please specify margin rates for different types of trade operations." },
  { id: "swaps", label: "Swaps", intro: "Please set up parameters of charging swaps by symbols for the group." },
];

function Num({ label, value, onChange, disabled, lots, suffix, wide }) {
  const shown = value == null ? "" : lots ? toLots(value) : value;
  return (
    <>
      <label>{label}</label>
      <span className={suffix ? "grp-suffixed" : undefined}>
        <input
          type="text"
          className={wide ? "grp-sym-wide" : undefined}
          disabled={disabled}
          value={shown}
          placeholder="Default"
          onChange={(e) => {
            const raw = e.target.value.trim();
            if (raw === "") return onChange(null);
            const n = Number(raw);
            onChange(Number.isNaN(n) ? null : lots ? fromLots(n) : n);
          }}
        />
        {suffix && <span className="grp-suffix">{suffix}</span>}
      </span>
    </>
  );
}

function Sel({ label, value, options, onChange, disabled }) {
  return (
    <>
      <label>{label}</label>
      <PropSelect
        fill
        disabled={disabled}
        value={value == null ? DEFAULT : value}
        options={options}
        onChange={(v) => onChange(v === DEFAULT ? null : v)}
      />
    </>
  );
}

function Check({ label, checked, onChange, disabled }) {
  return (
    <label className="sym-check">
      <input type="checkbox" checked={checked} disabled={disabled} onChange={(e) => onChange(e.target.checked)} />{" "}
      {label}
    </label>
  );
}

/**
 * One group-symbol override row: every cluster carries a Use-default control that greys its
 * fields; a section whose clusters are all defaulted writes back as NULL via use_default_*.
 */
export function GroupSymbolDialog({ groupId, row, onClose, onSaved }) {
  const isNew = !row;
  const [activeTab, setActiveTab] = useState("common");
  const [draft, setDraft] = useState(() => (isNew ? { path: "*" } : { ...row }));
  const [paths, setPaths] = useState([]);
  const [useDefault, setUseDefault] = useState(() =>
    Object.fromEntries(
      Object.entries(CLUSTERS).map(([id, c]) => [id, isNew || c.keys.every((k) => row[k] == null)]),
    ),
  );
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(row?.symbol_id ?? "new");
  const close = useDialogStack(onClose);

  useEffect(() => {
    fetchSymbols().then((res) => {
      if (!res.ok) return;
      const folders = new Set((res.data || []).map((s) => String(s.path || "").split("\\")[0]).filter(Boolean));
      setPaths([...folders].sort().map((f) => `${f}\\*`));
    });
  }, []);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));
  const bit = (key, mask) => (draft[key] == null ? false : (draft[key] & mask) !== 0);
  const setBit = (key, mask, on) => set(key, on ? (draft[key] ?? 0) | mask : (draft[key] ?? 0) & ~mask);

  async function handleOk() {
    const patch = { path: draft.path || "*" };
    for (const section of ["common", "trade", "execution", "margin", "margin_rate", "swaps"]) {
      const ids = Object.keys(CLUSTERS).filter((id) => CLUSTERS[id].section === section);
      if (ids.every((id) => useDefault[id])) {
        patch[`use_default_${section}`] = true;
        continue;
      }
      for (const id of ids) {
        if (useDefault[id]) continue;
        for (const key of CLUSTERS[id].keys) if (draft[key] != null) patch[key] = draft[key];
      }
    }
    const res = isNew
      ? await createGroupSymbol(groupId, patch)
      : await updateGroupSymbol(groupId, row.symbol_id, patch);
    if (!res.ok) {
      setError(res.message || "save failed");
      return;
    }
    onSaved();
    close();
  }

  function Default({ id, label }) {
    return (
      <Check
        label={label}
        checked={useDefault[id]}
        onChange={(on) => setUseDefault({ ...useDefault, [id]: on })}
      />
    );
  }

  function setSwapRates(values) {
    setDraft((prev) => ({ ...prev, ...Object.fromEntries(SWAP_DAYS.map(([key], i) => [key, values[i]])) }));
  }

  function ratesBlock(base, title) {
    return (
      <>
        <tr className="grp-sym-rate-head">
          <td colSpan={5}>{title}</td>
        </tr>
        {["buy", "sell"].map((side) => (
          <tr key={side}>
            <td className="grp-sym-rate-side">{side === "buy" ? "Buy" : "Sell"}</td>
            {RATE_COLUMNS.map(([sfx]) => {
              const key = `${base}_${side}${sfx}`;
              return (
                <td key={key}>
                  <input
                    type="text"
                    className="df-cell-input grp-sym-rate-cell"
                    disabled={useDefault.margin_rate}
                    placeholder="Default"
                    value={draft[key] == null ? "" : draft[key]}
                    onChange={(e) => {
                      const raw = e.target.value.trim();
                      set(key, raw === "" ? null : Number(raw) || 0);
                    }}
                  />
                </td>
              );
            })}
          </tr>
        ))}
      </>
    );
  }

  function panel(tab) {
    switch (tab) {
      case "common":
        return (
          <div className="form-grid sym-form-two-col">
            <Sel
              label="Symbol"
              value={draft.path ?? "*"}
              options={[...new Set(["*", draft.path || "*", ...paths])].map((p) => ({ value: p, label: p }))}
              onChange={(v) => set("path", v || "*")}
            />
            <span />
            <span />
            <Check
              label="Enable market depth"
              disabled={useDefault.spread}
              checked={bit("permissions_flags", 1)}
              onChange={(on) => setBit("permissions_flags", 1, on)}
            />
            <Sel
              label="Market depth limit"
              value={draft.permissions_book_depth}
              options={BOOK_DEPTH_OPTIONS}
              disabled={useDefault.spread}
              onChange={(v) => set("permissions_book_depth", v)}
            />
            <span />
            <span />
            <Default id="spread" label="Use default spread" />
            <Num
              label="Spread difference"
              suffix="pt"
              value={draft.spread_diff}
              disabled={useDefault.spread}
              onChange={(v) => set("spread_diff", v)}
            />
            <Num
              label="Difference balance"
              value={draft.spread_diff_balance}
              disabled={useDefault.spread}
              onChange={(v) => set("spread_diff_balance", v)}
              suffix={`${draft.spread_diff_balance ?? 0} bid / ${(draft.spread_diff ?? 0) - (draft.spread_diff_balance ?? 0)} ask`}
            />
            <Default id="volumes" label="Use default volumes" />
            <Num label="Minimum" lots value={draft.volume_min} disabled={useDefault.volumes} onChange={(v) => set("volume_min", v)} />
            <Num label="Step" lots value={draft.volume_step} disabled={useDefault.volumes} onChange={(v) => set("volume_step", v)} />
            <Num label="Maximum" lots value={draft.volume_max} disabled={useDefault.volumes} onChange={(v) => set("volume_max", v)} />
            <span />
            <span />
            <Default id="limit" label="Use default limit" />
            <Num label="Limit" lots value={draft.volume_limit} disabled={useDefault.limit} onChange={(v) => set("volume_limit", v)} />
          </div>
        );
      case "trade":
        return (
          <div className="form-grid sym-form-two-col">
            <Default id="trade" label="Use default trade settings" />
            <Sel
              label="Trade"
              value={draft.trade_mode}
              options={withDefault(enumOptions(TradeMode_name))}
              disabled={useDefault.trade}
              onChange={(v) => set("trade_mode", v)}
            />
            <span />
            <span />
            <Sel
              label="Filling"
              value={draft.fill_flags}
              options={flagOptions(FILLING_LABELS)}
              disabled={useDefault.trade}
              onChange={(v) => set("fill_flags", v)}
            />
            <span />
            <span />
            <Sel
              label="Expiration"
              value={draft.expir_flags}
              options={flagOptions(EXPIRATION_LABELS)}
              disabled={useDefault.trade}
              onChange={(v) => set("expir_flags", v)}
            />
            <span />
            <span />
            <Sel
              label="Orders"
              value={draft.order_flags}
              options={flagOptions(ORDER_LABELS)}
              disabled={useDefault.trade}
              onChange={(v) => set("order_flags", v)}
            />
            <span />
            <span />
            <Default id="trade_level" label="Use default trade level settings" />
            <Num
              label="Limit & stop level"
              suffix="pt"
              value={draft.stops_level}
              disabled={useDefault.trade_level}
              onChange={(v) => set("stops_level", v)}
            />
            <Num
              label="Freeze level"
              suffix="pt"
              value={draft.freeze_level}
              disabled={useDefault.trade_level}
              onChange={(v) => set("freeze_level", v)}
            />
          </div>
        );
      case "execution": {
        const off = useDefault.execution;
        return (
          <div className="form-grid sym-form-two-col">
            <Default id="execution" label="Use default execution settings" />
            <Sel
              label="Execution"
              value={draft.exec_mode}
              options={withDefault(enumOptions(ExecMode_name))}
              disabled={off}
              onChange={(v) => set("exec_mode", v)}
            />
            <span />
            <span />
            <Num label="Max time deviation" suffix="seconds" value={draft.ie_timeout} disabled={off} onChange={(v) => set("ie_timeout", v)} />
            <span />
            <span />
            <Num label="Max profit deviation" suffix="points" value={draft.ie_slip_profit} disabled={off} onChange={(v) => set("ie_slip_profit", v)} />
            <span />
            <span />
            <Num label="Max losing deviation" suffix="points" value={draft.ie_slip_losing} disabled={off} onChange={(v) => set("ie_slip_losing", v)} />
            <span />
            <span />
            <Check
              label="Fast confirmation of requotes within client deviation"
              disabled={off}
              checked={bit("ie_flags", IE_FAST_CONFIRMATION)}
              onChange={(on) => setBit("ie_flags", IE_FAST_CONFIRMATION, on)}
            />
            <Num
              label="Maximum volume"
              lots
              suffix="before switching to Request execution"
              value={draft.ie_volume_max}
              disabled={off}
              onChange={(v) => set("ie_volume_max", v)}
            />
            <Num label="Timeout" suffix="seconds" value={draft.re_timeout} disabled={off} onChange={(v) => set("re_timeout", v)} />
            <span />
            <span />
            <Check
              label="Confirm orders"
              disabled={off}
              checked={bit("re_flags", RE_CONFIRM_ORDERS)}
              onChange={(on) => setBit("re_flags", RE_CONFIRM_ORDERS, on)}
            />
          </div>
        );
      }
      case "margin":
        return (
          <div className="form-grid sym-form-two-col">
            <Default id="margin_values" label="Use default margin values" />
            <Num label="Initial margin" value={draft.margin_initial} disabled={useDefault.margin_values} onChange={(v) => set("margin_initial", v)} />
            <span />
            <span />
            <Num label="Hedged margin" value={draft.margin_hedged} disabled={useDefault.margin_values} onChange={(v) => set("margin_hedged", v)} />
            <span />
            <span />
            <Num
              label="Maintenance margin"
              value={draft.margin_maintenance}
              disabled={useDefault.margin_values}
              onChange={(v) => set("margin_maintenance", v)}
            />
            <span />
            <span />
            <Default id="margin_settings" label="Use default margin settings" />
            <Check
              label="Calculate hedged margin using larger leg"
              disabled={useDefault.margin_settings}
              checked={bit("margin_flags", MARGIN_HEDGE_LARGE_LEG)}
              onChange={(on) => setBit("margin_flags", MARGIN_HEDGE_LARGE_LEG, on)}
            />
            <Check
              label="Exclude long position PnL from free margin and margin level"
              disabled={useDefault.margin_settings}
              checked={bit("margin_flags", MARGIN_EXCLUDE_PL)}
              onChange={(on) => setBit("margin_flags", MARGIN_EXCLUDE_PL, on)}
            />
            <Check
              label="Recalculate margin exchange rate at the End of Day"
              disabled={useDefault.margin_settings}
              checked={bit("margin_flags", MARGIN_RECALC_RATES)}
              onChange={(on) => setBit("margin_flags", MARGIN_RECALC_RATES, on)}
            />
            <Sel
              label="Additional margin checks"
              value={draft.margin_flags == null ? null : draft.margin_flags & MARGIN_CHECK_BITS}
              options={MARGIN_CHECK_OPTIONS}
              disabled={useDefault.margin_settings}
              onChange={(v) =>
                set("margin_flags", v == null ? null : ((draft.margin_flags ?? 0) & ~MARGIN_CHECK_BITS) | v)
              }
            />
          </div>
        );
      case "margin_rate":
        return (
          <>
            <div className="form-grid sym-form-two-col">
              <Num
                label="Liquidity margin rate"
                value={draft.margin_liquidity}
                disabled={useDefault.margin_rate}
                onChange={(v) => set("margin_liquidity", v)}
              />
              <span />
              <span />
              <label>Currency margin rate</label>
              <input
                type="text"
                disabled={useDefault.margin_rate}
                placeholder="Default"
                value={draft.margin_currency ?? ""}
                onChange={(e) => set("margin_currency", e.target.value.toUpperCase() || null)}
              />
            </div>
            <table className="data-table data-table-grid grp-sym-rates">
              <thead>
                <tr>
                  <th />
                  {RATE_COLUMNS.map(([, label]) => (
                    <th key={label}>{label}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {ratesBlock("margin_initial", "Initial margin")}
                {ratesBlock("margin_maintenance", "Maintenance margin")}
              </tbody>
            </table>
            <div className="grp-sym-default-btn">
              <button type="button" onClick={() => setUseDefault({ ...useDefault, margin_rate: true })}>
                Use default rates settings
              </button>
            </div>
          </>
        );
      case "swaps": {
        const off = useDefault.swaps;
        return (
          <>
            <div className="form-grid sym-form-two-col">
              <Sel
                label="Type"
                value={draft.swap_mode}
                options={withDefault(enumOptions(SwapMode_name))}
                disabled={off}
                onChange={(v) => set("swap_mode", v)}
              />
              <span />
              <span />
              <Num label="Long positions" value={draft.swap_long} disabled={off} onChange={(v) => set("swap_long", v)} />
              <Num label="Short positions" value={draft.swap_short} disabled={off} onChange={(v) => set("swap_short", v)} />
              <Sel
                label="Days in year"
                value={draft.swap_year_day}
                options={withDefault(SwapYearDays_options.map((d) => ({ value: d, label: String(d) })))}
                disabled={off}
                onChange={(v) => set("swap_year_day", v)}
              />
              <Sel
                label="Consider holidays"
                value={draft.swap_flags == null ? null : draft.swap_flags & SWAP_CONSIDER_HOLIDAYS}
                options={YES_NO}
                disabled={off}
                onChange={(v) => set("swap_flags", v)}
              />
            </div>
            <div className="grp-sym-swap-layout">
              <span className="grp-sym-swap-caption">Swap multipliers</span>
              <table className="data-table data-table-grid grp-sym-swap-table">
                <thead>
                  <tr>
                    <th>Day of week</th>
                    <th>Multiplier</th>
                  </tr>
                </thead>
                <tbody>
                  {SWAP_DAYS.map(([key, day]) => (
                    <tr key={key}>
                      <td>{day}</td>
                      <td>
                        <input
                          type="text"
                          className="df-cell-input grp-sym-rate-cell"
                          disabled={off}
                          placeholder="Default"
                          value={draft[key] == null ? "" : draft[key]}
                          onChange={(e) => {
                            const raw = e.target.value.trim();
                            set(key, raw === "" ? null : Number(raw) || 0);
                          }}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div className="grp-sym-swap-btns">
                <button type="button" disabled={off} onClick={() => setSwapRates([null, null, null, null, null, null, null])}>
                  Default
                </button>
                <button type="button" disabled={off} onClick={() => setSwapRates([0, 1, 1, 3, 1, 1, 0])}>
                  Forex
                </button>
                <button type="button" disabled={off} onClick={() => setSwapRates([1, 1, 1, 1, 1, 1, 1])}>
                  All week
                </button>
                <button type="button" disabled title="No source symbol route">
                  From symbol
                </button>
                <PropSelect fill disabled value="" options={[{ value: "", label: "" }]} />
              </div>
            </div>
            <div className="grp-sym-default-btn">
              <button type="button" onClick={() => setUseDefault({ ...useDefault, swaps: true })}>
                Use default swap settings
              </button>
            </div>
          </>
        );
      }
    }
  }

  const intro = TABS.find((t) => t.id === activeTab)?.intro;

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          width={700}
          height={480}
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          title={`Symbol: ${draft.path || "*"}`}
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
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="symbols-tree" size={48} />
              </span>
              <p>{intro}</p>
            </div>
            {panel(activeTab)}
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}
