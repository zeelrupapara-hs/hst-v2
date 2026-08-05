import { request } from "@/api/client.js";

export const fetchSymbols = (search = "") =>
  request(`/api/v1/symbols?limit=500${search ? `&search=${encodeURIComponent(search)}` : ""}`);

export const fetchSymbol = (id) => request(`/api/v1/symbols/${id}`);

export const createSymbol = (body) =>
  request("/api/v1/symbols", { method: "POST", body });

export const updateSymbol = (id, patch) =>
  request(`/api/v1/symbols/${id}`, { method: "PATCH", body: patch });

export const deleteSymbol = (id) =>
  request(`/api/v1/symbols/${id}`, { method: "DELETE" });
