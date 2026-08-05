import { useMemo, useState } from "react";
import {
  annotateTreeCounts,
  buildFolderTree,
  sortedTreeChildKeys,
} from "../../lib/symbolTree.js";

function TreeNode({ node, depth, expanded, selectedPath, onToggle, onSelect, onContextMenu }) {
  const keys = sortedTreeChildKeys(node);
  const hasKids = keys.length > 0;
  const path = node.path ?? "";
  const isExpanded = expanded[path] !== false;

  return (
    <>
      <div
        className={`sym-tree-row${selectedPath === path ? " sel" : ""}`}
        style={{ paddingLeft: `${4 + depth * 14}px` }}
        onClick={() => onSelect(path)}
        onContextMenu={(e) => {
          e.preventDefault();
          onContextMenu?.(e, path);
        }}
      >
        {hasKids ? (
          <span
            className="sym-tree-toggle"
            onClick={(e) => {
              e.stopPropagation();
              onToggle(path, isExpanded);
            }}
          >
            {isExpanded ? "▼" : "▶"}
          </span>
        ) : (
          <span className="sym-tree-toggle sym-tree-spacer" />
        )}
        <span
          className={node.isRoot ? "sym-tree-root-icon" : "sym-tree-folder"}
        />
        <span className="sym-tree-label">
          {node.name}
          {node.count != null && node.count > 0 && !node.isRoot
            ? ` (${node.count})`
            : ""}
          {node.isRoot && node.count != null ? ` (${node.count})` : ""}
        </span>
      </div>
      {hasKids &&
        isExpanded &&
        keys.map((key) => (
          <TreeNode
            key={node.children[key].path}
            node={node.children[key]}
            depth={depth + 1}
            expanded={expanded}
            selectedPath={selectedPath}
            onToggle={onToggle}
            onSelect={onSelect}
            onContextMenu={onContextMenu}
          />
        ))}
    </>
  );
}

export function SymbolPathTree({
  symbols,
  selectedPath,
  onSelectPath,
  onContextMenu,
}) {
  const [expanded, setExpanded] = useState({ "": true });

  const tree = useMemo(() => {
    const root = buildFolderTree(symbols);
    return annotateTreeCounts(root, symbols);
  }, [symbols]);

  function handleToggle(path, isOpen) {
    setExpanded((prev) => ({ ...prev, [path]: !isOpen }));
  }

  return (
    <div className="module-split-nav symbol-tree">
      <TreeNode
        node={tree}
        depth={0}
        expanded={expanded}
        selectedPath={selectedPath}
        onToggle={handleToggle}
        onSelect={onSelectPath}
        onContextMenu={onContextMenu}
      />
    </div>
  );
}
