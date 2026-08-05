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
};

export const ExecMode_name = {
  0: "Request Execution",
  1: "Instant Execution",
  2: "Market Execution",
  3: "Exchange Execution",
};

export const GTCMode_name = {
  0: "Good till cancelled",
  1: "Orders daily",
  2: "Orders & Stops daily",
};

export const SwapMode_name = {
  0: "Disabled",
  1: "In points",
  2: "By symbol currency",
  3: "By margin currency",
  4: "By group currency",
  5: "In percent (current price)",
  6: "In percent (open price)",
  7: "Reopen by close price",
  8: "Reopen by bid price",
  9: "In percent annual",
};

export const SwapDays_options = [360, 365, 366];

export const SymbolSector_name = { 0: "Undefined", 12: "Currency" };

export const FillFlag_labels = [
  { bit: 1, label: "Fill or Kill" },
  { bit: 2, label: "Immediate or Cancel" },
];

export const ExpirFlag_labels = [
  { bit: 1, label: "Good till cancelled" },
  { bit: 2, label: "Good till today" },
  { bit: 4, label: "Good till specified date" },
  { bit: 8, label: "Good till specified day" },
];

export const OrderFlag_labels = [
  { bit: 1, label: "Market orders" },
  { bit: 2, label: "Limit orders" },
  { bit: 4, label: "Stop orders" },
  { bit: 8, label: "Stop limit orders" },
  { bit: 16, label: "Stop loss" },
  { bit: 32, label: "Take profit" },
  { bit: 64, label: "Close by" },
];

// Internal volume units per lot on the admin symbol wire.
export const VOLUME_UNITS = 10000;
export const toLots = (v) => (v == null ? 0 : v / VOLUME_UNITS);
export const fromLots = (v) => Math.round(Number(v || 0) * VOLUME_UNITS);
