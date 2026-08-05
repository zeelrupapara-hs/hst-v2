// The only file that talks HTTP. Components call api/endpoints/*, which call this.

const STORAGE = {
  apiBase: "hst_api_base",
  token: "hst_access_token",
  refresh: "hst_refresh_token",
};

const DEFAULT_BASE = "http://localhost:8080";

export const getApiBase = () =>
  (localStorage.getItem(STORAGE.apiBase) || DEFAULT_BASE).replace(/\/$/, "");

export const getToken = () => localStorage.getItem(STORAGE.token) || "";

export function setSession({ access_token, refresh_token }) {
  if (access_token) localStorage.setItem(STORAGE.token, access_token);
  if (refresh_token) localStorage.setItem(STORAGE.refresh, refresh_token);
}

export function clearSession() {
  localStorage.removeItem(STORAGE.token);
  localStorage.removeItem(STORAGE.refresh);
}

/**
 * One shape for every call; never throws. A 401 gets one refresh attempt, then the
 * session is cleared and the caller's `ok:false` routes the app to /login.
 * @returns {Promise<{ok: boolean, status: number, data: any, message: string}>}
 */
export async function request(path, { method = "GET", body, retry = true } = {}) {
  let res;
  try {
    res = await fetch(getApiBase() + path, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...(getToken() ? { Authorization: `Bearer ${getToken()}` } : {}),
      },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    return { ok: false, status: 0, data: null, message: "server unreachable" };
  }

  if (res.status === 401 && retry && (await refreshSession())) {
    return request(path, { method, body, retry: false });
  }

  const envelope = await res.json().catch(() => ({}));

  return {
    ok: res.ok && envelope.success !== false,
    status: res.status,
    data: envelope.data ?? null,
    message: envelope.message || envelope.error || "",
  };
}

let refreshing = null;

// The server rotates refresh tokens and revokes the family on reuse, so
// concurrent 401s must share one refresh call.
function refreshSession() {
  refreshing ??= doRefresh().finally(() => (refreshing = null));
  return refreshing;
}

async function doRefresh() {
  const refresh = localStorage.getItem(STORAGE.refresh);
  if (!refresh) return clearSession(), false;

  const res = await fetch(getApiBase() + "/auth/v1/refresh", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refresh }),
  }).catch(() => null);

  const data = res?.ok ? (await res.json().catch(() => ({})))?.data : null;
  if (!data?.access_token) return clearSession(), false;

  setSession(data);
  return true;
}
