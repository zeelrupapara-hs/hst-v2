import {
  annotateTreeCounts,
  buildFolderTree,
  sortedTreeChildKeys,
} from "./symbolTree.js";
import { SYMBOL_FOLDER_ICON } from "./icons.js";

/** Deep-clone static nav nodes before injecting dynamic children. */
export function cloneNav(nodes) {
  return (nodes || []).map((n) => ({
    ...n,
    children: n.children ? cloneNav(n.children) : undefined,
  }));
}

function folderNodesToNav(folderNode, indent) {
  return sortedTreeChildKeys(folderNode).map((key) => {
    const child = folderNode.children[key];
    const subKeys = sortedTreeChildKeys(child);
    const navNode = {
      moduleId: "symbols",
      folderPath: child.path,
      label: child.name,
      icon: SYMBOL_FOLDER_ICON,
      indent,
      count: child.count != null ? `(${child.count})` : undefined,
    };
    if (subKeys.length) {
      navNode.children = folderNodesToNav(child, indent + 1);
    }
    return navNode;
  });
}

/** Expand ancestors of the selected folder path. */
function markOpenAlongPath(node, activeFolder) {
  if (!activeFolder || !node.children?.length) return;
  for (const child of node.children) {
    if (child.folderPath === undefined) continue;
    const isOnPath =
      activeFolder === child.folderPath ||
      activeFolder.startsWith(`${child.folderPath}\\`);
    if (isOnPath) {
      node.open = true;
      child.open = true;
      markOpenAlongPath(child, activeFolder);
    }
  }
}

/**
 * Inject symbol folder hierarchy under the Symbols nav node (MT5 Administrator).
 * @param {Array} nav
 * @param {Array<{ path?: string }>} symbols
 * @param {string} [activeFolder]
 */
export function injectSymbolFolders(nav, symbols, activeFolder = "") {
  const root = annotateTreeCounts(buildFolderTree(symbols), symbols);

  function walk(nodes) {
    return nodes.map((node) => {
      if (node.moduleId === "symbols") {
        const merged = {
          ...node,
          folderPath: "",
          count: root.count != null ? `(${root.count})` : node.count,
          open: node.open !== false,
          children: folderNodesToNav(root, (node.indent ?? 0) + 1),
        };
        markOpenAlongPath(merged, activeFolder);
        return merged;
      }
      if (node.children) {
        return { ...node, children: walk(node.children) };
      }
      return node;
    });
  }

  return walk(nav);
}

/** Build a symbols module link with optional folder query. */
export function symbolNavLink(panel, folderPath = "") {
  const base = `/${panel}/symbols`;
  if (!folderPath) return base;
  return `${base}?folder=${encodeURIComponent(folderPath)}`;
}
