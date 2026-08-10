import { request } from "@/api/client.js";

export const fetchRouting = () => request("/api/v1/routing");
export const fetchRoutingRule = (id) => request(`/api/v1/routing/${id}`);
export const createRoutingRule = (body) => request("/api/v1/routing", { method: "POST", body });
export const updateRoutingRule = (id, patch) =>
  request(`/api/v1/routing/${id}`, { method: "PATCH", body: patch });
export const deleteRoutingRule = (id) => request(`/api/v1/routing/${id}`, { method: "DELETE" });
export const moveRoutingRule = (id, dir) =>
  request(`/api/v1/routing/${id}/move-${dir}`, { method: "POST" });
export const reorderRouting = (order) =>
  request("/api/v1/routing/order", { method: "PUT", body: { routing_ids: order } });

export const fetchRoutingDealers = (id) => request(`/api/v1/routing/${id}/dealers`);
export const addRoutingDealer = (id, login) =>
  request(`/api/v1/routing/${id}/dealers`, { method: "POST", body: { login } });
export const removeRoutingDealer = (id, login) =>
  request(`/api/v1/routing/${id}/dealers/${login}`, { method: "DELETE" });

export const moveRoutingDealer = (id, login, dealerIndex) =>
  request(`/api/v1/routing/${id}/dealers/${login}/move`, {
    method: "PUT",
    body: { dealer_index: dealerIndex },
  });
