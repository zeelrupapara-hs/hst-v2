import { cloneNav } from "./symbolNavTree.js";

export const GROUP_FOLDER_ICON = "groups";

/**
 * A group name is a path: demo\forex-hedge-usd-02. Nothing stores the hierarchy, so the folders
 * are read back out of the paths. A node can be a folder and a group at once, and a folder can
 * exist with no group of its own, which is the same rule the server applies.
 *
 * @param {Array<{ group?: string }>} groups flat list of groups
 */
export function buildGroupTree(groups) {
  const root = { name: "", path: "", children: {}, exists: false, count: 0 };

  for (const g of groups || []) {
    const path = String(g?.group || "").trim();
    if (!path) continue;

    let node = root;
    let walked = "";

    for (const part of path.split("\\")) {
      if (!part) continue;
      walked = walked ? `${walked}\\${part}` : part;
      node.children[part] ??= {
        name: part,
        path: walked,
        children: {},
        exists: false,
        count: 0,
      };
      node = node.children[part];
    }

    node.exists = true;
  }

  // a folder counts every group beneath it, and itself when a group sits at its own path
  const count = (node) => {
    node.count = node.exists ? 1 : 0;
    for (const key of Object.keys(node.children)) node.count += count(node.children[key]);
    return node.count;
  };
  count(root);

  return root;
}

const sortedKeys = (node) => Object.keys(node.children).sort((a, b) => a.localeCompare(b));

function toNav(folder, indent) {
  return sortedKeys(folder).map((key) => {
    const child = folder.children[key];
    const node = {
      moduleId: "groups",
      groupPath: child.path,
      label: child.name,
      icon: GROUP_FOLDER_ICON,
      indent,
      count: child.count ? `(${child.count})` : undefined,
    };

    if (sortedKeys(child).length) node.children = toNav(child, indent + 1);

    return node;
  });
}

/** Expand the folders above whichever one is open. */
function openAlongPath(node, activePath) {
  if (!activePath || !node.children?.length) return;

  for (const child of node.children) {
    if (child.groupPath === undefined) continue;

    if (activePath === child.groupPath || activePath.startsWith(`${child.groupPath}\\`)) {
      node.open = true;
      child.open = true;
      openAlongPath(child, activePath);
    }
  }
}

/** Hang the group folders under the Groups node, the way the symbol folders hang under Symbols. */
export function injectGroupFolders(nav, groups, activePath = "") {
  const root = buildGroupTree(groups);

  const walk = (nodes) =>
    nodes.map((node) => {
      if (node.moduleId === "groups") {
        const merged = {
          ...node,
          groupPath: "",
          count: root.count ? `(${root.count})` : node.count,
          open: node.open !== false,
          children: toNav(root, (node.indent ?? 0) + 1),
        };
        openAlongPath(merged, activePath);
        return merged;
      }

      if (node.children) return { ...node, children: walk(node.children) };

      return node;
    });

  return walk(nav);
}

/** A groups module link, narrowed to one folder. */
export function groupNavLink(panel, groupPath = "") {
  const base = `/${panel}/groups`;

  return groupPath ? `${base}?folder=${encodeURIComponent(groupPath)}` : base;
}

export { cloneNav };
