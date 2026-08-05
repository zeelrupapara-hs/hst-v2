// Symbol paths are `folder\subfolder\SYMBOL`; the navigator tree is folders only.

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

/** @param {Array<{path?: string}>} symbols */
export function buildFolderTree(symbols) {
  const root = { name: "All symbols", path: "", children: {}, isRoot: true };

  for (const row of symbols || []) {
    const folder = symbolFolder(row.path || "");
    if (!folder) continue;
    let node = root;
    let acc = "";
    for (const part of folder.split("\\").filter(Boolean)) {
      acc = acc ? `${acc}\\${part}` : part;
      node.children[part] ??= { name: part, path: acc, children: {} };
      node = node.children[part];
    }
  }

  return root;
}

export function countSymbolsInFolder(symbols, folderPath) {
  if (!folderPath) return symbols.length;
  const prefix = `${folderPath}\\`;
  return symbols.filter((r) => (r.path || "").startsWith(prefix)).length;
}

/** @param {string} folderPath empty string = all symbols */
export function filterSymbolsByFolder(symbols, folderPath) {
  if (!folderPath) return symbols.slice();
  const prefix = `${folderPath}\\`;
  return symbols.filter((r) => (r.path || "").startsWith(prefix));
}

export function sortedTreeChildKeys(node) {
  return Object.keys(node.children || {}).sort((a, b) =>
    a.localeCompare(b, undefined, { sensitivity: "base" }),
  );
}

export function annotateTreeCounts(node, symbols) {
  node.count = countSymbolsInFolder(symbols, node.isRoot ? "" : node.path);
  for (const key of sortedTreeChildKeys(node)) annotateTreeCounts(node.children[key], symbols);
  return node;
}
