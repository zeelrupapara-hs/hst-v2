export function StatusBar({ demo, moduleLabel }) {
  return (
    <footer className="status-bar">
      <span className="connected">● {demo ? "Demo mode" : "Connected"}</span>
      <span>Server: {demo ? "Sample data" : "Live API"}</span>
      <span>{moduleLabel}</span>
    </footer>
  );
}
