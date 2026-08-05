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

export async function deleteSymbols(ids) {
  const mock = getMock();
  const count = mock?.deleteSymbols?.(ids) ?? 0;
  return { ok: true, count, elapsed: 0, demo: true };
}

export async function createSymbol(folder, name, description) {
  const mock = getMock();
  const data = mock?.createSymbol?.(folder, name, description) ?? null;
  return { ok: !!data, data, elapsed: 0, demo: true };
}

export async function moveSymbol(id, direction) {
  const mock = getMock();
  const ok = mock?.moveSymbolInList?.(id, direction) ?? false;
  return { ok, elapsed: 0, demo: true };
}

export async function fetchSymbolGroups() {
  const mock = getMock();
  return {
    ok: true,
    data: mock?.listSymbolGroups?.() ?? [],
    elapsed: 0,
    demo: true,
  };
}

export async function createSymbolGroup(parentPath, name) {
  const mock = getMock();
  const path = mock?.createSymbolGroup?.(parentPath, name) ?? null;
  return { ok: !!path, data: path, elapsed: 0, demo: true };
}

export async function fetchDatafeed(id) {
  const mock = getMock();
  if (isDemoMode() || !mock) {
    return { ok: true, data: mock?.getDatafeed?.(id) ?? null, elapsed: 0, demo: true };
  }
  const res = await get("/api/v1/datafeeds/" + encodeURIComponent(id));
  if (!res.ok) {
    return {
      ok: true,
      data: mock.getDatafeed(id),
      elapsed: res.elapsed,
      demo: true,
      fallback: res.status,
    };
  }
  return res;
}

export async function saveDatafeed(id, patch) {
  const mock = getMock();
  const data = mock?.updateDatafeed?.(id, patch) ?? patch;
  return { ok: true, data, elapsed: 0, demo: true };
}

export async function deleteDatafeeds(ids) {
  const mock = getMock();
  const count = mock?.deleteDatafeeds?.(ids) ?? 0;
  return { ok: true, count, elapsed: 0, demo: true };
}

export async function createDatafeed(data) {
  const mock = getMock();
  const created = mock?.createDatafeed?.(data) ?? null;
  return { ok: !!created, data: created, elapsed: 0, demo: true };
}

export async function moveDatafeed(id, direction) {
  const mock = getMock();
  const ok = mock?.moveDatafeedInList?.(id, direction) ?? false;
  return { ok, elapsed: 0, demo: true };
}

export async function setDatafeedEnable(id, enable) {
  const mock = getMock();
  const data = mock?.setDatafeedEnable?.(id, enable) ?? null;
  return { ok: !!data, data, elapsed: 0, demo: true };
}

export async function fetchDatafeedModules(mode = 1) {
  const mock = getMock();
  if (isDemoMode() || !mock) {
    return {
      ok: true,
      data: mock?.listDatafeedModules?.(mode) ?? [],
      elapsed: 0,
      demo: true,
    };
  }
  const path =
    "/api/v1/datafeeds/modules" + buildQuery({ mode: String(mode) });
  const res = await get(path);
  if (!res.ok) {
    return {
      ok: true,
      data: mock?.listDatafeedModules?.(mode) ?? [],
      elapsed: res.elapsed,
      demo: true,
      fallback: res.status,
    };
  }
  return res;
}

export function getSymbolSessionsForList(symbolId) {
  const mock = getMock();
  return mock?.getSymbolSessions?.(symbolId) ?? [];
}
