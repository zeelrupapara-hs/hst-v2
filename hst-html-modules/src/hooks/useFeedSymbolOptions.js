import { useEffect, useState } from "react";
import { fetchList, fetchSymbolGroups } from "../lib/data.js";
import { buildFeedSymbolOptions } from "../lib/feedSymbolOptions.js";

/** Load symbol/group choices for data feed Symbols and Translations tabs. */
export function useFeedSymbolOptions() {
  const [options, setOptions] = useState([{ value: "*", label: "*" }]);

  useEffect(() => {
    let cancelled = false;

    (async () => {
      const [symRes, groupRes] = await Promise.all([
        fetchList({ endpoint: "/api/v1/symbols", query: { limit: "500" } }),
        fetchSymbolGroups(),
      ]);
      if (cancelled) return;
      setOptions(
        buildFeedSymbolOptions(symRes.data || [], groupRes.data || [])
      );
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  return options;
}
