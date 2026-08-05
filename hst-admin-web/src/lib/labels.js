// Wire labels are lower_snake contract strings; humanising them happens only here.

/** `sell_limit` -> `Sell Limit` */
export const humanize = (label) =>
  String(label ?? "")
    .split("_")
    .map((w) => (w ? w[0].toUpperCase() + w.slice(1) : w))
    .join(" ");
