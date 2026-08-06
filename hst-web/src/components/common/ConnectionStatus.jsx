import { useEffect, useState } from "react";
import { useSocket } from "../../socket";

// A socket can be open while nothing comes down it, which is the state that reads as healthy and
// is not. Prices are the proof, so the indicator watches them rather than the connection alone.
const STALE_AFTER_MS = 8000;

const STATE = {
  live: { label: "Live", dot: "bg-green-500", title: "Prices are arriving" },
  stale: {
    label: "No data",
    dot: "bg-amber-500",
    title: "Connected, but no prices have arrived recently",
  },
  offline: {
    label: "Reconnecting",
    dot: "bg-red-500 animate-pulse",
    title: "Not connected to the server, retrying",
  },
};

const ConnectionStatus = () => {
  const { readyState, lastTickAt } = useSocket() || {};
  const [state, setState] = useState("offline");

  useEffect(() => {
    const tick = () => {
      if (readyState !== 1) return setState("offline");

      const since = Date.now() - (lastTickAt?.current || 0);
      setState(since > STALE_AFTER_MS ? "stale" : "live");
    };

    tick();
    const t = setInterval(tick, 1000);
    return () => clearInterval(t);
  }, [readyState, lastTickAt]);

  const s = STATE[state];

  return (
    <div
      className="flex items-center gap-1.5 px-2 text-xs text-theme-text"
      title={s.title}
    >
      <span className={`h-2 w-2 rounded-full ${s.dot}`} />
      <span className="hidden sm:inline">{s.label}</span>
    </div>
  );
};

export default ConnectionStatus;
