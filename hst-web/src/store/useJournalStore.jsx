import { create } from "zustand";
import { getAllJournals } from "../api/request/activity";

const useJournalStore = create((set) => ({
  logs: [],
  loading: false,
  error: null,

  fetchJournals: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllJournals();
      set({ logs: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  addJournal: (journal) => {
    set((state) => ({
      logs: [journal, ...state.logs],
    }));
  },
}));

export default useJournalStore;
