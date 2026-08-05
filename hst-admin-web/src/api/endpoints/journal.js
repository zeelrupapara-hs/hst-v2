import { request } from "@/api/client.js";

/** Search the server journal: message text, time bounds (ns), errors-only mode. */
export function searchJournal({ search = "", from, to, errorsOnly, limit = 200 } = {}) {
  const q = new URLSearchParams({ limit: String(limit) });
  if (search) q.set("search", search);
  if (from) q.set("from", String(from));
  if (to) q.set("to", String(to));
  if (errorsOnly) q.set("mode", "errors_only");
  return request(`/api/v1/journal?${q}`);
}

export const fetchJournal = ({ limit = 100 } = {}) =>
  request(`/api/v1/journal?limit=${limit}`);
