import { create } from "zustand";
import useSymbolStore from "./useSymbolStore";
import {
  getWatchlists,
  createWatchlist,
  setWatchlistSymbols,
} from "../api/request/watchlist";

// Toggles are applied locally at once and pushed a moment later in one request, because the
// server replaces the whole set on every write. One request in flight at a time; a toggle that
// lands mid-flight marks the set dirty and is pushed when the flight returns.
const SYNC_DELAY = 400;

const useWatchlistStore = create((set, get) => ({
  // null means the account owns no Market Watch list yet, which shows everything. The server
  // seeds the broker's default set on first fetch, so this stays null only when no default is
  // configured or none of it is granted to the account's group.
  watchlistId: null,
  // ordered symbol names, what the UI trusts
  wishlist: [],
  // the last set the server acknowledged, the rollback target
  serverWishlist: [],
  loading: false,
  syncing: false,

  _timer: null,
  _inflight: false,
  _dirty: false,
  _retry: false,

  fetchWatchlist: async () => {
    try {
      set({ loading: true });
      const { data } = await getWatchlists();
      const favourites = (data?.data ?? []).find((w) => w.kind === 0);

      if (favourites) {
        set({
          watchlistId: favourites.id,
          wishlist: favourites.symbols,
          serverWishlist: favourites.symbols,
        });
      }
    } catch (error) {
      console.error("Error loading watchlist:", error);
    } finally {
      set({ loading: false });
    }
  },

  addSymbol: (name) => {
    const { wishlist, _scheduleSync } = get();
    if (!name || wishlist.includes(name)) return;

    set({ wishlist: [...wishlist, name] });
    _scheduleSync();
  },

  removeSymbol: (name) => {
    const { wishlist, _scheduleSync } = get();
    if (!wishlist.includes(name)) return;

    set({ wishlist: wishlist.filter((s) => s !== name) });
    _scheduleSync();
  },

  _scheduleSync: () => {
    clearTimeout(get()._timer);
    set({ _timer: setTimeout(() => get()._sync(), SYNC_DELAY) });
  },

  _sync: async (isRetry = false) => {
    if (get()._inflight) {
      set({ _dirty: true });
      return;
    }
    set({ _inflight: true, syncing: true });

    try {
      let id = get().watchlistId;
      if (id == null) {
        const { data } = await createWatchlist();
        id = data?.data?.id;
        if (!id) throw new Error("watchlist was not created");
        set({ watchlistId: id });
      }

      // the current set, not a snapshot from schedule time, so late toggles are not lost
      const sent = get().wishlist;
      const { data } = await setWatchlistSymbols(id, sent);
      set({ serverWishlist: data?.data?.symbols ?? sent });
    } catch (error) {
      if (!isRetry && error?.response?.status === 400) {
        // a name the server no longer knows or grants: keep only what this account can see
        const known = useSymbolStore.getState().symbols;
        set({ wishlist: get().wishlist.filter((name) => known[name]), _retry: true });
      } else {
        // the api interceptor already toasts the failure; this only takes the optimistic set back
        set({ wishlist: get().serverWishlist });
      }
    } finally {
      set({ _inflight: false, syncing: false });
      const { _dirty, _retry, _sync } = get();
      if (_retry) {
        set({ _retry: false, _dirty: false });
        _sync(true);
      } else if (_dirty) {
        set({ _dirty: false });
        _sync();
      }
    }
  },
}));

export default useWatchlistStore;
