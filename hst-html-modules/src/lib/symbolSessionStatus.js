import { SESSION_QUOTE, SESSION_TRADE, windowsForDay } from "./symbolSessions.js";

/** @param {Array<{type:number,day:number,open:number,close:number}>} sessions */
function inSessionWindows(sessions, kind, at = new Date()) {
  const day = at.getDay();
  const windows = windowsForDay(sessions, kind, day);
  if (!windows.length) return false;
  const minutes = at.getHours() * 60 + at.getMinutes();
  return windows.some((w) => minutes >= w.open && minutes < w.close);
}

/**
 * @returns {"open"|"quote"|"closed"}
 */
export function sessionDotStatus(sessions, at = new Date()) {
  if (!sessions?.length) return "closed";
  const quote = inSessionWindows(sessions, SESSION_QUOTE, at);
  const trade = inSessionWindows(sessions, SESSION_TRADE, at);
  if (trade) return "open";
  if (quote) return "quote";
  return "closed";
}

export function sessionDotClass(status) {
  if (status === "quote") return "sym-session-quote";
  if (status === "closed") return "sym-session-closed";
  return "sym-session-open";
}
