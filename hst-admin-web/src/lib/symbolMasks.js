/** Turn an MT5-style mask (`EUR*`, `!GBPUSD`) into a case-insensitive RegExp. */
function maskToRegex(mask) {
  const neg = mask.startsWith("!");
  const raw = neg ? mask.slice(1) : mask;
  let re = "^";
  for (const ch of raw) {
    if (ch === "*") re += ".*";
    else if (/[.+?^${}()|[\]\\]/.test(ch)) re += `\\${ch}`;
    else re += ch;
  }
  re += "$";
  return new RegExp(re, "i");
}

/**
 * Filter symbols by comma-separated masks (MT5 filter tab).
 * Positives are OR'd; exclusions (`!`) always apply.
 * @param {Array<{symbol?: string}>} rows
 * @param {string} expr
 */
export function filterSymbolsByMasks(rows, expr) {
  const masks = expr
    .split(",")
    .map((m) => m.trim())
    .filter(Boolean);
  if (!masks.length) return rows.slice();

  const positive = masks.filter((m) => !m.startsWith("!")).map(maskToRegex);
  const negative = masks.filter((m) => m.startsWith("!")).map((m) => maskToRegex(m.slice(1)));

  return rows.filter((row) => {
    const name = row.symbol ?? "";
    if (positive.length && !positive.some((re) => re.test(name))) return false;
    if (negative.some((re) => re.test(name))) return false;
    return true;
  });
}

/** Human label for an applied mask expression. */
export function maskFilterLabel(expr) {
  const trimmed = expr.trim();
  return trimmed.includes(",") || trimmed.includes("*") || trimmed.startsWith("!")
    ? trimmed
    : `containing '${trimmed}'`;
}
