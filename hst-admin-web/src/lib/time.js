// Most timestamps on the wire are unix nanoseconds. Symbol session limits
// (time_start, time_expiration) use unix seconds — see fromUnixSec / formatUnixSec.

/** @param {number} ns unix nanoseconds @returns {Date} */
export const fromNs = (ns) => new Date(ns / 1e6);

/** @param {number} sec unix seconds @returns {Date} */
export const fromUnixSec = (sec) => new Date(sec * 1000);

const pad = (n) => String(n).padStart(2, "0");

/** `2026-08-05 14:03:21`, the journal/table format. */
export function formatNs(ns) {
  if (!ns) return "";
  const d = fromNs(ns);
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  );
}

/** MT5 datetime from unix seconds: `1970.01.01 00:00`. */
export function formatUnixSec(sec) {
  if (!sec) return "";
  const d = fromUnixSec(sec);
  return `${d.getFullYear()}.${pad(d.getMonth() + 1)}.${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** Parse MT5 datetime string to unix seconds, or null when invalid. */
export function parseMt5DateTimeToSec(raw) {
  const s = raw.trim();
  const m = s.match(/^(\d{4})\.(\d{2})\.(\d{2})\s+(\d{2}):(\d{2})$/);
  if (!m) return null;
  const d = new Date(+m[1], +m[2] - 1, +m[3], +m[4], +m[5]);
  if (Number.isNaN(d.getTime())) return null;
  return Math.floor(d.getTime() / 1000);
}
