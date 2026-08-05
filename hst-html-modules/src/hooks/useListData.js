import { useCallback, useEffect, useState } from "react";
import { statusMessage } from "../lib/formatters.js";
import { fetchList, normalizeRows } from "../lib/data.js";

export function useListData(config, search) {
  const [rows, setRows] = useState([]);
  const [apiState, setApiState] = useState({ loading: true, demo: true, message: "…" });

  const load = useCallback(async () => {
    setApiState({ loading: true, demo: true, message: "Loading…" });
    const q = { ...(config.query || {}) };
    if (search.trim()) q.search = search.trim();

    const res = await fetchList({ endpoint: config.endpoint, query: q });
    let nextRows = normalizeRows(res.data, config.transform);
    if (config.filterEnabledOnly) {
      nextRows = nextRows.filter(
        (r) => r.enable !== false && r.enabled !== false && r.status !== "disabled"
      );
    }
    setRows(nextRows);
    setApiState({
      loading: false,
      demo: res.demo,
      message:
        statusMessage(config.title, res, nextRows.length) +
        (res.fallback ? " (API fallback)" : ""),
    });
    return nextRows;
  }, [config, search]);

  useEffect(() => {
    load();
  }, [load]);

  return { rows, apiState, reload: load };
}
