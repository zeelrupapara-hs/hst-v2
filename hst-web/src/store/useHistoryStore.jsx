import { create } from "zustand";
import {
  getAllDeals,
  getAllOrders,
  getAllClosedPositions,
} from "../api/request/activity";

const useHistoryStore = create((set, get) => ({
  positions: [],
  orders: [],
  deals: [],
  loading: false,
  error: null,

  fetchHistory: async () => {
    const { fetchPositions, fetchOrders, fetchDeals } = get();

    try {
      set({ loading: true });
      await Promise.all([fetchPositions(), fetchOrders(), fetchDeals()]);
    } catch (error) {
      set({ error: error });
    } finally {
      set({ loading: false });
    }
  },

  fetchPositions: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllClosedPositions();
      set({ positions: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  fetchOrders: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllOrders({ active: false });
      set({ orders: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  fetchDeals: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllDeals();
      set({ deals: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  addPosition: (position) => {
    set((state) => ({ positions: [...state.positions, position] }));
  },

  addOrder: (order) => {
    set((state) => ({
      orders: [...state.orders, order],
    }));
  },

  updateOrder: (id, data) => {
    set((state) => ({
      orders: state.orders.map((order) => (order?.id === id ? data : order)),
    }));
  },

  addDeal: (deal) => {
    set((state) => ({
      deals: [...state.deals, deal],
    }));
  },
}));

export default useHistoryStore;
