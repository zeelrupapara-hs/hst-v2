import { request } from "@/api/client.js";

export const fetchDatafeeds = () => request("/api/v1/datafeeds");
export const fetchDatafeed = (id) => request(`/api/v1/datafeeds/${id}`);
export const fetchDatafeedModules = (mode) => request(`/api/v1/datafeeds/modules?mode=${mode}`);
export const createDatafeed = (body) => request("/api/v1/datafeeds", { method: "POST", body });
export const updateDatafeed = (id, patch) =>
  request(`/api/v1/datafeeds/${id}`, { method: "PATCH", body: patch });
export const deleteDatafeed = (id) => request(`/api/v1/datafeeds/${id}`, { method: "DELETE" });
export const activateDatafeed = (id) =>
  request(`/api/v1/datafeeds/${id}/activate`, { method: "POST" });

export const createDatafeedSymbol = (id, body) =>
  request(`/api/v1/datafeeds/${id}/symbols`, { method: "POST", body });
export const deleteDatafeedSymbol = (id, rowId) =>
  request(`/api/v1/datafeeds/${id}/symbols/${rowId}`, { method: "DELETE" });
export const resolveDatafeedSymbols = (id) =>
  request(`/api/v1/datafeeds/${id}/symbols/resolve`);

export const createDatafeedParam = (id, body) =>
  request(`/api/v1/datafeeds/${id}/params`, { method: "POST", body });
export const updateDatafeedParam = (id, rowId, patch) =>
  request(`/api/v1/datafeeds/${id}/params/${rowId}`, { method: "PATCH", body: patch });
export const deleteDatafeedParam = (id, rowId) =>
  request(`/api/v1/datafeeds/${id}/params/${rowId}`, { method: "DELETE" });

export const createDatafeedTranslate = (id, body) =>
  request(`/api/v1/datafeeds/${id}/translates`, { method: "POST", body });
export const updateDatafeedTranslate = (id, rowId, patch) =>
  request(`/api/v1/datafeeds/${id}/translates/${rowId}`, { method: "PATCH", body: patch });
export const deleteDatafeedTranslate = (id, rowId) =>
  request(`/api/v1/datafeeds/${id}/translates/${rowId}`, { method: "DELETE" });

export const reorderDatafeeds = (datafeedIds) =>
  request("/api/v1/datafeeds/order", { method: "PUT", body: { datafeed_ids: datafeedIds } });
