import { Fragment, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Icon } from "@/components/ui/Icon.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { useExclusiveDropdown } from "@/hooks/useExclusiveDropdown.js";
import { useSymbols } from "@/hooks/useSymbols.js";
import { fetchSymbolSwaps } from "@/api/endpoints/symbols.js";
import { SymbolTreeSelectField } from "@/components/ui/SymbolTreeSelectField.jsx";
import {
  digitsForCurrencyField,
  isForexDerived,
  isMarginDerived,
} from "@/lib/symbolCurrency.js";
import { COUNTRY_options } from "@/constants/countries.js";
import {
  BackgroundColor_options,
  BookVolume_options,
  CalcMode_name,
  CalcMode_order,
  ChartMode_name,
  CURRENCY_options,
  Deviation_options,
  DIGITS_options,
  ExecMode_name,
  ExpirFlag_labels,
  FillFlag_labels,
  FilterTicks_options,
  GTCMode_name,
  INSTANT_FAST_CONFIRMATION,
  MARGIN_CHECK_PROCESS,
  MARGIN_CHECK_SLTP,
  MARGIN_EXCLUDE_PL,
  MARGIN_HEDGE_LARGE_LEG,
  MARGIN_RECALC_RATES,
  MarginCheck_name,
  MarketDepth_options,
  OrderFlag_labels,
  REQUEST_ORDER,
  SWAP_CONSIDER_HOLIDAYS,
  SubscriptionDelay_options,
  SwapYearDays_options,
  SwapMode_name,
  SymbolIndustry_name,
  SymbolIndustries_bySector,
  SymbolSector_name,
  TICK_COLLECT_RAW,
  TICK_FEED_STATS,
  TICK_NEGATIVE_PRICES,
  TICK_REALTIME,
  TRADE_ALLOW_SIGNALS,
  TRADE_PROFIT_BY_MARKET,
  TradeMode_name,
  colorToCss,
  COLOR_NONE,
  toLots,
  fromLots,
} from "@/constants/symbols.js";

export function TabIntro({ icon = "symbols-tree", children }) {
  return (
    <div className="sym-sessions-intro">
      <span className="sym-tab-intro-icon" aria-hidden="true">
        <Icon id={icon} size={48} />
      </span>
      <p>{children}</p>
    </div>
  );
}

const enumOptions = (names, order) =>
  (order ?? Object.keys(names).map(Number)).map((value) => ({ value, label: names[value] }));

const plain = (values) => values.map((v) => ({ value: v, label: String(v) }));

/** MT5 SpreadBalance → bid/ask point shifts shown under the slider. */
function spreadBalanceShifts(spread, balance) {
  const s = Number(spread) || 0;
  const b = Number(balance) || 0;
  if (s === 0) return { bid: b, ask: b };
  const lo = Math.floor(s / 2);
  const hi = s - lo;
  return { bid: -lo + b, ask: hi + b };
}

function spreadBalanceCaption(spread, balance) {
  const { bid, ask } = spreadBalanceShifts(spread, balance);
  return `${bid} bid / ${ask} ask`;
}

const fmt = (v, digits) => (v == null ? "" : Number(v).toFixed(digits));

function Field({ label, value, onChange, wide, readOnly, combo, suffix, disabled, lockField, fieldKey }) {
  const locked = lockField?.(fieldKey);
  const input = (
    <input
      type="text"
      readOnly={readOnly || locked || !onChange}
      disabled={disabled || locked}
      value={value ?? ""}
      className={wide ? "wide" : ""}
      onChange={onChange ? (e) => onChange(e.target.value) : undefined}
    />
  );
  return (
    <>
      <label>{label}</label>
      {combo || suffix ? (
        <span className={combo ? "sym-combo" : "sym-with-suffix"}>
          {input}
          {suffix && <span className="sym-suffix">{suffix}</span>}
        </span>
      ) : (
        input
      )}
    </>
  );
}

/** Keeps the reference's fixed decimals without fighting the caret while typing. */
function FmtInput({ value, digits, onChange, className = "" }) {
  const [raw, setRaw] = useState(null);
  return (
    <input
      type="text"
      className={className}
      value={raw ?? fmt(value, digits)}
      onFocus={() => setRaw(String(value ?? ""))}
      onBlur={() => setRaw(null)}
      onChange={(e) => {
        setRaw(e.target.value);
        onChange(e.target.value === "" ? 0 : Number(e.target.value) || 0);
      }}
    />
  );
}

function NumField({ label, value, onChange, readOnly, digits, suffix, offWhenZero, lockField, fieldKey }) {
  const locked = lockField?.(fieldKey);
  if (locked) {
    return (
      <Field label={label} suffix={suffix} value={offWhenZero && !value ? "off" : value} readOnly />
    );
  }
  if (digits != null && onChange) {
    return (
      <>
        <label>{label}</label>
        <span className="sym-with-suffix">
          <FmtInput value={value} digits={digits} onChange={onChange} />
          {suffix && <span className="sym-suffix">{suffix}</span>}
        </span>
      </>
    );
  }
  return (
    <Field
      label={label}
      suffix={suffix}
      value={offWhenZero && !value ? "off" : value}
      readOnly={readOnly}
      onChange={
        onChange ? (v) => onChange(v === "" || v === "off" ? 0 : Number(v) || 0) : undefined
      }
    />
  );
}

function SelectField({ label, value, names, order, options, onChange, disabled, suffix, fallback, lockField, fieldKey, emptyValue }) {
  const locked = lockField?.(fieldKey);
  const opts = options ?? enumOptions(names, order);
  const list =
    fallback != null && !opts.some((o) => String(o.value) === String(value))
      ? [{ value, label: String(value ?? "") }, ...opts]
      : opts;
  const defaultValue =
    emptyValue !== undefined
      ? emptyValue
      : opts.length && typeof opts[0]?.value === "string"
        ? ""
        : 0;
  const select = (
    <PropSelect fill value={value ?? defaultValue} options={list} onChange={onChange} disabled={disabled || locked} />
  );
  return (
    <>
      <label>{label}</label>
      {suffix ? (
        <span className="sym-with-suffix">
          {select}
          <span className="sym-suffix">{suffix}</span>
        </span>
      ) : (
        select
      )}
    </>
  );
}

function formatMarketDepth(value) {
  const n = Number(value);
  return !n ? "off" : String(n);
}

function parseMarketDepth(text) {
  const t = String(text ?? "").trim().toLowerCase();
  if (t === "" || t === "off") return 0;
  const n = Number.parseInt(t, 10);
  return Number.isFinite(n) && n >= 0 ? n : 0;
}

/** MT5-style combo: presets in the list, any value may be typed. */
function EditableSelectField({ label, value, options, onChange, format, parse, lockField, fieldKey, filterOptions }) {
  const locked = lockField?.(fieldKey);
  const [open, setOpen] = useState(false);
  const [raw, setRaw] = useState(null);
  const [pos, setPos] = useState(null);
  const rootRef = useRef(null);
  const btnRef = useRef(null);
  const menuRef = useRef(null);
  const openRef = useRef(false);
  const announceOpen = useExclusiveDropdown(open, setOpen);
  const fmt = format ?? String;
  const par = parse ?? ((text) => Number(text) || 0);
  const items = options ?? [];
  const display = raw ?? fmt(value);
  const needle = String(raw ?? fmt(value) ?? "").trim().toLowerCase();
  const visibleItems =
    filterOptions && needle
      ? items.filter((o) => String(o.label ?? o.value ?? "").toLowerCase().includes(needle))
      : items;

  useEffect(() => {
    openRef.current = open;
  }, [open]);

  useEffect(() => {
    if (!open) return;
    function close(e) {
      if (rootRef.current?.contains(e.target)) return;
      if (btnRef.current?.contains(e.target)) return;
      if (menuRef.current?.contains(e.target)) return;
      setOpen(false);
    }
    const onKey = (e) => e.key === "Escape" && setOpen(false);
    const t = setTimeout(() => document.addEventListener("mousedown", close), 0);
    document.addEventListener("keydown", onKey);
    return () => {
      clearTimeout(t);
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  function commit(text) {
    onChange?.(par(text));
    setRaw(null);
  }

  function openMenu(e) {
    if (locked) return;
    e.preventDefault();
    e.stopPropagation();
    if (openRef.current) {
      setOpen(false);
      return;
    }
    announceOpen();
    const r = rootRef.current.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: Math.max(r.width, 180) });
    setOpen(true);
  }

  return (
    <>
      <label>{label}</label>
      <div
        ref={rootRef}
        className={`prop-select prop-select-fill editable-select${locked ? " prop-select-disabled" : ""}`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="prop-select-box editable-select-box">
          <input
            type="text"
            className="editable-select-input"
            value={display}
            disabled={locked}
            onFocus={() => {
              setRaw(fmt(value));
            }}
            onBlur={(e) => {
              if (btnRef.current?.contains(e.relatedTarget)) return;
              commit(e.target.value);
              setOpen(false);
            }}
            onChange={(e) => {
              setRaw(e.target.value);
              if (filterOptions && !openRef.current) {
                announceOpen();
                const r = rootRef.current.getBoundingClientRect();
                setPos({ top: r.bottom, left: r.left, width: Math.max(r.width, 180) });
                setOpen(true);
              }
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                commit(e.currentTarget.value);
                e.currentTarget.blur();
                setOpen(false);
              } else if (e.key === "Escape") {
                setOpen(false);
              } else if (e.key === "ArrowDown" && !open) {
                e.preventDefault();
                openMenu(e);
              }
            }}
          />
          <span className="prop-select-btn-wrap" ref={btnRef} onMouseDown={openMenu}>
            <button
              type="button"
              className="prop-select-btn"
              aria-label="Open list"
              disabled={locked}
              tabIndex={-1}
            />
          </span>
        </div>
      </div>
      {open &&
        pos &&
        createPortal(
          <ul
            ref={menuRef}
            className="prop-select-menu prop-select-menu-front"
            style={{ position: "fixed", top: pos.top, left: pos.left, minWidth: pos.width }}
          >
            {visibleItems.length ? (
              visibleItems.map((opt) => (
                <li
                  key={`${String(opt.value)}:${String(opt.label)}`}
                  className={String(opt.value) === String(value) ? "sel" : undefined}
                  onMouseDown={(e) => {
                    e.preventDefault();
                    onChange?.(opt.value);
                    setRaw(null);
                    setOpen(false);
                  }}
                >
                  {opt.label}
                </li>
              ))
            ) : (
              <li className="prop-select-menu-empty">No matches</li>
            )}
          </ul>,
          document.body,
        )}
    </>
  );
}

function CheckField({ label, checked, onChange, disabled, lockField, fieldKey, className = "" }) {
  const locked = lockField?.(fieldKey);
  return (
    <label className={`sym-check${locked ? " sym-field-locked" : ""}${className ? ` ${className}` : ""}`}>
      <input
        type="checkbox"
        checked={!!checked}
        disabled={disabled || locked}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span>{label}</span>
    </label>
  );
}

const hasBit = (v, bit) => ((v ?? 0) & bit) !== 0;
const setBit = (v, bit, on) => (on ? (v ?? 0) | bit : (v ?? 0) & ~bit);

/** The reference shows order/filling/expiration sets as one combo summarising the selection. */
function FlagCombo({ label, labels, value, onChange, lockField, fieldKey }) {
  const locked = lockField?.(fieldKey);
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState(null);
  const rootRef = useRef(null);
  const menuRef = useRef(null);
  const openRef = useRef(false);
  const announceOpen = useExclusiveDropdown(open, setOpen);
  const all = labels.reduce((a, l) => a | l.bit, 0);
  const picked = labels.filter((l) => hasBit(value, l.bit));
  const summary =
    picked.length === labels.length
      ? "All"
      : picked.length
        ? picked.map((l) => l.label).join(", ")
        : "None";

  useEffect(() => {
    openRef.current = open;
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const close = (e) => {
      if (rootRef.current?.contains(e.target)) return;
      if (menuRef.current?.contains(e.target)) return;
      setOpen(false);
    };
    const onKey = (e) => e.key === "Escape" && setOpen(false);
    const t = setTimeout(() => document.addEventListener("mousedown", close), 0);
    document.addEventListener("keydown", onKey);
    return () => {
      clearTimeout(t);
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  function toggle(e) {
    if (locked) return;
    e.stopPropagation();
    if (openRef.current) {
      setOpen(false);
      return;
    }
    announceOpen();
    const r = rootRef.current.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: r.width });
    setOpen(true);
  }

  return (
    <>
      <label>{label}</label>
      <span className="sym-flag-combo" ref={rootRef}>
        <button type="button" className="sym-flag-combo-box" disabled={locked} onClick={toggle}>
          <span className="sym-flag-combo-value">{summary}</span>
          <span className="sym-flag-combo-arrow" />
        </button>
      </span>
      {open &&
        pos &&
        createPortal(
          <div
            ref={menuRef}
            className="sym-flag-combo-menu"
            style={{ position: "fixed", top: pos.top, left: pos.left, minWidth: pos.width }}
            onMouseDown={(e) => e.stopPropagation()}
          >
            {labels.map((l) => (
              <label key={l.bit} className="sym-flag-combo-item">
                <input
                  type="checkbox"
                  checked={hasBit(value, l.bit)}
                  onChange={(e) => onChange(setBit(value, l.bit, e.target.checked) & all)}
                />
                <span>{l.label}</span>
              </label>
            ))}
          </div>,
          document.body,
        )}
    </>
  );
}

export function CommonTab({ s, set, isNew, lockField }) {
  const domActive = Number(s.tick_book_depth) > 0;
  const spreadLocked = domActive || lockField?.("spread");
  const balanceLocked = domActive || lockField?.("spread_balance");

  // Basis and Source point at another symbol — tree only, plus "-" for none.
  const symbolPickerHeaders = useMemo(() => [{ value: "", label: "-" }], []);

  const sector = s.sector ?? 0;
  const industryIds = SymbolIndustries_bySector[sector] ?? [0];
  const industryNames = useMemo(() => {
    const names = { 0: SymbolIndustry_name[0] };
    for (const id of industryIds) {
      if (id !== 0 && SymbolIndustry_name[id]) names[id] = SymbolIndustry_name[id];
    }
    return names;
  }, [industryIds]);

  const onSectorChange = (v) => {
    set("sector", v);
    const allowed = SymbolIndustries_bySector[v ?? 0] ?? [0];
    if (!allowed.includes(Number(s.industry ?? 0))) set("industry", 0);
  };

  return (
    <>
      <TabIntro>
        The setting up of main parameters of the symbol. Please specify its name, description, and
        other parameters.
      </TabIntro>
      <div className="form-grid sym-form-two-col">
        <Field label="Symbol" fieldKey="symbol" lockField={lockField} value={s.symbol} onChange={isNew ? (v) => set("symbol", v) : undefined} />
        <Field label="Description" fieldKey="description" lockField={lockField} value={s.description} onChange={(v) => set("description", v)} />
        <Field label="Exchange" fieldKey="exchange" lockField={lockField} value={s.exchange} onChange={(v) => set("exchange", v)} />
        <Field label="International" fieldKey="international" lockField={lockField} value={s.international} onChange={(v) => set("international", v)} />
        <Field label="ISIN" fieldKey="isin" lockField={lockField} value={s.isin} onChange={(v) => set("isin", v)} />
        <SelectField label="Sector" fieldKey="sector" lockField={lockField} value={s.sector} names={SymbolSector_name} onChange={onSectorChange} />
        <Field label="CFI" fieldKey="cfi" lockField={lockField} value={s.cfi} onChange={(v) => set("cfi", v)} />
        <SelectField
          label="Industry"
          fieldKey="industry"
          lockField={lockField}
          value={s.industry ?? 0}
          names={industryNames}
          order={industryIds}
          fallback
          onChange={(v) => set("industry", v)}
        />
        <SymbolTreeSelectField
          label="Basis"
          fieldKey="basis"
          lockField={lockField}
          value={s.basis ?? ""}
          headerItems={symbolPickerHeaders}
          onChange={(v) => set("basis", v)}
        />
        <EditableSelectField
          label="Country"
          fieldKey="country"
          lockField={lockField}
          value={s.country ?? ""}
          options={COUNTRY_options}
          filterOptions
          format={(v) => (v == null || v === "" ? "" : String(v))}
          parse={(text) => String(text ?? "")}
          onChange={(v) => set("country", v)}
        />
        <SymbolTreeSelectField
          label="Source"
          fieldKey="source"
          lockField={lockField}
          value={s.source ?? ""}
          headerItems={symbolPickerHeaders}
          onChange={(v) => set("source", v)}
        />
        <Field label="Category" fieldKey="category" lockField={lockField} value={s.category} onChange={(v) => set("category", v)} />
        <label>Background</label>
        <span className="sym-color-select">
          <span
            className="sym-color-swatch"
            style={{
              background:
                Number(s.color_background) === COLOR_NONE || s.color_background == null
                  ? "#fff"
                  : colorToCss(s.color_background),
            }}
          />
          <PropSelect
            fill
            value={s.color_background ?? COLOR_NONE}
            options={BackgroundColor_options}
            onChange={(v) => set("color_background", v)}
          />
        </span>
        <Field label="Page" value={s.page} onChange={(v) => set("page", v)} />
        <SelectField label="Digits" fieldKey="digits" lockField={lockField} value={s.digits} options={plain(DIGITS_options)} onChange={(v) => set("digits", v)} />
        <EditableSelectField
          label="Market depth"
          fieldKey="tick_book_depth"
          lockField={lockField}
          value={s.tick_book_depth ?? 0}
          options={MarketDepth_options}
          format={formatMarketDepth}
          parse={parseMarketDepth}
          onChange={(v) => set("tick_book_depth", v)}
        />
        <NumField
          label="Spread"
          fieldKey="spread"
          lockField={lockField}
          value={s.spread}
          offWhenZero
          readOnly={spreadLocked}
          onChange={spreadLocked ? undefined : (v) => set("spread", v)}
        />
        <SelectField
          label="Market depth volume"
          value={s.tick_book_volume ?? 0}
          options={BookVolume_options}
          onChange={(v) => set("tick_book_volume", v)}
        />
        <label className="sym-spread-balance-label">Spread balance</label>
        <span className={`sym-spread-balance${balanceLocked ? " sym-spread-balance-disabled" : ""}`}>
          <span className="sym-spread-balance-track">
            <input
              type="range"
              min={-100}
              max={100}
              disabled={balanceLocked}
              value={s.spread_balance ?? 0}
              onChange={(e) => set("spread_balance", Number(e.target.value))}
            />
          </span>
          <span className="sym-spread-balance-caption">
            {spreadBalanceCaption(s.spread, s.spread_balance)}
          </span>
        </span>
        <SelectField
          label="Chart mode"
          value={s.tick_chart_mode}
          names={ChartMode_name}
          onChange={(v) => set("tick_chart_mode", v)}
        />
      </div>
    </>
  );
}

export function CurrencyTab({ s, set, lockField }) {
  const forexDerived = isForexDerived(s.calc_mode);
  const marginDerived = isMarginDerived(s.calc_mode);
  const cur = (v) => plain(CURRENCY_options).concat(
    CURRENCY_options.includes(v) || !v ? [] : [{ value: v, label: v }],
  );

  const setCurrency = (fieldKey, value) => {
    set(fieldKey, value);
    const digitPatch = digitsForCurrencyField(fieldKey, value);
    for (const [k, v] of Object.entries(digitPatch)) set(k, v);
  };

  return (
    <>
      <TabIntro>The setting up of base, profit, and margin currencies of the symbol.</TabIntro>
      <div className="form-grid sym-currency-block">
        <SelectField
          label="Base currency"
          fieldKey="currency_base"
          lockField={lockField}
          value={s.currency_base}
          options={cur(s.currency_base)}
          disabled={forexDerived}
          onChange={(v) => setCurrency("currency_base", v)}
        />
        <SelectField
          label="Base currency digits"
          fieldKey="currency_base_digits"
          lockField={lockField}
          value={s.currency_base_digits}
          options={plain(DIGITS_options)}
          disabled={forexDerived}
          onChange={(v) => set("currency_base_digits", v)}
        />
      </div>
      <div className="form-grid sym-currency-block">
        <SelectField
          label="Profit currency"
          fieldKey="currency_profit"
          lockField={lockField}
          value={s.currency_profit}
          options={cur(s.currency_profit)}
          disabled={forexDerived}
          onChange={(v) => setCurrency("currency_profit", v)}
        />
        <SelectField
          label="Profit currency digits"
          fieldKey="currency_profit_digits"
          lockField={lockField}
          value={s.currency_profit_digits}
          options={plain(DIGITS_options)}
          disabled={forexDerived}
          onChange={(v) => set("currency_profit_digits", v)}
        />
      </div>
      <div className="form-grid sym-currency-block">
        <SelectField
          label="Margin currency"
          fieldKey="currency_margin"
          lockField={lockField}
          value={s.currency_margin}
          options={cur(s.currency_margin)}
          disabled={marginDerived}
          onChange={(v) => setCurrency("currency_margin", v)}
        />
        <SelectField
          label="Margin currency digits"
          fieldKey="currency_margin_digits"
          lockField={lockField}
          value={s.currency_margin_digits}
          options={plain(DIGITS_options)}
          disabled={marginDerived}
          onChange={(v) => set("currency_margin_digits", v)}
        />
      </div>
    </>
  );
}

export function QuotesTab({ s, set, lockField }) {
  const tick = (bit, on) => {
    if (lockField?.("tick_flags")) return;
    set("tick_flags", setBit(s.tick_flags, bit, on));
  };
  return (
    <>
      <TabIntro>
        The setting up of different filtration levels is intended for cutting off the incorrect
        data. Please specify the deviation of quotes and their repetition for filters to trigger.
      </TabIntro>
      <div className="sym-quotes-checks">
        <CheckField
          label="Allow realtime quotes from datafeeds"
          checked={hasBit(s.tick_flags, TICK_REALTIME)}
          onChange={(on) => tick(TICK_REALTIME, on)}
        />
        <CheckField
          label="Allow negative quotes"
          checked={hasBit(s.tick_flags, TICK_NEGATIVE_PRICES)}
          onChange={(on) => tick(TICK_NEGATIVE_PRICES, on)}
        />
        <CheckField
          label="Receive market statistics from datafeeds"
          checked={hasBit(s.tick_flags, TICK_FEED_STATS)}
          onChange={(on) => tick(TICK_FEED_STATS, on)}
        />
        <CheckField
          label="Save raw prices"
          checked={hasBit(s.tick_flags, TICK_COLLECT_RAW)}
          onChange={(on) => tick(TICK_COLLECT_RAW, on)}
        />
      </div>
      <div className="form-grid sym-quotes-form">
        <NumField
          label="Soft filtration level"
          value={s.filter_soft}
          suffix="points (must not be less than 4 or 5 times spread by default)"
          onChange={(v) => set("filter_soft", v)}
        />
        <SelectField
          label="Filter"
          value={s.filter_soft_ticks}
          options={plain(FilterTicks_options)}
          fallback
          suffix="wrong quotes coming one after another"
          onChange={(v) => set("filter_soft_ticks", v)}
        />
        <NumField
          label="Hard filtration level"
          value={s.filter_hard}
          suffix="points (must not be less than soft filtration level)"
          onChange={(v) => set("filter_hard", v)}
        />
        <SelectField
          label="Filter"
          value={s.filter_hard_ticks}
          options={plain(FilterTicks_options)}
          fallback
          suffix="wrong quotes coming one after another"
          onChange={(v) => set("filter_hard_ticks", v)}
        />
        <NumField
          label="Discard filtration level"
          value={s.filter_discard}
          suffix="points (must not be less than hard filtration level)"
          onChange={(v) => set("filter_discard", v)}
        />
      </div>
      <div className="form-grid sym-quotes-pairs">
        <NumField label="Gap mode level" value={s.filter_gap} offWhenZero suffix="points" onChange={(v) => set("filter_gap", v)} />
        <NumField label="Disable gap after" value={s.filter_gap_ticks} offWhenZero suffix="ticks" onChange={(v) => set("filter_gap_ticks", v)} />
        <NumField label="Minimum spread" value={s.filter_spread_min} offWhenZero onChange={(v) => set("filter_spread_min", v)} />
        <NumField label="Maximum spread" value={s.filter_spread_max} offWhenZero suffix="points" onChange={(v) => set("filter_spread_max", v)} />
      </div>
      <div className="form-grid sym-quotes-form">
        <SelectField
          label="Delay for subscriptions"
          value={s.subscriptions_delay}
          options={plain(SubscriptionDelay_options)}
          fallback
          suffix="min"
          onChange={(v) => set("subscriptions_delay", v)}
        />
      </div>
    </>
  );
}

export function TradeTab({ s, set, lockField }) {
  const exchangeFutures = [33, 34].includes(Number(s.calc_mode));
  return (
    <>
      <TabIntro>
        The setting up of symbol trading. Please specify the type of profit calculation, type of
        placed orders, allowed trade volumes, etc.
      </TabIntro>
      <div className="form-grid sym-form-two-col sym-trade-grid">
        <NumField label="Contract size" fieldKey="contract_size" lockField={lockField} value={s.contract_size} onChange={(v) => set("contract_size", v)} />
        <NumField label="Limit & stop level" fieldKey="stops_level" lockField={lockField} value={s.stops_level} suffix="pt" onChange={(v) => set("stops_level", v)} />
        <SelectField label="Calculation" fieldKey="calc_mode" lockField={lockField} value={s.calc_mode} names={CalcMode_name} order={CalcMode_order} onChange={(v) => set("calc_mode", v)} />
        <NumField label="Freeze level" fieldKey="freeze_level" lockField={lockField} value={s.freeze_level} suffix="pt" onChange={(v) => set("freeze_level", v)} />
        <SelectField label="Trade" fieldKey="trade_mode" lockField={lockField} value={s.trade_mode} names={TradeMode_name} onChange={(v) => set("trade_mode", v)} />
        <NumField label="Max quote delay" fieldKey="quotes_timeout" lockField={lockField} value={s.quotes_timeout} offWhenZero suffix="sec" onChange={(v) => set("quotes_timeout", v)} />
        <SelectField label="GTC" fieldKey="gtc_mode" lockField={lockField} value={s.gtc_mode} names={GTCMode_name} onChange={(v) => set("gtc_mode", v)} />
        <SelectField
          label="Convert profit"
          fieldKey="trade_flags"
          lockField={lockField}
          value={hasBit(s.trade_flags, TRADE_PROFIT_BY_MARKET) ? 1 : 0}
          options={[
            { value: 0, label: "by deal" },
            { value: 1, label: "by market" },
          ]}
          onChange={(v) => set("trade_flags", setBit(s.trade_flags, TRADE_PROFIT_BY_MARKET, v === 1))}
        />
        <FlagCombo label="Filling" fieldKey="fill_flags" lockField={lockField} labels={FillFlag_labels} value={s.fill_flags} onChange={(v) => set("fill_flags", v)} />
        <label className="sym-signals-label">Enable Trading Signals</label>
        <input
          type="checkbox"
          className="sym-trade-signal-check"
          disabled={lockField?.("trade_flags")}
          checked={hasBit(s.trade_flags, TRADE_ALLOW_SIGNALS)}
          onChange={(e) => set("trade_flags", setBit(s.trade_flags, TRADE_ALLOW_SIGNALS, e.target.checked))}
        />
        <FlagCombo label="Expiration" fieldKey="expir_flags" lockField={lockField} labels={ExpirFlag_labels} value={s.expir_flags} onChange={(v) => set("expir_flags", v)} />
        <NumField label="Tick size" fieldKey="tick_size" lockField={lockField} value={s.tick_size} onChange={(v) => set("tick_size", v)} />
        <FlagCombo label="Orders" fieldKey="order_flags" lockField={lockField} labels={OrderFlag_labels} value={s.order_flags} onChange={(v) => set("order_flags", v)} />
        <NumField label="Tick value" fieldKey="tick_value" lockField={lockField} value={s.tick_value} onChange={(v) => set("tick_value", v)} />
      </div>
      <fieldset className="fieldset sym-volumes">
        <legend>Volumes</legend>
        <div className="form-grid sym-volumes-grid">
          <NumField label="Minimum" fieldKey="volume_min" lockField={lockField} value={toLots(s.volume_min)} digits={2} onChange={(v) => set("volume_min", fromLots(v))} />
          <NumField label="Step" fieldKey="volume_step" lockField={lockField} value={toLots(s.volume_step)} digits={2} onChange={(v) => set("volume_step", fromLots(v))} />
          <NumField label="Maximum" fieldKey="volume_max" lockField={lockField} value={toLots(s.volume_max)} digits={2} onChange={(v) => set("volume_max", fromLots(v))} />
          <NumField
            label="Limit"
            fieldKey="volume_limit"
            lockField={lockField}
            value={s.volume_limit ? toLots(s.volume_limit) : ""}
            onChange={(v) => set("volume_limit", fromLots(v))}
          />
        </div>
      </fieldset>
      {exchangeFutures && (
        <fieldset className="fieldset">
          <legend>Prices</legend>
          <div className="form-grid sym-volumes-grid">
            <NumField label="Settlement" value={s.price_settle} onChange={(v) => set("price_settle", v)} />
            <NumField label="Minimum" value={s.price_limit_min} onChange={(v) => set("price_limit_min", v)} />
            <NumField label="Maximum" value={s.price_limit_max} onChange={(v) => set("price_limit_max", v)} />
          </div>
        </fieldset>
      )}
    </>
  );
}

export function ExecutionTab({ s, set, lockField }) {
  const mode = Number(s.exec_mode);
  return (
    <>
      <TabIntro>
        The setting up of execution of orders by the symbol. Please specify the execution type and
        its parameters.
      </TabIntro>
      <div className="form-grid sym-exec-grid">
        <SelectField label="Execution" fieldKey="exec_mode" lockField={lockField} value={s.exec_mode} names={ExecMode_name} onChange={(v) => set("exec_mode", v)} />
        {mode === 1 && (
          <>
            <NumField label="Max time deviation" value={s.ie_timeout} suffix="seconds" onChange={(v) => set("ie_timeout", v)} />
            <SelectField
              label="Max profit deviation"
              value={s.ie_slip_profit}
              options={plain(Deviation_options)}
              fallback
              suffix="points"
              onChange={(v) => set("ie_slip_profit", v)}
            />
            <SelectField
              label="Max losing deviation"
              value={s.ie_slip_losing}
              options={plain(Deviation_options)}
              fallback
              suffix="points"
              onChange={(v) => set("ie_slip_losing", v)}
            />
            <label />
            <CheckField
              label="Fast confirmation of requotes within client deviation"
              checked={hasBit(s.ie_flags, INSTANT_FAST_CONFIRMATION)}
              onChange={(on) => set("ie_flags", setBit(s.ie_flags, INSTANT_FAST_CONFIRMATION, on))}
            />
            <NumField
              label="Maximum volume"
              value={toLots(s.ie_volume_max)}
              digits={2}
              suffix="before switching to Request execution"
              onChange={(v) => set("ie_volume_max", fromLots(v))}
            />
          </>
        )}
        {mode === 0 && (
          <>
            <NumField label="Timeout" value={s.re_timeout} suffix="seconds" onChange={(v) => set("re_timeout", v)} />
            <label />
            <CheckField
              label="Confirm orders"
              checked={hasBit(s.re_flags, REQUEST_ORDER)}
              onChange={(on) => set("re_flags", setBit(s.re_flags, REQUEST_ORDER, on))}
            />
          </>
        )}
      </div>
    </>
  );
}

export function MarginTab({ s, set }) {
  const checks = hasBit(s.margin_flags, MARGIN_CHECK_SLTP)
    ? MARGIN_CHECK_SLTP
    : hasBit(s.margin_flags, MARGIN_CHECK_PROCESS)
      ? MARGIN_CHECK_PROCESS
      : 0;
  const flag = (bit, on) => set("margin_flags", setBit(s.margin_flags, bit, on));
  return (
    <>
      <TabIntro>
        The setting up of margin requirements for the symbol. Please specify the initial and
        maintenance margin and multipliers for different types of trade operations.
      </TabIntro>
      <div className="form-grid sym-margin-grid">
        <NumField label="Initial margin" value={s.margin_initial} digits={2} onChange={(v) => set("margin_initial", v)} />
        <NumField label="Hedged margin" value={s.margin_hedged} onChange={(v) => set("margin_hedged", v)} />
        <NumField label="Maintenance margin" value={s.margin_maintenance} digits={2} onChange={(v) => set("margin_maintenance", v)} />
      </div>
      <div className="form-grid grp-check-stack">
        <CheckField
          label="Calculate hedged margin using larger leg"
          checked={hasBit(s.margin_flags, MARGIN_HEDGE_LARGE_LEG)}
          onChange={(on) => flag(MARGIN_HEDGE_LARGE_LEG, on)}
        />
        <CheckField
          label="Exclude long position PnL from free margin and margin level"
          checked={hasBit(s.margin_flags, MARGIN_EXCLUDE_PL)}
          onChange={(on) => flag(MARGIN_EXCLUDE_PL, on)}
        />
        <CheckField
          label="Recalculate margin exchange rate at the End of Day"
          checked={hasBit(s.margin_flags, MARGIN_RECALC_RATES)}
          onChange={(on) => flag(MARGIN_RECALC_RATES, on)}
        />
      </div>
      <div className="form-grid sym-margin-grid">
        <SelectField
          label="Additional margin checks"
          value={checks}
          names={MarginCheck_name}
          onChange={(v) => {
            const base = setBit(setBit(s.margin_flags, MARGIN_CHECK_PROCESS, false), MARGIN_CHECK_SLTP, false);
            set("margin_flags", v ? base | v : base);
          }}
        />
      </div>
    </>
  );
}

const RATE_GROUPS = [
  [
    "Initial margin",
    ["Buy", "buy", "margin_initial_buy", "margin_initial_buy_limit", "margin_initial_buy_stop", "margin_initial_buy_stop_limit"],
    ["Sell", "sell", "margin_initial_sell", "margin_initial_sell_limit", "margin_initial_sell_stop", "margin_initial_sell_stop_limit"],
  ],
  [
    "Maintenance margin",
    ["Buy", "buy", "margin_maintenance_buy", "margin_maintenance_buy_limit", "margin_maintenance_buy_stop", "margin_maintenance_buy_stop_limit"],
    ["Sell", "sell", "margin_maintenance_sell", "margin_maintenance_sell_limit", "margin_maintenance_sell_stop", "margin_maintenance_sell_stop_limit"],
  ],
];

export function MarginRatesTab({ s, set }) {
  return (
    <>
      <TabIntro>Please specify margin rates for different types of trade operations.</TabIntro>
      <div className="form-grid sym-margin-grid">
        <NumField label="Liquidity margin rate" value={s.margin_rate_liquidity} digits={3} onChange={(v) => set("margin_rate_liquidity", v)} />
        <NumField label="Currency margin rate" value={s.margin_rate_currency} digits={8} onChange={(v) => set("margin_rate_currency", v)} />
      </div>
      <div className="sym-margin-rates-caption">Margin rates:</div>
      <table className="sym-margin-rates-table">
        <thead>
          <tr>
            <th />
            <th>Market Order</th>
            <th>Limit Order</th>
            <th>Stop Order</th>
            <th>Stop Limit Order</th>
          </tr>
        </thead>
        <tbody>
          {RATE_GROUPS.map(([group, ...rows]) => (
            <Fragment key={group}>
              <tr className="sym-rate-group">
                <td colSpan={5}>{group}</td>
              </tr>
              {rows.map(([label, side, ...keys]) => (
                <tr key={`${group}-${label}`}>
                  <td>
                    <span className={`sym-side-icon sym-side-${side}`} aria-hidden="true" />
                    {label}
                  </td>
                  {keys.map((key) => (
                    <td key={key}>
                      <FmtInput
                        className="df-cell-input sym-rate-input"
                        value={s[key] ?? 0}
                        digits={7}
                        onChange={(v) => set(key, v)}
                      />
                    </td>
                  ))}
                </tr>
              ))}
            </Fragment>
          ))}
        </tbody>
      </table>
    </>
  );
}

const SWAP_RATE_KEYS = [
  ["Sunday", "swap_rate_sunday"],
  ["Monday", "swap_rate_monday"],
  ["Tuesday", "swap_rate_tuesday"],
  ["Wednesday", "swap_rate_wednesday"],
  ["Thursday", "swap_rate_thursday"],
  ["Friday", "swap_rate_friday"],
  ["Saturday", "swap_rate_saturday"],
];

const FOREX_MULTIPLIERS = [0, 1, 1, 3, 1, 1, 0];

const SWAP_COPY_FIELDS = [
  "swap_mode",
  "swap_long",
  "swap_short",
  "swap_year_day",
  "swap_flags",
  ...SWAP_RATE_KEYS.map(([, key]) => key),
];

function swapFieldsFromRow(row) {
  if (!row || row.swap_mode === undefined) return null;
  const out = {};
  for (const key of SWAP_COPY_FIELDS) {
    if (row[key] !== undefined) out[key] = row[key];
  }
  return Object.keys(out).length ? out : null;
}

function swapYearDayValue(v) {
  const n = Number(v);
  return SwapYearDays_options.includes(n) ? n : 360;
}

export function SwapsTab({ s, set, lockField }) {
  const enabled = Number(s.swap_mode) !== 0;
  const [selectedDay, setSelectedDay] = useState(0);
  const [fromSymbol, setFromSymbol] = useState("");
  const { symbols, reload } = useSymbols();

  const applyMultipliers = (values) =>
    SWAP_RATE_KEYS.forEach(([, key], i) => set(key, values[i]));

  async function copySwapFromSymbol() {
    const name = fromSymbol.trim();
    if (!name) return;
    if (name === s.symbol) {
      window.alert("Choose a different symbol to copy from");
      return;
    }

    let list = symbols;
    if (list.length && !("swap_mode" in list[0])) {
      list = await reload();
    }

    const row = list.find((sym) => sym.symbol === name);
    if (!row?.symbol_id) {
      window.alert(`Symbol '${name}' not found`);
      return;
    }

    let fields = swapFieldsFromRow(row);
    if (!fields) {
      const res = await fetchSymbolSwaps(row.symbol_id);
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
    set(fields);
  }

  return (
    <>
      <TabIntro>
        The setting up of charging swaps by the symbol. Please specify swaps for long and short
        positions and swap multipliers for each day of the week.
      </TabIntro>
      <div className="form-grid grp-check-stack">
        <CheckField
          label="Enable swaps"
          fieldKey="swap_mode"
          lockField={lockField}
          checked={enabled}
          onChange={(on) => set("swap_mode", on ? 1 : 0)}
        />
      </div>
      <div className="form-grid sym-swaps-type">
        <SelectField label="Type" value={s.swap_mode} names={SwapMode_name} disabled={!enabled} onChange={(v) => set("swap_mode", v)} />
      </div>
      <div className="form-grid sym-form-two-col sym-swaps-positions">
        <NumField label="Long positions" value={s.swap_long} onChange={(v) => set("swap_long", v)} />
        <NumField label="Short positions" value={s.swap_short} onChange={(v) => set("swap_short", v)} />
      </div>
      <div className="form-grid sym-form-two-col sym-swaps-year-row">
        <SelectField
          label="Days in year"
          value={swapYearDayValue(s.swap_year_day)}
          options={plain(SwapYearDays_options)}
          onChange={(v) => set("swap_year_day", v)}
        />
        <CheckField
          label="Automatically consider holidays"
          className="sym-swaps-holidays-check"
          checked={hasBit(s.swap_flags, SWAP_CONSIDER_HOLIDAYS)}
          onChange={(on) => set("swap_flags", setBit(s.swap_flags, SWAP_CONSIDER_HOLIDAYS, on))}
        />
      </div>
      <div className="sym-swaps-row">
        <span className="sym-swaps-caption">Swap multipliers</span>
        <div className="sym-swaps-table-wrap">
          <table className="sym-sessions-table sym-swaps-table">
            <thead>
              <tr>
                <th>Day of week</th>
                <th>Multiplier</th>
              </tr>
            </thead>
            <tbody>
              {SWAP_RATE_KEYS.map(([name, key], day) => (
                <tr
                  key={key}
                  className={selectedDay === day ? "selected" : ""}
                  onClick={() => setSelectedDay(day)}
                >
                  <td>
                    <span className="sym-swaps-day-mark" aria-hidden="true">
                      ✓
                    </span>
                    {name}
                  </td>
                  <td className="sym-swaps-mult-cell">
                    <input
                      type="text"
                      className="sym-swaps-mult-input"
                      disabled={!enabled}
                      value={s[key] != null && s[key] !== "" ? s[key] : ""}
                      onClick={(e) => e.stopPropagation()}
                      onFocus={() => setSelectedDay(day)}
                      onChange={(e) => set(key, Number(e.target.value) || 0)}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="sym-swaps-presets">
          <button type="button" disabled={!enabled} onClick={() => applyMultipliers(FOREX_MULTIPLIERS)}>
            Forex
          </button>
          <button type="button" disabled={!enabled} onClick={() => applyMultipliers([1, 1, 1, 1, 1, 1, 1])}>
            All week
          </button>
          <button
            type="button"
            disabled={!enabled || !fromSymbol.trim()}
            onClick={copySwapFromSymbol}
          >
            From symbol
          </button>
          <div className="sym-swaps-from-picker">
            <SymbolTreeSelectField
              label=""
              fieldKey="swap_from"
              lockField={() => !enabled}
              value={fromSymbol}
              headerItems={[{ value: "", label: "" }]}
              onChange={setFromSymbol}
            />
          </div>
        </div>
      </div>
    </>
  );
}
