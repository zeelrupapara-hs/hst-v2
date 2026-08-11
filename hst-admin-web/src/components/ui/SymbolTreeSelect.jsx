import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { fetchSymbols } from "@/api/endpoints/symbols.js";
import { buildScopeTree, scopeTreeOpenPaths, SymbolScopeTreePanel } from "@/lib/symbolScopeTree.jsx";

/**
 * Inline cell editor whose dropdown is the symbol scope tree (datafeed tables).
 * A folder picks the whole branch as a pattern; a leaf picks that one symbol.
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
    fetchSymbols().then((res) => res.ok && setTree(buildScopeTree(res.data || [])));
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
    // a collapsed root is a dead end: open it, and the branch the current value sits on, so the
    // list arrives showing folders rather than one word
    setOpen((prev) => (prev.size ? prev : new Set(["", ...scopeTreeOpenPaths(value)])));
    setOpenList(true);
  }

  const commit = (v) => onCommit(v.trim() || (leafOnly ? "" : "*"));

  return (
    <span className="symtree-editor">
      <input
        ref={inputRef}
        type="text"
        className="df-cell-input"
        defaultValue={value ?? ""}
        onBlur={(e) => {
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
              <SymbolScopeTreePanel
                tree={tree}
                open={open}
                toggle={toggle}
                onPick={commit}
                leafOnly={leafOnly}
              />
            </div>
          </>,
          document.body,
        )}
    </span>
  );
}
