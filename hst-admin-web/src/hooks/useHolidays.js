import { useEffect, useState } from "react";
import { fetchHolidays } from "@/api/endpoints/holidays.js";
import { onEvent } from "@/api/socket.js";

// Shared holiday list, refreshed by the holiday events every terminal receives.
let cache = null;
let inflight = null;
const listeners = new Set();
let watching = false;

async function load() {
  inflight ??= fetchHolidays().then((res) => {
    cache = res.ok ? res.data || [] : cache || [];
    inflight = null;
    listeners.forEach((fn) => fn(cache));
    return cache;
  });
  return inflight;
}

function watch() {
  if (watching) return;
  watching = true;
  onEvent("*", (event) => {
    if (!/^holiday_/.test(event.type)) return;
    cache = null;
    load();
  });
}

/** @returns {{holidays: Array}} the enabled and disabled calendar rows, live */
export function useHolidays() {
  const [holidays, setHolidays] = useState(cache || []);

  useEffect(() => {
    watch();
    const fn = (rows) => setHolidays(rows);
    listeners.add(fn);
    if (cache === null) load();
    else setHolidays(cache);
    return () => listeners.delete(fn);
  }, []);

  return { holidays };
}

// holidayToday mirrors the engine's Covers: a row whose date and symbol mask match closes the
// day, except inside its from..to work window; several rows may match and their windows add up.
export function holidayToday(holidays, path, symbol, at = new Date()) {
  const rows = (holidays || []).filter((h) => {
    if (h.mode !== 1) return false;
    if (h.year !== 0 && h.year !== at.getFullYear()) return false;
    if (h.month !== at.getMonth() + 1 || h.day !== at.getDate()) return false;
    return maskCovers(h.symbols, path, symbol);
  });
  if (!rows.length) return null;

  const windows = rows.filter((h) => h.from !== 0 || h.to !== 0).map((h) => ({ from: h.from, to: h.to }));
  return { windows, description: rows.map((h) => h.description).filter(Boolean).join("; ") };
}

function maskCovers(masks, path, symbol) {
  if (!masks?.length) return true;
  return masks.some((raw) => {
    const m = (raw || "").trim();
    if (m === "" || m === "*") return true;
    if (m.endsWith("*")) {
      const prefix = m.slice(0, -1);
      return (path || "").startsWith(prefix) || (symbol || "").startsWith(prefix);
    }
    return m === symbol || m === path;
  });
}
