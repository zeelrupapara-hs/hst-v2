import { adminModules } from "./definitions/admin.js";
import { managerModules } from "./definitions/manager.js";

/**
 * @typedef {Object} ModuleDef
 * @property {'list'|'legacy'|'settings'} type
 * @property {string} title
 * @property {string} [endpoint]
 * @property {Record<string, string>} [query]
 * @property {string} [idKey]
 * @property {Array} [columns]
 * @property {string} [legacyDetail]
 * @property {string} [legacyDetailPanel]
 * @property {string} [legacySrc]
 * @property {string} [settingsId]
 * @property {string} [editLabel]
 * @property {string} [note]
 * @property {string} [searchPlaceholder]
 * @property {string} [detailParam]
 * @property {boolean} [search]
 * @property {boolean} [toolbar]
 */

const MODULES = Object.fromEntries([
  ...Object.entries(adminModules).map(([id, def]) => [`admin.${id}`, def]),
  ...Object.entries(managerModules).map(([id, def]) => [`manager.${id}`, def]),
]);

/** @param {'admin'|'manager'} panel @param {string} moduleId */
export function getModuleDef(panel, moduleId) {
  return MODULES[`${panel}.${moduleId}`] || null;
}

export function getModuleLabel(panel, moduleId) {
  return getModuleDef(panel, moduleId)?.title || moduleId;
}

export function getLegacyDetailSrc(panel, moduleId, recordId) {
  const def = getModuleDef(panel, moduleId);
  if (!def?.legacyDetail) return null;
  const param = def.detailParam || "id";
  const detailPanel = def.legacyDetailPanel || panel;
  return `/modules/${detailPanel}/${def.legacyDetail}?${param}=${encodeURIComponent(recordId)}`;
}

export const DEFAULT_MODULE = {
  admin: "clients",
  manager: "clients",
};

export { adminModules, managerModules };
