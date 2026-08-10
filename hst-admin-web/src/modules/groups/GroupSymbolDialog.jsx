import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack, prevTabEscape } from "@/hooks/useDialogStack.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { SymbolScopeSelectField } from "@/components/ui/SymbolScopeSelectField.jsx";
import { SymbolTreeSelectField } from "@/components/ui/SymbolTreeSelectField.jsx";
import { useSymbols } from "@/hooks/useSymbols.js";
import { fetchSymbolSwaps } from "@/api/endpoints/symbols.js";
import { createGroupSymbol, updateGroupSymbol } from "@/api/endpoints/groups.js";
import {
  Deviation_options,
  ExecMode_name,
  ExpirFlag_labels,
  FillFlag_labels,
  INSTANT_FAST_CONFIRMATION,
  OrderFlag_labels,
  REQUEST_ORDER,
  SwapMode_name,
  TradeMode_name,
  toLots,
  fromLots,
  SWAP_CONSIDER_HOLIDAYS,
} from "@/constants/symbols.js";
import {
  FOREX_SWAP_MULTIPLIERS,
  SWAP_DAY_KEYS,
  SWAP_HOLIDAYS_OPTIONS,
  formatSwapMultiplier,
  parseSwapAmount,
  parseSwapMultiplier,
  parseSwapYearDay,
  swapAmountOptions,
  swapFieldsFromRow,
  swapMultiplierOptions,
  swapYearDayOptions,
} from "@/lib/swapConfig.js";
import { FlagCombo, TabIntro, EditableSelectField } from "@/modules/symbols/SymbolTabs.jsx";
import {
  defaultGroupDiffBalance,
  groupDiffBalanceCaption,
  groupDiffBalanceFromSlider,
  groupDiffBalanceSliderValue,
  groupSpreadMagnitude,
} from "@/lib/groupSpreadBalance.js";

const DEFAULT = "";

const SPREAD_DIFF_OPTIONS = [-3, -2, -1, 0, 1, 2, 3].map((v) => ({ value: v, label: String(v) }));

function parseSpreadDiff(text) {
  const raw = String(text ?? "").trim();
  if (raw === "") return null;
  const n = parseInt(raw, 10);
  return Number.isNaN(n) ? null : n;
}

function parseDeviation(text) {
  const raw = String(text ?? "").trim();
  if (raw === "") return null;
  const n = Number.parseInt(raw, 10);
  return Number.isFinite(n) && n >= 0 ? n : null;
}

const DEVIATION_PRESET_OPTIONS = Deviation_options.map((v) => ({ value: v, label: String(v) }));

const CATALOG_DEFAULTS = {
  spread_diff: 0,
  spread_diff_balance: 0,
  volume_min: 10000,
  volume_max: 10000000000,
  volume_step: 10000,
  volume_limit: 0,
};

// MT5 greys these strings on the `*` row when Use default is checked.
const WILDCARD_DISPLAY = {
  spread_diff: 0,
  spread_diff_balance: 0,
  volume_min: "0.00000001",
  volume_max: "100000000",
  volume_step: "0.00000001",
  volume_limit: "0",
};

function isRootPath(path) {
  return (path ?? "*") === "*";
}

/** The "*" row defaults market depth on when permissions_flags is unset. */
function marketDepthOn(flags, path) {
  if (flags != null) return (flags & 1) !== 0;
  return isRootPath(path);
}

// Every override control offers the literal Default entry; it writes NULL (inherit).
const withDefault = (options) => [{ value: DEFAULT, label: "Default" }, ...options];

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

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

const EXEC_REQUEST = 0;
const EXEC_INSTANT = 1;

// permissions_book_depth is a plain count; 0 means no limit.
const BOOK_DEPTH_OPTIONS = Array.from({ length: 33 }, (_, i) => ({
  value: i,
  label: i === 0 ? "unlimited" : String(i),
}));

const RATE_COLUMNS = [
  ["", "Market Order"],
  ["_limit", "Limit Order"],
  ["_stop", "Stop Order"],
  ["_stop_limit", "Stop Limit Order"],
];

const rateKeys = (base) =>
  ["buy", "sell"].flatMap((side) => RATE_COLUMNS.map(([sfx]) => `${base}_${side}${sfx}`));

const fmtRate = (v, digits) => (v == null || v === "" ? "" : Number(v).toFixed(digits));

function parseMarginRate(text) {
  const raw = String(text ?? "").trim();
  if (raw === "" || raw.toLowerCase() === "default") return null;
  const n = Number(raw);
  return Number.isNaN(n) ? null : n;
}

function parseMarginCurrency(text) {
  const raw = String(text ?? "").trim();
  if (raw === "" || raw.toLowerCase() === "default") return null;
  const n = Number(raw);
  return Number.isNaN(n) ? null : fmtRate(n, 8);
}

function marginRateOptions(value, digits, presets = [0, 1]) {
  const items = [{ value: null, label: "Default" }];
  for (const n of presets) items.push({ value: n, label: fmtRate(n, digits) });
  const num = value == null || value === "" ? null : Number(value);
  if (num != null && !Number.isNaN(num) && !presets.includes(num)) {
    items.push({ value: num, label: fmtRate(num, digits) });
  }
  return items;
}

// One Use-default control per reference cluster; the wire clears whole sections only.
const CLUSTERS = {
  spread: { section: "common", keys: ["spread_diff", "spread_diff_balance"] },
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
      ...SWAP_DAY_KEYS.map(([, key]) => key),
    ],
  },
};

function buildResolved(row) {
  const resolved = { ...CATALOG_DEFAULTS };
  if (!row) return resolved;
  for (const c of Object.values(CLUSTERS)) {
    for (const key of c.keys) {
      if (row[key] != null) resolved[key] = row[key];
    }
  }
  return resolved;
}

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

function Num({ label, value, inherit, onChange, disabled, lots, suffix, wide, span, execGrid }) {
  const showPlaceholder = disabled && inherit == null;
  let shown = "";
  if (!showPlaceholder) {
    if (inherit != null) {
      shown = typeof inherit === "string" ? inherit : lots ? toLots(inherit) : inherit;
    } else if (lots) {
      shown = toLots(value ?? 0);
    } else if (value != null) {
      shown = value;
    }
  }
  const suffixCls = suffix ? (execGrid ? "sym-with-suffix" : "grp-suffixed") : null;
  return (
    <>
      <label>{label}</label>
      <span className={[suffixCls, span && !execGrid && "grp-sym-span"].filter(Boolean).join(" ") || undefined}>
        <span className="grp-sym-spin">
          <input
            type="text"
            className={wide ? "grp-sym-wide" : undefined}
            disabled={disabled}
            value={shown}
            placeholder={showPlaceholder ? "Default" : undefined}
            onChange={(e) => {
              const raw = e.target.value.trim();
              if (raw === "") return onChange(null);
              const n = Number(raw);
              onChange(Number.isNaN(n) ? null : lots ? fromLots(n) : n);
            }}
          />
        </span>
        {suffix && <span className="grp-suffix">{suffix}</span>}
      </span>
    </>
  );
}

function Sel({ label, value, options, onChange, disabled, span, depthLimit, suffix }) {
  const select = (
    <PropSelect
      fill
      disabled={disabled}
      value={value == null ? DEFAULT : value}
      options={options}
      onChange={(v) => onChange(v === DEFAULT ? null : v)}
    />
  );
  const control = suffix ? (
    <span className="grp-suffixed">
      {select}
      <span className="grp-suffix">{suffix}</span>
    </span>
  ) : (
    select
  );
  if (depthLimit) {
    return (
      <>
        <label className="grp-sym-depth-limit-label">{label}</label>
        <span className="grp-sym-depth-limit-field">{control}</span>
      </>
    );
  }
  return (
    <>
      <label>{label}</label>
      {span ? <span className="grp-sym-span">{control}</span> : control}
    </>
  );
}

function Check({ label, checked, onChange, disabled, half, inline, depth, execGrid }) {
  const cls = execGrid
    ? "sym-check"
    : depth
      ? "sym-check grp-sym-depth-check"
      : inline
        ? "sym-check grp-sym-inline-check"
        : half
          ? "sym-check grp-sym-half"
          : "sym-check grp-sym-full";
  return (
    <label className={cls}>
      <input type="checkbox" checked={checked} disabled={disabled} onChange={(e) => onChange(e.target.checked)} />{" "}
      {label}
    </label>
  );
}

/** Difference balance — distributes spread_diff points between bid and ask (group symbols Common tab). */
function DiffBalanceField({ spread, balance, disabled, onChange }) {
  const s = groupSpreadMagnitude(spread);
  const sliderVal = groupDiffBalanceSliderValue(spread, balance);
  const locked = disabled || s === 0;

  function commit(nextBid) {
    onChange(groupDiffBalanceFromSlider(spread, nextBid));
  }

  return (
    <>
      <label className="sym-spread-balance-label grp-sym-diff-balance-label">Difference balance</label>
      <span className={`sym-spread-balance grp-sym-diff-balance${locked ? " sym-spread-balance-disabled" : ""}`}>
        <span className="sym-spread-balance-track grp-sym-diff-balance-track">
          <input
            type="range"
            className="grp-sym-diff-balance-range"
            min={0}
            max={s}
            step={1}
            disabled={locked}
            value={sliderVal}
            onInput={(e) => commit(Number(e.target.value))}
            onChange={(e) => commit(Number(e.target.value))}
          />
        </span>
        <span className="sym-spread-balance-caption grp-sym-diff-balance-caption">
          {groupDiffBalanceCaption(spread, balance)}
        </span>
      </span>
    </>
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
  const [useDefault, setUseDefault] = useState(() =>
    Object.fromEntries(
      Object.entries(CLUSTERS).map(([id, c]) => [id, isNew || c.keys.every((k) => row[k] == null)]),
    ),
  );
  const [resolved] = useState(() => buildResolved(row));
  const [error, setError] = useState("");
  const [fromSymbol, setFromSymbol] = useState("");
  const [selectedSwapDay, setSelectedSwapDay] = useState(0);
  const { symbols, reload } = useSymbols();
  const { offset, onTitlePointerDown } = useDialogDrag(row?.symbol_id ?? "new");
  const close = useDialogStack(onClose, {
    onEscape: () => prevTabEscape(TABS, activeTab, setActiveTab),
  });

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));
  const bit = (key, mask) => (draft[key] == null ? false : (draft[key] & mask) !== 0);
  const setBit = (key, mask, on) => set(key, on ? (draft[key] ?? 0) | mask : (draft[key] ?? 0) & ~mask);

  async function handleOk() {
    const patch = { path: draft.path || "*" };
    // market depth has no Use-default control of its own, so it is written whenever it is set
    const DEPTH = ["permissions_flags", "permissions_book_depth"];
    for (const key of DEPTH) if (draft[key] != null) patch[key] = draft[key];

    for (const section of ["common", "trade", "execution", "margin", "margin_rate", "swaps"]) {
      const ids = Object.keys(CLUSTERS).filter((id) => CLUSTERS[id].section === section);
      if (ids.every((id) => useDefault[id]) && !(section === "common" && DEPTH.some((k) => draft[k] != null))) {
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
        onChange={(on) => {
          setUseDefault({ ...useDefault, [id]: on });
          setDraft((prev) => {
            const next = { ...prev };
            for (const key of CLUSTERS[id].keys) {
              if (on) next[key] = null;
              else if (next[key] == null) next[key] = resolved[key] ?? CATALOG_DEFAULTS[key] ?? 0;
            }
            return next;
          });
        }}
      />
    );
  }

  function setSwapRates(values) {
    setDraft((prev) => ({ ...prev, ...Object.fromEntries(SWAP_DAY_KEYS.map(([, key], i) => [key, values[i]])) }));
  }

  function resetSwapDefaults() {
    setUseDefault({ ...useDefault, swaps: true });
    setDraft((prev) => {
      const next = { ...prev };
      for (const key of CLUSTERS.swaps.keys) next[key] = null;
      return next;
    });
  }

  function swapHolidaysValue() {
    if (draft.swap_flags == null) return null;
    return draft.swap_flags & SWAP_CONSIDER_HOLIDAYS ? 1 : 0;
  }

  function swapDayChecked(key) {
    const v = useDefault.swaps ? resolved[key] : draft[key];
    return v != null && Number(v) !== 0;
  }

  function swapDayMultiplier(key) {
    return useDefault.swaps ? null : draft[key];
  }

  async function copySwapFromSymbol() {
    const name = fromSymbol.trim();
    if (!name) return;

    let list = symbols;
    if (list.length && !("swap_mode" in list[0])) {
      list = await reload();
    }

    const sym = list.find((item) => item.symbol === name);
    if (!sym?.symbol_id) {
      window.alert(`Symbol '${name}' not found`);
      return;
    }

    let fields = swapFieldsFromRow(sym);
    if (!fields) {
      const res = await fetchSymbolSwaps(sym.symbol_id);
      if (!res.ok) {
        window.alert(res.message || "failed to load swap settings");
        return;
      }
      fields = swapFieldsFromRow(res.data);
    }
    if (!fields) {
      window.alert("No swap settings found for that symbol");
      return;
    }

    setUseDefault({ ...useDefault, swaps: false });
    setDraft((prev) => ({ ...prev, ...fields }));
  }

  function resetMarginRateDefaults() {
    setUseDefault({ ...useDefault, margin_rate: true });
    setDraft((prev) => {
      const next = { ...prev };
      for (const key of CLUSTERS.margin_rate.keys) next[key] = null;
      return next;
    });
  }

  function ratesBlock(base, title) {
    const off = useDefault.margin_rate;
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
              const value = off ? null : draft[key];
              return (
                <td key={key}>
                  <EditableSelectField
                    hideLabel
                    className="grp-sym-rate-select"
                    placeholder="Default"
                    disabled={off}
                    value={value}
                    options={marginRateOptions(value, 7)}
                    format={(v) => fmtRate(v, 7)}
                    parse={parseMarginRate}
                    onChange={(v) => set(key, v)}
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
      case "common": {
        const root = isRootPath(draft.path);
        const depthOn = marketDepthOn(draft.permissions_flags, draft.path);
        const inherit = (clusterId, key) =>
          useDefault[clusterId] && root ? (WILDCARD_DISPLAY[key] ?? CATALOG_DEFAULTS[key] ?? 0) : null;
        return (
          <div className="form-grid sym-form-two-col grp-sym-form grp-sym-common">
            <SymbolScopeSelectField
              label="Symbol"
              className="grp-sym-symbol-wide"
              value={draft.path ?? "*"}
              onChange={(v) => set("path", v || "*")}
              disabled={!isNew}
            />
            <Check
              depth
              label="Enable market depth"
              checked={depthOn}
              onChange={(on) => {
                if (draft.permissions_flags == null) set("permissions_flags", on ? 1 : 0);
                else setBit("permissions_flags", 1, on);
              }}
            />
            <Sel
              depthLimit
              label="Market depth limit"
              value={draft.permissions_book_depth ?? 0}
              options={BOOK_DEPTH_OPTIONS}
              disabled={!depthOn}
              onChange={(v) => set("permissions_book_depth", v)}
            />
            <Default id="spread" label="Use default spread" />
            <EditableSelectField
              label="Spread difference"
              suffix="pt"
              placeholder="Default"
              liveCommit
              disabled={useDefault.spread}
              value={useDefault.spread ? inherit("spread", "spread_diff") : draft.spread_diff}
              options={SPREAD_DIFF_OPTIONS}
              format={(v) => (v == null ? "" : String(v))}
              parse={parseSpreadDiff}
              onChange={(v) =>
                setDraft((prev) => ({
                  ...prev,
                  spread_diff: v,
                  // MT5 resets difference balance when spread difference changes.
                  spread_diff_balance: defaultGroupDiffBalance(v),
                }))
              }
            />
            <DiffBalanceField
              spread={
                useDefault.spread
                  ? (inherit("spread", "spread_diff") ?? 0)
                  : (draft.spread_diff ?? 0)
              }
              balance={
                useDefault.spread
                  ? (inherit("spread", "spread_diff_balance") ?? 0)
                  : (draft.spread_diff_balance ?? 0)
              }
              disabled={useDefault.spread}
              onChange={(v) => set("spread_diff_balance", v)}
            />
            <Default id="volumes" label="Use default volumes" />
            <div className="grp-sym-triple">
              <Num label="Minimum" lots value={draft.volume_min} inherit={inherit("volumes", "volume_min")} disabled={useDefault.volumes} onChange={(v) => set("volume_min", v)} />
              <Num label="Step" lots value={draft.volume_step} inherit={inherit("volumes", "volume_step")} disabled={useDefault.volumes} onChange={(v) => set("volume_step", v)} />
              <Num label="Maximum" lots value={draft.volume_max} inherit={inherit("volumes", "volume_max")} disabled={useDefault.volumes} onChange={(v) => set("volume_max", v)} />
            </div>
            <Default id="limit" label="Use default limit" />
            <Num label="Limit" lots value={draft.volume_limit} inherit={inherit("limit", "volume_limit")} disabled={useDefault.limit} onChange={(v) => set("volume_limit", v)} />
          </div>
        );
      }
      case "trade":
        return (
          <div className="form-grid sym-form-two-col grp-sym-form">
            <Default id="trade" label="Use default trade settings" />
            <Sel
              span
              label="Trade"
              value={draft.trade_mode}
              options={withDefault(enumOptions(TradeMode_name))}
              disabled={useDefault.trade}
              onChange={(v) => set("trade_mode", v)}
            />
            <FlagCombo
              label="Filling"
              labels={FillFlag_labels}
              value={draft.fill_flags ?? 0}
              disabled={useDefault.trade}
              onChange={(v) => set("fill_flags", v)}
            />
            <FlagCombo
              label="Expiration"
              labels={ExpirFlag_labels}
              value={draft.expir_flags ?? 0}
              disabled={useDefault.trade}
              onChange={(v) => set("expir_flags", v)}
            />
            <FlagCombo
              label="Orders"
              labels={OrderFlag_labels}
              value={draft.order_flags ?? 0}
              disabled={useDefault.trade}
              onChange={(v) => set("order_flags", v)}
            />
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
        const mode = draft.exec_mode != null ? Number(draft.exec_mode) : null;
        return (
          <div className="form-grid sym-form-two-col grp-sym-form">
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
            {mode === EXEC_INSTANT && (
              <>
                <EditableSelectField
                  label="Max time deviation"
                  suffix="seconds"
                  placeholder="Default"
                  disabled={off}
                  value={off ? null : draft.ie_timeout}
                  options={DEVIATION_PRESET_OPTIONS}
                  format={(v) => (v == null ? "" : String(v))}
                  parse={parseDeviation}
                  onChange={(v) => set("ie_timeout", v)}
                />
                <span />
                <span />
                <EditableSelectField
                  label="Max profit deviation"
                  suffix="points"
                  placeholder="Default"
                  disabled={off}
                  value={off ? null : draft.ie_slip_profit}
                  options={DEVIATION_PRESET_OPTIONS}
                  format={(v) => (v == null ? "" : String(v))}
                  parse={parseDeviation}
                  onChange={(v) => set("ie_slip_profit", v)}
                />
                <span />
                <span />
                <EditableSelectField
                  label="Max losing deviation"
                  suffix="points"
                  placeholder="Default"
                  disabled={off}
                  value={off ? null : draft.ie_slip_losing}
                  options={DEVIATION_PRESET_OPTIONS}
                  format={(v) => (v == null ? "" : String(v))}
                  parse={parseDeviation}
                  onChange={(v) => set("ie_slip_losing", v)}
                />
                <span />
                <span />
                <div className="form-grid sym-exec-grid grp-sym-exec-instant">
                  <label />
                  <Check
                    execGrid
                    label="Fast confirmation of requotes within client deviation"
                    disabled={off}
                    checked={bit("ie_flags", INSTANT_FAST_CONFIRMATION)}
                    onChange={(on) => setBit("ie_flags", INSTANT_FAST_CONFIRMATION, on)}
                  />
                  <Num
                    execGrid
                    label="Maximum volume"
                    lots
                    suffix="before switching to Request execution"
                    value={draft.ie_volume_max}
                    disabled={off}
                    onChange={(v) => set("ie_volume_max", v)}
                  />
                </div>
              </>
            )}
            {mode === EXEC_REQUEST && (
              <>
                <Num label="Timeout" suffix="seconds" value={draft.re_timeout} disabled={off} onChange={(v) => set("re_timeout", v)} />
                <span />
                <span />
                <Check
                  label="Confirm orders"
                  disabled={off}
                  checked={bit("re_flags", REQUEST_ORDER)}
                  onChange={(on) => setBit("re_flags", REQUEST_ORDER, on)}
                />
              </>
            )}
          </div>
        );
      }
      case "margin":
        return (
          <div className="form-grid sym-form-two-col grp-sym-form">
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
              span
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
      case "margin_rate": {
        const off = useDefault.margin_rate;
        return (
          <>
            <div className="form-grid sym-form-two-col grp-sym-form">
              <Default id="margin_rate" label="Use default margin rate settings" />
              <EditableSelectField
                label="Liquidity margin rate"
                placeholder="Default"
                disabled={off}
                value={off ? null : draft.margin_liquidity}
                options={marginRateOptions(draft.margin_liquidity, 3)}
                format={(v) => fmtRate(v, 3)}
                parse={parseMarginRate}
                onChange={(v) => set("margin_liquidity", v)}
              />
              <span />
              <span />
              <EditableSelectField
                label="Currency margin rate"
                placeholder="Default"
                disabled={off}
                value={off ? null : draft.margin_currency}
                options={marginRateOptions(draft.margin_currency, 8)}
                format={(v) => fmtRate(v, 8)}
                parse={parseMarginCurrency}
                onChange={(v) => set("margin_currency", v)}
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
              <button type="button" onClick={resetMarginRateDefaults}>
                Use default rates settings
              </button>
            </div>
          </>
        );
      }
      case "swaps": {
        const off = useDefault.swaps;
        return (
          <>
            <div className="form-grid sym-form-two-col grp-sym-form">
              <Default id="swaps" label="Use default swap settings" />
              <Sel
                span
                label="Type"
                value={draft.swap_mode}
                options={withDefault(enumOptions(SwapMode_name))}
                disabled={off}
                onChange={(v) => set("swap_mode", v)}
              />
              <EditableSelectField
                label="Long positions"
                placeholder="Default"
                disabled={off}
                value={off ? null : draft.swap_long}
                options={swapAmountOptions(draft.swap_long)}
                format={(v) => (v == null ? "" : String(v))}
                parse={parseSwapAmount}
                onChange={(v) => set("swap_long", v)}
              />
              <EditableSelectField
                label="Short positions"
                placeholder="Default"
                disabled={off}
                value={off ? null : draft.swap_short}
                options={swapAmountOptions(draft.swap_short)}
                format={(v) => (v == null ? "" : String(v))}
                parse={parseSwapAmount}
                onChange={(v) => set("swap_short", v)}
              />
              <EditableSelectField
                label="Days in year"
                placeholder="Default"
                disabled={off}
                value={off ? null : draft.swap_year_day}
                options={swapYearDayOptions(draft.swap_year_day)}
                format={(v) => (v == null ? "" : String(v))}
                parse={parseSwapYearDay}
                onChange={(v) => set("swap_year_day", v)}
              />
              <Sel
                label="Consider holidays"
                value={off ? null : swapHolidaysValue()}
                options={SWAP_HOLIDAYS_OPTIONS}
                disabled={off}
                onChange={(v) => set("swap_flags", v === DEFAULT ? null : v === 1 ? SWAP_CONSIDER_HOLIDAYS : 0)}
              />
            </div>
            <div className="grp-sym-swap-layout">
              <span className="grp-sym-swap-caption">Swap multipliers</span>
              <table className="data-table data-table-grid grp-sym-swap-table">
                <thead>
                  <tr>
                    <th className="grp-sym-swap-check-col" />
                    <th>Day of week</th>
                    <th>Multiplier</th>
                  </tr>
                </thead>
                <tbody>
                  {SWAP_DAY_KEYS.map(([day, key], idx) => (
                    <tr
                      key={key}
                      className={selectedSwapDay === idx ? "selected" : ""}
                      onClick={() => setSelectedSwapDay(idx)}
                    >
                      <td className="grp-sym-swap-check-col">
                        <input
                          type="checkbox"
                          disabled={off}
                          checked={swapDayChecked(key)}
                          onClick={(e) => e.stopPropagation()}
                          onChange={(e) => set(key, e.target.checked ? (draft[key] > 0 ? draft[key] : 1) : 0)}
                        />
                      </td>
                      <td>{day}</td>
                      <td>
                        <EditableSelectField
                          hideLabel
                          className="grp-sym-swap-mult-select"
                          placeholder="Default"
                          disabled={off}
                          value={swapDayMultiplier(key)}
                          options={swapMultiplierOptions(draft[key])}
                          format={formatSwapMultiplier}
                          parse={parseSwapMultiplier}
                          onChange={(v) => set(key, v)}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div className="grp-sym-swap-btns">
                <button type="button" disabled={off} onClick={() => setSwapRates(SWAP_DAY_KEYS.map(() => null))}>
                  Default
                </button>
                <button type="button" disabled={off} onClick={() => setSwapRates(FOREX_SWAP_MULTIPLIERS)}>
                  Forex
                </button>
                <button type="button" disabled={off} onClick={() => setSwapRates([1, 1, 1, 1, 1, 1, 1])}>
                  All week
                </button>
                <button type="button" disabled={off || !fromSymbol.trim()} onClick={copySwapFromSymbol}>
                  From symbol
                </button>
                <SymbolTreeSelectField
                  label=""
                  value={fromSymbol}
                  headerItems={[{ value: "", label: "" }]}
                  lockField={() => off}
                  onChange={setFromSymbol}
                />
              </div>
            </div>
            <div className="grp-sym-default-btn">
              <button type="button" onClick={resetSwapDefaults}>
                Use default swap settings
              </button>
            </div>
          </>
        );
      }
    }
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          className="sym-config-window"
          width={787}
          height={620}
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          title={`Symbol: ${draft.path || "*"}`}
          tabs={
            <div className="config-tabs sym-config-tabs">
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
          {TABS.map((tab) => (
            <div key={tab.id} className={`config-panel${activeTab === tab.id ? " active" : ""}`}>
              <TabIntro>{tab.intro}</TabIntro>
              {panel(tab.id)}
            </div>
          ))}
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}
