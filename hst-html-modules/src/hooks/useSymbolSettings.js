import { useCallback, useEffect, useState } from "react";
import { createSymbol, fetchSymbol, saveSymbol } from "../lib/data.js";
import { mergeDaySessions } from "../lib/symbolSessions.js";
import { statusMessage } from "../lib/formatters.js";
import { isNewSymbolId, newSymbolDraft } from "../lib/symbolDraft.js";
import { symbolFolder } from "../lib/symbolTree.js";

export function useSymbolSettings(symbolId, folderPath = "") {
  const isNew = isNewSymbolId(symbolId);
  const [symbol, setSymbol] = useState(null);
  const [apiState, setApiState] = useState({
    loading: !isNew,
    demo: true,
    message: isNew ? "New symbol" : "…",
  });
  const [toast, setToast] = useState(null);

  const load = useCallback(async () => {
    if (!symbolId) return null;
    if (isNew) {
      const draft = newSymbolDraft(folderPath);
      setSymbol(draft);
      setApiState({ loading: false, demo: true, message: "New symbol" });
      return draft;
    }
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
  }, [symbolId, folderPath, isNew]);

  useEffect(() => {
    load();
  }, [load]);

  const updateField = useCallback((key, value) => {
    setSymbol((prev) => {
      if (!prev) return prev;
      const next = { ...prev, [key]: value };
      if (key === "symbol") {
        const folder = folderPath || symbolFolder(prev.path || "");
        next.path = folder ? `${folder}\\${value}` : value;
      }
      return next;
    });
  }, [folderPath]);

  const persist = useCallback(
    async (id, patch) => {
      const res = await saveSymbol(id, patch);
      if (res.data) setSymbol(res.data);
      return res.data;
    },
    []
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
      if (isNew) {
        setSymbol((s) => ({ ...s, sessions: merged }));
        return;
      }
      await persist(symbolId, { sessions: merged });
    },
    [symbol, isNew, persist, symbolId]
  );

  const updateTimeLimits = useCallback(
    async (enabled, timeStart, timeExpiration) => {
      const patch = {
        time_start: enabled ? timeStart : 0,
        time_expiration: enabled ? timeExpiration : 0,
      };
      if (isNew) {
        setSymbol((s) => ({ ...s, ...patch }));
        return;
      }
      await persist(symbolId, patch);
    },
    [isNew, persist, symbolId]
  );

  const saveAll = useCallback(async () => {
    if (!symbol) return null;
    if (isNew) {
      const name = String(symbol.symbol || "").trim();
      if (!name) {
        setToast("Symbol name is required");
        setTimeout(() => setToast(null), 3000);
        return null;
      }
      const folder = folderPath || symbolFolder(symbol.path || "");
      const created = await createSymbol(
        folder,
        name,
        symbol.description || name
      );
      if (!created.data) return null;
      const saved = await persist(created.data.symbol_id, {
        ...symbol,
        symbol: name,
        path: created.data.path,
      });
      setToast("Symbol created");
      setTimeout(() => setToast(null), 3000);
      return saved;
    }
    await persist(symbolId, symbol);
    setToast("Symbol settings saved");
    setTimeout(() => setToast(null), 3000);
    return symbol;
  }, [symbol, isNew, folderPath, persist, symbolId]);

  return {
    symbol,
    isNew,
    apiState,
    toast,
    reload: load,
    updateField,
    updateSessions,
    updateTimeLimits,
    saveAll,
  };
}
