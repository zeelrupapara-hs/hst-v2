import { request } from "@/api/client.js";

const PAGE = 500;
const MAX_ROWS = 5000;

/** One page of the journal: message text, time bounds (ns), mode, module type, own or server scope. */
export function searchJournal({ search = "", from, to, mode, type, channel, scope, page, limit = 200, order } = {}) {
  const q = new URLSearchParams({ limit: String(limit) });
  if (search) q.set("search", search);
  if (from) q.set("from", String(from));
  if (to) q.set("to", String(to));
  if (mode && mode !== "full") q.set("mode", mode);
  if (type) q.set("type", String(type));
  if (channel) q.set("channel", channel);
  if (scope) q.set("scope", scope);
  if (page) q.set("page", String(page));
  if (order) q.set("order", order);
  return request(`/api/v1/journal?${q}`);
}

/** Every row of a request, oldest first, paged until the window is drained or MAX_ROWS is reached. */
export async function requestJournal(params) {
  const rows = [];
  for (let page = 1; rows.length < MAX_ROWS; page += 1) {
    const res = await searchJournal({ ...params, page, limit: PAGE, order: "asc" });
    if (!res.ok) return res;
    rows.push(...(res.data || []));
    if ((res.data || []).length < PAGE) break;
  }
  return { ok: true, data: rows, truncated: rows.length >= MAX_ROWS };
}

export const fetchJournal = ({ limit = 100 } = {}) =>
  request(`/api/v1/journal?limit=${limit}`);
