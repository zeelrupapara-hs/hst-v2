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

/** Turn persisted empty-folder paths into the parent→names map the navigator uses. */
export function emptyFolderPathsToMap(paths) {
  const map = {};
  for (const path of paths || []) {
    const parts = path.split("\\").filter(Boolean);
    if (!parts.length) continue;
    const name = parts.pop();
    const parent = parts.join("\\");
    map[parent] ??= [];
    if (!map[parent].includes(name)) map[parent].push(name);
  }
  for (const key of Object.keys(map)) {
    map[key].sort((a, b) => a.localeCompare(b, undefined, { sensitivity: "base" }));
  }
  return map;
}

/** Last segment of a backslash folder path. */
export function folderLeafName(path) {
  if (!path) return "";
  const i = path.lastIndexOf("\\");
  return i >= 0 ? path.slice(i + 1) : path;
}

/** Rebuild a full path after renaming the leaf segment. */
export function folderWithLeafName(path, leaf) {
  const i = path.lastIndexOf("\\");
  const parent = i >= 0 ? path.slice(0, i) : "";
  return parent ? `${parent}\\${leaf}` : leaf;
}

/** @param {Array<{symbol?: string, path?: string}>} symbols */
export function indexSymbolsByFolder(symbols) {
  /** @type {Record<string, Array<{symbol?: string, path?: string}>>} */
  const byFolder = {};
  for (const row of symbols || []) {
    const folder = symbolFolder(row.path || "");
    byFolder[folder] ??= [];
    byFolder[folder].push(row);
  }
  for (const key of Object.keys(byFolder)) {
    byFolder[key].sort((a, b) =>
      (a.symbol || "").localeCompare(b.symbol || "", undefined, { sensitivity: "base" }),
    );
  }
  return byFolder;
}

/**
 * MT5-style folder tree with symbol leaves for Basis/Source combos.
 * @param {Array<{symbol?: string, path?: string}>} symbols
 * @param {Record<string, string[]>} emptyFolders
 */
export function buildSymbolSelectTree(symbols, emptyFolders = {}) {
  const byFolder = indexSymbolsByFolder(symbols);
  const root = buildFolderTree(symbols);

  /** @param {{path?: string, children?: Record<string, unknown>, isRoot?: boolean}} folderNode */
  function buildFolderNodes(folderNode) {
    const parentPath = folderNode.isRoot ? "" : folderNode.path;
    const seen = new Set(sortedTreeChildKeys(folderNode));
    const nodes = [];

    for (const key of sortedTreeChildKeys(folderNode)) {
      const child = folderNode.children[key];
      const subfolders = buildFolderNodes(child);
      const folderSymbols = (byFolder[child.path] || []).map((row) => ({
        kind: "symbol",
        key: `sym:${row.path || row.symbol}`,
        symbol: row.symbol,
        path: row.path || row.symbol,
        label: row.symbol,
      }));
      const children = [...subfolders, ...folderSymbols];
      if (children.length > 0 || (emptyFolders[child.path]?.length ?? 0) > 0) {
        nodes.push({
          kind: "folder",
          key: `dir:${child.path}`,
          path: child.path,
          label: child.name,
          children,
        });
      }
    }

    for (const name of emptyFolders[parentPath] || []) {
      if (seen.has(name)) continue;
      const path = parentPath ? `${parentPath}\\${name}` : name;
      nodes.push({
        kind: "folder",
        key: `dir:${path}`,
        path,
        label: name,
        children: buildFolderNodes({ path, children: {} }),
      });
    }

    return nodes.sort((a, b) => a.label.localeCompare(b.label, undefined, { sensitivity: "base" }));
  }

  const rootSymbols = (byFolder[""] || []).map((row) => ({
    kind: "symbol",
    key: `sym:${row.path || row.symbol}`,
    symbol: row.symbol,
    path: row.path || row.symbol,
    label: row.symbol,
  }));

  return [...buildFolderNodes(root), ...rootSymbols];
}

/** Flat symbol matches for combo filter — label shows full stored path. */
export function filterSymbolPickerRows(symbols, needle, headerItems = []) {
  const q = String(needle ?? "").trim().toLowerCase();
  if (!q) return null;

  const rows = [];
  const seen = new Set();

  for (const item of headerItems) {
    const value = item.value ?? item;
    const label = item.label ?? String(value);
    const text = String(label || value).toLowerCase();
    if (!text.includes(q)) continue;
    const key = `hdr:${value}`;
    if (seen.has(key)) continue;
    seen.add(key);
    rows.push({ kind: "header", key, value, label });
  }

  for (const row of symbols || []) {
    const sym = row.symbol || "";
    const path = row.path || sym;
    if (!sym.toLowerCase().includes(q) && !path.toLowerCase().includes(q)) continue;
    const key = `sym:${path}`;
    if (seen.has(key)) continue;
    seen.add(key);
    rows.push({
      kind: "symbol",
      key,
      value: sym,
      label: path,
      symbol: sym,
      path,
    });
  }

  return rows;
}
