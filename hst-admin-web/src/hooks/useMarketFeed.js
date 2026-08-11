import { useEffect, useState } from "react";
import { onEvent, sendEvent } from "@/api/socket.js";

// Live prices, parsed from the tick lines the market feed streams:
// symbol,bid,ask,last,volume,time,open,high,low,close,change,change_percent
// The feed starts when the first consumer mounts and stops when the last one leaves.
const ticks = new Map();
const listeners = new Set();
let consumers = 0;
let wired = false;
let notifyTimer = 0;

// a burst of ticks paints once, not once per line
function notifySoon() {
  if (notifyTimer) return;
  notifyTimer = setTimeout(() => {
    notifyTimer = 0;
    listeners.forEach((fn) => fn());
  }, 100);
}

function readLine(line) {
  const parts = line.split(",");
  if (parts.length < 12) return;

  const symbol = parts[0];
  const bid = Number(parts[1]);
  const previous = ticks.get(symbol);

  ticks.set(symbol, {
    symbol,
    bid,
    ask: Number(parts[2]),
    last: Number(parts[3]),
    high: Number(parts[7]),
    low: Number(parts[8]),
    close: Number(parts[9]),
    change: Number(parts[10]),
    changePercent: Number(parts[11]),
    // up or down against the previous bid, for the blue or red flash
    dir: !previous || bid === previous.bid ? (previous?.dir ?? 0) : bid > previous.bid ? 1 : -1,
    at: Date.now(),
  });

  notifySoon();
}

function wire() {
  if (wired) return;
  wired = true;
  onEvent("market_feed", (event) => readLine(event.payload));
  // a reconnect must ask for the feed again
  onEvent("socket.open", () => consumers > 0 && sendEvent("start_market_feed"));
}

/** @returns {Map<string, object>} live prices per symbol while any consumer is mounted */
export function useMarketFeed() {
  const [, bump] = useState(0);

  useEffect(() => {
    wire();
    if (consumers === 0) sendEvent("start_market_feed");
    consumers += 1;

    const fn = () => bump((n) => n + 1);
    listeners.add(fn);

    return () => {
      listeners.delete(fn);
      consumers -= 1;
      if (consumers === 0) sendEvent("stop_market_feed");
    };
  }, []);

  return ticks;
}
