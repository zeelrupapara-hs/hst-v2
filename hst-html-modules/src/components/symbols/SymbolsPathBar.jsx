export function SymbolsPathBar({ path }) {
  if (!path) {
    return <div className="symbols-path-bar">All symbols</div>;
  }
  const parts = path.split("\\");
  return (
    <div className="symbols-path-bar">
      {parts.map((p, i) => (
        <span key={i}>
          {i > 0 && <span className="sym-path-sep">\</span>}
          <span className="sym-path-seg">{p}</span>
        </span>
      ))}
    </div>
  );
}
