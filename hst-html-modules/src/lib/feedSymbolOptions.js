import {
  buildFolderTree,
  sortedTreeChildKeys,
} from "./symbolTree.js";

/**
 * Build PropSelect options for data feed Symbol / Group masks.
 * Includes *, folder masks (Forex\\Majors\\*), and individual symbols.
 */
export function buildFeedSymbolOptions(symbols = [], extraFolderPaths = []) {
  const options = [];
  const seen = new Set();

  function add(value, label) {
    const v = value == null ? "" : String(value);
    if (!v || seen.has(v)) return;
    seen.add(v);
    options.push({ value: v, label: label || v });
  }

  add("*", "*");

  const root = buildFolderTree(symbols, extraFolderPaths);

  function walkFolders(node) {
    if (!node.isRoot && node.path) {
      add(`${node.path}\\*`, `${node.path}\\*`);
    }
    for (const key of sortedTreeChildKeys(node)) {
      walkFolders(node.children[key]);
    }
  }
  walkFolders(root);

  const sorted = [...(symbols || [])].sort((a, b) =>
    String(a.symbol || "").localeCompare(String(b.symbol || ""), undefined, {
      sensitivity: "base",
    })
  );

  for (const row of sorted) {
    if (row.symbol) add(row.symbol, row.symbol);
  }

  return options;
}

/** Ensure the current cell value appears in the dropdown even if not in catalog. */
export function feedSymbolOptionsWithValue(selectOptions, value) {
  const opts = selectOptions || [];
  const current = value == null ? "" : String(value);
  if (!current || opts.some((o) => String(o.value) === current)) return opts;
  return [...opts, { value: current, label: current }];
}
