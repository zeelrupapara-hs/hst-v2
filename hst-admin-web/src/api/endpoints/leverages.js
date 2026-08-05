import { request } from "@/api/client.js";

export const fetchLeverages = () => request("/api/v1/leverage-profiles");
export const fetchLeverage = (id) => request(`/api/v1/leverage-profiles/${id}`);
export const createLeverage = (body) =>
  request("/api/v1/leverage-profiles", { method: "POST", body });
export const updateLeverage = (id, body) =>
  request(`/api/v1/leverage-profiles/${id}`, { method: "PUT", body });
export const deleteLeverage = (id) =>
  request(`/api/v1/leverage-profiles/${id}`, { method: "DELETE" });
