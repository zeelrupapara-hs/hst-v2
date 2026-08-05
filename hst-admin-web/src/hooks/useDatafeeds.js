import { useEffect, useState } from "react";
import { fetchDatafeeds } from "@/api/endpoints/datafeeds.js";

// Shared list: the navigator feed children and the datafeeds table read the same fetch.
let cache = null;
let inflight = null;
const listeners = new Set();

async function load() {
  inflight ??= fetchDatafeeds().then((res) => {
    cache = res.ok ? res.data || [] : cache || [];
    inflight = null;
    listeners.forEach((fn) => fn(cache));
    return cache;
  });
  return inflight;
}

export function reloadDatafeeds() {
  cache = null;
  return load();
}

/** @returns {{datafeeds: Array, loading: boolean, reload: () => Promise<Array>}} */
export function useDatafeeds() {
  const [datafeeds, setDatafeeds] = useState(cache || []);
  const [loading, setLoading] = useState(cache === null);

  useEffect(() => {
    const fn = (rows) => {
      setDatafeeds(rows);
      setLoading(false);
    };
    listeners.add(fn);
    if (cache === null) load();
    else setDatafeeds(cache);
    return () => listeners.delete(fn);
  }, []);

  return { datafeeds, loading, reload: reloadDatafeeds };
}
