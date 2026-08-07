/** ISO 4217 minor units — MT5-aligned defaults for currency_*_digits. */
export const CURRENCY_DIGITS = {
  JPY: 0,
  HUF: 0,
  KRW: 0,
  VND: 0,
  CLP: 0,
  ISK: 0,
  TWD: 0,
  BHD: 3,
  JOD: 3,
  KWD: 3,
  OMR: 3,
  TND: 3,
};

export const FOREX_CALC_MODES = new Set([0, 5]);
export const CFD_CALC_MODES = new Set([2, 4]);

/** Quote currencies that use 3 price digits instead of 5 (matches seed). */
const SMALL_DIGIT_QUOTES = new Set(["JPY", "HUF"]);

const FOREX_PAIR = /^([A-Z]{3})([A-Z]{3})([._#&][A-Z0-9._#&]*)?$/;

export function currencyDigits(code) {
  const c = String(code ?? "").trim().toUpperCase();
  if (!c) return 2;
  return CURRENCY_DIGITS[c] ?? 2;
}

export function isForexDerived(calcMode) {
  return FOREX_CALC_MODES.has(Number(calcMode));
}

export function isMarginDerived(calcMode) {
  return isForexDerived(calcMode) || CFD_CALC_MODES.has(Number(calcMode));
}

export function parseForexPair(symbol) {
  const m = String(symbol ?? "").trim().toUpperCase().match(FOREX_PAIR);
  if (!m) return null;
  return { base: m[1], profit: m[2] };
}

function withDigits(base, profit, margin) {
  return {
    currency_base: base,
    currency_profit: profit,
    currency_margin: margin,
    currency_base_digits: currencyDigits(base),
    currency_profit_digits: currencyDigits(profit),
    currency_margin_digits: currencyDigits(margin),
  };
}

/**
 * Derive currency fields from symbol name + calc mode (v1 / MT5 parity).
 * @returns {Record<string, string|number>|null} partial patch, or null if no auto rule applies
 */
export function deriveSymbolCurrencies({ symbol, calc_mode, current = {} }) {
  const mode = Number(calc_mode);
  const cur = current ?? {};

  if (isForexDerived(mode)) {
    const pair = parseForexPair(symbol);
    if (!pair) return null;
    const patch = withDigits(pair.base, pair.profit, pair.base);
    patch.digits = SMALL_DIGIT_QUOTES.has(pair.profit) ? 3 : 5;
    return patch;
  }

  if (CFD_CALC_MODES.has(mode)) {
    const profit = String(cur.currency_profit ?? "").trim() || "USD";
    const patch = {
      currency_profit: profit,
      currency_margin: profit,
      currency_profit_digits: currencyDigits(profit),
      currency_margin_digits: currencyDigits(profit),
    };
    const pair = parseForexPair(symbol);
    if (pair && !String(cur.currency_base ?? "").trim()) {
      Object.assign(patch, {
        currency_base: pair.base,
        currency_base_digits: currencyDigits(pair.base),
      });
    } else if (cur.currency_base) {
      patch.currency_base_digits = currencyDigits(cur.currency_base);
    }
    return patch;
  }

  return null;
}

/** Apply currency digits when user picks a currency code manually. */
export function digitsForCurrencyField(fieldKey, currencyCode) {
  const d = currencyDigits(currencyCode);
  switch (fieldKey) {
    case "currency_base":
      return { currency_base_digits: d };
    case "currency_profit":
      return { currency_profit_digits: d };
    case "currency_margin":
      return { currency_margin_digits: d };
    default:
      return {};
  }
}
