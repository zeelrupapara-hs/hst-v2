import api from "../api";
import { adaptWatchlist, mapList, mapOne } from "../adapt";

export const getWatchlists = () =>
  api.get("/watchlists").then((r) => mapList(r, adaptWatchlist));

export const createWatchlist = (name = "Favourites") =>
  api.post("/watchlists", { name, kind: 0 }).then((r) => mapOne(r, adaptWatchlist));

export const updateWatchlist = (id, name) =>
  api.put(`/watchlists/${id}`, { name }).then((r) => mapOne(r, adaptWatchlist));

export const deleteWatchlist = (id) => api.delete(`/watchlists/${id}`);

export const setWatchlistSymbols = (id, symbols) =>
  api.put(`/watchlists/${id}/symbols`, { symbols }).then((r) => mapOne(r, adaptWatchlist));
