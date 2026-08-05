/** Feeder mode flags — matches hst-server model.FeederFlags */
export const FEEDER_MODE_QUOTES = 1;
export const FEEDER_MODE_NEWS = 2;

/** Admin UI categories (quotes-only or news-only feeds). */
export const FEEDER_CATEGORIES = [
  { mode: FEEDER_MODE_QUOTES, label: "Quotes" },
  { mode: FEEDER_MODE_NEWS, label: "News" },
];

export function isNewsMode(mode) {
  return Number(mode) === FEEDER_MODE_NEWS;
}

export function isQuotesMode(mode) {
  return Number(mode) === FEEDER_MODE_QUOTES;
}

export function moduleLabel(modules, moduleId) {
  if (!moduleId) return "—";
  const hit = (modules || []).find((m) => m.module === moduleId);
  return hit?.label || moduleId;
}
