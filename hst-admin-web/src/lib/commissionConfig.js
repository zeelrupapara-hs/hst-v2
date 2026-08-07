/** Commission enums — mirrors hst-server/model/commission.go + MT5 Manager labels. */

export const COMMISSION_MODE = { 0: "Standard commission", 1: "Agent commission", 2: "Fee" };
export const COMMISSION_MODE_SHORT = { 0: "Standard", 1: "Agent", 2: "Fee" };

export const COMMISSION_RANGE = {
  0: "Volume",
  1: "Turnover money",
  2: "Turnover volume",
};

export const COMMISSION_CHARGE = { 0: "Daily", 1: "Monthly", 2: "Instant" };

export const COMMISSION_ENTRY = { 0: "All", 1: "In", 2: "Out" };
export const COMMISSION_ACTION = { 0: "All", 1: "Buy", 2: "Sell" };
export const COMMISSION_PROFIT = { 0: "All", 1: "Profit", 2: "Loss" };

/** Deal reason flags — bitmask on mode_reason (EnCommReasonFlags). */
export const COMMISSION_REASON_FLAGS = [
  { bit: 0x01, label: "Client" },
  { bit: 0x02, label: "Expert" },
  { bit: 0x04, label: "Dealer" },
  { bit: 0x08, label: "External client" },
  { bit: 0x10, label: "Mobile" },
  { bit: 0x20, label: "Web" },
  { bit: 0x40, label: "Signal" },
  { bit: 0x80, label: "Gateway" },
  { bit: 0x100, label: "Ultency" },
];

export const COMMISSION_REASON_ALL_MASK = COMMISSION_REASON_FLAGS.reduce((m, l) => m | l.bit, 0);

/** Wire 0 = all reasons (runtime treats unset mask as match-all). */
export function reasonFlagsForDisplay(value) {
  if (value == null || value === 0) return COMMISSION_REASON_ALL_MASK;
  return value;
}

export function reasonFlagsForSave(value) {
  if ((value & COMMISSION_REASON_ALL_MASK) === COMMISSION_REASON_ALL_MASK) return 0;
  return value;
}

export const TIER_MODE = {
  0: "deposit ccy",
  1: "base ccy",
  2: "profit ccy",
  3: "margin ccy",
  4: "points",
  5: "percents",
  6: "specified ccy",
};

export const TIER_TYPE = { 0: "per trade", 1: "per volume" };

export const COMMISSION_MODE_FEE = 2;
export const COMMISSION_CHARGE_INSTANT = 2;

export const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

export function newCommissionTier() {
  return {
    mode: 0,
    type: 0,
    value: 0,
    range_from: 0,
    range_to: 0,
    minimal: 0,
    maximal: 0,
    currency: "",
  };
}

export function newCommissionDraft() {
  return {
    name: "",
    description: "",
    path: "*",
    mode: 0,
    mode_range: 0,
    mode_charge: COMMISSION_CHARGE_INSTANT,
    turnover_currency: "",
    mode_entry: 0,
    mode_action: 0,
    mode_profit: 0,
    mode_reason: 0,
    tiers: [newCommissionTier()],
  };
}

export function commissionFromRow(row) {
  if (!row) return newCommissionDraft();
  return {
    ...row,
    mode_entry: row.mode_entry ?? 0,
    mode_action: row.mode_action ?? 0,
    mode_profit: row.mode_profit ?? 0,
    mode_reason: row.mode_reason ?? 0,
    tiers: (row.tiers || []).map((t) => ({
      ...t,
      maximal: t.maximal ?? 0,
    })),
  };
}

export function commissionToBody(draft) {
  return {
    name: draft.name.trim(),
    description: draft.description || "",
    path: draft.path || "*",
    mode: draft.mode ?? 0,
    mode_range: draft.mode_range ?? 0,
    mode_charge: draft.mode === COMMISSION_MODE_FEE ? COMMISSION_CHARGE_INSTANT : (draft.mode_charge ?? 0),
    turnover_currency: draft.turnover_currency || "",
    mode_entry: draft.mode_entry ?? 0,
    mode_action: draft.mode_action ?? 0,
    mode_profit: draft.mode_profit ?? 0,
    mode_reason: draft.mode_reason ?? 0,
    tiers: draft.tiers.map((t) => ({
      mode: t.mode ?? 0,
      type: t.type ?? 0,
      value: t.value ?? 0,
      range_from: t.range_from ?? 0,
      range_to: t.range_to ?? 0,
      minimal: t.minimal ?? 0,
      currency: t.currency || "",
    })),
  };
}
