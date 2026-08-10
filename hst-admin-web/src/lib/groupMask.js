/** Whether a group path matches a bulk-operation mask (All, real*, demo*, …). */
export function groupMatchesMask(groupPath, mask) {
  const path = String(groupPath ?? "");
  if (mask === "All") return true;
  if (!mask.endsWith("*")) {
    return path === mask || path.startsWith(`${mask}\\`);
  }
  const prefix = mask.slice(0, -1);
  if (!prefix) return true;
  return path === prefix || path.startsWith(`${prefix}\\`);
}

/** Real group records only — skip navigator folder nodes without a group row. */
export function realGroups(groups) {
  return (groups || []).filter((g) => g.group_id);
}

export function groupsMatchingMask(groups, mask) {
  return realGroups(groups).filter((g) => groupMatchesMask(g.group, mask));
}
