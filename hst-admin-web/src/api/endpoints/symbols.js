import { request } from "@/api/client.js";

export const fetchSymbols = ({ search = "", folder = "", page = 0, limit = 500 } = {}) => {
  const q = new URLSearchParams({ limit: String(limit), page: String(page) });
  if (search) q.set("search", search);
  if (folder) q.set("folder", folder);
  return request(`/api/v1/symbols?${q}`);
};

export const fetchSymbol = (id) => request(`/api/v1/symbols/${id}`);

export const createSymbol = (body) =>
  request("/api/v1/symbols", { method: "POST", body });

export const updateSymbol = (id, patch) =>
  request(`/api/v1/symbols/${id}`, { method: "PATCH", body: patch });

export const deleteSymbol = (id) =>
  request(`/api/v1/symbols/${id}`, { method: "DELETE" });

export const cloneSymbols = ({ postfix, path = "", copy_to = "", symbols = [] }) =>
  request("/api/v1/symbols/clone", { method: "POST", body: { postfix, path, copy_to, symbols } });

export const fetchSymbolFolders = () => request("/api/v1/symbols/folders");

export const createSymbolFolder = (parent, name) =>
  request("/api/v1/symbols/folders", { method: "POST", body: { parent, name } });

export const renameSymbolFolder = (from, to) =>
  request("/api/v1/symbols/folders", { method: "PATCH", body: { from, to } });

export const deleteSymbolFolder = (path, cascade = false) =>
  request("/api/v1/symbols/folders", { method: "DELETE", body: { path, cascade } });
