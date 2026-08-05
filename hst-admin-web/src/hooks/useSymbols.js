import { useEffect, useState } from "react";
import { fetchSymbols } from "@/api/endpoints/symbols.js";

// One shared list: the navigator folder tree and the symbols table read the same fetch.
let cache = null;
let inflight = null;
const listeners = new Set();

async function load() {
  inflight ??= fetchSymbols().then((res) => {
    cache = res.ok ? res.data || [] : cache || [];
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
