/** Default draft for Add → Symbol settings (MT5 opens settings dialog on Add). */
export function newSymbolDraft(folderPath = "") {
  const prefix = folderPath ? `${folderPath}\\` : "";
  return {
    symbol: "",
    description: "",
    path: prefix,
    exchange: "",
    international: "",
    isin: "",
    sector: 12,
    cfi: "",
    industry: 0,
    digits: 5,
    spread: 0,
    tick_book_depth: 0,
    currency_base: "USD",
    currency_base_digits: 2,
    currency_profit: "USD",
    currency_profit_digits: 2,
    currency_margin: "USD",
    currency_margin_digits: 2,
    exec_mode: 3,
    trade_mode: 4,
    calc_mode: 0,
    contract_size: 100000,
    sessions: [],
    time_start: 0,
    time_expiration: 0,
  };
}

export function isNewSymbolId(id) {
  return id === "new" || id === "0";
}
