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

// The server pushes the full symbol -> is-price-flowing map every sweep; null until the first.
let liveness = null;
const livenessListeners = new Set();

export function applySymbolLiveness(map) {
  liveness = map || {};
  livenessListeners.forEach((fn) => fn(liveness));
}

/** @returns {{isLive: (symbol: string) => boolean}} everything is live until the first sweep */
export function useSymbolLiveness() {
  const [map, setMap] = useState(liveness);

  useEffect(() => {
    const fn = (m) => setMap({ ...m });
    livenessListeners.add(fn);
    return () => livenessListeners.delete(fn);
  }, []);

  return { isLive: (symbol) => (map ? map[symbol] === true : true) };
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
