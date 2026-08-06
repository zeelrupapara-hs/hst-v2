import { request } from "@/api/client.js";

export const fetchMailServers = () => request("/api/v1/mail-servers");
export const createMailServer = (body) => request("/api/v1/mail-servers", { method: "POST", body });
export const updateMailServer = (id, patch) =>
  request(`/api/v1/mail-servers/${id}`, { method: "PATCH", body: patch });
export const deleteMailServer = (id) => request(`/api/v1/mail-servers/${id}`, { method: "DELETE" });
