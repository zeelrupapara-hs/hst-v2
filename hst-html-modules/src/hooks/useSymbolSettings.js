import { useCallback, useEffect, useState } from "react";
import { fetchSymbol, saveSymbol } from "../lib/data.js";
import { mergeDaySessions } from "../lib/symbolSessions.js";
import { statusMessage } from "../lib/formatters.js";

export function useSymbolSettings(symbolId) {
  const [symbol, setSymbol] = useState(null);
  const [apiState, setApiState] = useState({
    loading: true,
    demo: true,
    message: "…",
  });
  const [toast, setToast] = useState(null);

  const load = useCallback(async () => {
    if (!symbolId) return null;
    setApiState({ loading: true, demo: true, message: "Loading…" });
    const res = await fetchSymbol(symbolId);
    setSymbol(res.data);
    setApiState({
      loading: false,
      demo: res.demo,
      message:
        statusMessage("Symbol", res, 1) + (res.fallback ? " (API fallback)" : ""),
    });
    return res.data;
  }, [symbolId]);

  useEffect(() => {
    load();
  }, [load]);

  const persist = useCallback(
    async (patch) => {
      const res = await saveSymbol(symbolId, patch);
      if (res.data) setSymbol(res.data);
      return res.data;
    },
    [symbolId]
  );

  const updateSessions = useCallback(
    async (dayIndexes, quoteWins, tradeWins, separate) => {
      if (!symbol) return;
      const merged = mergeDaySessions(
        symbol.sessions || [],
        dayIndexes,
        quoteWins,
        tradeWins,
        separate
      );
      await persist({ sessions: merged });
    },
    [symbol, persist]
  );

  const updateTimeLimits = useCallback(
    async (enabled, timeStart, timeExpiration) => {
      await persist({
        time_start: enabled ? timeStart : 0,
        time_expiration: enabled ? timeExpiration : 0,
      });
    },
    [persist]
  );

  const saveAll = useCallback(async () => {
    if (!symbol) return;
    await persist(symbol);
    setToast("Symbol settings saved");
    setTimeout(() => setToast(null), 3000);
  }, [symbol, persist]);

  return {
    symbol,
    apiState,
    toast,
    reload: load,
    updateSessions,
    updateTimeLimits,
    saveAll,
  };
}
