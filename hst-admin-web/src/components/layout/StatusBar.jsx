import { useEffect, useState } from "react";
import { isSocketOpen, onEvent } from "@/api/socket.js";

/** Hints left, connection truth right — the dot follows the live socket, not just the token. */
export function StatusBar({ connected, moduleLabel }) {
  const [socketUp, setSocketUp] = useState(isSocketOpen());

  useEffect(() => {
    const offOpen = onEvent("socket.open", () => setSocketUp(true));
    const offClose = onEvent("socket.close", () => setSocketUp(false));
    return () => {
      offOpen();
      offClose();
    };
  }, []);

  const up = connected && socketUp;

  return (
    <footer className="status-bar">
      <span>For Help, press F1</span>
      <span>{moduleLabel}</span>
      <span className={up ? "connected" : "disconnected"}>
        ● {up ? "Connected" : "Disconnected"}
      </span>
    </footer>
  );
}
