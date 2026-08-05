export const DAY_NAMES = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

export const SESSION_QUOTE = 0;
export const SESSION_TRADE = 1;

export function minutesToTime(m) {
  m = Number(m);
  if (m >= 1440) return "24:00";
  const h = Math.floor(m / 60);
  const min = m % 60;
  return `${h < 10 ? "0" : ""}${h}:${min < 10 ? "0" : ""}${min}`;
}

export function windowsForDay(sessions, type, day) {
  return (sessions || [])
    .filter((s) => s.type === type && s.day === day)
    .sort((a, b) => a.open - b.open);
}

export function formatDaySessions(sessions, type, day) {
  const wins = windowsForDay(sessions, type, day);
  return wins.map((w) => `${minutesToTime(w.open)}-${minutesToTime(w.close)}`).join(", ");
}

function sessionsEqual(a, b) {
  if (a.length !== b.length) return false;
  return a.every((w, i) => w.open === b[i].open && w.close === b[i].close);
}

export function hasSeparateTrade(sessions, day) {
  return !sessionsEqual(
    windowsForDay(sessions, SESSION_QUOTE, day),
    windowsForDay(sessions, SESSION_TRADE, day),
  );
}

/** Replace sessions for given days; keep other days unchanged. */
export function mergeDaySessions(allSessions, dayIndexes, quoteWins, tradeWins, separate) {
  const days = new Set(dayIndexes);
  const out = (allSessions || []).filter((s) => !days.has(s.day));
  for (const day of dayIndexes) {
    for (const w of quoteWins) out.push({ type: SESSION_QUOTE, day, open: w.open, close: w.close });
    for (const w of separate ? tradeWins : quoteWins)
      out.push({ type: SESSION_TRADE, day, open: w.open, close: w.close });
  }
  return out;
}

export function tradeWithinQuote(quoteWins, tradeWins) {
  if (!tradeWins.length) return true;
  if (!quoteWins.length) return false;
  return tradeWins.every((t) => quoteWins.some((q) => t.open >= q.open && t.close <= q.close));
}

export function dayLabel(dayIndexes) {
  return dayIndexes.map((d) => DAY_NAMES[d] || `Day ${d}`).join(", ");
}
