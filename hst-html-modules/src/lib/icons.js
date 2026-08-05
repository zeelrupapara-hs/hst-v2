import manifest from "../../assets/icons/manifest.json";

/** Bump when regenerating icons (`npm run build-icons`). */
export const ICON_VERSION = 4;

const ICON_IDS = new Set(manifest.icons.map((entry) => entry.id));

/** MT5 Administrator / Manager module → generated icon id. */
export const MODULE_NAV_ICON = {
  time: "time",
  holidays: "holidays",
  leverages: "leverages",
  groups: "groups",
  clients: "clients-tree",
  managers: "managers-tree",
  users: "accounts-tree",
  positions: "positions",
  orders: "orders-tree",
  deals: "deals-tree",
  datafeeds: "datafeeds",
  routing: "routing",
  symbols: "symbols-tree",
  dealing: "dealing",
  balance: "balance",
  journal: "journal",
};

/** MT5 nav: all symbol folder nodes use the tree folder icon. */
export const SYMBOL_FOLDER_ICON = "symbols-tree";

/** Entity / table row icons from manifest. */
export const ENTITY_ICON = {
  administrator: "administrator",
  manager: "manager",
  client: "client",
  account: "account",
  order: "order",
  deal: "deal",
  "deal-out": "deal-out",
};

export function isGeneratedIcon(id) {
  return ICON_IDS.has(id);
}

export function normalizeIconId(id, fallback = "symbols-tree") {
  if (!id) return fallback;
  return ICON_IDS.has(id) ? id : fallback;
}

export function iconUrl(id) {
  const safe = normalizeIconId(id);
  return `/assets/icons/svg/${safe}.svg?v=${ICON_VERSION}`;
}

export function spriteRef(id) {
  const safe = normalizeIconId(id);
  return `/assets/icons/sprite.svg?v=${ICON_VERSION}#icon-${safe}`;
}

/** Resolve the generated icon for a navigator tree node. */
export function resolveNavIcon(node) {
  if (node?.icon && isGeneratedIcon(node.icon)) return node.icon;
  if (node?.folderPath !== undefined) return SYMBOL_FOLDER_ICON;
  if (node?.moduleId && MODULE_NAV_ICON[node.moduleId]) {
    return MODULE_NAV_ICON[node.moduleId];
  }
  return node?.icon ? normalizeIconId(node.icon) : null;
}

export function resolveEntityIcon(kind) {
  return ENTITY_ICON[kind] || kind;
}

/** Legacy HTML table cells — same generated assets as React. */
export function iconImgHtml(id, title) {
  const safe = normalizeIconId(id);
  const url = iconUrl(safe);
  const t = title
    ? ` title="${String(title).replace(/"/g, "&quot;")}"`
    : "";
  return `<img class="hst-icon" src="${url}" width="16" height="16" alt=""${t}>`;
}
