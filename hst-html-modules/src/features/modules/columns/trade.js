import { fmtTs } from "../../../lib/formatters.js";
import { getTradeUi } from "../../../lib/legacy.js";

function tu() {
  return getTradeUi() || {};
}

export function tradeIconCol(getAction) {
  return {
    key: "icon",
    label: "",
    html: true,
    render: (r) => tu().actionIcon?.(...getAction(r)) || "",
  };
}

export const adminPositionCols = () => [
  tradeIconCol((r) => [r.action]),
  { key: "position_id", label: "Position" },
  { key: "login", label: "Login" },
  { key: "symbol", label: "Symbol" },
  {
    key: "action",
    label: "Action",
    render: (r) => tu().labelPosAction?.(r.action) ?? r.action,
  },
  {
    key: "volume",
    label: "Volume",
    render: (r) => tu().fmtVol?.(r.volume) ?? r.volume,
  },
  { key: "price_open", label: "Open" },
  { key: "price_current", label: "Current" },
  { key: "price_sl", label: "S/L" },
  { key: "price_tp", label: "T/P" },
  {
    key: "reason",
    label: "Reason",
    render: (r) => tu().labelReason?.(r.reason) ?? r.reason,
  },
  { key: "profit", label: "Profit" },
  {
    key: "time_create",
    label: "Created",
    render: (r) => tu().fmtTs?.(r.time_create, r.time_create_ms) ?? fmtTs(r.time_create),
  },
];

export const adminOrderCols = () => [
  tradeIconCol((r) => [r.type === 0 ? 0 : 1]),
  {
    key: "time_setup",
    label: "Setup",
    render: (r) => tu().fmtTs?.(r.time_setup, r.time_setup_ms) ?? fmtTs(r.time_setup),
  },
  { key: "order_id", label: "Order" },
  { key: "login", label: "Login" },
  { key: "symbol", label: "Symbol" },
  {
    key: "type",
    label: "Type",
    render: (r) => tu().labelOrderType?.(r.type) ?? r.type,
  },
  {
    key: "state",
    label: "State",
    render: (r) => tu().labelOrderState?.(r.state) ?? r.state,
  },
  {
    key: "volume",
    label: "Volume",
    render: (r) =>
      r.volume_initial != null
        ? `${tu().fmtVol?.(r.volume)} / ${tu().fmtVol?.(r.volume_initial)}`
        : tu().fmtVol?.(r.volume),
  },
  { key: "price_order", label: "Price" },
  { key: "price_sl", label: "S/L" },
  { key: "price_tp", label: "T/P" },
  {
    key: "reason",
    label: "Reason",
    render: (r) => tu().labelReason?.(r.reason) ?? r.reason,
  },
  {
    key: "comment",
    label: "Comment",
    render: (r) => r.comment || "—",
  },
];

export const adminDealCols = () => [
  tradeIconCol((r) => [r.action, r.entry]),
  {
    key: "time",
    label: "Time",
    render: (r) => tu().fmtTs?.(r.time, r.time_ms) ?? fmtTs(r.time),
  },
  { key: "deal_id", label: "Deal" },
  { key: "login", label: "Login" },
  { key: "order_id", label: "Order" },
  { key: "position_id", label: "Position" },
  { key: "symbol", label: "Symbol" },
  {
    key: "action",
    label: "Action",
    render: (r) => tu().labelDealAction?.(r.action) ?? r.action,
  },
  {
    key: "entry",
    label: "Entry",
    render: (r) => tu().labelDealEntry?.(r.entry) ?? r.entry,
  },
  {
    key: "volume",
    label: "Volume",
    render: (r) => tu().fmtVol?.(r.volume) ?? r.volume,
  },
  { key: "price", label: "Price" },
  {
    key: "reason",
    label: "Reason",
    render: (r) => tu().labelReason?.(r.reason) ?? r.reason,
  },
  { key: "profit", label: "Profit" },
];

export function managerTradeCols(type, panel) {
  const login = {
    key: "login",
    label: "Login",
    link: (r) => `/${panel}/users/${r.login}`,
  };
  if (type === "positions") {
    return [
      { key: "position_id", label: "Position" },
      login,
      { key: "symbol", label: "Symbol" },
      { key: "action", label: "Action" },
      { key: "volume", label: "Volume" },
      { key: "profit", label: "Profit" },
      { key: "price_open", label: "Open" },
    ];
  }
  if (type === "orders") {
    return [
      { key: "order_id", label: "Order" },
      login,
      { key: "symbol", label: "Symbol" },
      { key: "type", label: "Type" },
      { key: "state", label: "State" },
      { key: "volume", label: "Volume" },
      { key: "price_order", label: "Price" },
      { key: "time_setup", label: "Setup", render: (r) => fmtTs(r.time_setup) },
    ];
  }
  return [
    { key: "deal_id", label: "Deal" },
    login,
    { key: "order_id", label: "Order" },
    { key: "symbol", label: "Symbol" },
    { key: "action", label: "Action" },
    { key: "volume", label: "Volume" },
    { key: "price", label: "Price" },
    { key: "profit", label: "Profit" },
    { key: "time", label: "Time", render: (r) => fmtTs(r.time) },
  ];
}
