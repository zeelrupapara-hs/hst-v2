// Mirrors hst-server/model/order.go OrderType values.

export const OrderType_name = {
  0: "buy",
  1: "sell",
  2: "buy limit",
  3: "sell limit",
  4: "buy stop",
  5: "sell stop",
  6: "buy stop limit",
  7: "sell stop limit",
  8: "close by",
};

export const OrderState_name = {
  0: "started",
  1: "placed",
  2: "canceled",
  3: "partial",
  4: "filled",
  5: "rejected",
  6: "expired",
  7: "request add",
  8: "request modify",
  9: "request cancel",
};

export const DealEntry_name = { 0: "in", 1: "out", 2: "in/out", 3: "out by" };

export const DealAction_name = {
  0: "buy",
  1: "sell",
  2: "balance",
  3: "credit",
  4: "charge",
  5: "correction",
  6: "bonus",
  7: "commission",
};
