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

export const RouteFlags_labels = [
  { bit: 1, label: "Price" },
  { bit: 2, label: "Request execution" },
  { bit: 4, label: "Instant execution" },
  { bit: 8, label: "Market execution" },
  { bit: 16, label: "Exchange execution" },
  { bit: 32, label: "Pending order" },
  { bit: 64, label: "SL & TP modification" },
  { bit: 128, label: "Modification" },
  { bit: 256, label: "Removal" },
  { bit: 512, label: "Pending order activation" },
  { bit: 1024, label: "Stop limit activation" },
  { bit: 2048, label: "Stop Loss activation" },
  { bit: 4096, label: "Take Profit activation" },
  { bit: 8192, label: "Stop out (orders)" },
  { bit: 16384, label: "Stop out (positions)" },
  { bit: 32768, label: "Expiration" },
  { bit: 16777216, label: "Close by" },
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

export const RouteCondition_name = {
  1: "Symbol",
  2: "Volume",
  3: "Deviation",
  4: "Time",
  5: "Weekday",
  6: "Comment",
  12: "Gap",
  14: "Request price",
  16: "Current spread",
  1000: "Login",
  1001: "Group",
  1002: "Country",
  1005: "Leverage",
  2000: "Margin",
  2001: "Margin level",
  2002: "Free margin",
  2003: "Equity",
  2004: "Balance",
  2005: "Profit",
  4000: "Position volume",
  4005: "Positions total",
  4007: "Orders total",
};

export const ConditionRule_name = {
  0: "=",
  1: "!=",
  2: ">",
  3: ">=",
  4: "<",
  5: "<=",
};
