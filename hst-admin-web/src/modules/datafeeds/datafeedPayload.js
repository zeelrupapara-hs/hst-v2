/** Main datafeed fields sent on create / patch (Common + Timeouts tabs). */
export const MAIN_FIELDS = [
  "name",
  "module",
  "enable",
  "mode",
  "gateway_server",
  "feed_server",
  "feed_login",
  "feed_password",
  "gateway_login",
  "gateway_password",
  "timeout",
  "timeout_reconnect",
  "timeout_sleep",
  "attempts_sleep",
  "company",
  "issuer",
  "allow_import_symbols",
];

const LOGIN_FIELDS = ["feed_login", "gateway_login"];
const PASSWORD_FIELDS = ["feed_password", "gateway_password"];

/** Counterparty login ids are opaque text, not numbers. */
export function loginField(value) {
  if (value == null || value === 0) return "";
  return String(value);
}

/** Normalize API / draft values for the form. */
export function normalizeDatafeedDraft(row) {
  if (!row) return row;
  const next = { ...row };
  for (const key of LOGIN_FIELDS) next[key] = loginField(next[key]);
  for (const key of PASSWORD_FIELDS) next[key] = next[key] == null ? "" : String(next[key]);
  return next;
}

export function newDatafeedDraft() {
  return normalizeDatafeedDraft({
    name: "",
    module: "",
    enable: 1,
    mode: 1,
    feed_server: "",
    feed_login: "",
    gateway_server: "",
    gateway_login: "",
    timeout: 0,
    timeout_reconnect: 5,
    timeout_sleep: 60,
    attempts_sleep: 10,
    feed_symbols: [],
    translates: [],
    params: [],
  });
}

/** Body for POST /datafeeds — login fields always strings. */
export function datafeedCreateBody(draft) {
  const d = normalizeDatafeedDraft(draft);
  const body = {};
  for (const key of MAIN_FIELDS) {
    const v = d[key];
    if (v === undefined) continue;
    if (PASSWORD_FIELDS.includes(key) && v === "") continue;
    body[key] = v;
  }
  return body;
}

/** Body for PATCH /datafeeds/:id — passwords only when non-empty. */
export function datafeedPatchBody(draft, original) {
  const d = normalizeDatafeedDraft(draft);
  const o = normalizeDatafeedDraft(original);
  const patch = {};
  for (const key of MAIN_FIELDS) {
    const v = d[key];
    if (v === undefined) continue;
    if (PASSWORD_FIELDS.includes(key) && v === "") continue;
    if (v !== o[key]) patch[key] = v;
  }
  return patch;
}
