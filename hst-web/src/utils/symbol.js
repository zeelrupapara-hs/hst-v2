export const TRADE_MODE = {
  DISABLED: 0,
  LONG_ONLY: 1,
  SHORT_ONLY: 2,
  CLOSE_ONLY: 3,
  FULL: 4,
};

// A symbol is live when it is enabled for trading, the feed marks it quoted, and at least
// one tick has arrived with non-zero bid and ask.
export const isSymbolLive = (symbol, live = {}) => {
  if (!symbol) return false;

  const mode = symbol.trade_level ?? symbol.trade_mode;
  if (mode === TRADE_MODE.DISABLED || mode === TRADE_MODE.CLOSE_ONLY) return false;
  if (!symbol.has_quote) return false;

  const bid = Number(live.last_bid);
  const ask = Number(live.last_ask);

  return bid > 0 && ask > 0;
};
