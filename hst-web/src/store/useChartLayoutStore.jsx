import { create } from "zustand";
import {
  getChartLayouts,
  saveChartLayout,
  getChartSettings,
  saveChartSettings,
} from "../api/request/charts";
import { getLocalItem } from "../utils/utils";
import useGlobalStore from "./useGlobalStore";

const LAYOUT_SAVE_DELAY = 1500;
const SETTINGS_SAVE_DELAY = 1000;

// One timer per chart slot so the four widgets never clobber each other's saves.
const layoutTimers = new Map();
let settingsTimer = null;
let settingsDirty = false;
let pagehideRegistered = false;

// The main series lives somewhere in the blob's panes; walk them rather than
// trusting a fixed index, since studies can add panes above the price pane.
const blobSymbol = (content) => {
  try {
    for (const chart of content?.charts ?? []) {
      for (const pane of chart?.panes ?? []) {
        for (const source of pane?.sources ?? []) {
          if (source?.type === "MainSeries") return source?.state?.symbol || null;
        }
      }
    }
  } catch {
    // a future library build may reshape the state; the slot then just keeps
    // whatever symbol the local store already had
  }
  return null;
};

// The axios client cannot run during pagehide; fetch keepalive can carry the
// Bearer header where sendBeacon cannot.
const keepalivePut = (path, content) => {
  const token = getLocalItem("token");
  if (!token) return;
  fetch(`${import.meta.env.VITE_BASE_URL}/api/trader/v1${path}`, {
    method: "PUT",
    keepalive: true,
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ content }),
  }).catch(() => {});
};

const useChartLayoutStore = create((set, get) => ({
  // Nothing renders a widget and nothing saves until the server state is in,
  // so a default empty chart can never overwrite a trader's saved one.
  loaded: false,
  layouts: {},
  settings: {},

  fetchChartState: async () => {
    const [rows, settings] = await Promise.all([
      getChartLayouts(),
      getChartSettings(),
    ]);

    const layouts = {};
    for (const row of rows) layouts[row.chart_id] = row.content;

    // On boot the saved blob's symbol wins over the locally persisted slot
    // assignment, the way an MT5 profile restores its own symbols. Direct
    // setGlobalStore: setChartSymbol's duplicate guard would block this sync.
    const { chartSymbols, setGlobalStore } = useGlobalStore.getState();
    const next = { ...(chartSymbols ?? {}) };
    let changed = false;
    for (const [chartId, content] of Object.entries(layouts)) {
      const symbol = blobSymbol(content);
      if (symbol && next[chartId] !== symbol) {
        next[chartId] = symbol;
        changed = true;
      }
    }
    if (changed) setGlobalStore({ chartSymbols: next });

    if (!pagehideRegistered) {
      pagehideRegistered = true;
      window.addEventListener("pagehide", () => {
        const state = get();
        for (const chartId of layoutTimers.keys()) {
          keepalivePut(`/charts/${chartId}`, state.layouts[chartId]);
        }
        layoutTimers.forEach((timer) => clearTimeout(timer));
        layoutTimers.clear();
        if (settingsDirty) {
          keepalivePut("/charts/settings", state.settings);
          settingsDirty = false;
        }
      });
    }

    set({ layouts, settings, loaded: true });
  },

  saveLayoutDebounced: (chartId, content) => {
    if (!get().loaded || !content) return;

    // Kept in memory synchronously so a widget rebuilt mid-debounce restores
    // the newest state, not the last persisted one.
    set((state) => ({ layouts: { ...state.layouts, [chartId]: content } }));

    clearTimeout(layoutTimers.get(chartId));
    layoutTimers.set(
      chartId,
      setTimeout(() => {
        layoutTimers.delete(chartId);
        saveChartLayout(chartId, get().layouts[chartId]).catch(() => {});
      }, LAYOUT_SAVE_DELAY)
    );
  },

  flushLayout: (chartId, content) => {
    if (!get().loaded) return;

    clearTimeout(layoutTimers.get(chartId));
    layoutTimers.delete(chartId);

    if (content) {
      set((state) => ({ layouts: { ...state.layouts, [chartId]: content } }));
    }
    const latest = get().layouts[chartId];
    if (latest) saveChartLayout(chartId, latest).catch(() => {});
  },

  setSetting: (key, value) => {
    set((state) => ({ settings: { ...state.settings, [key]: value } }));
    get().saveSettingsDebounced();
  },

  removeSetting: (key) => {
    set((state) => {
      const settings = { ...state.settings };
      delete settings[key];
      return { settings };
    });
    get().saveSettingsDebounced();
  },

  // The four widgets share one settings doc; the whole doc is written from the
  // single in-memory copy, so interleaved setValue calls converge to one PUT.
  saveSettingsDebounced: () => {
    if (!get().loaded) return;

    settingsDirty = true;
    clearTimeout(settingsTimer);
    settingsTimer = setTimeout(() => {
      settingsDirty = false;
      saveChartSettings(get().settings).catch(() => {});
    }, SETTINGS_SAVE_DELAY);
  },
}));

export default useChartLayoutStore;
