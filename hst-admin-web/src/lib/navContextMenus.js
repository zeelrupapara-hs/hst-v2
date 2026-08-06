/**
 * Navigator context menus — the tree's right-click menus differ from list menus.
 * Symbols: admin_symbols.htm#context (Add group, export/import, refresh).
 */

/**
 * @param {{
 *   canEdit?: boolean,
 *   isFolder?: boolean,
 *   onAdd?: Function,
 *   onEdit?: Function,
 *   onDelete?: Function,
 *   onRefresh?: Function,
 * }} handlers
 */
export function symbolsNavMenuItems({
  canEdit = true,
  isFolder = false,
  onAdd,
  onEdit,
  onDelete,
  onRefresh,
} = {}) {
  const head = isFolder
    ? [
        { label: "Add", icon: "add", shortcut: "Ctrl+N", disabled: !canEdit, onClick: onAdd },
        { label: "Edit", icon: "edit", shortcut: "Ctrl+U", disabled: !canEdit, onClick: onEdit },
        { label: "Delete", icon: "delete", shortcut: "Del", disabled: !canEdit, onClick: onDelete },
      ]
    : [{ label: "Add", icon: "add", shortcut: "Ctrl+N", disabled: !canEdit, onClick: onAdd }];

  return [
    ...head,
    "sep",
    {
      label: "Automation Triggers",
      disabled: true,
      items: [{ label: "(none)", disabled: true }],
    },
    {
      label: "Automation Actions",
      disabled: true,
      items: [{ label: "(none)", disabled: true }],
    },
    "sep",
    { label: "Export Configurations to File", disabled: true },
    { label: "Import Configurations from File", disabled: true },
    { label: "Import Configurations from Server", disabled: true },
    "sep",
    { label: "Refresh", icon: "refresh", shortcut: "F5", onClick: onRefresh },
  ];
}

/** Folder path carried by a symbols navigator node ("" = root under Symbols). */
export function symbolsNavFolderPath(node) {
  if (!node) return null;
  if (node.key === "symbols") return "";
  if (node.key?.startsWith("symfolder:")) return node.key.slice("symfolder:".length);
  return null;
}

export function isSymbolsNavNode(node) {
  return symbolsNavFolderPath(node) !== null;
}

export function isSymbolFolderNode(node) {
  return node?.key?.startsWith("symfolder:");
}
