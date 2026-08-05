import { request } from "@/api/client.js";

export const fetchHolidays = () => request("/api/v1/holidays");
export const createHoliday = (body) => request("/api/v1/holidays", { method: "POST", body });
export const updateHoliday = (id, patch) =>
  request(`/api/v1/holidays/${id}`, { method: "PATCH", body: patch });
export const deleteHoliday = (id) => request(`/api/v1/holidays/${id}`, { method: "DELETE" });
