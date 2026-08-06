// Translates hst-server's payloads into the shapes this terminal was written against.
// Everything here is naming only: no business rules, no derived numbers.

// Our symbols have no numeric id, and the tick stream keys on the symbol name, so the name IS
// the id. Keeping one rule for both transports avoids a lookup table that could disagree.
export const symbolId = (name) => name;

// A folder in the symbol tree is identified by its path; the terminal wants a number, so the
// path is hashed. Same path always yields the same id, which is what the group filter needs.
export const groupId = (path) => {
  let h = 0;
  for (const c of String(path)) h = (h * 31 + c.charCodeAt(0)) >>> 0;
  return h % 1e9;
};

// order type 0..8 -> the terminal's [type, side], which keeps side in its own field
const ORDER_TYPE = {
  0: [0, 0],
  1: [0, 1],
  2: [1, 0],
  3: [3, 1],
  4: [2, 0],
  5: [4, 1],
  6: [2, 0],
  7: [4, 1],
  8: [0, 0],
};

// our OrderState -> the terminal's ORDER_STATUS, which orders its values differently
const ORDER_STATE = { 0: 0, 1: 1, 2: 4, 3: 2, 4: 3, 5: 5, 6: 6, 7: 1, 8: 1, 9: 1 };

// the terminal's channel codes, ours are names
const CHANNEL = { mobile: 0, web: 1, desktop: 2, api: 3, script: 4, system: 5, sdk: 6 };

// zero means "not set" on the wire, but the terminal renders a literal 0 unless it is null
const unset = (v) => v || null;

export const adaptSymbol = (s) => {
  const parts = String(s.path || "").split("\\");
  const folder = parts.slice(0, -1);

  return {
    ...s,
    id: symbolId(s.symbol),
    last_bid: s.bid,
    last_ask: s.ask,
    high_bid: s.high,
    low_bid: s.low,
    low_ask: s.low,
    spread: s.spread_diff,
    spread_balance: 0,
    min_value: s.volume_min,
    max_value: s.volume_max,
    step: s.volume_step,
    stop_level: s.stops_level,
    calculation: s.calc_mode,
    trade_level: s.trade_mode,
    swap_type: s.swap_mode,
    execution: s.exec_mode,
    symbol_class: folder.length
      ? {
          id: groupId(folder.join("\\")),
          parent_id: folder.length > 1 ? groupId(folder.slice(0, -1).join("\\")) : null,
          desc: folder[folder.length - 1],
        }
      : null,
  };
};

// Every symbol carries the path it hangs under, so the folders can be read off the symbols the
// terminal already has. Asking the server for the tree as well would fetch the whole instrument
// set a second time, since its leaves carry the symbols in full.
export const groupsFromSymbols = (symbols = []) => {
  const groups = {};

  for (const s of symbols) {
    const parts = String(s.path || "").split("\\").slice(0, -1);

    parts.forEach((name, i) => {
      const path = parts.slice(0, i + 1).join("\\");
      groups[groupId(path)] = {
        id: groupId(path),
        desc: name,
        parent_id: i ? groupId(parts.slice(0, i).join("\\")) : null,
      };
    });
  }

  return groups;
};

export const adaptPosition = (p) => ({
  id: p.position_id,
  symbol_id: symbolId(p.symbol),
  type: 0,
  side: p.action,
  volume: p.volume,
  open_price: p.price_open,
  close_price: p.price_current,
  stop_loss: unset(p.price_sl),
  take_profit: unset(p.price_tp),
  swaps: p.storage,
  created_at: p.time_create,
  updated_at: p.time_update,
  profit: p.profit,
  comment: p.comment,
});

export const adaptOrder = (o) => {
  const [type, side] = ORDER_TYPE[o.type] ?? [0, 0];

  return {
    id: o.order_id,
    symbol_id: symbolId(o.symbol),
    type,
    side,
    status: ORDER_STATE[o.state] ?? o.state,
    volume: o.volume,
    order_limit_price: unset(o.price_order),
    filled_price: unset(o.price_current),
    stop_loss: unset(o.price_sl),
    take_profit: unset(o.price_tp),
    created_at: o.time_setup,
    expiry_at: unset(o.time_expiration),
    comment: o.comment,
  };
};

export const adaptDeal = (d) => ({
  id: d.deal_id,
  symbol_id: symbolId(d.symbol),
  type: 0,
  side: d.action,
  volume: d.volume,
  created_at: d.time,
  profit: d.profit,
  comment: d.comment,
});

export const adaptClosedPosition = (p) => ({
  id: p.position_id,
  symbol_id: symbolId(p.symbol),
  type: 0,
  side: p.action,
  volume: p.volume,
  open_price: p.price_open,
  close_price: p.price_close,
  stop_loss: unset(p.price_sl),
  take_profit: unset(p.price_tp),
  created_at: p.time_open,
  updated_at: p.time_close,
  profit: p.profit,
  comment: p.comment ?? "",
});

// The terminal carries an alert as one comma separated formula; ours carries typed fields.
const ALERT_KIND = { 1: "market_ask", 2: "market_bid", 3: "balance", 4: "equity", 5: "margin_level" };
const ALERT_COND = { 1: "greater_than", 2: "less_than" };

const reverse = (m) => Object.fromEntries(Object.entries(m).map(([k, v]) => [v, Number(k)]));

export const adaptAlert = (a) => ({
  ...a,
  id: a.alert_id,
  status: a.triggered_at ? 2 : 0,
  formula: [ALERT_KIND[a.kind], ALERT_COND[a.condition], a.value, a.symbol ?? ""].join(","),
});

// an account alert carries no symbol, and sending one back would be refused
export const alertFromFormula = ({ formula }) => {
  const [kind, condition, value, symbol] = String(formula ?? "").split(",");

  return {
    kind: reverse(ALERT_KIND)[kind] ?? 0,
    condition: reverse(ALERT_COND)[condition] ?? 0,
    value: Number(value),
    symbol: symbol || "",
  };
};

export const adaptSummary = (a) => ({
  ...a,
  used_margin: a.margin,
  free_margin: a.margin_free,
});

// the table keys its rows on id, and a virtual table with no key renders nothing at all
export const adaptJournal = (j) => ({
  ...j,
  id: j.journal_id,
  desc: j.message,
  channel: CHANNEL[j.channel] ?? 5,
});

// axios hands back the whole response; only the inner data list is rewritten
export const mapList = (res, fn) => ({
  ...res,
  data: { ...res.data, data: (res.data?.data ?? []).map(fn) },
});

export const mapOne = (res, fn) => ({
  ...res,
  data: { ...res.data, data: res.data?.data ? fn(res.data.data) : res.data?.data },
});
