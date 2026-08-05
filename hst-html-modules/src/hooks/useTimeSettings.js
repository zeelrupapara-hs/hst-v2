import { useCallback, useEffect, useState } from "react";
import { fetchTimeSettings, saveTimeSettings } from "../lib/data.js";
import { statusMessage } from "../lib/formatters.js";

/** @typedef {{ time_zone: string, daylight_saving: boolean, sync_servers: string, schedule: boolean[][] }} TimeSettings */

export function useTimeSettings() {
  const [settings, setSettings] = useState(null);
  const [selectedRowId, setSelectedRowId] = useState(null);
  const [apiState, setApiState] = useState({
    loading: true,
    demo: true,
    message: "…",
  });
  const [toast, setToast] = useState(null);

  const load = useCallback(async () => {
    setApiState({ loading: true, demo: true, message: "Loading…" });
    const res = await fetchTimeSettings();
    setSettings(res.data);
    setApiState({
      loading: false,
      demo: res.demo,
      message: statusMessage("Time", res, 1) + (res.fallback ? " (API fallback)" : ""),
    });
    return res.data;
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const persist = useCallback(async (patch, showRestartToast = false) => {
    const res = await saveTimeSettings(patch);
    if (res.data) setSettings(res.data);
    if (showRestartToast) {
      setToast("Changes take effect after server restart");
      setTimeout(() => setToast(null), 4000);
    }
    return res.data;
  }, []);

  const updateField = useCallback(
    async (field, value) => {
      const patch = { [field]: value };
      const restart = field === "time_zone" || field === "daylight_saving";
      await persist(patch, restart);
    },
    [persist]
  );

  const updateDaySchedule = useCallback(
    async (schedule) => {
      await persist({ schedule });
    },
    [persist]
  );

  const importSettings = useCallback(
    async (data) => {
      await persist(data);
      setToast("Settings imported");
      setTimeout(() => setToast(null), 3000);
    },
    [persist]
  );

  return {
    settings,
    selectedRowId,
    setSelectedRowId,
    apiState,
    toast,
    reload: load,
    updateField,
    updateDaySchedule,
    importSettings,
  };
}
