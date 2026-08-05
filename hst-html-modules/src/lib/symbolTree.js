/**
 * Admin Symbols navigation tree — folders only (MT5 Administrator).
 *
 * API stores path as `folder\subfolder\SYMBOL`. There is no admin tree endpoint;
 * build the left nav from flat GET /api/v1/symbols, matching symbolFolder() in
 * hst-server/internal/server/v1/admin/symbols.go and folder nesting in
 * hst-server/internal/server/v1/trader/symbols.go symbolTree().
 */

/** Drop the symbol name from a stored path. */
export function symbolFolder(path) {
  if (!path) return "";
  const i = path.lastIndexOf("\\");
  return i >= 0 ? path.slice(0, i) : "";
}

/** Top-level group for the Type column (first path segment). */
export function topType(path) {
  if (!path) return "—";
  const folder = symbolFolder(path) || path;
  return folder.split("\\")[0] || "—";
}

/**
 * @typedef {{ name: string, path: string, children: Record<string, SymbolTreeNode>, isRoot?: boolean, count?: number }} SymbolTreeNode
 */

/**
 * Build folder-only tree from a flat symbol list.
 * Does NOT add symbol names as nodes (admin MT5: symbols appear in the table only).
 *
 * @param {Array<{ path?: string }>} symbols
 * @param {string[]} [extraFolderPaths] empty groups persisted without symbols yet
 * @returns {SymbolTreeNode}
 */
export function buildFolderTree(symbols, extraFolderPaths = []) {
  const root = { name: "All symbols", path: "", children: {}, isRoot: true };

  function addFolderPath(folder) {
    if (!folder) return;
    const parts = folder.split("\\").filter(Boolean);
    let node = root;
    let acc = "";

    for (const part of parts) {
      acc = acc ? `${acc}\\${part}` : part;
      if (!node.children[part]) {
        node.children[part] = { name: part, path: acc, children: {} };
      }
      node = node.children[part];
    }
  }

  for (const row of symbols || []) {
    addFolderPath(symbolFolder(row.path || ""));
  }
  for (const folder of extraFolderPaths || []) {
    addFolderPath(folder);
  }

  return root;
}

/**
 * Count symbols whose stored path falls under a folder (prefix match).
 * @param {Array<{ path?: string }>} symbols
 * @param {string} folderPath empty = all
 */
export function countSymbolsInFolder(symbols, folderPath) {
  if (!folderPath) return symbols.length;
  const prefix = `${folderPath}\\`;
  return symbols.filter((r) => {
    const p = r.path || "";
    return p.startsWith(prefix) || symbolFolder(p) === folderPath;
  }).length;
}

/**
 * Filter flat symbol list for the selected tree folder.
 * @param {Array<{ path?: string }>} symbols
 * @param {string} folderPath empty string = all symbols
 */
export function filterSymbolsByFolder(symbols, folderPath) {
  if (!folderPath) return symbols.slice();
  const prefix = `${folderPath}\\`;
  return symbols.filter((r) => {
    const p = r.path || "";
    return p.startsWith(prefix);
  });
}

/** Sorted child keys: folders alphabetically (matches server sortNodes). */
export function sortedTreeChildKeys(node) {
  return Object.keys(node.children || {}).sort((a, b) =>
    a.localeCompare(b, undefined, { sensitivity: "base" })
  );
}

/**
 * Annotate each folder node with direct+nested symbol count (MT5 shows counts).
 * @param {SymbolTreeNode} node
 * @param {Array<{ path?: string }>} symbols
 */
export function annotateTreeCounts(node, symbols) {
  node.count = countSymbolsInFolder(symbols, node.isRoot ? "" : node.path);
  for (const key of sortedTreeChildKeys(node)) {
    annotateTreeCounts(node.children[key], symbols);
  }
  return node;
}
