import { useCallback, useEffect, useMemo, useState } from "react";
import { fetchList, getSymbolSessionsForList } from "../lib/data.js";
import { filterSymbolsByFolder } from "../lib/symbolTree.js";
import { statusMessage } from "../lib/formatters.js";

export function useSymbolsList(config, search, folderPath) {
  const [allRows, setAllRows] = useState([]);
  const [sessionMap, setSessionMap] = useState({});
  const [apiState, setApiState] = useState({
    loading: true,
    demo: true,
    message: "…",
  });

  const load = useCallback(async () => {
    setApiState({ loading: true, demo: true, message: "Loading…" });
    const q = { ...(config.query || {}), limit: "500" };
    if (search.trim()) q.search = search.trim();
    const res = await fetchList({ ...config, query: q });
    const rows = res.data || [];
    setAllRows(rows);
    const sm = {};
    rows.forEach((r) => {
      sm[r.symbol_id] = getSymbolSessionsForList(r.symbol_id);
    });
    setSessionMap(sm);
    setApiState({
      loading: false,
      demo: res.demo,
      message:
        statusMessage(config.title || "Symbols", res, rows.length) +
        (res.fallback ? " (API fallback)" : ""),
    });
    return rows;
  }, [config, search]);

  useEffect(() => {
    load();
  }, [load]);

  const rows = useMemo(() => {
    let list = filterSymbolsByFolder(allRows, folderPath);
    if (search.trim()) {
      const s = search.trim().toLowerCase();
      list = list.filter((r) =>
        JSON.stringify(r).toLowerCase().includes(s)
      );
    }
    return list;
  }, [allRows, folderPath, search]);

  return { rows, allRows, sessionMap, apiState, reload: load };
}
