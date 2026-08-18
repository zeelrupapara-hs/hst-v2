import { create } from "zustand";
import { persist } from "zustand/middleware";

const useGlobalStore = create(
  persist(
    (set, get) => ({
      symbol: null,
      panels: {
        marketWatch: true,
        chart: true,
        order: false,
        activity: true,
      },
      visibleColumns: ["symbol", "bid", "ask", "spread"],
      chartLayout: 1,
      chartSymbols: {},
      isDrag: false,
      oneClick: false,
      symbolOneClick: false,
      volume: {},
      orderModals: [],
      // the Market Watch selection: a watchlist id, or "all" for the whole symbol list
      activeWatchlistId: null,

      setGlobalStore: (data) => set(data),

      togglePanel: (panelName, value) => {
        set((state) => ({
          panels: {
            ...state.panels,
            [panelName]: value || !state.panels[panelName],
          },
        }));
      },

      setChartLayout: (chartLayout) => {
        set({ chartLayout, oneClick: chartLayout === 1 });
      },

      setChartSymbol: (chartId, symbolId) => {
        if (!(chartId || symbolId)) return;

        const { chartSymbols, chartLayout } = get();

        const liveCharts = Object.values(chartSymbols ?? {})?.slice(0, chartLayout);
        if (liveCharts?.includes(String(symbolId))) return;

        set((state) => ({
          chartSymbols: { ...(state.chartSymbols ?? {}), [chartId]: String(symbolId) },
        }));
      },

      setSymbolVolume: (key, value) => {
        set((state) => ({ volume: { ...state.volume, [key]: value } }));
      },

      openOrderModal: (symbolId) => {
        set((state) => ({ orderModals: [...state.orderModals, symbolId] }));
      },

      closeOrderModal: (symbolId) => {
        set((state) => ({
          orderModals: state.orderModals.filter((id) => id !== symbolId),
        }));
      },
    }),
    { name: "global-store" } // localStorage key
  )
);

export default useGlobalStore;
