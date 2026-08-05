/** Hints left, connection truth right — a dead session must not look connected. */
export function StatusBar({ connected, moduleLabel }) {
  return (
    <footer className="status-bar">
      <span>For Help, press F1</span>
      <span>{moduleLabel}</span>
      <span className={connected ? "connected" : "disconnected"}>
        ● {connected ? "Connected" : "Disconnected"}
      </span>
    </footer>
  );
}
