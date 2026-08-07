import { SwapYearDays_options } from "@/constants/symbols.js";

/** Sun–Sat swap_rate_* keys paired with display labels. */
export const SWAP_DAY_KEYS = [
  ["Sunday", "swap_rate_sunday"],
  ["Monday", "swap_rate_monday"],
  ["Tuesday", "swap_rate_tuesday"],
  ["Wednesday", "swap_rate_wednesday"],
  ["Thursday", "swap_rate_thursday"],
  ["Friday", "swap_rate_friday"],
  ["Saturday", "swap_rate_saturday"],
];

/** MT5 forex preset: no swap Sun/Sat, triple Wed. */
export const FOREX_SWAP_MULTIPLIERS = [0, 1, 1, 3, 1, 1, 0];

export const SWAP_COPY_FIELDS = [
  "swap_mode",
  "swap_long",
  "swap_short",
  "swap_year_day",
  "swap_flags",
  ...SWAP_DAY_KEYS.map(([, key]) => key),
];

export function swapFieldsFromRow(row) {
  if (!row || row.swap_mode === undefined) return null;
  const out = {};
  for (const key of SWAP_COPY_FIELDS) {
    if (row[key] !== undefined) out[key] = row[key];
  }
  return Object.keys(out).length ? out : null;
}

export function swapYearDayValue(v) {
  const n = Number(v);
  return SwapYearDays_options.includes(n) ? n : 360;
}

const SWAP_MULT_PRESETS = [1, 2, 3];

/** Group-symbol multiplier dropdown: Default, None, 1–3, plus the current value. */
export function swapMultiplierOptions(value) {
  const items = [
    { value: null, label: "Default" },
    { value: 0, label: "None" },
  ];
  for (const n of SWAP_MULT_PRESETS) items.push({ value: n, label: String(n) });
  const num = value == null || value === "" ? null : Number(value);
  if (num != null && !Number.isNaN(num) && num !== 0 && !SWAP_MULT_PRESETS.includes(num)) {
    items.push({ value: num, label: String(num) });
  }
  return items;
}

export function parseSwapMultiplier(text) {
  const raw = String(text ?? "").trim();
  if (raw === "" || raw.toLowerCase() === "default") return null;
  if (raw.toLowerCase() === "none") return 0;
  const n = Number(raw);
  return Number.isNaN(n) ? null : n;
}

export function formatSwapMultiplier(v) {
  if (v == null || v === "") return "";
  if (Number(v) === 0) return "None";
  return String(v);
}

/** Nullable long/short swap amount for group overrides. */
export function parseSwapAmount(text) {
  const raw = String(text ?? "").trim();
  if (raw === "" || raw.toLowerCase() === "default") return null;
  const n = Number(raw);
  return Number.isNaN(n) ? null : n;
}

export function swapAmountOptions(value) {
  const items = [{ value: null, label: "Default" }];
  const num = value == null || value === "" ? null : Number(value);
  if (num != null && !Number.isNaN(num) && !items.some((o) => o.value === num)) {
    items.push({ value: num, label: String(num) });
  }
  return items;
}

/** Days-in-year select for group overrides (Default + catalog presets). */
export function swapYearDayOptions(value) {
  const items = [{ value: null, label: "Default" }];
  for (const d of SwapYearDays_options) items.push({ value: d, label: String(d) });
  const num = value == null || value === "" ? null : Number(value);
  if (num != null && !Number.isNaN(num) && !SwapYearDays_options.includes(num)) {
    items.push({ value: num, label: String(num) });
  }
  return items;
}

export function parseSwapYearDay(text) {
  const raw = String(text ?? "").trim();
  if (raw === "" || raw.toLowerCase() === "default") return null;
  const n = Number.parseInt(raw, 10);
  return Number.isFinite(n) && n > 0 ? n : null;
}

/** MT5 group-symbol Consider holidays: Default / Disabled / Automatically. */
export const SWAP_HOLIDAYS_OPTIONS = [
  { value: "", label: "Default" },
  { value: 0, label: "Disabled" },
  { value: 1, label: "Automatically" },
];
