import { request } from "@/api/client.js";

export const fetchManagers = () => request("/api/v1/managers");
export const fetchManager = (login) => request(`/api/v1/managers/${login}`);
export const fetchManagerRights = (login) => request(`/api/v1/managers/${login}/rights`);
export const createManager = (body) => request("/api/v1/managers", { method: "POST", body });
export const updateManager = (login, body) =>
  request(`/api/v1/managers/${login}`, { method: "PATCH", body });
export const deleteManager = (login) => request(`/api/v1/managers/${login}`, { method: "DELETE" });
