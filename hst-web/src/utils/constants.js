export const POSITION_SIDE = {
  0: "Buy",
  1: "Sell",
};

export const ORDER_STATUS = {
  0: "Started",
  1: "Placed",
  2: "Partially Filled",
  3: "Filled",
  4: "Canceled",
  5: "Rejected",
  6: "Expired",
};

export const ORDER_TYPES = {
  0: "Instant Execution", // price = last_ask
  1: "Buy Limit", // price < last_ask
  2: "Buy Stop", // price > last_ask
  3: "Sell Limit", // price > last_bid
  4: "Sell Stop", // price < last_bid
};

export const FILL_POLICY = {
  0: "Fill or Kill",
  1: "Immediate or Cancel",
  2: "Return",
};

export const EXPIRATION_POLICY = {
  0: "GTC",
  1: "Today",
  2: "Specified Time",
  3: "Specified Day",
};

export const CHANNEL = {
  0: "Mobile",
  1: "Web",
  2: "Desktop",
  3: "Api",
  4: "Script",
  5: "System",
  6: "Sdk",
};

export const ALERT_STATUS = {
  0: "Not Triggered",
  1: "In Queue",
  2: "Triggered",
};

export const ALERT_TYPES = {
  market_ask: "Market Ask",
  market_bid: "Market Bid",
  balance: "Balance",
  equity: "Equity",
  margin_level: "Margin Level",
};

export const ALERT_CONDITIONS = {
  greater_than: "Greater Than",
  less_than: "Less Than",
};

export const REPORT_STATUS = {
  0: "Pending",
  1: "Completed",
  2: "Failed",
};

export const SYMBOL_DETAILS = {
  CALCULATION: {
    0: "Forex",
    1: "Forex No Leverage",
    2: "CFD",
    3: "CFD Leverage",
    4: "Futures",
  },
  TRADE_LEVEL: {
    0: "Full",
    1: "Buy",
    2: "Sell",
    3: "Close",
    4: "Disabled",
  },
  SWAP_TYPE: {
    0: "Points",
    1: "In Money",
    2: "Current Price",
    3: "Open Price",
  },
  EXECUTION: {
    0: "Instance",
    1: "Request",
    2: "Execute",
    3: "Exchange",
  },
};
