// Money and price formatting shared by every module.
export const money = (v) => (v ?? 0).toFixed(2);
export const price = (v, digits = 5) => (v ? v.toFixed(digits) : "");
