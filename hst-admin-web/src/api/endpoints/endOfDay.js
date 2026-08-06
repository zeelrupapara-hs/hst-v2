import { request } from "@/api/client.js";

export const fetchEndOfDay = () => request("/api/v1/system/end-of-day");
export const updateEndOfDay = (at) =>
  request("/api/v1/system/end-of-day", { method: "PUT", body: { at } });
export const runEndOfDay = () => request("/api/v1/system/end-of-day/run", { method: "POST" });

export const fetchTimeSettings = () => request("/api/v1/system/time");
export const updateTimeSettings = (body) =>
  request("/api/v1/system/time", { method: "PUT", body });
