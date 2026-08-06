import { useEffect, useState } from "react";
import { fetchSymbolFolders } from "@/api/endpoints/symbols.js";

let cache = null;
let inflight = null;
const listeners = new Set();

async function load() {
  inflight ??= fetchSymbolFolders().then((res) => {
    cache = res.ok ? (res.data || []).map((r) => r.path) : [];
    inflight = null;
    listeners.forEach((fn) => fn(cache));
    return cache;
  });
  return inflight;
}

export function reloadSymbolFolders() {
  cache = null;
  return load();
}

/** @returns {{folders: string[], loading: boolean, reload: () => Promise<string[]>}} */
export function useSymbolFolders() {
  const [folders, setFolders] = useState(cache || []);
  const [loading, setLoading] = useState(cache === null);

  useEffect(() => {
    const fn = (paths) => {
      setFolders(paths);
      setLoading(false);
    };
    listeners.add(fn);
    if (cache === null) load();
    else setFolders(cache);
    return () => listeners.delete(fn);
  }, []);

  return { folders, loading, reload: reloadSymbolFolders };
}
