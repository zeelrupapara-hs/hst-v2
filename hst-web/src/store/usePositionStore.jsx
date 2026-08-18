import { create } from "zustand";
import {
  getAccountDetails,
  getAllOrders,
  getAllPositions,
} from "../api/request/activity";

const usePositionStore = create((set, get) => ({
  positions: [],
  positionPL: {},
  summary: {},
  // how many decimals money is shown to, decided by the account's group and changeable by an admin
  currencyDigits: 2,
  orders: [],
  loading: false,
  error: null,

  fetchPositions: async () => {
    const { fetchActiveOrders, fetchSummary, fetchPendingOrders } = get();

    try {
      set({ loading: true });
      await Promise.all([
        fetchActiveOrders(),
        fetchSummary(),
        fetchPendingOrders(),
      ]);
    } catch (error) {
      set({ error: error });
    } finally {
      set({ loading: false });
    }
  },

  fetchActiveOrders: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllPositions({ active: true });
      if (!data?.data?.length) return;

      const positionPL = data?.data?.reduce((acc, position) => {
        acc[position.id] = position.profit;
        return acc;
      }, {});

      set({ positions: data?.data, positionPL });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  fetchSummary: async () => {
    set({ loading: true });
    try {
      const { data } = await getAccountDetails();
      if (!data?.data) return;

      set({
        summary: data?.data,
        currencyDigits: data?.data?.currency_digits ?? 2,
      });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  fetchPendingOrders: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllOrders({ active: true });
      if (!data?.data?.length) return;

      set({ orders: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  getPositionsByType: (type = "all") => {
    const { positions, positionPL } = get();
    return positions.filter((position) => {
      const pl = Number(positionPL[position.id]);

      if (type === "profit") return pl > 0;
      if (type === "loss") return pl < 0;
      return true; // for "all"
    });
  },

  addPosition: (position) => {
    set((state) => ({ positions: [...state.positions, position] }));
  },

  updatePosition: (id, data) => {
    set((state) => ({
      positions: state.positions.map((position) =>
        position?.id === id ? data : position
      ),
    }));
  },

  closePosition: (id) => {
    set((state) => ({
      positions: state.positions.filter((position) => position?.id !== id),
    }));
  },

  updatePositionPL: (id, profitLoss) => {
    set((state) => ({ positionPL: { ...state.positionPL, [id]: profitLoss } }));
  },

  // merged, not replaced: the live frame carries the money only, and would otherwise drop the
  // fields the http account call supplied
  updateSummary: (summary) =>
    set((state) => ({ summary: { ...state.summary, ...summary } })),

  setCurrencyDigits: (digits) => {
    const d = Number(digits);
    set({ currencyDigits: Number.isInteger(d) && d >= 0 ? d : 2 });
  },

  // upsert: the engine re-announces an order on order_create for every change it makes, so a
  // modified order would otherwise be listed twice
  addOrder: (order) => {
    set((state) => ({
      orders: state.orders.some((o) => o?.id === order?.id)
        ? state.orders.map((o) => (o?.id === order?.id ? order : o))
        : [...state.orders, order],
    }));
  },

  updateOrder: (id, data) => {
    set((state) => ({
      orders: state.orders.map((order) => (order?.id === id ? data : order)),
    }));
  },

  removeOrder: (id) => {
    set((state) => ({
      orders: state.orders.filter((order) => order?.id !== id),
    }));
  },
}));

export default usePositionStore;
