import { create } from "zustand";
import { getAllAlerts } from "../api/request/activity";

const useAlertStore = create((set) => ({
  alerts: [],
  loading: false,
  error: null,

  fetchAlerts: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllAlerts();

      const alerts = data?.data?.map((alert) => {
        const [type, condition, triggerPrice, symbolId] =
          alert?.formula?.split(",");

        return {
          ...alert,
          condition: type + " " + condition,
          trigger_price: triggerPrice,
          symbol_id: symbolId,
        };
      });

      set({ alerts });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  addAlert: (alert) => {
    set((state) => ({ alerts: [...state.alerts, alert] }));
  },

  updateAlert: (id, data) => {
    set((state) => ({
      alerts: state.alerts.map((alert) => (alert?.id === id ? data : alert)),
    }));
  },

  deleteAlert: (id) => {
    set((state) => ({
      alerts: state.alerts.filter((alert) => alert?.id !== id),
    }));
  },
}));

export default useAlertStore;
