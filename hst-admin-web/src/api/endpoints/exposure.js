import { request } from "@/api/client.js";

export const fetchExposure = () => request("/api/v1/exposure");
