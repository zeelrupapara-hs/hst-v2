import { create } from "zustand";
import useSymbolStore from "./useSymbolStore";
import useGlobalStore from "./useGlobalStore";
import {
  getWatchlists,
  createWatchlist,
  updateWatchlist,
  deleteWatchlist,
  setWatchlistSymbols,
} from "../api/request/watchlist";

// Symbol toggles are applied locally at once and pushed a moment later in one request, because
// the server replaces a list's whole set on every write. One request in flight at a time; a
// toggle that lands mid-flight marks the set dirty and is pushed when the flight returns.
const SYNC_DELAY = 400;

// writeActive mirrors the active list's optimistic symbols into the lists array, so the
// switcher and the Market Watch read the same truth.
const writeActive = (state, symbols) => ({
  wishlist: symbols,
  watchlists: state.watchlists.map((w) =>
    w.id === state.activeId ? { ...w, symbols } : w
  ),
});

const useWatchlistStore = create((set, get) => ({
  // every list the trader owns; symbols inside are optimistic
  watchlists: [],
  // the id shown in the Market Watch; null while the trader owns no list
  activeId: null,
  // true shows the whole symbol table even though lists exist
  viewAll: false,
  // the active list's ordered symbol names, what the UI trusts
  wishlist: [],
  // per list id, the last set the server acknowledged: the rollback target
  serverSymbols: {},
  loading: false,
  syncing: false,

  _timer: null,
  _pendingSync: false,
  _inflight: false,
  _dirty: false,
  _retry: false,

  fetchWatchlists: async () => {
    try {
      set({ loading: true });
      const { data } = await getWatchlists();
      const lists = data?.data ?? [];

      const persisted = useGlobalStore.getState().activeWatchlistId;
      const active = lists.find((w) => w.id === persisted) ?? lists[0] ?? null;

      set({
        watchlists: lists,
        serverSymbols: Object.fromEntries(lists.map((w) => [w.id, w.symbols])),
        activeId: active?.id ?? null,
        viewAll: persisted === "all" && lists.length > 0,
        wishlist: active?.symbols ?? [],
      });
    } catch (error) {
      console.error("Error loading watchlists:", error);
    } finally {
      set({ loading: false });
    }
  },

  // id is a watchlist id, or "all" for the whole symbol table
  setActive: (id) => {
    // a toggle still waiting on its debounce belongs to the list being left; push it now
    if (get()._pendingSync) {
      clearTimeout(get()._timer);
      get()._sync();
    }

    if (id === "all") {
      set({ viewAll: true });
    } else {
      const list = get().watchlists.find((w) => w.id === id);
      if (!list) return;
      set({ viewAll: false, activeId: id, wishlist: list.symbols });
    }
    useGlobalStore.getState().setGlobalStore({ activeWatchlistId: id });
  },

  createList: async (name) => {
    try {
      const { data } = await createWatchlist(name);
      const list = data?.data;
      if (!list) return null;

      set((state) => ({
        watchlists: [...state.watchlists, list],
        serverSymbols: { ...state.serverSymbols, [list.id]: list.symbols },
      }));
      get().setActive(list.id);
      return list;
    } catch {
      return null; // the api interceptor already toasted the failure
    }
  },

  renameList: async (id, name) => {
    try {
      await updateWatchlist(id, name);
      set((state) => ({
        watchlists: state.watchlists.map((w) => (w.id === id ? { ...w, name } : w)),
      }));
      return true;
    } catch {
      return false;
    }
  },

  deleteList: async (id) => {
    try {
      await deleteWatchlist(id);
    } catch {
      return false;
    }

    const remaining = get().watchlists.filter((w) => w.id !== id);
    const active = get().activeId === id ? remaining[0] ?? null : null;
    set((state) => ({
      watchlists: remaining,
      serverSymbols: Object.fromEntries(
        Object.entries(state.serverSymbols).filter(([key]) => Number(key) !== id)
      ),
      ...(state.activeId === id
        ? { activeId: active?.id ?? null, wishlist: active?.symbols ?? [] }
        : {}),
    }));
    if (useGlobalStore.getState().activeWatchlistId === id) {
      useGlobalStore.getState().setGlobalStore({ activeWatchlistId: active?.id ?? null });
    }
    return true;
  },

  addSymbol: (name) => {
    const { wishlist, _scheduleSync } = get();
    if (!name || wishlist.includes(name)) return;

    set((state) => writeActive(state, [...state.wishlist, name]));
    _scheduleSync();
  },

  removeSymbol: (name) => {
    const { wishlist, _scheduleSync } = get();
    if (!wishlist.includes(name)) return;

    set((state) => writeActive(state, state.wishlist.filter((s) => s !== name)));
    _scheduleSync();
  },

  _scheduleSync: () => {
    clearTimeout(get()._timer);
    set({
      _pendingSync: true,
      _timer: setTimeout(() => get()._sync(), SYNC_DELAY),
    });
  },

  _sync: async (isRetry = false) => {
    set({ _pendingSync: false });
    if (get()._inflight) {
      set({ _dirty: true });
      return;
    }
    set({ _inflight: true, syncing: true });

    // captured before any await, so a list switch mid-flight cannot redirect this write
    let id = get().activeId;
    const sent = get().wishlist;

    try {
      if (id == null) {
        const { data } = await createWatchlist();
        const list = data?.data;
        if (!list) throw new Error("watchlist was not created");
        id = list.id;
        set((state) => ({
          watchlists: [...state.watchlists, { ...list, symbols: sent }],
          serverSymbols: { ...state.serverSymbols, [id]: [] },
          activeId: id,
        }));
        useGlobalStore.getState().setGlobalStore({ activeWatchlistId: id });
      }

      const { data } = await setWatchlistSymbols(id, sent);
      set((state) => ({
        serverSymbols: { ...state.serverSymbols, [id]: data?.data?.symbols ?? sent },
      }));
    } catch (error) {
      if (!isRetry && error?.response?.status === 400) {
        // a name the server no longer knows or grants: keep only what this account can see
        const known = useSymbolStore.getState().symbols;
        set((state) => ({
          ...writeActive(state, state.wishlist.filter((name) => known[name])),
          _retry: true,
        }));
      } else {
        // the api interceptor already toasts the failure; this only takes the optimistic set back
        set((state) => writeActive(state, state.serverSymbols[id] ?? []));
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
