import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { fetchSymbols } from "@/api/endpoints/symbols.js";

// The symbol tree, built once per mount from every symbol's backslash path.
function buildTree(symbols) {
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

function Branch({ node, path, depth, open, toggle, onPick, leafOnly }) {
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
          onClick={() => (leafOnly ? toggle(key) : onPick(depth === 0 ? "*" : `${key}\\*`))}
        >
          <span className="df-row-icon" aria-hidden="true">$</span>
          {node.name}
        </button>
      </div>
      {expanded && (
        <>
          {[...node.children.values()]
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((child) => (
              <Branch
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
                <button type="button" className="symtree-item" onClick={() => onPick(leafOnly ? s.symbol : [...path, s.symbol].join("\\"))}>
                  <span className="df-row-icon" aria-hidden="true">$</span>
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

/**
 * A mask editor whose dropdown is the symbol tree, folders closed at first: a folder
 * picks the whole branch as a pattern, a leaf picks that one symbol.
 */
export function SymbolTreeSelect({ value, onCommit, onCancel, leafOnly = false }) {
  const inputRef = useRef(null);
  const [tree, setTree] = useState(null);
  const [openList, setOpenList] = useState(false);
  const [open, setOpen] = useState(() => new Set());
  const [rect, setRect] = useState(null);

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select?.();
    fetchSymbols().then((res) => res.ok && setTree(buildTree(res.data || [])));
  }, []);

  const toggle = (key) =>
    setOpen((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });

  function showList() {
    setRect(inputRef.current.getBoundingClientRect());
    setOpenList(true);
  }

  // an empty scope means everything; an empty symbol stays empty
  const commit = (v) => onCommit(v.trim() || (leafOnly ? "" : "*"));

  return (
    <span className="symtree-editor">
      <input
        ref={inputRef}
        type="text"
        className="df-cell-input"
        defaultValue={value ?? ""}
        onBlur={(e) => {
          // picking from the list steals focus; the pick commits, not the blur
          if (!openList) commit(e.target.value);
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            commit(e.currentTarget.value);
          } else if (e.key === "Escape") {
            e.preventDefault();
            onCancel();
          }
        }}
        onClick={(e) => e.stopPropagation()}
        onDoubleClick={(e) => e.stopPropagation()}
      />
      <button
        type="button"
        className="symtree-drop"
        aria-label="browse symbols"
        onMouseDown={(e) => e.preventDefault()}
        onClick={(e) => {
          e.stopPropagation();
          if (openList) setOpenList(false);
          else showList();
        }}
      >
        ▾
      </button>
      {openList &&
        rect &&
        createPortal(
          <>
            <div className="symtree-backdrop" onClick={() => setOpenList(false)} />
            <div
              className="symtree-pop"
              style={{ left: rect.left, top: rect.bottom + 1, minWidth: rect.width }}
            >
              {tree ? (
                <Branch node={tree} path={[]} depth={0} open={open} toggle={toggle} onPick={commit} leafOnly={leafOnly} />
              ) : (
                <div className="symtree-row symtree-empty">loading…</div>
              )}
            </div>
          </>,
          document.body,
        )}
    </span>
  );
}
