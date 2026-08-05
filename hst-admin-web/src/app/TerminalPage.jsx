/** @param {{panel: "admin"|"manager"}} props */
export function TerminalPage({ panel }) {
  return (
    <div style={{ padding: 24 }}>
      <h3>{panel === "admin" ? "HST Administrator" : "HST Manager"} — Trade Server</h3>
      <p>The shell is the first build phase; see ../ADMIN_FRONTEND_PROMPT.md §5.</p>
    </div>
  );
}
