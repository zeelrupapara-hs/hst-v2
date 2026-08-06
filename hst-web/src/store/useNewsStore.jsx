import { create } from "zustand";
import { getAllNews } from "../api/request/activity";

const useNewsStore = create((set) => ({
  news: [],
  loading: false,
  error: null,

  fetchNews: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllNews();
      set({ news: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },
}));

export default useNewsStore;
