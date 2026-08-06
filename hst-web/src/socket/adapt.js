// Translates socket payloads between this terminal's vocabulary and the server's.
// The server speaks MT5: one order type covering side, prices named price_sl/price_tp, volume
// in lots. The terminal splits type and side and names things its own way.

import { adaptDeal, adaptJournal, adaptOrder, adaptPosition } from "../api/adapt";
import { SOCKET_EVENTS } from "./events";

// The terminal's order type, with its side, becomes MT5's single type.
// terminal: 0 instant, 1 buy limit, 2 buy stop, 3 sell limit, 4 sell stop
// server:   0 buy, 1 sell, 2 buy limit, 3 sell limit, 4 buy stop, 5 sell stop
const SERVER_TYPE = { 1: 2, 2: 4, 3: 3, 4: 5 };

const serverType = (type, side) => {
  if (!type) return side ? 1 : 0;
  return SERVER_TYPE[type] ?? 0;
};

// a pending order is priced, a market order is not
const isPending = (type) => Boolean(type);

// the terminal leaves an unset price null; the server reads zero as "none"
const price = (v) => Number(v) || 0;

const toServer = {
  [SOCKET_EVENTS.ORDER_CREATE]: (p) => ({
    symbol: p.symbol_id,
    type: serverType(p.type, p.side),
    volume: p.volume,
    price: isPending(p.type) ? price(p.order_price) : 0,
    price_sl: price(p.stop_loss),
    price_tp: price(p.take_profit),
    type_fill: p.fill_policy ?? 0,
    type_time: p.expiration_policy ?? 0,
    ...(p.expiry_at ? { expiry_at: p.expiry_at } : {}),
    deviation: p.deviation ?? 0,
    comment: (p.comment ?? "").slice(0, 64),
  }),

  [SOCKET_EVENTS.ORDER_UPDATE]: (p) => ({
    order_id: p.order_id ?? p.id,
    price: price(p.order_limit_price),
    price_trigger: price(p.price_trigger),
    price_sl: price(p.stop_loss),
    price_tp: price(p.take_profit),
    type_time: p.expiration_policy ?? 0,
    ...(p.expiry_at ? { expiry_at: p.expiry_at } : {}),
    comment: (p.comment ?? "").slice(0, 64),
  }),

  [SOCKET_EVENTS.ORDER_CANCEL]: (p) => ({
    order_id: p.order_id ?? p.id,
    comment: (p.comment ?? "").slice(0, 64),
  }),

  [SOCKET_EVENTS.POSITION_UPDATE]: (p) => ({
    position_id: p.position_id ?? p.id,
    price_sl: price(p.stop_loss),
    price_tp: price(p.take_profit),
    comment: (p.comment ?? "").slice(0, 64),
  }),

  // volume 0 closes the whole position, which is what the terminal means by no volume
  [SOCKET_EVENTS.POSITION_CLOSE]: (p) => ({
    position_id: p.position_id ?? p.id,
    volume: p.volume ?? 0,
    price: price(p.price),
    deviation: p.deviation ?? 0,
    comment: (p.comment ?? "").slice(0, 64),
  }),

  [SOCKET_EVENTS.POSITION_HEDGE]: (p) => ({
    position_id: p.position ?? p.position_id,
    position_by_id: p.hedge_position ?? p.position_by_id,
    comment: (p.comment ?? "").slice(0, 64),
  }),
};

// Outbound: what this terminal sends becomes what the server accepts.
export const toServerPayload = (type, payload) => {
  const fn = toServer[type];
  return fn && payload ? fn(payload) : payload;
};

const fromServer = {
  [SOCKET_EVENTS.POSITION_CREATE]: adaptPosition,
  [SOCKET_EVENTS.POSITION_UPDATE]: adaptPosition,
  [SOCKET_EVENTS.POSITION_CLOSE]: adaptPosition,
  [SOCKET_EVENTS.ORDER_CREATE]: adaptOrder,
  [SOCKET_EVENTS.ORDER_UPDATE]: adaptOrder,
  [SOCKET_EVENTS.ORDER_CANCEL]: adaptOrder,
  [SOCKET_EVENTS.DEAL_CREATE]: adaptDeal,
  [SOCKET_EVENTS.JOURNAL_CREATE]: adaptJournal,
};

// Inbound: what the server announces becomes what the tables already render.
export const fromServerPayload = (type, payload) => {
  const fn = fromServer[type];
  return fn && payload ? fn(payload) : payload;
};
