import { create } from "zustand";
import useLiveSymbolStore from "./useLiveSymbolStore";
import useGlobalStore from "./useGlobalStore";
import { getAllSymbols } from "../api/request/symbol";
import { groupsFromSymbols } from "../api/adapt";
import { calculateSpread } from "../utils/calculations";

const useSymbolStore = create((set, get) => ({
  symbols: {},
  symbolGroups: {},
  loading: false,
  error: null,

  fetchSymbols: async () => {
    const { fetchAllSymbols } = get();

    try {
      set({ loading: true });
      await fetchAllSymbols();
    } catch (error) {
      set({ error: error });
    } finally {
      set({ loading: false });
    }
  },

  fetchAllSymbols: async () => {
    const { setLiveSymbols } = useLiveSymbolStore.getState();
    const { symbol, chartSymbols, setGlobalStore } = useGlobalStore.getState();

    set({ loading: true });
    try {
      const response = await getAllSymbols();
      const list = response?.data?.data;

      if (!Array.isArray(list) || !list.length) {
        console.error("[symbols] no instruments returned", response?.data);
        set({ symbols: {}, symbolGroups: {}, error: "No symbols available" });
        return;
      }

      const symbols = list.reduce((acc, symbol) => {
        const { newSpread } = calculateSpread(symbol);
        acc[symbol.id] = { ...symbol, newSpread };
        return acc;
      }, {});

      // The server says which instruments are actually being fed; the rest keep a last price for
      // ever and would open a chart that never moves. Freshest first, so the terminal opens on
      // something live whatever the feed happens to carry.
      const ranked = Object.keys(symbols).sort((a, b) => {
        const q = Number(Boolean(symbols[b]?.has_quote)) - Number(Boolean(symbols[a]?.has_quote));
        return q || (symbols[b]?.time ?? 0) - (symbols[a]?.time ?? 0);
      });
      const [symbol1, symbol2, symbol3, symbol4] = ranked;

      // A chart kept from a previous visit can name an instrument this account no longer reaches,
      // because its group's access changed or its feed stopped. Those are replaced; a chart the
      // trader chose that is still live is left alone.
      const usable = (id) => id && symbols[id]?.has_quote;
      const chosen = chartSymbols
        ? Object.fromEntries(
            Object.entries(chartSymbols).map(([slot, id], i) => [
              slot,
              usable(id) ? id : ranked[i] ?? symbol1,
            ])
          )
        : { 1: symbol1, 2: symbol2, 3: symbol3, 4: symbol4 };

      set({ symbols, symbolGroups: groupsFromSymbols(list) });
      setLiveSymbols(symbols);

      setGlobalStore({
        symbol: usable(symbol) ? symbol : symbol1,
        chartSymbols: chosen,
      });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

}));

export default useSymbolStore;
