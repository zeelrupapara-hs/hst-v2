import { getApiBase, getToken } from "@/api/client.js";

// One shared connection; consumers subscribe to event types, never to the raw socket.
let ws = null;
let wantOpen = false;
let retryDelay = 1000;
const handlers = new Map();

function dispatch(type, event) {
  handlers.get(type)?.forEach((fn) => fn(event));
  handlers.get("*")?.forEach((fn) => fn(event));
}

function connect() {
  const token = getToken();
  if (!token) return;

  const base = getApiBase().replace(/^http/, "ws");
  ws = new WebSocket(`${base}/ws?access_token=${encodeURIComponent(token)}`);
  ws.binaryType = "arraybuffer";

  ws.onopen = () => {
    retryDelay = 1000;
    dispatch("socket.open", {});
  };

  ws.onmessage = (msg) => {
    // ticks and account summaries arrive as one comma line, everything else as json
    const raw = typeof msg.data === "string" ? msg.data : new TextDecoder().decode(msg.data);
    if (raw.startsWith("summary,")) {
      dispatch("account_summary", { type: "account_summary", payload: raw });
      return;
    }
    let event;
    try {
      event = JSON.parse(raw);
    } catch {
      dispatch("market_feed", { type: "market_feed", payload: raw });
      return;
    }
    if (event?.type === "account_summary" && typeof event.payload === "string") {
      dispatch("account_summary", event);
      return;
    }
    if (event?.type) dispatch(event.type, event);
  };

  ws.onclose = () => {
    ws = null;
    dispatch("socket.close", {});
    if (!wantOpen) return;
    setTimeout(() => wantOpen && connect(), retryDelay);
    retryDelay = Math.min(retryDelay * 2, 30000);
  };

  ws.onerror = () => ws?.close();
}

export function startSocket() {
  if (wantOpen) return;
  wantOpen = true;
  connect();
}

export function stopSocket() {
  wantOpen = false;
  ws?.close();
  ws = null;
}

export const isSocketOpen = () => ws?.readyState === WebSocket.OPEN;

/** Sends one event to the server, silently dropped while the socket is down. */
export function sendEvent(type, payload) {
  if (!isSocketOpen()) return;
  ws.send(JSON.stringify(payload === undefined ? { type } : { type, payload }));
}

/** @returns {() => void} unsubscribe */
export function onEvent(type, fn) {
  handlers.get(type) ?? handlers.set(type, new Set());
  handlers.get(type).add(fn);
  return () => handlers.get(type)?.delete(fn);
}
