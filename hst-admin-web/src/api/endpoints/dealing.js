import { request } from "@/api/client.js";

export const fetchDealingQueue = () => request("/api/v1/dealing");
export const fetchDealerState = () => request("/api/v1/dealing/state");
export const connectDealer = () => request("/api/v1/dealing/connect", { method: "POST" });
export const disconnectDealer = () => request("/api/v1/dealing/disconnect", { method: "POST" });
export const heartbeatDealer = () => request("/api/v1/dealing/heartbeat", { method: "POST" });

export const confirmRequest = (requestId, login, price = 0) =>
  request(`/api/v1/dealing/${requestId}/confirm`, { method: "POST", body: { request_id: requestId, login, price } });
export const requoteRequest = (requestId, login, price) =>
  request(`/api/v1/dealing/${requestId}/requote`, { method: "POST", body: { request_id: requestId, login, price } });
export const rejectRequest = (requestId, login, reason = "") =>
  request(`/api/v1/dealing/${requestId}/reject`, { method: "POST", body: { request_id: requestId, login, reason } });
