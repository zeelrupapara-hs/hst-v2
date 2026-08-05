import { get, isLoggedIn } from "./api.js";
import { buildQuery } from "./formatters.js";
import { getMock } from "./legacy.js";

function isDemoMode() {
  return !isLoggedIn();
}

export async function fetchList(cfg) {
  const q = { ...(cfg.query || {}) };
  const mock = getMock();
  if (isDemoMode()) {
    return { ok: true, data: mock.list(cfg.endpoint, q), elapsed: 0, demo: true };
  }
  const path = cfg.endpoint + buildQuery(q);
  const res = await get(path);
  if (!res.ok) {
    return {
      ok: true,
      data: mock.list(cfg.endpoint, q),
      elapsed: res.elapsed,
      demo: true,
      fallback: res.error || res.status,
    };
  }
  return res;
}

export async function fetchOne(endpoint, id) {
  const mock = getMock();
  if (isDemoMode()) {
    return { ok: true, data: mock.get(endpoint, id), elapsed: 0, demo: true };
  }
  const res = await get(endpoint + "/" + encodeURIComponent(id));
  if (!res.ok) {
    return {
      ok: true,
      data: mock.get(endpoint, id),
      elapsed: res.elapsed,
      demo: true,
      fallback: res.status,
    };
  }
  return res;
}

export function normalizeRows(data, transform) {
  if (transform) return transform(data) || [];
  return Array.isArray(data) ? data : (data && data.items) || [];
}

export async function fetchTimeSettings() {
  const mock = getMock();
  if (isDemoMode() || !mock) {
    return { ok: true, data: mock?.getTimeSettings?.() ?? null, elapsed: 0, demo: true };
  }
  const res = await get("/api/v1/system/time");
  if (!res.ok) {
    return {
      ok: true,
      data: mock.getTimeSettings(),
      elapsed: res.elapsed,
      demo: true,
      fallback: res.error || res.status,
    };
  }
  return res;
}

export async function saveTimeSettings(patch) {
  const mock = getMock();
  const data = mock?.updateTimeSettings?.(patch) ?? patch;
  return { ok: true, data, elapsed: 0, demo: true };
}

export async function fetchSymbol(id) {
  const mock = getMock();
  if (isDemoMode() || !mock) {
    return { ok: true, data: mock?.getSymbol?.(id) ?? null, elapsed: 0, demo: true };
  }
  const res = await get("/api/v1/symbols/" + encodeURIComponent(id));
  if (!res.ok) {
    return {
      ok: true,
      data: mock.getSymbol(id),
      elapsed: res.elapsed,
      demo: true,
      fallback: res.status,
    };
  }
  return res;
}

export async function saveSymbol(id, patch) {
  const mock = getMock();
  const data = mock?.updateSymbol?.(id, patch) ?? patch;
  return { ok: true, data, elapsed: 0, demo: true };
}

export function getSymbolSessionsForList(symbolId) {
  const mock = getMock();
  return mock?.getSymbolSessions?.(symbolId) ?? [];
}
