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

function SymbolTreeRow({ node, depth, value, onSelect, openPaths, togglePath }) {
  if (node.kind === "symbol") {
    const selected = node.symbol === value;
    return (
      <li className="sym-folder-select-branch">
        <div
          className={`sym-folder-select-row sym-folder-select-depth-${Math.min(depth, 6)} sym-symbol-select-row${selected ? " sel" : ""}`}
          onMouseDown={(e) => {
            e.preventDefault();
            onSelect(node.symbol);
          }}
        >
          <span className="sym-folder-select-chevron empty" aria-hidden="true" />
          <span className="sym-coin-icon" aria-hidden="true" />
          <span className="sym-folder-select-label">{node.label}</span>
        </div>
      </li>
    );
  }

  const hasChildren = node.children?.length > 0;
  const open = openPaths.has(node.path ?? "");
  const selected = false;

  return (
    <li className="sym-folder-select-branch">
      <div
        className={`sym-folder-select-row sym-folder-select-depth-${Math.min(depth, 6)}${selected ? " sel" : ""}`}
        onMouseDown={(e) => {
          if (e.target.closest(".sym-folder-select-chevron")) return;
          e.preventDefault();
          if (hasChildren) togglePath(node.path ?? "");
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
        <span className="sym-folder-select-label">{node.label}</span>
      </div>
      {hasChildren && open && (
        <ul>
          {node.children.map((child) => (
            <SymbolTreeRow
              key={child.key}
              node={child}
              depth={depth + 1}
              value={value}
              onSelect={onSelect}
              openPaths={openPaths}
              togglePath={togglePath}
            />
          ))}
        </ul>
      )}
    </li>
  );
}

function FilterRow({ row, value, onSelect }) {
  const selected = String(row.value) === String(value);
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
          <span className="sym-folder-select-label" style={{ width: 0, margin: 0 }} aria-hidden="true" />
        )}
        <span className="sym-folder-select-label" title={row.label}>
          {row.label}
        </span>
      </div>
    </li>
  );
}

/**
 * MT5-style editable combo whose list is the symbol folder tree (not a flat dump).
 * @param {{
 *   label: string,
 *   value: string,
 *   onChange: Function,
 *   headerItems?: Array<{value: string, label?: string}|string>,
 *   lockField?: Function,
 *   fieldKey?: string,
 * }} props
 */
export function SymbolTreeSelectField({
  label,
  value,
  onChange,
  headerItems = [],
  lockField,
  fieldKey,
}) {
  const locked = lockField?.(fieldKey);
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

  const normalizedHeader = useMemo(
    () =>
      headerItems.map((item) =>
        typeof item === "string" ? { value: item, label: item } : { value: item.value, label: item.label ?? item.value },
      ),
    [headerItems],
  );

  const display = raw ?? (value == null || value === "" ? "" : String(value));
  const filtering = raw !== null;
  const filterRows = useMemo(
    () => (filtering ? filterSymbolPickerRows(symbols, display, normalizedHeader) : null),
    [filtering, symbols, display, normalizedHeader],
  );

  useEffect(() => {
    if (!open || filtering) return;
    const expand = new Set([""]);
    // Expand folders that contain the current basis/source value.
    const needle = String(value ?? "").trim();
    if (needle) {
      for (const row of symbols) {
        if (row.symbol !== needle) continue;
        const parts = (row.path || "").split("\\").filter(Boolean);
        parts.pop();
        let acc = "";
        for (const part of parts) {
          acc = acc ? `${acc}\\${part}` : part;
          expand.add(acc);
        }
        break;
      }
    }
    // Show first-level folders expanded so the tree is usable immediately.
    for (const node of tree) {
      if (node.kind === "folder" && node.path) expand.add(node.path);
    }
    setOpenPaths(expand);
  }, [open, filtering, symbols, value, tree]);

  useEffect(() => {
    openRef.current = open;
  }, [open]);

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
    onChange?.(String(text ?? ""));
    setRaw(null);
  }

  function openMenu(e) {
    if (locked) return;
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

  function pick(next) {
    onChange?.(next == null || next === "" ? "" : String(next));
    setRaw(null);
    setOpen(false);
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
        className={`prop-select prop-select-fill editable-select${locked ? " prop-select-disabled" : ""}`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="prop-select-box editable-select-box">
          <input
            type="text"
            className="editable-select-input"
            value={display}
            disabled={locked}
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
              aria-label="Open symbol tree"
              disabled={locked}
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
                {normalizedHeader.map((item) => (
                  <FilterRow
                    key={`hdr:${item.value}`}
                    row={{ kind: "header", value: item.value, label: item.label }}
                    value={value}
                    onSelect={pick}
                  />
                ))}
                {normalizedHeader.length > 0 && tree.length > 0 && (
                  <li className="sym-symbol-select-divider" aria-hidden="true" />
                )}
                {tree.map((node) => (
                  <SymbolTreeRow
                    key={node.key}
                    node={node}
                    depth={0}
                    value={value}
                    onSelect={pick}
                    openPaths={openPaths}
                    togglePath={togglePath}
                  />
                ))}
                {!normalizedHeader.length && !tree.length && (
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
