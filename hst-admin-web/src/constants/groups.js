// Mirrors hst-server/model/group.go enum values; keys are the wire ints.

export const AuthMode_name = {
  0: "Normal",
  1: "1024-bit RSA SSL certificate",
  2: "2048-bit RSA SSL certificate",
};

export const MarginMode_name = {
  0: "for Retail Forex, CFD, Futures",
  1: "for Stock Exchange, based on margin discount rates",
  2: "for Retail Forex, CFD, Futures with hedging",
};

/** The reference lists netting, then hedging, then exchange — not numeric order. */
export const MarginMode_order = [0, 2, 1];

export const MarginMode_short = { 0: "Netting", 1: "Exchange", 2: "Hedged" };

export const NewsMode_name = {
  0: "no news",
  1: "only headers",
  2: "full package",
};

export const ReportsMode_name = {
  0: "Disabled",
  1: "End of Day and End of Month",
  2: "End of Day",
  3: "End of Month",
};

export const TransferMode_name = {
  0: "Disable transfer of funds between accounts",
  1: "Enable between accounts with the same name",
  2: "Enable between account from the same group and sub-groups",
  3: "Enable between account with the same name and group",
};

export const FreeMarginMode_name = {
  0: "Do not use unrealized profit/loss",
  1: "Use unrealized profit/loss",
  2: "Use unrealized profit",
  3: "Use unrealized loss",
};

export const StopOutMode_name = { 0: "%, percent", 1: "money" };

export const MarginFreeProfitMode_name = {
  0: "Use daily fixed profit/loss",
  1: "Use daily fixed loss",
};

export const HistoryLimit_name = {
  0: "All",
  1: "1 month",
  2: "3 months",
  3: "6 months",
  4: "1 year",
  5: "2 years",
  6: "3 years",
};

/** Strict MT5 presets for orders/positions — dropdown only; any value may be typed. */
export const Limit_strict_options = [
  { value: 0, label: "unlimited" },
  { value: 10, label: "10" },
  { value: 30, label: "30" },
  { value: 50, label: "50" },
];

/** Demo Maximum symbols shows blank when unlimited. */
export const Limit_demo_symbols_options = [
  { value: 0, label: "" },
  { value: 10, label: "10" },
  { value: 30, label: "30" },
  { value: 50, label: "50" },
];

function appendCustomLimit(base, current) {
  const n = Number(current) || 0;
  const next = n !== 0 && !base.some((o) => o.value === n)
    ? [...base, { value: n, label: String(n) }]
    : [...base];
  return next.sort((a, b) => a.value - b.value);
}

export function limitStrictOptions() {
  return Limit_strict_options;
}

export function limitSymbolsOptions(current, { demo = false } = {}) {
  const base = demo ? Limit_demo_symbols_options : Limit_strict_options;
  return appendCustomLimit(base, current);
}

export function formatLimit(value, { demoSymbols = false } = {}) {
  const n = Number(value) || 0;
  if (n === 0) return demoSymbols ? "" : "unlimited";
  return String(n);
}

export function parseLimit(text) {
  const raw = String(text ?? "").trim();
  if (raw === "" || raw.toLowerCase() === "unlimited") return 0;
  const n = Number.parseInt(raw, 10);
  return Number.isFinite(n) && n >= 0 ? n : 0;
}

/** Demo account default leverage presets (MT5 "1 : N" display) — strict dropdown only. */
export const DemoLeverage_options = [
  { value: "", label: "" },
  { value: 1, label: "1 : 1" },
  { value: 10, label: "1 : 10" },
  { value: 20, label: "1 : 20" },
  { value: 50, label: "1 : 50" },
  { value: 100, label: "1 : 100" },
  { value: 200, label: "1 : 200" },
  { value: 500, label: "1 : 500" },
  { value: 1000, label: "1 : 1000" },
];

export const PermissionFlag_labels = [
  { bit: 2, label: "Enable connections" },
  { bit: 1, label: "Enable certificate confirmation" },
  { bit: 4, label: "Change password at first login" },
  { bit: 16, label: "Show the risk warning window after connection" },
  { bit: 32, label: "Enforce country-specific regulatory restrictions for retail clients" },
];

export const PermissionFlag_forceOtp = 8;

/** notify_deals | notify_orders | notify_balances — multi-select in the reference UI. */
export const PermissionFlag_notifyMask = 64 | 128 | 256;
export const NotifyFlag_deals = 64;
export const NotifyFlag_orders = 128;
export const NotifyFlag_balances = 256;
export const NotifyFlag_all = NotifyFlag_deals | NotifyFlag_orders | NotifyFlag_balances;
export const NotifyFlag_labels = [
  { bit: NotifyFlag_deals, label: "Deals" },
  { bit: NotifyFlag_orders, label: "Orders" },
  { bit: NotifyFlag_balances, label: "Balance" },
];

/** Legacy single-select mapping; the dialog uses NotifyFlag_labels via NotifyCombo. */
export const NotifyMode_options = [
  { value: 64 | 128 | 256, label: "All" },
  { value: 64, label: "Deals" },
  { value: 128, label: "Orders" },
  { value: 256, label: "Balance operations" },
  { value: 0, label: "Disabled" },
];

/** Laid out in two columns, row by row. */
export const TradeFlag_labels = [
  { bit: 4, label: "Enable trading by Expert Advisors" },
  { bit: 256, label: "Enable position closing according to FIFO rule" },
  { bit: 2, label: "Enable trailing stops" },
  { bit: 512, label: "Prohibit hedge positions" },
  { bit: 1, label: "Enable charge of swaps" },
  { bit: 1024, label: "Enable deal cost calculation" },
];

export const TradeFlag_signalsMask = 16 | 32;
export const SignalsMode_options = [
  { value: 0, label: "Disabled" },
  { value: 16, label: "Enable all signals from all brokers" },
  { value: 32, label: "Enable signals from my servers only" },
];

export const TradeFlag_soFullyHedged = 128;
export const TradeFlag_soCompensation = 64;
export const TradeFlag_soCompensationCredit = 2048;

export const MarginFlag_clearAcc = 1;

export const ReportsFlag_statements = 4;
export const ReportsFlag_email = 1;
export const ReportsFlag_support = 2;

/** The reference offers these deposit currencies; the field stays free-form on the wire. */
export const Currency_options = ["USD", "EUR", "GBP", "JPY", "CHF", "AUD", "CAD", "RUR"];

/** Demo groups live under a top-level `demo\\...` path in MT5. */
export function isDemoGroup(path) {
  const first = String(path ?? "").split("\\").filter(Boolean)[0]?.toLowerCase() ?? "";
  return first === "demo";
}

/** The platform derives the group type from the first path segment. */
export function groupKind(path) {
  const first = String(path ?? "").split("\\").filter(Boolean)[0]?.toLowerCase() ?? "";
  if (first === "demo") return "demo";
  if (first === "manager" || first === "managers") return "manager";
  if (first === "contest") return "contest";
  if (first === "coverage") return "coverage";
  if (first === "preliminary") return "preliminary";
  return "real";
}
