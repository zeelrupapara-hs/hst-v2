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

/** Vertical grid lines every 2 hours across the full MT5 frame. */
export const GRID_DISPLAY_MINUTES = [
  ...PRIMARY_SCALE_HOURS.map((h) => (h === 24 ? PRIMARY_MIN : h * 60)),
  ...EXTEND_SCALE_HOURS.map((h) => PRIMARY_MIN + h * 60),
  ...TAIL_SCALE_HOURS.map((h) => DISPLAY_MIN + (h - 12) * 60),
];

export function minutesToTime(m) {
  m = Number(m);
  if (m >= 1440) return "24:00";
  const h = Math.floor(m / 60);
  const min = m % 60;
  return `${h < 10 ? "0" : ""}${h}:${min < 10 ? "0" : ""}${min}`;
}

/** MT5-style scale label (00:00 at day start, then 02…24). */
export function formatScaleHour(hour, zone) {
  if (zone === "primary" && hour === 0) return "00:00";
  if (hour === 24) return "24";
  return hour < 10 ? `0${hour}` : String(hour);
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

function hasMidnightEnd(windows, skipIndex = -1) {
  return windows.some((w, i) => i !== skipIndex && w.close === PRIMARY_MIN);
}

/** Classify window as primary day or post-midnight extension (MT5 overnight layout). */
export function classifySessionZone(w, windows, index) {
  if (w.close === PRIMARY_MIN) return "primary";
  if (hasMidnightEnd(windows, index) && w.open < EXTEND_MIN && w.close <= EXTEND_MIN) {
    return "extension";
  }
  return "primary";
}

export function classifyWindows(windows) {
  return windows.map((w, i) => ({
    ...w,
    zone: classifySessionZone(w, windows, i),
  }));
}

/** Windows in the same zone may overlap; primary and extension never overlap each other. */
export function sessionOverlaps(windows, open, close, skipIndex = -1, zone = "primary") {
  const lo = Math.min(open, close);
  const hi = Math.max(open, close);
  if (hi - lo < 1) return true;
  return windows.some((w, i) => {
    if (i === skipIndex) return false;
    if (classifySessionZone(w, windows, i) !== zone) return false;
    return lo < w.close && hi > w.open;
  });
}

export function toDisplayMinute(storageMinute, zone) {
  return zone === "extension" ? PRIMARY_MIN + storageMinute : storageMinute;
}

/** True when timeline shows post-midnight extension + tail (dynamic MT5 layout). */
export function usesOvernightScale(windows) {
  const list = windows?.length ? windows : [];
  if (!list.length) return false;
  const classified = classifyWindows(list);
  return (
    classified.some((w) => w.zone === "extension") ||
    classified.some((w) => w.zone === "primary" && w.close === PRIMARY_MIN)
  );
}

export function scaleTotalMin(windows) {
  return usesOvernightScale(windows) ? TOTAL_SCALE_MIN : PRIMARY_MIN;
}

export function displayPct(displayMinute, windows = null) {
  const total =
    windows && windows.length !== undefined
      ? scaleTotalMin(windows)
      : TOTAL_SCALE_MIN;
  return `${(displayMinute / total) * 100}%`;
}

export function gridDisplayMinutes(windows) {
  if (!usesOvernightScale(windows)) {
    return PRIMARY_SCALE_HOURS.map((h) => (h === 24 ? PRIMARY_MIN : h * 60));
  }
  return GRID_DISPLAY_MINUTES;
}

export function primaryScaleLeft(hour, windows = null) {
  const m = hour === 24 ? PRIMARY_MIN : hour * 60;
  return displayPct(m, windows);
}

export function extendScaleLeft(hour, windows = null) {
  const storage = Math.min(hour * 60, EXTEND_MIN);
  return displayPct(PRIMARY_MIN + storage, windows);
}

export function tailScaleLeft(hour, windows = null) {
  const m = DISPLAY_MIN + (hour - 12) * 60;
  return displayPct(m, windows);
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
    return hourInWindows(hour, primaryWins) ? "sym-timeline-tick" : "sym-timeline-tick off";
  }
  if (scaleZone === "extend") {
    const storageStart = Math.min(hour * 60, EXTEND_MIN);
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

/** Last editable display minute (end of extension zone). */
export function maxEditableDisplayMinute() {
  return DISPLAY_MIN;
}

/**
 * Map a mouse position on the track to storage minute + zone.
 * Editable range: primary (0–1440) and extension (1440+0 – 1440+720 display).
 * Returns null in the gray tail preview zone (not editable in MT5).
 * Pass clampTail=true while dragging to pin at the extension limit instead of ignoring.
 */
export function minuteFromTrack(clientX, trackEl, shiftKey = false, clampTail = false, windows = []) {
  const overnight = usesOvernightScale(windows);
  const total = scaleTotalMin(windows);
  const rect = trackEl.getBoundingClientRect();
  let displayM = ((clientX - rect.left) / rect.width) * total;

  if (!overnight) {
    displayM = Math.round(Math.max(0, Math.min(PRIMARY_MIN, displayM)));
    if (shiftKey) displayM = Math.round(displayM / 5) * 5;
    return { minute: displayM, zone: "primary", displayMinute: displayM };
  }

  if (displayM > DISPLAY_MIN) {
    if (!clampTail) return null;
    displayM = DISPLAY_MIN;
  }
  displayM = Math.round(Math.max(0, Math.min(DISPLAY_MIN, displayM)));
  if (shiftKey) displayM = Math.round(displayM / 5) * 5;

  if (displayM >= PRIMARY_MIN) {
    return { minute: displayM - PRIMARY_MIN, zone: "extension", displayMinute: displayM };
  }
  return { minute: displayM, zone: "primary", displayMinute: displayM };
}

export function canAddWindow(windows, open, close, zone = "primary") {
  const lo = Math.min(open, close);
  const hi = Math.max(open, close);
  if (hi - lo < 1) return false;
  return !sessionOverlaps(windows, lo, hi, -1, zone);
}

function sortWindows(windows) {
  return [...windows].sort((a, b) => a.open - b.open);
}

/** Merge overlapping or touching windows within each zone (MT5 auto-correct on release). */
export function mergeSessionWindows(windows) {
  if (!windows.length) return windows;

  const fold = (list) => {
    if (!list.length) return [];
    const sorted = [...list].sort((a, b) => a.open - b.open);
    const out = [{ ...sorted[0] }];
    for (let i = 1; i < sorted.length; i++) {
      const w = sorted[i];
      const last = out[out.length - 1];
      if (w.open <= last.close) {
        last.close = Math.max(last.close, w.close);
      } else {
        out.push({ ...w });
      }
    }
    return out;
  };

  const primary = [];
  const extension = [];
  for (let i = 0; i < windows.length; i++) {
    const w = windows[i];
    const z = classifySessionZone(w, windows, i);
    if (z === "extension") extension.push({ open: w.open, close: w.close });
    else primary.push({ open: w.open, close: w.close });
  }

  return sortWindows([...fold(primary), ...fold(extension)]);
}

/**
 * Drag-create: union with any overlapping/touching same-zone windows (never split).
 * Release merges fragments into one continuous session per connected group.
 */
export function insertSessionWindow(windows, open, close, zone = "primary") {
  const lo = Math.min(open, close);
  const hi = Math.max(open, close);
  if (hi - lo < 1) return windows;

  const keep = [];
  let mergeLo = lo;
  let mergeHi = hi;

  for (let i = 0; i < windows.length; i++) {
    const w = windows[i];
    const wZone = classifySessionZone(w, windows, i);
    if (wZone !== zone) {
      keep.push({ ...w });
      continue;
    }
    if (w.open <= mergeHi && w.close >= mergeLo) {
      mergeLo = Math.min(mergeLo, w.open);
      mergeHi = Math.max(mergeHi, w.close);
      continue;
    }
    keep.push({ ...w });
  }
  keep.push({ open: mergeLo, close: mergeHi });
  return mergeSessionWindows(keep);
}

function stripExtensionWindows(windows) {
  return windows.filter((w, i) => classifySessionZone(w, windows, i) !== "extension");
}

function findExtensionIndex(windows) {
  return windows.findIndex((w, i) => classifySessionZone(w, windows, i) === "extension");
}

/** Apply close-marker drag, including drag past 24:00 into the extension zone. */
export function applyCloseDrag(windows, index, trackPos) {
  const classified = classifyWindows(windows);
  const zone = classified[index]?.zone ?? "primary";
  const w = windows[index];

  if (trackPos.zone === "extension") {
    const extClose = Math.max(1, Math.min(EXTEND_MIN, trackPos.minute));
    if (zone === "primary") {
      const base = stripExtensionWindows(windows.filter((_, i) => i !== index));
      const primary = { open: w.open, close: PRIMARY_MIN };
      const extension = { open: 0, close: extClose };
      if (sessionOverlaps(base, primary.open, primary.close, -1, "primary")) return windows;
      if (sessionOverlaps(base, extension.open, extension.close, -1, "extension")) return windows;
      return sortWindows([...base, primary, extension]);
    }
    const extIdx = zone === "extension" ? index : findExtensionIndex(windows);
    if (extIdx < 0) return windows;
    const next = windows.map((win, i) => (i === extIdx ? { ...win, close: extClose } : { ...win }));
    if (sessionOverlaps(next, next[extIdx].open, extClose, extIdx, "extension")) return windows;
    return next;
  }

  if (zone === "extension") return windows;

  const close = Math.max(w.open + 1, Math.min(PRIMARY_MIN, trackPos.minute));
  let next = windows.map((win, i) => (i === index ? { ...win, close } : { ...win }));
  if (close < PRIMARY_MIN) {
    next = stripExtensionWindows(next);
  }
  if (sessionOverlaps(next, next[index].open, close, index, "primary")) return windows;
  return next;
}

/** Apply open-marker drag within primary or extension zones. */
export function applyOpenDrag(windows, index, trackPos) {
  const classified = classifyWindows(windows);
  const zone = classified[index]?.zone ?? "primary";
  const w = windows[index];

  if (zone === "extension") {
    if (trackPos.zone !== "extension") return windows;
    let open = Math.max(0, Math.min(trackPos.minute, EXTEND_MIN - 1));
    let close = w.close;
    if (open >= close) close = Math.min(EXTEND_MIN, open + 1);
    const next = windows.map((win, i) => (i === index ? { ...win, open, close } : { ...win }));
    if (sessionOverlaps(next, open, close, index, "extension")) return windows;
    return next;
  }

  if (trackPos.zone === "extension") return windows;
  const open = Math.max(0, Math.min(trackPos.minute, w.close - 1));
  const next = windows.map((win, i) => (i === index ? { ...win, open } : { ...win }));
  if (sessionOverlaps(next, open, w.close, index, "primary")) return windows;
  return next;
}

/** Clamp marker drag for a classified window (keyboard nudge). */
export function clampMarkerMinute(minute, edge, window, zone) {
  const w = { ...window };
  if (zone === "extension") {
    if (edge === "open") {
      w.open = Math.max(0, Math.min(minute, EXTEND_MIN - 1));
      if (w.open >= w.close) w.close = Math.min(EXTEND_MIN, w.open + 1);
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

/** Apply keyboard nudge to one marker on a classified window list. */
export function nudgeSessionMarker(windows, index, edge, delta) {
  const classified = classifyWindows(windows);
  const zone = classified[index]?.zone ?? "primary";
  const w = windows[index];
  const raw = edge === "open" ? w.open + delta : w.close + delta;
  const minute =
    zone === "extension"
      ? Math.max(0, Math.min(EXTEND_MIN, raw))
      : Math.max(0, Math.min(PRIMARY_MIN, raw));
  const trackPos = {
    minute,
    zone: zone === "extension" ? "extension" : "primary",
    displayMinute: zone === "extension" ? PRIMARY_MIN + minute : minute,
  };
  if (edge === "close") return mergeSessionWindows(applyCloseDrag(windows, index, trackPos));
  return mergeSessionWindows(applyOpenDrag(windows, index, trackPos));
}
