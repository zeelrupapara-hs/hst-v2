import { request } from "@/api/client.js";

export const fetchOnlineUsers = () => request("/api/v1/online");
export const disconnectSession = (sessionId) =>
  request(`/api/v1/online/${sessionId}`, { method: "DELETE" });
