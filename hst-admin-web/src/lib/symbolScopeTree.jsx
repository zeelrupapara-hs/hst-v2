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

/** Every folder and symbol whose name, path or description contains the text, flat, folders first. */
export function filterScopeTree(tree, needle, leafOnly = false) {
  const q = String(needle ?? "").trim().toLowerCase();
  if (!tree || !q) return [];
  const out = [];
  const walk = (node, path) => {
    for (const child of [...node.children.values()].sort((a, b) => a.name.localeCompare(b.name))) {
      const key = [...path, child.name].join("\\");
      if (!leafOnly && key.toLowerCase().includes(q)) out.push({ kind: "folder", value: `${key}\\*`, label: key });
      walk(child, [...path, child.name]);
    }
    for (const s of node.symbols.slice().sort((a, b) => a.symbol.localeCompare(b.symbol))) {
      const full = [...path, s.symbol].join("\\");
      if (full.toLowerCase().includes(q) || s.description.toLowerCase().includes(q)) {
        out.push({ kind: "symbol", value: leafOnly ? s.symbol : full, label: full, description: s.description });
      }
    }
  };
  walk(tree, []);
  return out;
}

/** The arrow keys walk the match list; returns the new index, or null when the key is not an arrow. */
export function stepMatch(key, index, count) {
  if (key !== "ArrowDown" && key !== "ArrowUp") return null;
  if (!count) return 0;
  return key === "ArrowDown" ? (index + 1) % count : (index - 1 + count) % count;
}

/**
 * What typed text commits to: a mask stays as typed, anything else must name a folder or symbol in
 * the tree (exact first, then the first match), and null says there is nothing to commit.
 */
export function resolveScopeText(tree, text, leafOnly = false) {
  const t = String(text ?? "").trim();
  if (t === "") return leafOnly ? "" : "*";
  if (t.includes("*") || t.includes("!") || t.includes(",")) return leafOnly ? null : t;
  const matches = filterScopeTree(tree, t, leafOnly);
  const exact = matches.find((m) => m.label.toLowerCase() === t.toLowerCase() || m.value.toLowerCase() === t.toLowerCase()
    || (m.kind === "symbol" && m.label.split("\\").pop().toLowerCase() === t.toLowerCase()));
  return (exact ?? matches[0])?.value ?? null;
}

function ScopeMatches({ matches, onPick, active }) {
  if (!matches.length) return <div className="symtree-row symtree-empty">no symbol matches</div>;
  return matches.map((m, i) => (
    <div key={`${m.kind}:${m.label}`} className="symtree-row" style={{ paddingLeft: 8 }}>
      <button
        type="button"
        className={`symtree-item${m.kind === "folder" ? " symtree-folder" : ""}${i === active ? " symtree-active" : ""}`}
        onMouseDown={(e) => {
          e.preventDefault();
          onPick(m.value);
        }}
      >
        <span className="df-row-icon" aria-hidden="true">
          $
        </span>
        {m.kind === "folder" ? `${m.label}\\*` : m.label}
        {m.description && <span className="symtree-desc">, {m.description}</span>}
      </button>
    </div>
  ));
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

export function SymbolScopeTreePanel({ tree, open, toggle, onPick, leafOnly = false, filter = "", active = 0 }) {
  if (!tree) {
    return <div className="symtree-row symtree-empty">loading…</div>;
  }
  if (String(filter).trim()) {
    return <ScopeMatches matches={filterScopeTree(tree, filter, leafOnly)} onPick={onPick} active={active} />;
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
