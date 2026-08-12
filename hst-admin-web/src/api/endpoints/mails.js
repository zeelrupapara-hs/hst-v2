import { request } from "@/api/client.js";

export const sendMail = (body) => request("/api/v1/mails", { method: "POST", body });
