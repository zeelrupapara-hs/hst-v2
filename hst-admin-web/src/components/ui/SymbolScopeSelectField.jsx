import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { fetchSymbols } from "@/api/endpoints/symbols.js";
import { useExclusiveDropdown } from "@/hooks/useExclusiveDropdown.js";
import {
  buildScopeTree,
  scopeTreeOpenPaths,
  SymbolScopeTreePanel,
} from "@/lib/symbolScopeTree.jsx";

/**
 * Group-symbol path mask picker using the same symbol tree as datafeed scopes:
 * `*`, folder\\*, or a full symbol path.
 */
export function SymbolScopeSelectField({
  label = "Symbol",
  value,
  onChange,
  disabled,
  className = "",
  leafOnly = false,
}) {
  const [tree, setTree] = useState(null);
  const [open, setOpen] = useState(false);
  const [openPaths, setOpenPaths] = useState(() => new Set());
  const [raw, setRaw] = useState(null);
  const [pos, setPos] = useState(null);
  const rootRef = useRef(null);
  const btnRef = useRef(null);
  const menuRef = useRef(null);
  const openRef = useRef(false);
  const suppressBlurRef = useRef(false);
  const announceOpen = useExclusiveDropdown(open, setOpen);

  useEffect(() => {
    fetchSymbols().then((res) => res.ok && setTree(buildScopeTree(res.data || [])));
  }, []);

  useEffect(() => {
    openRef.current = open;
  }, [open]);

  useEffect(() => {
    if (!open) return;
    setOpenPaths(scopeTreeOpenPaths(value));
    function close(e) {
      if (rootRef.current?.contains(e.target)) return;
      if (btnRef.current?.contains(e.target)) return;
      if (menuRef.current?.contains(e.target)) return;
      setOpen(false);
    }
    const onKey = (e) => e.key === "Escape" && setOpen(false);
    const t = setTimeout(() => document.addEventListener("mousedown", close), 0);
    document.addEventListener("keydown", onKey);
    return () => {
      clearTimeout(t);
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", onKey);
    };
  }, [open, value]);

  const display = raw ?? (value == null || value === "" ? (leafOnly ? "" : "*") : String(value));

  function commit(text) {
    const next = String(text ?? "").trim();
    onChange?.(next === "" ? (leafOnly ? "" : "*") : next);
    setRaw(null);
  }

  function pick(next) {
    onChange?.(next == null || next === "" ? (leafOnly ? "" : "*") : String(next));
    setRaw(null);
    setOpen(false);
  }

  function openMenu(e) {
    if (disabled) return;
    e.preventDefault();
    e.stopPropagation();
    suppressBlurRef.current = true;
    setRaw(null);
    if (openRef.current) {
      setOpen(false);
      return;
    }
    announceOpen();
    const r = rootRef.current.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: Math.max(r.width, 260) });
    setOpen(true);
  }

  const togglePath = (key) =>
    setOpenPaths((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });

  return (
    <>
      <label>{label}</label>
      <div
        ref={rootRef}
        className={`prop-select prop-select-fill editable-select sym-common-left ${className}${disabled ? " prop-select-disabled" : ""}`.trim()}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="prop-select-box editable-select-box">
          <input
            type="text"
            className="editable-select-input"
            value={display}
            disabled={disabled}
            onBlur={(e) => {
              if (suppressBlurRef.current) {
                suppressBlurRef.current = false;
                return;
              }
              if (open) return;
              commit(e.target.value);
              setRaw(null);
            }}
            onChange={(e) => setRaw(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                commit(e.currentTarget.value);
                e.currentTarget.blur();
                setOpen(false);
              } else if (e.key === "Escape") {
                setOpen(false);
              } else if (e.key === "ArrowDown" && !open) {
                e.preventDefault();
                openMenu(e);
              }
            }}
          />
          <span className="prop-select-btn-wrap" ref={btnRef} onMouseDown={openMenu}>
            <button
              type="button"
              className="prop-select-btn"
              aria-label="Open symbol tree"
              disabled={disabled}
              tabIndex={-1}
            />
          </span>
        </div>
      </div>
      {open &&
        pos &&
        createPortal(
          <div
            ref={menuRef}
            className="symtree-pop prop-select-menu-front"
            style={{ position: "fixed", top: pos.top, left: pos.left, minWidth: pos.width }}
          >
            <SymbolScopeTreePanel
              tree={tree}
              open={openPaths}
              toggle={togglePath}
              onPick={pick}
              leafOnly={leafOnly}
            />
          </div>,
          document.body,
        )}
    </>
  );
}
