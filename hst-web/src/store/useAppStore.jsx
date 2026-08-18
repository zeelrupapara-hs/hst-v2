import { create } from "zustand";
import useSymbolStore from "./useSymbolStore";
import useWatchlistStore from "./useWatchlistStore";
import usePositionStore from "./usePositionStore";
import useHistoryStore from "./useHistoryStore";
import useAlertStore from "./useAlertStore";
import useMailStore from "./useMailStore";
import useReportStore from "./useReportStore";
import useNewsStore from "./useNewsStore";
import useJournalStore from "./useJournalStore";

const useAppStore = create((set) => ({
  loading: false,
  error: null,

  fetchAppData: async () => {
    const { fetchSymbols } = useSymbolStore.getState();
    const { fetchWatchlists } = useWatchlistStore.getState();
    const { fetchPositions } = usePositionStore.getState();
    const { fetchHistory } = useHistoryStore.getState();
    const { fetchAlerts } = useAlertStore.getState();
    const { fetchMails } = useMailStore.getState();
    const { fetchReports } = useReportStore.getState();
    const { fetchNews } = useNewsStore.getState();
    const { fetchJournals } = useJournalStore.getState();

    try {
      set({ loading: true });
      await fetchSymbols();
      await Promise.all([
        fetchWatchlists(),
        fetchPositions(),
        fetchHistory(),
        fetchAlerts(),
        fetchMails(),
        fetchReports(),
        fetchNews(),
        fetchJournals(),
      ]);
    } catch (error) {
      console.error("Error loading app data:", error);
      set({ error: error });
    } finally {
      set({ loading: false });
    }
  },
}));

export default useAppStore;
