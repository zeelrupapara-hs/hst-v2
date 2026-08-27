export const TRADE_MODE = {
  DISABLED: 0,
  LONG_ONLY: 1,
  SHORT_ONLY: 2,
  CLOSE_ONLY: 3,
  FULL: 4,
};

// Symbol flag bits, as the admin's Trade tab sets them (MT5 EnOrderFlags / EnExpirationFlags / EnFillingFlags).
export const ORDER_FLAG = { MARKET: 1, LIMIT: 2, STOP: 4, STOP_LIMIT: 8, SL: 16, TP: 32, CLOSE_BY: 64 };
export const EXPIR_FLAG = { GTC: 1, DAY: 2, SPECIFIED: 4, SPECIFIED_DAY: 8 };
export const FILL_FLAG = { FOK: 1, IOC: 2, BOC: 4 };

const tradeMode = (symbol) => symbol.trade_level ?? symbol.trade_mode;

// The feed is on and a price has arrived: charts, market watch and closing use this.
export const isSymbolQuoted = (symbol, live = {}) => {
  if (!symbol || !symbol.has_quote) return false;
  return Number(live.last_bid) !== 0 && Number(live.last_ask) !== 0;
};

// Opening is allowed: quoted, and the trade mode is not disabled / close only.
export const isSymbolLive = (symbol, live = {}) => {
  if (!isSymbolQuoted(symbol, live)) return false;
  const mode = tradeMode(symbol);
  return mode !== TRADE_MODE.DISABLED && mode !== TRADE_MODE.CLOSE_ONLY;
};

export const canBuy = (symbol) => symbol && tradeMode(symbol) !== TRADE_MODE.SHORT_ONLY;
export const canSell = (symbol) => symbol && tradeMode(symbol) !== TRADE_MODE.LONG_ONLY;

export const hasFlag = (value, bit) => (Number(value) & bit) !== 0;
