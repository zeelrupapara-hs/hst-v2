/**
 * HST API client — shared across module iframes.
 */
(function (global) {
  var STORAGE = {
    apiBase: "hst_api_base",
    token: "hst_access_token",
    refreshToken: "hst_refresh_token",
    login: "hst_login",
    connectionType: "hst_connection_type",
  };

  var DEFAULT_BASE = "http://localhost:8080";

  function getApiBase() {
    return (localStorage.getItem(STORAGE.apiBase) || DEFAULT_BASE).replace(/\/$/, "");
  }

  function setApiBase(url) {
    localStorage.setItem(STORAGE.apiBase, (url || DEFAULT_BASE).replace(/\/$/, ""));
  }

  function getToken() {
    return localStorage.getItem(STORAGE.token) || "";
  }

  function isLoggedIn() {
    return !!getToken();
  }

  function setTokens(access, refresh, login, connectionType) {
    if (access) localStorage.setItem(STORAGE.token, access);
    if (refresh) localStorage.setItem(STORAGE.refreshToken, refresh);
    if (login != null) localStorage.setItem(STORAGE.login, String(login));
    if (connectionType != null) localStorage.setItem(STORAGE.connectionType, String(connectionType));
  }

  function clearAuth() {
    Object.keys(STORAGE).forEach(function (k) { localStorage.removeItem(STORAGE[k]); });
  }

  async function request(method, path, body) {
    var url = path.indexOf("http") === 0 ? path : getApiBase() + path;
    var headers = { Accept: "application/json" };
    if (body != null) headers["Content-Type"] = "application/json";
    var token = getToken();
    if (token) headers.Authorization = "Bearer " + token;

    var init = { method: method, headers: headers };
    if (body != null) init.body = JSON.stringify(body);

    var started = performance.now();
    try {
      var res = await fetch(url, init);
      var text = await res.text();
      var json = null;
      try { json = text ? JSON.parse(text) : null; } catch (_) { json = { raw: text }; }

      return {
        ok: res.ok,
        status: res.status,
        elapsed: Math.round(performance.now() - started),
        data: json && json.data !== undefined ? json.data : json,
        envelope: json,
        url: url,
      };
    } catch (err) {
      return {
        ok: false,
        status: 0,
        elapsed: Math.round(performance.now() - started),
        error: err.message || String(err),
        data: null,
        url: url,
      };
    }
  }

  function get(path) { return request("GET", path); }
  function post(path, body) { return request("POST", path, body); }
  function patch(path, body) { return request("PATCH", path, body); }
  function del(path) { return request("DELETE", path); }

  async function login(loginId, password, connectionType) {
    var base = getApiBase();
    var res = await fetch(base + "/auth/v1/login", {
      method: "POST",
      headers: {
        Authorization: "Basic " + btoa(loginId + ":" + password),
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ connection_type: Number(connectionType) }),
    });
    var json = await res.json();
    if (res.ok && json.data) {
      setTokens(json.data.access_token, json.data.refresh_token, json.data.login, connectionType);
    }
    return { ok: res.ok, status: res.status, data: json };
  }

  global.HSTApi = {
    STORAGE: STORAGE,
    getApiBase: getApiBase,
    setApiBase: setApiBase,
    getToken: getToken,
    isLoggedIn: isLoggedIn,
    setTokens: setTokens,
    clearAuth: clearAuth,
    login: login,
    get: get,
    post: post,
    patch: patch,
    del: del,
    request: request,
  };
})(window);
