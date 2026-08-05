import { useEffect, useState } from "react";
import { fetchGroups } from "@/api/endpoints/groups.js";

// Shared flat group list: the navigator folder tree and the groups table read the same fetch.
let cache = null;
let inflight = null;
const listeners = new Set();

async function load() {
  inflight ??= fetchGroups().then((res) => {
    cache = res.ok ? res.data || [] : cache || [];
    inflight = null;
    listeners.forEach((fn) => fn(cache));
    return cache;
  });
  return inflight;
}

export function reloadGroups() {
  cache = null;
  return load();
}

/** @returns {{groups: Array, loading: boolean, reload: () => Promise<Array>}} */
export function useGroups() {
  const [groups, setGroups] = useState(cache || []);
  const [loading, setLoading] = useState(cache === null);

  useEffect(() => {
    const fn = (rows) => {
      setGroups(rows);
      setLoading(false);
    };
    listeners.add(fn);
    if (cache === null) load();
    else setGroups(cache);
    return () => listeners.delete(fn);
  }, []);

  return { groups, loading, reload: reloadGroups };
}
