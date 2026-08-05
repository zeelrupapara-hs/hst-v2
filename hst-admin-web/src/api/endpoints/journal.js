import { request } from "@/api/client.js";

export const fetchJournal = ({ limit = 100 } = {}) =>
  request(`/api/v1/journal?limit=${limit}`);
