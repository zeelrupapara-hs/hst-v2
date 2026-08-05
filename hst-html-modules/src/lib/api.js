const STORAGE = {
  apiBase: "hst_api_base",
  token: "hst_access_token",
  refreshToken: "hst_refresh_token",
  login: "hst_login",
  connectionType: "hst_connection_type",
};

const DEFAULT_BASE = "http://localhost:8080";

export function getApiBase() {
  return (localStorage.getItem(STORAGE.apiBase) || DEFAULT_BASE).replace(/\/$/, "");
}

export function setApiBase(url) {
  localStorage.setItem(STORAGE.apiBase, (url || DEFAULT_BASE).replace(/\/$/, ""));
}

export function getToken() {
  return localStorage.getItem(STORAGE.token) || "";
}

export function isLoggedIn() {
  return !!getToken();
}

export function setTokens(access, refresh, login, connectionType) {
  if (access) localStorage.setItem(STORAGE.token, access);
  if (refresh) localStorage.setItem(STORAGE.refreshToken, refresh);
  if (login != null) localStorage.setItem(STORAGE.login, String(login));
  if (connectionType != null) localStorage.setItem(STORAGE.connectionType, String(connectionType));
}

export function clearAuth() {
  Object.keys(STORAGE).forEach((k) => localStorage.removeItem(STORAGE[k]));
}

export async function request(method, path, body) {
  const url = path.indexOf("http") === 0 ? path : getApiBase() + path;
  const headers = { Accept: "application/json" };
  if (body != null) headers["Content-Type"] = "application/json";
  const token = getToken();
  if (token) headers.Authorization = "Bearer " + token;

  const init = { method, headers };
  if (body != null) init.body = JSON.stringify(body);

  const started = performance.now();
  try {
    const res = await fetch(url, init);
    const text = await res.text();
    let json = null;
    try {
      json = text ? JSON.parse(text) : null;
    } catch {
      json = { raw: text };
    }

    return {
      ok: res.ok,
      status: res.status,
      elapsed: Math.round(performance.now() - started),
      data: json && json.data !== undefined ? json.data : json,
      envelope: json,
      url,
    };
  } catch (err) {
    return {
      ok: false,
      status: 0,
      elapsed: Math.round(performance.now() - started),
      error: err.message || String(err),
      data: null,
      url,
    };
  }
}

export function get(path) {
  return request("GET", path);
}

export function post(path, body) {
  return request("POST", path, body);
}

export function patch(path, body) {
  return request("PATCH", path, body);
}

export function del(path) {
  return request("DELETE", path);
}

export async function login(loginId, password, connectionType) {
  const base = getApiBase();
  const res = await fetch(base + "/auth/v1/login", {
    method: "POST",
    headers: {
      Authorization: "Basic " + btoa(loginId + ":" + password),
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({ connection_type: Number(connectionType) }),
  });
  const json = await res.json();
  if (res.ok && json.data) {
    setTokens(json.data.access_token, json.data.refresh_token, json.data.login, connectionType);
  }
  return { ok: res.ok, status: res.status, data: json };
}
