import { useEffect, useState } from "react";
import { fetchSymbols } from "@/api/endpoints/symbols.js";

// One shared list: the navigator folder tree and the symbols table read the same fetch.
let cache = null;
let inflight = null;
const listeners = new Set();

async function loadAll() {
  const all = [];
  let page = 1;
  for (;;) {
    const res = await fetchSymbols({ page, limit: 500 });
    if (!res.ok) return cache || [];
    const batch = res.data || [];
    all.push(...batch);
    if (batch.length < 500) break;
    page += 1;
  }
  return all;
}

async function load() {
  inflight ??= loadAll().then((rows) => {
    cache = rows;
    inflight = null;
    listeners.forEach((fn) => fn(cache));
    return cache;
  });
  return inflight;
}

export function reloadSymbols() {
  cache = null;
  return load();
}

// The server pushes the full symbol -> is-price-flowing map every sweep; null until the first,
// which reads as nothing being live rather than everything.
let liveness = null;
const livenessListeners = new Set();

export function applySymbolLiveness(map) {
  liveness = map || {};
  livenessListeners.forEach((fn) => fn(liveness));
}

/**
 * Nothing counts as live until a sweep says so: the yellow coin means prices are arriving now,
 * so assuming it before the first sweep flashes a live feed that may not exist.
 * @returns {{isLive: (symbol: string) => boolean}}
 */
export function useSymbolLiveness() {
  const [map, setMap] = useState(liveness);

  useEffect(() => {
    const fn = (m) => setMap({ ...m });
    livenessListeners.add(fn);
    return () => livenessListeners.delete(fn);
  }, []);

  return { isLive: (symbol) => map?.[symbol] === true };
}

/** @returns {{symbols: Array, loading: boolean, reload: () => Promise<Array>}} */
export function useSymbols() {
  const [symbols, setSymbols] = useState(cache || []);
  const [loading, setLoading] = useState(cache === null);

  useEffect(() => {
    const fn = (rows) => {
      setSymbols(rows);
      setLoading(false);
    };
    listeners.add(fn);
    if (cache === null) load();
    else setSymbols(cache);
    return () => listeners.delete(fn);
  }, []);

  return { symbols, loading, reload: reloadSymbols };
}
