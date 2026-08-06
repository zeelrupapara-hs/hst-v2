import { create } from "zustand";
import { getAllReports } from "../api/request/activity";

const useReportStore = create((set) => ({
  reports: [],
  loading: false,
  error: null,

  fetchReports: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllReports();
      set({ reports: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  addReport: (report) => {
    set((state) => ({ reports: [...state.reports, report] }));
  },

  updateReport: (id, data) => {
    set((state) => ({
      reports: state.reports.map((report) =>
        report?.id === id ? data : report
      ),
    }));
  },

  deleteReport: (id) => {
    set((state) => ({
      reports: state.reports.filter((report) => report?.id !== id),
    }));
  },
}));

export default useReportStore;
