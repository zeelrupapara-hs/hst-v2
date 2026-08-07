// Mirrors hst-server/model/symbol.go enum values; keys are the wire ints.

export const TradeMode_name = {
  0: "Disabled",
  1: "Long only",
  2: "Short only",
  3: "Close only",
  4: "Full access",
};

export const CalcMode_name = {
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

// The reference lists calculation modes by meaning, not by wire value.
export const CalcMode_order = [0, 5, 2, 1, 3, 4, 32, 37, 33, 34, 35, 36, 64];

export const ExecMode_name = {
  0: "Request",
  1: "Instant",
  2: "Market",
  3: "Exchange",
};

export const GTCMode_name = {
  0: "Good till canceled",
  1: "Good till today including SL/TP",
  2: "Good till today excluding SL/TP",
};

export const SwapMode_name = {
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

export const SwapYearDays_options = [360, 365, 366];

/** @deprecated use SwapYearDays_options */
export const SwapDays_options = SwapYearDays_options;

export const SymbolSector_name = {
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

export { SymbolIndustry_name, SymbolIndustries_bySector } from "./symbol_industry.generated.js";

export const ChartMode_name = { 0: "by bid price", 1: "by last price" };

export const FillFlag_labels = [
  { bit: 1, label: "Fill or Kill" },
  { bit: 2, label: "Immediate or Cancel" },
  { bit: 4, label: "Book or Cancel" },
];

export const ExpirFlag_labels = [
  { bit: 1, label: "Good till canceled" },
  { bit: 2, label: "Day" },
  { bit: 4, label: "Specified time" },
  { bit: 8, label: "Specified day" },
];

export const OrderFlag_labels = [
  { bit: 1, label: "Market" },
  { bit: 2, label: "Limit" },
  { bit: 4, label: "Stop" },
  { bit: 8, label: "Stop Limit" },
  { bit: 16, label: "Stop Loss" },
  { bit: 32, label: "Take Profit" },
  { bit: 64, label: "Close By" },
];

// tick_flags bits (TickFlags in the model).
export const TICK_REALTIME = 1;
export const TICK_COLLECT_RAW = 2;
export const TICK_FEED_STATS = 4;
export const TICK_NEGATIVE_PRICES = 8;

// ie_flags bits (InstantFlags).
export const INSTANT_FAST_CONFIRMATION = 1;

// trade_flags bits (SymbolTradeFlags).
export const TRADE_PROFIT_BY_MARKET = 1;
export const TRADE_ALLOW_SIGNALS = 2;

// margin_flags bits (SymbolMarginFlags).
export const MARGIN_CHECK_PROCESS = 1;
export const MARGIN_CHECK_SLTP = 2;
export const MARGIN_HEDGE_LARGE_LEG = 4;
export const MARGIN_EXCLUDE_PL = 8;
export const MARGIN_RECALC_RATES = 16;

export const MarginCheck_name = {
  0: "None",
  1: "Check before executing orders",
  2: "Check on SL/TP trigger",
};

// swap_flags bits (SwapFlags).
export const SWAP_CONSIDER_HOLIDAYS = 1;

// re_flags bits (RequestFlags).
export const REQUEST_ORDER = 1;

// color_background is a wire int; the reference calls the unset value "None".
export const COLOR_NONE = 4294967295;

export const BackgroundColor_options = [
  { value: COLOR_NONE, label: "None" },
  { value: 0xffffff, label: "White" },
  { value: 0xc0c0c0, label: "Silver" },
  { value: 0x00ffff, label: "Yellow" },
  { value: 0x00a5ff, label: "Orange" },
  { value: 0x8080f0, label: "Salmon" },
  { value: 0x90ee90, label: "Light Green" },
  { value: 0xe6d8ad, label: "Light Blue" },
  { value: 0xd8bfd8, label: "Thistle" },
];

/** The wire colour is BGR, CSS wants RGB. */
export const colorToCss = (v) => {
  const n = Number(v) >>> 0;
  return `#${(n & 0xff).toString(16).padStart(2, "0")}${((n >> 8) & 0xff).toString(16).padStart(2, "0")}${((n >> 16) & 0xff).toString(16).padStart(2, "0")}`;
};

/** Row/cell background from symbol settings; null when unset ("None"). */
export const symbolBackgroundCss = (colorBackground) => {
  const n = Number(colorBackground);
  if (colorBackground == null || Number.isNaN(n) || (n >>> 0) === COLOR_NONE) return null;
  return colorToCss(colorBackground);
};

export const CURRENCY_options = [
  "AUD", "CAD", "CHF", "CNH", "CNY", "CZK", "DKK", "EUR", "GBP", "HKD", "HUF",
  "ILS", "JPY", "MXN", "NOK", "NZD", "PLN", "RUB", "SEK", "SGD", "THB", "TRY", "USD", "ZAR",
];

export { CURRENCY_DIGITS, currencyDigits } from "@/lib/symbolCurrency.js";

export const DIGITS_options = [0, 1, 2, 3, 4, 5, 6, 7, 8];

export const MarketDepth_options = [
  { value: 0, label: "off" },
  { value: 10, label: "10" },
  { value: 16, label: "16" },
  { value: 32, label: "32" },
];

export const BookVolume_options = [
  { value: 0, label: "default" },
  ...Array.from({ length: 32 }, (_, i) => ({ value: i + 1, label: String(i + 1) })),
];

export const FilterTicks_options = Array.from({ length: 10 }, (_, i) => i + 1);

export const SubscriptionDelay_options = [0, 1, 5, 10, 15, 30, 60];

// MT5 instant-execution deviation presets (0–5); any non-negative integer may be typed.
export const Deviation_options = Array.from({ length: 6 }, (_, i) => i);

// Internal volume units per lot on the admin symbol wire.
export const VOLUME_UNITS = 10000;
export const toLots = (v) => (v == null ? 0 : v / VOLUME_UNITS);
export const fromLots = (v) => Math.round(Number(v || 0) * VOLUME_UNITS);
