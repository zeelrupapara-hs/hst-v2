import {
  AUTOMATION_ITEMS,
  buildContextMenu,
  CONFIG_LIST_CORE,
  copyToClipboard,
  EXPORT_IMPORT_ITEMS,
  MENU_SEP,
  VIEW_TOGGLE_ITEMS,
} from "./buildContextMenu.js";

export const DATAFEED_TABLE_COLUMNS = [
  { id: "name", label: "Name" },
  { id: "source", label: "Source" },
  { id: "server", label: "Server" },
  { id: "symbols", label: "Symbols" },
  { id: "last_active", label: "Last active" },
  { id: "state", label: "State" },
];

const DATAFEED_LIST_MENU_DEF = [
  ...CONFIG_LIST_CORE,
  MENU_SEP,
  {
    id: "enable",
    label: "Enable",
    icon: "✓",
    action: "enable",
    disabledWhen: (ctx) => !ctx.hasSelection,
  },
  {
    id: "disable",
    label: "Disable",
    icon: "✕",
    action: "disable",
    disabledWhen: (ctx) => !ctx.hasSelection,
  },
  ...AUTOMATION_ITEMS,
  ...EXPORT_IMPORT_ITEMS,
  MENU_SEP,
  {
    id: "journal",
    label: "Journal",
    iconId: "journal",
    action: "journal",
    disabledWhen: (ctx) => !ctx.singleSelection,
  },
  {
    id: "find",
    label: "Find",
    icon: "🔍",
    shortcut: "Ctrl+F",
    action: "find",
  },
  ...VIEW_TOGGLE_ITEMS(DATAFEED_TABLE_COLUMNS),
];

/** MT5 Data Feeds list context menu — admin_feeds.htm#context */
export function buildDatafeedTableMenu(ctx) {
  return buildContextMenu(DATAFEED_LIST_MENU_DEF, ctx);
}
