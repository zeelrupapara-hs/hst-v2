// Generated icon set: 51 sprite entries cropped from the reference screenshots.
import spriteUrl from "@/assets/icons/sprite.svg";

/** Backend navigation key → sprite id. Keys the server does not know map to null. */
export const NAV_ICON = {
  servers: "servers",
  server: "server",
  integrations: "integrations",
  clients_accounts: "clients-accounts",
  clients_orders: "clients-accounts",
  orders_deals: "orders-deals",
  accounts: "accounts-tree",
  clients: "clients-tree",
  managers: "managers-tree",
  positions: "positions",
  orders: "orders-tree",
  deals: "deals-tree",
  groups: "groups",
  group_folder: "groups-folder",
  symbols: "symbols-tree",
  leverages: "leverages",
  routing: "routing",
  datafeeds: "datafeeds",
  mail_servers: "mailbox",
  mailbox: "mailbox",
  holidays: "holidays",
  end_of_day: "time",
  journal: "journal",
  dealing: "dealing",
  balance: "balance",
  margin_calls: "security",
  online: "client",
  market_watch: "symbols",
};

export const spriteRef = (id) => `${spriteUrl}#icon-${id}`;

export const navIcon = (key) => {
  if (NAV_ICON[key]) return NAV_ICON[key];
  if (key?.startsWith("group:") || key?.startsWith("groupfolder:")) return "folder";
  if (key?.startsWith("symfolder:")) return "folder";
  if (key?.startsWith("feed:")) return "datafeeds";
  return null;
};
