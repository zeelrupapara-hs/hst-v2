import { create } from "zustand";

const useLiveSymbolStore = create((set, get) => ({
  liveSymbols: {},

  setLiveSymbols: (liveSymbols) => {
    set({ liveSymbols });
  },

  updateSymbolPrice: (symbolId, priceData) => {
    set((state) => ({
      liveSymbols: {
        ...state.liveSymbols,
        [symbolId]: {
          ...state.liveSymbols[symbolId],
          ...priceData,
        },
      },
    }));
  },

  getSymbolInfo: (symbolId) => {
    const { liveSymbols } = get();
    return liveSymbols[symbolId] || {};
  },
}));

export default useLiveSymbolStore;
