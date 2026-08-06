import { request } from "@/api/client.js";

export const fetchUsers = () => request("/api/v1/users?limit=500");
export const fetchUser = (login) => request(`/api/v1/users/${login}`);
export const createUser = (body) => request("/api/v1/users", { method: "POST", body });
export const updateUser = (login, patch) =>
  request(`/api/v1/users/${login}`, { method: "PATCH", body: patch });
export const deleteUser = (login) => request(`/api/v1/users/${login}`, { method: "DELETE" });

export const fetchClients = () => request("/api/v1/clients?limit=500");
export const fetchClient = (id) => request(`/api/v1/clients/${id}`);
export const createClient = (body) => request("/api/v1/clients", { method: "POST", body });
export const updateClient = (id, patch) =>
  request(`/api/v1/clients/${id}`, { method: "PATCH", body: patch });
export const deleteClient = (id) => request(`/api/v1/clients/${id}`, { method: "DELETE" });

export const resetUserPassword = (login, kind, password) =>
  request(`/api/v1/users/${login}/password`, { method: "POST", body: { kind, password } });
