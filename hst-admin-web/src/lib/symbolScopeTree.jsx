/** Symbol scope tree — shared by datafeed mask editors and group path pickers. */

export function buildScopeTree(symbols) {
  const root = { name: "Symbols", children: new Map(), symbols: [] };
  for (const s of symbols) {
    const parts = String(s.path || s.symbol || "").split("\\").filter(Boolean);
    if (!parts.length) continue;
    let node = root;
    for (const part of parts.slice(0, -1)) {
      if (!node.children.has(part)) node.children.set(part, { name: part, children: new Map(), symbols: [] });
      node = node.children.get(part);
    }
    node.symbols.push({ symbol: parts[parts.length - 1], description: s.description || "" });
  }
  return root;
}

/** Expand folder keys that contain the current scope/path value. */
export function scopeTreeOpenPaths(value) {
  const open = new Set();
  const needle = String(value ?? "").trim();
  if (!needle || needle === "*") return open;
  const base = needle.endsWith("\\*") ? needle.slice(0, -2) : needle;
  let acc = "";
  for (const part of base.split("\\").filter(Boolean)) {
    acc = acc ? `${acc}\\${part}` : part;
    open.add(acc);
  }
  return open;
}

function ScopeTreeBranch({ node, path, depth, open, toggle, onPick, leafOnly }) {
  const key = path.join("\\");
  const expanded = open.has(key);
  return (
    <>
      <div className="symtree-row" style={{ paddingLeft: depth * 16 }}>
        <button
          type="button"
          className="symtree-arrow"
          aria-label={expanded ? "collapse" : "expand"}
          onClick={(e) => {
            e.stopPropagation();
            toggle(key);
          }}
        >
          {expanded ? "▾" : "▸"}
        </button>
        <button
          type="button"
          className="symtree-item symtree-folder"
          onMouseDown={(e) => {
            e.preventDefault();
            if (leafOnly) toggle(key);
            else onPick(depth === 0 ? "*" : `${key}\\*`);
          }}
        >
          <span className="df-row-icon" aria-hidden="true">
            $
          </span>
          {node.name}
        </button>
      </div>
      {expanded && (
        <>
          {[...node.children.values()]
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((child) => (
              <ScopeTreeBranch
                key={child.name}
                node={child}
                path={[...path, child.name]}
                depth={depth + 1}
                open={open}
                toggle={toggle}
                onPick={onPick}
                leafOnly={leafOnly}
              />
            ))}
          {node.symbols
            .slice()
            .sort((a, b) => a.symbol.localeCompare(b.symbol))
            .map((s) => (
              <div key={s.symbol} className="symtree-row" style={{ paddingLeft: (depth + 1) * 16 + 18 }}>
                <button
                  type="button"
                  className="symtree-item"
                  onMouseDown={(e) => {
                    e.preventDefault();
                    onPick(leafOnly ? s.symbol : [...path, s.symbol].join("\\"));
                  }}
                >
                  <span className="df-row-icon" aria-hidden="true">
                    $
                  </span>
                  {s.symbol}
                  {s.description && <span className="symtree-desc">, {s.description}</span>}
                </button>
              </div>
            ))}
        </>
      )}
    </>
  );
}

export function SymbolScopeTreePanel({ tree, open, toggle, onPick, leafOnly = false }) {
  if (!tree) {
    return <div className="symtree-row symtree-empty">loading…</div>;
  }
  return (
    <ScopeTreeBranch
      node={tree}
      path={[]}
      depth={0}
      open={open}
      toggle={toggle}
      onPick={onPick}
      leafOnly={leafOnly}
    />
  );
}
