// Mirrors hst-server/model/routing.go; keys are the wire ints.

export const RouteAction_name = {
  0: "Delay in milliseconds",
  1: "Delay in ticks",
  2: "Clear Take Profit",
  3: "Clear Stop Loss",
  4: "Clear Stop Loss and Take Profit",
  1001: "Process to dealers",
  1002: "Process to online dealers",
  1003: "Reject",
  1004: "Requote",
  1005: "Confirm by request price",
  1006: "Confirm by market price",
  1007: "Cancel order",
};

export const RouteActions_withValue = new Set([0, 1, 1003, 1004]);

// these two carry action_value as the "skip if no dealers online" flag, not a number
export const RouteActions_dealer = new Set([1001, 1002]);

export const RouteFlags_labels = [
  { bit: 1, label: "Price" },
  { bit: 2, label: "Request Execution" },
  { bit: 4, label: "Instant Execution" },
  { bit: 8, label: "Market Execution" },
  { bit: 16, label: "Exchange Execution" },
  { bit: 32, label: "Pending order" },
  { bit: 64, label: "SL & TP modification" },
  { bit: 128, label: "Order modification" },
  { bit: 256, label: "Order removal" },
  { bit: 512, label: "Order activation" },
  { bit: 1024, label: "Stop-Limit activation" },
  { bit: 2048, label: "Stop Loss activation" },
  { bit: 4096, label: "Take Profit activation" },
  { bit: 8192, label: "Stop out (orders)" },
  { bit: 16384, label: "Stop out (positions)" },
  { bit: 32768, label: "Expiration" },
  { bit: 65536, label: "Dealer position execute" },
  { bit: 131072, label: "Dealer pending order" },
  { bit: 262144, label: "Dealer position modify" },
  { bit: 524288, label: "Dealer order modify" },
  { bit: 1048576, label: "Dealer order remove" },
  { bit: 2097152, label: "Dealer order activate" },
  { bit: 4194304, label: "Dealer Stop-Limit activate" },
  { bit: 8388608, label: "Dealer Close By" },
  { bit: 16777216, label: "Close By" },
];

export const TypeFlags_labels = [
  { bit: 1, label: "Buy" },
  { bit: 2, label: "Sell" },
  { bit: 4, label: "Buy Limit" },
  { bit: 8, label: "Sell Limit" },
  { bit: 16, label: "Buy Stop" },
  { bit: 32, label: "Sell Stop" },
  { bit: 64, label: "Buy Stop Limit" },
  { bit: 128, label: "Sell Stop Limit" },
];

// Only the conditions the engine can actually evaluate are offered: a rule keyed on a
// condition the engine cannot read would never match anything.
export const RouteCondition_name = {
  0: "Date and time",
  1: "Symbols",
  2: "Request volume",
  3: "Deviation from market",
  4: "Time",
  5: "Weekday",
  6: "Request comment",
  7: "Expert ID",
  9: "Dealer login",
  11: "Deviation from spread",
  12: "Gap",
  13: "Request reason",
  14: "Request price",
  15: "Request value",
  16: "Current spread",
  1000: "Client login",
  1001: "Client group",
  1002: "Client country",
  1003: "Client city",
  1005: "Client leverage",
  1008: "Client status",
  1009: "Client ID",
  2000: "Margin",
  2001: "Margin level",
  2002: "Free margin",
  2003: "Equity",
  2004: "Balance",
  2005: "Profit",
  4000: "Position volume",
  4001: "Position profit",
  4002: "Position age",
  4003: "Position modification time",
  4005: "Positions total",
  4006: "Positions total by symbol",
  4007: "Orders total",
  4008: "Orders total by symbol",
  4013: "Position value",
};

// text-valued conditions get the "ab" glyph, everything else the numeric "01" one
export const RouteCondition_text = new Set([1, 6, 1001, 1002, 1003, 1008]);

// calendar-valued conditions get the date glyph, as the reference draws them
export const RouteCondition_date = new Set([0, 4, 5]);

// the reference names each condition under its branch: Request \ Account \ Position \ Order
export function routeConditionGroup(id) {
  if (id === 12 || id === 16) return "Symbol";
  if (id >= 4007 && id <= 4008) return "Order";
  if (id >= 4000) return "Position";
  if (id >= 1000) return "Account";
  return "Request";
}

export const RouteConditionGroups = ["Request", "Account", "Position", "Order", "Symbol"];

export const ConditionRule_name = {
  0: "Equal (=)",
  1: "Not equal (!=)",
  2: "Greater (>)",
  3: "Greater or equal (>=)",
  4: "Less (<)",
  5: "Less or equal (<=)",
};
