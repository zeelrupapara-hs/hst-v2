import { useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useSymbols } from "@/hooks/useSymbols.js";
import { useSymbolFolders } from "@/hooks/useSymbolFolders.js";
import {
  buildFolderTree,
  emptyFolderPathsToMap,
  folderLeafName,
  sortedTreeChildKeys,
} from "@/lib/symbolTree.js";

function folderLabel(path) {
  if (!path) return "Symbols";
  return folderLeafName(path);
}

function buildSelectTree(folderNode, emptyFolders) {
  const parentPath = folderNode.isRoot ? "" : folderNode.path;
  const seen = new Set(sortedTreeChildKeys(folderNode));

  const nodes = sortedTreeChildKeys(folderNode).map((key) => {
    const child = folderNode.children[key];
    const hasKids =
      sortedTreeChildKeys(child).length > 0 || (emptyFolders[child.path]?.length ?? 0) > 0;
    return {
      path: child.path,
      label: child.name,
      children: hasKids ? buildSelectTree(child, emptyFolders) : [],
    };
  });

  for (const name of emptyFolders[parentPath] || []) {
    if (seen.has(name)) continue;
    const path = parentPath ? `${parentPath}\\${name}` : name;
    nodes.push({
      path,
      label: name,
      children: buildSelectTree({ path, children: {} }, emptyFolders),
    });
  }

  return nodes.sort((a, b) => a.label.localeCompare(b.label, undefined, { sensitivity: "base" }));
}

function FolderTreeRow({ node, depth, value, onSelect, openPaths, togglePath }) {
  const hasChildren = node.children?.length > 0;
  const open = openPaths.has(node.path);
  const selected = node.path === value;

  return (
    <li className="sym-folder-select-branch">
      <div
        className={`sym-folder-select-row sym-folder-select-depth-${Math.min(depth, 6)}${selected ? " sel" : ""}`}
        onMouseDown={(e) => {
          e.preventDefault();
          onSelect(node.path);
        }}
      >
        <span
          className={`sym-folder-select-chevron${hasChildren ? "" : " empty"}`}
          onMouseDown={(e) => {
            e.preventDefault();
            e.stopPropagation();
            if (hasChildren) togglePath(node.path);
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
            <FolderTreeRow
              key={child.path}
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

/**
 * MT5-style folder tree picker for symbol paths (backslash-separated, no symbol name).
 * @param {{value: string, onChange: Function, disabled?: boolean, className?: string}} props
 */
export function SymbolFolderSelect({ value, onChange, disabled, className = "" }) {
  const { symbols } = useSymbols();
  const { folders } = useSymbolFolders();
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState(null);
  const [openPaths, setOpenPaths] = useState(() => new Set());
  const rootRef = useRef(null);

  const emptyFolders = useMemo(() => emptyFolderPathsToMap(folders), [folders]);
  const tree = useMemo(() => {
    const root = buildFolderTree(symbols);
    return {
      path: "",
      label: "Symbols",
      children: buildSelectTree(root, emptyFolders),
    };
  }, [symbols, emptyFolders]);

  useEffect(() => {
    if (!open) return;
    const expand = new Set([""]);
    if (value) {
      let acc = "";
      for (const part of value.split("\\").filter(Boolean)) {
        acc = acc ? `${acc}\\${part}` : part;
        expand.add(acc);
      }
    }
    setOpenPaths(expand);
  }, [open, value]);

  useEffect(() => {
    if (!open) return;
    function close(e) {
      if (rootRef.current?.contains(e.target)) return;
      if (e.target.closest?.(".sym-folder-select-menu")) return;
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

  function openMenu(e) {
    if (disabled) return;
    e.stopPropagation();
    const r = rootRef.current.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: Math.max(r.width, 220) });
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

  const cls = ["prop-select", "prop-select-fill", disabled && "prop-select-disabled", className]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <div ref={rootRef} className={cls} onClick={(e) => e.stopPropagation()}>
        <div className="prop-select-box" onClick={openMenu}>
          <span className="prop-select-value sym-folder-select-value" title={value || "Symbols"}>
            <span className="sym-clone-folder-icon" aria-hidden="true" />
            {folderLabel(value)}
          </span>
          <button type="button" className="prop-select-btn" aria-label="Open folder list" disabled={disabled} onClick={openMenu} />
        </div>
      </div>
      {open &&
        pos &&
        createPortal(
          <ul
            className="sym-folder-select-menu"
            style={{ position: "fixed", top: pos.top, left: pos.left, minWidth: pos.width }}
          >
            <FolderTreeRow
              node={tree}
              depth={0}
              value={value}
              onSelect={(path) => {
                onChange?.(path);
                setOpen(false);
              }}
              openPaths={openPaths}
              togglePath={togglePath}
            />
          </ul>,
          document.body,
        )}
    </>
  );
}
