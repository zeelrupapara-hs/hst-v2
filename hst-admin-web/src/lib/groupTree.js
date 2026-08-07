// A group name is a path (demo\forex-usd); folders are read back out of the paths.
// A node can be a folder and a group at once; the tree counts groups beneath each folder.

/** @param {Array<{group?: string, exists?: boolean}>} groups flat list */
export function buildGroupTree(groups) {
  const root = { name: "", path: "", children: {}, exists: false, count: 0 };

  for (const g of groups || []) {
    const path = String(g.group || "").trim();
    if (!path) continue;
    let node = root;
    let walked = "";
    for (const part of path.split("\\")) {
      if (!part) continue;
      walked = walked ? `${walked}\\${part}` : part;
      node.children[part] ??= { name: part, path: walked, children: {}, exists: false, count: 0 };
      node = node.children[part];
    }
    node.exists = true;
  }

  const count = (node) => {
    node.count = node.exists ? 1 : 0;
    for (const key of Object.keys(node.children)) node.count += count(node.children[key]);
    return node.count;
  };
  count(root);

  return root;
}

export const sortedGroupKeys = (node) =>
  Object.keys(node.children).sort((a, b) => a.localeCompare(b));

/** Groups whose path sits at or under a folder. */
export function filterGroupsByFolder(groups, folderPath) {
  const real = groups || [];
  if (!folderPath) return real;
  const prefix = `${folderPath}\\`;
  return real.filter((g) => g.group === folderPath || (g.group || "").startsWith(prefix));
}

/** All group records to remove when deleting a navigator section. */
export function groupsUnderFolder(groups, folderPath) {
  return filterGroupsByFolder(groups, folderPath);
}
