import { useCallback, useEffect, useState } from "react";
import {
  createDatafeed,
  fetchDatafeed,
  fetchDatafeedModules,
  saveDatafeed,
} from "../lib/data.js";
import { FEEDER_MODE_QUOTES, isNewsMode } from "../lib/datafeedModules.js";
import { statusMessage } from "../lib/formatters.js";
import { isNewDatafeedId, newDatafeedDraft } from "../lib/datafeedDraft.js";

function pickModuleForMode(modules, current) {
  if (!modules?.length) return current || "";
  if (modules.some((m) => m.module === current)) return current;
  return modules[0].module;
}

export function useDatafeedSettings(datafeedId) {
  const isNew = isNewDatafeedId(datafeedId);
  const [datafeed, setDatafeed] = useState(null);
  const [modules, setModules] = useState([]);
  const [apiState, setApiState] = useState({
    loading: !isNew,
    demo: true,
    message: isNew ? "New data feed" : "…",
  });
  const [toast, setToast] = useState(null);

  const loadModules = useCallback(async (mode) => {
    const res = await fetchDatafeedModules(mode ?? FEEDER_MODE_QUOTES);
    setModules(res.data || []);
    return res.data || [];
  }, []);

  const load = useCallback(async () => {
    if (!datafeedId) return null;
    if (isNew) {
      const draft = newDatafeedDraft();
      setDatafeed(draft);
      await loadModules(draft.mode);
      setApiState({ loading: false, demo: true, message: "New data feed" });
      return draft;
    }
    setApiState({ loading: true, demo: true, message: "Loading…" });
    const res = await fetchDatafeed(datafeedId);
    const row = res.data;
    setDatafeed(row);
    if (row?.mode != null) {
      const opts = await loadModules(row.mode);
      if (row.module && !opts.some((m) => m.module === row.module)) {
        setModules((prev) => [
          ...prev,
          { module: row.module, label: row.module, description: "" },
        ]);
      }
    }
    setApiState({
      loading: false,
      demo: res.demo,
      message:
        statusMessage("Data feed", res, 1) +
        (res.fallback ? " (API fallback)" : ""),
    });
    return row;
  }, [datafeedId, isNew, loadModules]);

  useEffect(() => {
    load();
  }, [load]);

  const updateField = useCallback((key, value) => {
    setDatafeed((prev) => (prev ? { ...prev, [key]: value } : prev));
  }, []);

  const changeMode = useCallback(
    async (mode) => {
      const nextMode = Number(mode) || FEEDER_MODE_QUOTES;
      const opts = await loadModules(nextMode);
      setDatafeed((prev) => {
        if (!prev) return prev;
        const next = {
          ...prev,
          mode: nextMode,
          module: pickModuleForMode(opts, prev.module),
        };
        if (isNewsMode(nextMode)) {
          next.feed_symbols = [];
          next.translates = [];
          next.allow_import_symbols = 0;
        }
        return next;
      });
    },
    [loadModules]
  );

  const saveAll = useCallback(async () => {
    if (!datafeed) return null;
    if (isNew) {
      const name = String(datafeed.name || "").trim();
      if (!name) {
        setToast("Name is required");
        setTimeout(() => setToast(null), 3000);
        return null;
      }
      if (!datafeed.module) {
        setToast("Module is required");
        setTimeout(() => setToast(null), 3000);
        return null;
      }
      const created = await createDatafeed(datafeed);
      if (!created.data) return null;
      const saved = await saveDatafeed(created.data.datafeed_id, datafeed);
      setToast("Data feed created");
      setTimeout(() => setToast(null), 3000);
      return saved.data;
    }
    const saved = await saveDatafeed(datafeedId, datafeed);
    setToast("Data feed settings saved");
    setTimeout(() => setToast(null), 3000);
    return saved.data;
  }, [datafeed, isNew, datafeedId]);

  return {
    datafeed,
    modules,
    isNew,
    apiState,
    toast,
    reload: load,
    updateField,
    changeMode,
    saveAll,
  };
}
