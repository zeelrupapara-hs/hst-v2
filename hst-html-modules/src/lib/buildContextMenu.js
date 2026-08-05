/** @typedef {import('../components/ui/ContextMenu.jsx').ContextMenuItem} ContextMenuItem */

/**
 * Resolve a menu definition into ContextMenu items.
 * Definition entries support:
 * - { type: 'separator' }
 * - { action: 'add', label, iconId, shortcut, disabledWhen: (ctx)=>bool, ... }
 * - { children: Definition[] | (ctx)=>Definition[] }
 * - function(ctx) => item | separator | null
 *
 * @param {Array<object|function>} definition
 * @param {object} ctx — handlers, state, helpers
 * @returns {ContextMenuItem[]}
 */
export function buildContextMenu(definition, ctx) {
  const items = [];
  for (const entry of definition || []) {
    const resolved = resolveMenuEntry(entry, ctx);
    if (resolved == null) continue;
    if (Array.isArray(resolved)) {
      items.push(...resolved);
    } else {
      items.push(resolved);
    }
  }
  return items;
}

function resolveMenuEntry(entry, ctx) {
  if (typeof entry === "function") {
    return resolveMenuEntry(entry(ctx), ctx);
  }
  if (!entry || entry.hidden?.(ctx)) return null;
  if (entry.type === "separator") {
    return { type: "separator" };
  }

  const item = { ...entry };
  delete item.action;
  delete item.disabledWhen;
  delete item.checkedWhen;
  delete item.hidden;
  delete item.childrenDef;

  if (entry.disabledWhen) {
    item.disabled = !!entry.disabledWhen(ctx);
  }
  if (entry.check && entry.checkedWhen) {
    item.checked = !!entry.checkedWhen(ctx);
  }
  if (entry.action && ctx.handlers?.[entry.action]) {
    item.onClick = () =>
      ctx.handlers[entry.action]({ ...ctx, item: entry, meta: entry.meta });
  }
  if (entry.childrenDef) {
    const childDef =
      typeof entry.childrenDef === "function"
        ? entry.childrenDef(ctx)
        : entry.childrenDef;
    item.children = () => buildContextMenu(childDef, ctx);
  }
  return item;
}

export async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand("copy");
      return true;
    } catch {
      return false;
    } finally {
      document.body.removeChild(ta);
    }
  }
}

export const MENU_SEP = { type: "separator" };

/** Shared config-list items (Add/Edit/Delete/Move/Sort/Export/Import/Find/View). */
export const CONFIG_LIST_CORE = [
  {
    id: "add",
    label: "Add",
    iconId: "add",
    shortcut: "Ctrl+N",
    action: "add",
  },
  {
    id: "edit",
    label: "Edit",
    iconId: "edit",
    shortcut: "Ctrl+U",
    action: "edit",
    disabledWhen: (ctx) => !ctx.hasSelection,
  },
  {
    id: "delete",
    label: "Delete",
    iconId: "delete",
    shortcut: "Ctrl+D",
    action: "delete",
    disabledWhen: (ctx) => !ctx.hasSelection,
  },
  MENU_SEP,
  {
    id: "move-up",
    label: "Move Up",
    icon: "↑",
    action: "moveUp",
    disabledWhen: (ctx) => !ctx.singleSelection || ctx.columnSort,
  },
  {
    id: "move-down",
    label: "Move Down",
    icon: "↓",
    action: "moveDown",
    disabledWhen: (ctx) => !ctx.singleSelection || ctx.columnSort,
  },
  {
    id: "sort",
    label: "Sort Alphabetically",
    icon: "A↓",
    action: "sort",
  },
];

export const AUTOMATION_ITEMS = [
  MENU_SEP,
  {
    id: "automation-triggers",
    label: "Automation Triggers",
    disabled: true,
    childrenDef: [{ label: "(none)", disabled: true }],
  },
  {
    id: "automation-actions",
    label: "Automation Actions",
    disabled: true,
    childrenDef: [{ label: "(none)", disabled: true }],
  },
];

export const EXPORT_IMPORT_ITEMS = [
  MENU_SEP,
  {
    id: "export",
    label: "Export to File",
    icon: "⭳",
    action: "exportFile",
    disabledWhen: (ctx) => !ctx.hasSelection,
  },
  {
    id: "import-file",
    label: "Import from File",
    icon: "⭱",
    action: "importFile",
  },
];

export const VIEW_TOGGLE_ITEMS = (columnsDef) => [
  MENU_SEP,
  {
    id: "auto-arrange",
    label: "Auto Arrange",
    check: true,
    checkedWhen: (ctx) => ctx.view?.autoArrange !== false,
    action: "toggleAutoArrange",
  },
  {
    id: "grid",
    label: "Grid",
    check: true,
    checkedWhen: (ctx) => ctx.view?.grid !== false,
    action: "toggleGrid",
  },
  {
    id: "columns",
    label: "Columns",
    childrenDef: () =>
      (columnsDef || []).map((col) => ({
        id: `col-${col.id}`,
        label: col.label,
        check: true,
        checkedWhen: (ctx) => !ctx.view?.hiddenCols?.[col.id],
        action: "toggleColumn",
        meta: { columnId: col.id },
      })),
  },
];
