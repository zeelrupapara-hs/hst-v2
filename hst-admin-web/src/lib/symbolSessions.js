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

/** Primary calendar day (00:00–24:00). */
export const PRIMARY_MIN = 1440;
/** Editable post-midnight extension (next-day 00:00–12:00). */
export const EXTEND_MIN = 720;
/** Primary + extension editable range. */
export const DISPLAY_MIN = PRIMARY_MIN + EXTEND_MIN;
/** Gray scale-only tail (next-day 12:00–24:00 preview). */
export const TAIL_MIN = 720;
export const TOTAL_SCALE_MIN = DISPLAY_MIN + TAIL_MIN;

export const PRIMARY_SCALE_HOURS = [0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24];
export const EXTEND_SCALE_HOURS = [2, 4, 6, 8, 10, 12];
export const TAIL_SCALE_HOURS = [14, 16, 18, 20, 22, 24];

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

/** Classify window as primary day or post-midnight extension as the overnight layout requires. */
export function classifySessionZone(w, windows, index) {
  const hasOvernight = windows.some((o, i) => i !== index && o.close === 1440);
  if (!hasOvernight) return "primary";
  if (w.close === 1440 || w.open >= 720) return "primary";
  // Post-midnight continuation: early-morning windows before 10:00 on the extension scale.
  if (w.open < 600) return "extension";
  return "primary";
}

export function classifyWindows(windows) {
  return windows.map((w, i) => ({
    ...w,
    zone: classifySessionZone(w, windows, i),
  }));
}

export function toDisplayMinute(storageMinute, zone) {
  return zone === "extension" ? PRIMARY_MIN + storageMinute : storageMinute;
}

export function displayPct(displayMinute) {
  return `${(displayMinute / TOTAL_SCALE_MIN) * 100}%`;
}

export function primaryScaleLeft(hour) {
  const m = hour === 24 ? PRIMARY_MIN : hour * 60;
  return displayPct(m);
}

export function extendScaleLeft(hour) {
  return displayPct(PRIMARY_MIN + hour * 60);
}

export function tailScaleLeft(hour) {
  const m = DISPLAY_MIN + (hour - 12) * 60;
  return displayPct(m);
}

function hourInWindows(hour, windows) {
  if (!windows?.length) return false;
  const start = hour * 60;
  const end = start + 60;
  return windows.some((w) => w.open < end && w.close > start);
}

/** Scale tick class for primary / extension / tail zones. */
export function hourTickClass(hour, scaleZone, windows) {
  const classified = classifyWindows(windows.length ? windows : []);
  const primaryWins = classified.filter((w) => w.zone === "primary");
  const extendWins = classified.filter((w) => w.zone === "extension");

  if (scaleZone === "primary") {
    if (hour === 24) return "sym-timeline-tick end";
    return hourInWindows(hour, primaryWins)
      ? "sym-timeline-tick"
      : "sym-timeline-tick off";
  }
  if (scaleZone === "extend") {
    const storageStart = hour * 60;
    const inSession = extendWins.some((w) => w.open <= storageStart && w.close > storageStart);
    return inSession ? "sym-timeline-tick extend-active" : "sym-timeline-tick extend";
  }
  return "sym-timeline-tick tail";
}

export function tradeWithinQuoteExtended(quoteWins, tradeWins) {
  if (!tradeWins.length) return true;
  if (!quoteWins.length) return false;
  const qRanges = quoteWins.map((w, i) => {
    const zone = classifySessionZone(w, quoteWins, i);
    return {
      open: toDisplayMinute(w.open, zone),
      close: toDisplayMinute(w.close, zone),
    };
  });
  const tRanges = tradeWins.map((w, i) => {
    const zone = classifySessionZone(w, tradeWins, i);
    return {
      open: toDisplayMinute(w.open, zone),
      close: toDisplayMinute(w.close, zone),
    };
  });
  return tRanges.every((t) => qRanges.some((q) => t.open >= q.open && t.close <= q.close));
}

export function dayLabel(dayIndexes) {
  return dayIndexes.map((d) => DAY_NAMES[d] || `Day ${d}`).join(", ");
}

/**
 * Map a mouse position on the track to storage minute + zone.
 * Editable range: primary (0–1440) and extension (1440+0 – 1440+720 display).
 */
export function minuteFromTrack(clientX, trackEl, shiftKey = false) {
  const rect = trackEl.getBoundingClientRect();
  let displayM = Math.round(((clientX - rect.left) / rect.width) * TOTAL_SCALE_MIN);
  displayM = Math.max(0, Math.min(DISPLAY_MIN, displayM));
  if (shiftKey) displayM = Math.round(displayM / 5) * 5;

  if (displayM >= PRIMARY_MIN) {
    return { minute: displayM - PRIMARY_MIN, zone: "extension", displayMinute: displayM };
  }
  return { minute: displayM, zone: "primary", displayMinute: displayM };
}

export function sessionOverlaps(windows, open, close, skipIndex = -1) {
  const lo = Math.min(open, close);
  const hi = Math.max(open, close);
  if (hi - lo < 1) return true;
  return windows.some((w, i) => i !== skipIndex && lo < w.close && hi > w.open);
}

export function canAddWindow(windows, open, close) {
  const lo = Math.min(open, close);
  const hi = Math.max(open, close);
  if (hi - lo < 1) return false;
  return !sessionOverlaps(windows, lo, hi);
}

/** Clamp marker drag for a classified window. */
export function clampMarkerMinute(minute, edge, window, zone) {
  const w = { ...window };
  if (zone === "extension") {
    if (edge === "open") {
      w.open = Math.max(0, Math.min(minute, w.close - 1));
    } else {
      w.close = Math.min(EXTEND_MIN, Math.max(minute, w.open + 1));
    }
  } else if (edge === "open") {
    w.open = Math.max(0, Math.min(minute, w.close - 1));
  } else {
    w.close = Math.min(PRIMARY_MIN, Math.max(minute, w.open + 1));
  }
  return w;
}
