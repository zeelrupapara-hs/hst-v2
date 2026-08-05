export function fmtTradeMode(n) {
  const m = {
    0: "Disabled",
    1: "Long only",
    2: "Short only",
    3: "Close only",
    4: "Full",
  };
  return m[n] ?? String(n);
}

export function fmtCalcMode(n) {
  const m = {
    0: "Forex",
    1: "Futures",
    2: "CFD",
    3: "CFD Index",
    4: "CFD Leverage",
    5: "Forex No Leverage",
  };
  return m[n] ?? String(n);
}

export function fmtGtcMode(n) {
  const m = { 0: "Good till cancelled", 1: "Daily", 2: "Daily no stops" };
  return m[n] ?? String(n);
}

export function fmtSwapMode(n) {
  const m = { 0: "Disabled", 1: "In points", 2: "By symbol currency" };
  return m[n] ?? String(n);
}

export function fmtSector(n) {
  const m = { 12: "Currency" };
  return m[n] ?? "Undefined";
}

export function fmtSpread(v) {
  return v === 0 || v == null ? "off" : String(v);
}

export function fmtOff(v, unit = "") {
  if (v === 0 || v == null || v === "") return "off";
  return unit ? `${v} ${unit}` : String(v);
}

export function tradeModeClass(mode) {
  if (mode === 0) return "sym-trade-disabled";
  if (mode === 1) return "sym-trade-long";
  if (mode === 2) return "sym-trade-short";
  if (mode === 3) return "sym-trade-close";
  return "sym-trade-full";
}
