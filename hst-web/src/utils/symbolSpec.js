// Label maps aligned with hst-server/model/symbol.go and admin constants.

export const SECTOR = {
  0: "Undefined",
  1: "Basic Materials",
  2: "Communication Services",
  3: "Consumer Cyclical",
  4: "Consumer Defensive",
  5: "Energy",
  6: "Financial",
  7: "Healthcare",
  8: "Industrials",
  9: "Real Estate",
  10: "Technology",
  11: "Utilities",
  12: "Currency",
  13: "Currency Crypto",
  14: "Indexes",
  15: "Commodities",
};

export const CALC_MODE = {
  0: "Forex",
  1: "Futures",
  2: "CFD",
  3: "CFD Index",
  4: "CFD Leverage",
  5: "Forex No Leverage",
  32: "Exchange Stocks",
  33: "Exchange Futures",
  34: "Exchange FORTS Futures",
  35: "Exchange Option",
  36: "Exchange Option Margin",
  37: "Exchange Bonds",
  64: "Collateral",
};

export const TRADE_MODE = {
  0: "Disabled",
  1: "Long only",
  2: "Short only",
  3: "Close only",
  4: "Full access",
};

export const CHART_MODE = {
  0: "By bid price",
  1: "By last price",
};

export const GTC_MODE = {
  0: "Good till cancelled",
  1: "Good till today including SL/TP",
  2: "Good till today excluding SL/TP",
};

export const SWAP_MODE = {
  0: "Disabled",
  1: "In points",
  2: "Using base currency",
  3: "Using margin currency",
  4: "Using group currency",
  5: "In percentage terms using current price",
  6: "In percentage terms using open price",
  7: "In points reopen position by close price",
  8: "In points reopen position by bid price",
  9: "Using profit currency",
};

const FILL_FLAGS = [
  { bit: 1, label: "Fill or Kill" },
  { bit: 2, label: "Immediate or Cancel" },
  { bit: 4, label: "Book or Cancel" },
];

const EXPIR_FLAGS = [
  { bit: 1, label: "Good till canceled" },
  { bit: 2, label: "Day" },
  { bit: 4, label: "Specified time" },
  { bit: 8, label: "Specified day" },
];

const ORDER_FLAGS = [
  { bit: 1, label: "Market" },
  { bit: 2, label: "Limit" },
  { bit: 4, label: "Stop" },
  { bit: 8, label: "Stop Limit" },
  { bit: 16, label: "Stop Loss" },
  { bit: 32, label: "Take Profit" },
  { bit: 64, label: "Close By" },
];

export const formatFlags = (value, labels) => {
  const n = Number(value);
  if (!Number.isFinite(n) || n === 0) return "—";
  const enabled = labels.filter(({ bit }) => n & bit).map(({ label }) => label);
  if (enabled.length === labels.length) return "All";
  return enabled.length ? enabled.join(", ") : "—";
};

export const formatFillFlags = (v) => formatFlags(v, FILL_FLAGS);
export const formatExpirFlags = (v) => formatFlags(v, EXPIR_FLAGS);
export const formatOrderFlags = (v) => formatFlags(v, ORDER_FLAGS);

export const lookup = (map, value, defaultKey = null) => {
  if (value == null || value === "") {
    if (defaultKey != null && map[defaultKey] != null) return map[defaultKey];
    return "—";
  }
  return map[value] ?? "—";
};

export const fmtNum = (value, digits = 2) => {
  if (value == null || value === "") return "—";
  const n = Number(value);
  if (Number.isNaN(n)) return String(value);
  if (digits === 0) return String(Math.round(n));
  return n.toFixed(digits);
};

/** Forex often shows contract size when fixed initial margin is unset. */
export const displayInitialMargin = (data) => {
  const margin = Number(data?.margin_initial);
  if (margin > 0) return fmtNum(margin, 0);
  const calc = Number(data?.calc_mode ?? data?.calculation);
  if (calc === 0 || calc === 5) return fmtNum(data?.contract_size, 0);
  return margin === 0 ? "—" : fmtNum(margin, 0);
};
