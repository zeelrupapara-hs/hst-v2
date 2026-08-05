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

export function fmtMt5DateTime(sec) {
  if (sec == null || sec === 0) return "1970.01.01 00:00";
  const d = new Date(Number(sec) * (String(sec).length > 10 ? 1 : 1000));
  if (isNaN(d.getTime())) return "1970.01.01 00:00";
  const pad = (n) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}.${pad(d.getMonth() + 1)}.${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function parseMt5DateTime(str) {
  if (!str || str === "1970.01.01 00:00") return 0;
  const m = str.match(/^(\d{4})\.(\d{2})\.(\d{2})\s+(\d{2}):(\d{2})$/);
  if (!m) return 0;
  return Math.floor(
    new Date(
      Number(m[1]),
      Number(m[2]) - 1,
      Number(m[3]),
      Number(m[4]),
      Number(m[5])
    ).getTime() / 1000
  );
}

export function windowsForDay(sessions, type, day) {
  return (sessions || [])
    .filter((s) => s.type === type && s.day === day)
    .sort((a, b) => a.open - b.open);
}

export function formatDaySessions(sessions, type, day) {
  const wins = windowsForDay(sessions, type, day);
  if (!wins.length) return "";
  return wins
    .map((w) => `${minutesToTime(w.open)}-${minutesToTime(w.close)}`)
    .join(", ");
}

export function sessionsEqual(a, b) {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    if (a[i].open !== b[i].open || a[i].close !== b[i].close) return false;
  }
  return true;
}

export function hasSeparateTrade(sessions, day) {
  return !sessionsEqual(
    windowsForDay(sessions, SESSION_QUOTE, day),
    windowsForDay(sessions, SESSION_TRADE, day)
  );
}

/** Replace sessions for given days; keep other days unchanged. */
export function mergeDaySessions(allSessions, dayIndexes, quoteWins, tradeWins, separate) {
  const days = new Set(dayIndexes);
  const kept = (allSessions || []).filter((s) => !days.has(s.day));
  const out = [...kept];
  for (const day of dayIndexes) {
    for (const w of quoteWins) {
      out.push({ type: SESSION_QUOTE, day, open: w.open, close: w.close });
    }
    const trade = separate ? tradeWins : quoteWins;
    for (const w of trade) {
      out.push({ type: SESSION_TRADE, day, open: w.open, close: w.close });
    }
  }
  return out;
}

export function tradeWithinQuote(quoteWins, tradeWins) {
  if (!tradeWins.length) return true;
  if (!quoteWins.length) return false;
  return tradeWins.every((t) =>
    quoteWins.some((q) => t.open >= q.open && t.close <= q.close)
  );
}

export function dayLabel(dayIndexes) {
  return dayIndexes.map((d) => DAY_NAMES[d] || `Day ${d}`).join(", ");
}
