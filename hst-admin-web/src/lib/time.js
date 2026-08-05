// Every timestamp on the wire is unix nanoseconds; this is the one place that knows it.

/** @param {number} ns unix nanoseconds @returns {Date} */
export const fromNs = (ns) => new Date(ns / 1e6);

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
