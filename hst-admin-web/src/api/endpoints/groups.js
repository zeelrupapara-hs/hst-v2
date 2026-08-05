import { request } from "@/api/client.js";

export const fetchGroups = () => request("/api/v1/groups?flat=1");
export const fetchGroup = (id) => request(`/api/v1/groups/${id}`);
export const createGroup = (body) => request("/api/v1/groups", { method: "POST", body });
export const updateGroup = (id, patch) =>
  request(`/api/v1/groups/${id}`, { method: "PATCH", body: patch });
export const deleteGroup = (id) => request(`/api/v1/groups/${id}`, { method: "DELETE" });

export const fetchGroupSymbols = (id) => request(`/api/v1/groups/${id}/symbols`);
export const createGroupSymbol = (id, body) =>
  request(`/api/v1/groups/${id}/symbols`, { method: "POST", body });
export const updateGroupSymbol = (id, symbolId, patch) =>
  request(`/api/v1/groups/${id}/symbols/${symbolId}`, { method: "PATCH", body: patch });
export const deleteGroupSymbol = (id, symbolId) =>
  request(`/api/v1/groups/${id}/symbols/${symbolId}`, { method: "DELETE" });

export const fetchGroupCommissions = (id) => request(`/api/v1/groups/${id}/commissions`);
export const createGroupCommission = (id, body) =>
  request(`/api/v1/groups/${id}/commissions`, { method: "POST", body });
export const updateGroupCommission = (id, commissionId, patch) =>
  request(`/api/v1/groups/${id}/commissions/${commissionId}`, { method: "PATCH", body: patch });
export const deleteGroupCommission = (id, commissionId) =>
  request(`/api/v1/groups/${id}/commissions/${commissionId}`, { method: "DELETE" });
