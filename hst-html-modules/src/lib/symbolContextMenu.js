/** MT5 Symbols context menus — admin_symbols.htm#context */
import {
  AUTOMATION_ITEMS,
  buildContextMenu,
  CONFIG_LIST_CORE,
  copyToClipboard,
  EXPORT_IMPORT_ITEMS,
  MENU_SEP,
  VIEW_TOGGLE_ITEMS,
} from "./buildContextMenu.js";

export const SYMBOL_TABLE_COLUMNS = [
  { id: "symbol", label: "Symbol" },
  { id: "description", label: "Description" },
  { id: "digits", label: "Digits" },
  { id: "type", label: "Type" },
  { id: "execution", label: "Execution" },
];

function copyLines(rows) {
  return rows
    .map((r) =>
      [r.symbol, r.description || "", r.digits, topType(r.path), r.exec_mode]
        .join("\t")
    )
    .join("\n");
}

function topType(path) {
  if (!path) return "—";
  return path.split("\\")[0] || "—";
}

function focusSymbolSearch() {
  const el = document.querySelector("[data-symbol-search]");
  if (el) {
    el.focus();
    el.select?.();
    return true;
  }
  return false;
}

const SYMBOL_TABLE_MENU_DEF = [
  ...CONFIG_LIST_CORE.slice(0, 1),
  {
    id: "add-copy",
    label: "Add Copy",
    iconId: "add",
    shortcut: "Ctrl+M",
    action: "addCopy",
    disabledWhen: (ctx) => !ctx.hasSelection,
  },
  ...CONFIG_LIST_CORE.slice(1),
  ...AUTOMATION_ITEMS,
  MENU_SEP,
  {
    id: "copy-as",
    label: "Copy As",
    disabledWhen: (ctx) => !ctx.hasSelection,
    childrenDef: (ctx) => [
      {
        label: "Lines",
        icon: "⎘",
        action: "copyLines",
      },
      {
        label: "List of Symbols",
        action: "copySymbolNames",
      },
    ],
  },
  ...EXPORT_IMPORT_ITEMS,
  {
    id: "import-server",
    label: "Import from Server",
    icon: "⭱",
    action: "importServer",
  },
  MENU_SEP,
  {
    id: "journal",
    label: "Journal",
    iconId: "journal",
    action: "journal",
    disabledWhen: (ctx) => !ctx.singleSelection,
  },
  {
    id: "charts",
    label: "Charts",
    iconId: "charts-ticks",
    action: "charts",
    disabledWhen: (ctx) => !ctx.singleSelection,
  },
  {
    id: "ticks",
    label: "Ticks",
    iconId: "charts-ticks",
    action: "ticks",
    disabledWhen: (ctx) => !ctx.singleSelection,
  },
  {
    id: "find",
    label: "Find",
    icon: "🔍",
    shortcut: "Ctrl+F",
    action: "find",
  },
  ...VIEW_TOGGLE_ITEMS(SYMBOL_TABLE_COLUMNS),
];

const SYMBOL_FOLDER_MENU_DEF = [
  ...CONFIG_LIST_CORE.slice(0, 3),
  MENU_SEP,
  {
    id: "sort",
    label: "Sort Alphabetically",
    icon: "A↓",
    action: "sort",
  },
  MENU_SEP,
  {
    id: "find",
    label: "Find",
    icon: "🔍",
    shortcut: "Ctrl+F",
    action: "findFolder",
  },
];

/** Context menu for symbol rows in the table (admin.symbols). */
export function buildSymbolTableMenu(ctx) {
  const handlers = {
    ...ctx.handlers,
    copyLines: async () => {
      await copyToClipboard(copyLines(ctx.selectedRows));
      ctx.toast?.(`Copied ${ctx.selectedRows.length} row(s)`);
    },
    copySymbolNames: async () => {
      await copyToClipboard(
        ctx.selectedRows.map((r) => r.symbol).filter(Boolean).join("\n")
      );
      ctx.toast?.("Copied symbol names");
    },
    find: () => {
      if (!focusSymbolSearch()) ctx.toast?.("Use the Search field above the table");
    },
  };
  return buildContextMenu(SYMBOL_TABLE_MENU_DEF, { ...ctx, handlers });
}

/** Context menu for symbol folder nodes in Navigator. */
export function buildSymbolFolderMenu(ctx) {
  const handlers = {
    ...ctx.handlers,
    findFolder: () => {
      if (!focusSymbolSearch()) {
        ctx.toast?.(
          `Open Symbols to search under ${ctx.folderLabel || ctx.folderPath || "All"}`
        );
      }
    },
  };
  return buildContextMenu(SYMBOL_FOLDER_MENU_DEF, { ...ctx, handlers });
}

export { copyToClipboard };
