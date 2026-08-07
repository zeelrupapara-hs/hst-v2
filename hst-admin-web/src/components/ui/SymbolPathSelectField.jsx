import { useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useSymbols } from "@/hooks/useSymbols.js";
import { useSymbolFolders } from "@/hooks/useSymbolFolders.js";
import { useExclusiveDropdown } from "@/hooks/useExclusiveDropdown.js";
import {
  buildSymbolSelectTree,
  emptyFolderPathsToMap,
  filterSymbolPickerRows,
} from "@/lib/symbolTree.js";

function PathFolderRow({ node, depth, value, onSelectMask, onSelectPath, openPaths, togglePath }) {
  const hasChildren = node.children?.length > 0;
  const open = openPaths.has(node.path ?? "");
  const mask = node.path ? `${node.path}\\*` : "*";
  const selected = value === mask;

  return (
    <li className="sym-folder-select-branch">
      <div
        className={`sym-folder-select-row sym-folder-select-depth-${Math.min(depth, 6)}${selected ? " sel" : ""}`}
        onMouseDown={(e) => {
          if (e.target.closest(".sym-folder-select-chevron")) return;
          e.preventDefault();
          onSelectMask(mask);
        }}
      >
        <span
          className={`sym-folder-select-chevron${hasChildren ? "" : " empty"}`}
          onMouseDown={(e) => {
            e.preventDefault();
            e.stopPropagation();
            if (hasChildren) togglePath(node.path ?? "");
          }}
        >
          {hasChildren ? (open ? "▼" : "▶") : ""}
        </span>
        <span className="sym-clone-folder-icon" aria-hidden="true" />
        <span className="sym-folder-select-label">{node.label}\\*</span>
      </div>
      {hasChildren && open && (
        <ul>
          {node.children.map((child) =>
            child.kind === "folder" ? (
              <PathFolderRow
                key={child.key}
                node={child}
                depth={depth + 1}
                value={value}
                onSelectMask={onSelectMask}
                onSelectPath={onSelectPath}
                openPaths={openPaths}
                togglePath={togglePath}
              />
            ) : (
              <PathSymbolRow
                key={child.key}
                node={child}
                depth={depth + 1}
                value={value}
                onSelectPath={onSelectPath}
              />
            ),
          )}
        </ul>
      )}
    </li>
  );
}

function PathSymbolRow({ node, depth, value, onSelectPath }) {
  const selected = node.path === value;
  return (
    <li className="sym-folder-select-branch">
      <div
        className={`sym-folder-select-row sym-folder-select-depth-${Math.min(depth, 6)} sym-symbol-select-row${selected ? " sel" : ""}`}
        onMouseDown={(e) => {
          e.preventDefault();
          onSelectPath(node.path || node.symbol);
        }}
      >
        <span className="sym-folder-select-chevron empty" aria-hidden="true" />
        <span className="sym-coin-icon" aria-hidden="true" />
        <span className="sym-folder-select-label">{node.label}</span>
      </div>
    </li>
  );
}

function FilterRow({ row, value, onSelect }) {
  const selected = row.value === value;
  return (
    <li className="sym-folder-select-branch">
      <div
        className={`sym-folder-select-row sym-folder-select-depth-0${selected ? " sel" : ""}`}
        onMouseDown={(e) => {
          e.preventDefault();
          onSelect(row.value);
        }}
      >
        <span className="sym-folder-select-chevron empty" aria-hidden="true" />
        {row.kind === "symbol" ? (
          <span className="sym-coin-icon" aria-hidden="true" />
        ) : (
          <span className="sym-clone-folder-icon" aria-hidden="true" />
        )}
        <span className="sym-folder-select-label" title={row.label}>
          {row.label}
        </span>
      </div>
    </li>
  );
}

/**
 * Group-symbol path mask picker: `*`, folder\\*, or a full symbol path.
 */
export function SymbolPathSelectField({ label = "Symbol", value, onChange, disabled, className = "" }) {
  const { symbols } = useSymbols();
  const { folders } = useSymbolFolders();
  const [open, setOpen] = useState(false);
  const [raw, setRaw] = useState(null);
  const [pos, setPos] = useState(null);
  const [openPaths, setOpenPaths] = useState(() => new Set([""]));
  const rootRef = useRef(null);
  const btnRef = useRef(null);
  const menuRef = useRef(null);
  const openRef = useRef(false);
  const suppressBlurRef = useRef(false);
  const announceOpen = useExclusiveDropdown(open, setOpen);

  const emptyFolders = useMemo(() => emptyFolderPathsToMap(folders), [folders]);
  const tree = useMemo(
    () => buildSymbolSelectTree(symbols, emptyFolders),
    [symbols, emptyFolders],
  );

  const headerItems = useMemo(() => [{ value: "*", label: "*" }], []);
  const display = raw ?? (value == null || value === "" ? "*" : String(value));
  const filtering = raw !== null;
  const filterRows = useMemo(() => {
    if (!filtering) return null;
    const rows = filterSymbolPickerRows(symbols, display, headerItems);
    const needle = display.trim().toLowerCase();
    if (!needle || needle === "*") return rows;
    const masks = new Set();
    for (const row of symbols) {
      const parts = String(row.path || "").split("\\").filter(Boolean);
      parts.pop();
      let acc = "";
      for (const part of parts) {
        acc = acc ? `${acc}\\${part}` : part;
        const mask = `${acc}\\*`;
        if (mask.toLowerCase().includes(needle)) masks.add(mask);
      }
    }
    const maskRows = [...masks].sort().map((mask) => ({
      kind: "header",
      key: `mask:${mask}`,
      value: mask,
      label: mask,
    }));
    return [...maskRows, ...rows];
  }, [filtering, symbols, display, headerItems]);

  useEffect(() => {
    openRef.current = open;
  }, [open]);

  useEffect(() => {
    if (!open || filtering) return;
    const expand = new Set([""]);
    const needle = String(value ?? "").trim();
    if (needle && needle !== "*") {
      const base = needle.endsWith("\\*") ? needle.slice(0, -2) : needle;
      let acc = "";
      for (const part of base.split("\\").filter(Boolean)) {
        acc = acc ? `${acc}\\${part}` : part;
        expand.add(acc);
      }
    }
    setOpenPaths(expand);
  }, [open, filtering, value, tree]);

  useEffect(() => {
    if (!open) return;
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
  }, [open]);

  function commit(text) {
    const next = String(text ?? "").trim();
    onChange?.(next === "" ? "*" : next);
    setRaw(null);
  }

  function pick(next) {
    onChange?.(next == null || next === "" ? "*" : String(next));
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

  function togglePath(path) {
    setOpenPaths((prev) => {
      const next = new Set(prev);
      if (next.has(path)) next.delete(path);
      else next.add(path);
      return next;
    });
  }

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
              aria-label="Open symbol path tree"
              disabled={disabled}
              tabIndex={-1}
            />
          </span>
        </div>
      </div>
      {open &&
        pos &&
        createPortal(
          <ul
            ref={menuRef}
            className="sym-folder-select-menu prop-select-menu-front"
            style={{ position: "fixed", top: pos.top, left: pos.left, minWidth: pos.width }}
          >
            {filtering ? (
              (filterRows?.length ?? 0) > 0 ? (
                filterRows.map((row) => (
                  <FilterRow key={row.key} row={row} value={value} onSelect={pick} />
                ))
              ) : (
                <li className="sym-folder-select-row sym-folder-select-depth-0 sym-symbol-select-empty">
                  No matches
                </li>
              )
            ) : (
              <>
                {headerItems.map((item) => (
                  <FilterRow
                    key={`hdr:${item.value}`}
                    row={{ kind: "header", value: item.value, label: item.label }}
                    value={value}
                    onSelect={pick}
                  />
                ))}
                {headerItems.length > 0 && tree.length > 0 && (
                  <li className="sym-symbol-select-divider" aria-hidden="true" />
                )}
                {tree.map((node) =>
                  node.kind === "folder" ? (
                    <PathFolderRow
                      key={node.key}
                      node={node}
                      depth={0}
                      value={value}
                      onSelectMask={pick}
                      onSelectPath={pick}
                      openPaths={openPaths}
                      togglePath={togglePath}
                    />
                  ) : (
                    <PathSymbolRow
                      key={node.key}
                      node={node}
                      depth={0}
                      value={value}
                      onSelectPath={pick}
                    />
                  ),
                )}
                {!tree.length && (
                  <li className="sym-folder-select-row sym-folder-select-depth-0 sym-symbol-select-empty">
                    No symbols
                  </li>
                )}
              </>
            )}
          </ul>,
          document.body,
        )}
    </>
  );
}
