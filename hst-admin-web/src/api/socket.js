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
    if (typeof msg.data !== "string") return;
    if (msg.data.startsWith("summary,")) return;
    let event;
    try {
      event = JSON.parse(msg.data);
    } catch {
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

/** @returns {() => void} unsubscribe */
export function onEvent(type, fn) {
  handlers.get(type) ?? handlers.set(type, new Set());
  handlers.get(type).add(fn);
  return () => handlers.get(type)?.delete(fn);
}
