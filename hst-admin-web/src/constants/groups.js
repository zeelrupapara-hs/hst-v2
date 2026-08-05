// Mirrors hst-server/model/group.go enum values; keys are the wire ints.

export const AuthMode_name = {
  0: "Normal",
  1: "1024-bit RSA",
  2: "2048-bit RSA",
};

export const MarginMode_name = {
  0: "for Retail Forex, Futures, CFD",
  1: "for Stock Exchange, based on margin discount rates",
  2: "for Retail Forex, Futures, CFD with hedging",
};

export const MarginMode_short = { 0: "Netting", 1: "Exchange", 2: "Hedged" };

export const NewsMode_name = {
  0: "Disabled",
  1: "Collect news headers",
  2: "Full news package",
};

export const MailMode_name = { 0: "Disabled", 1: "Enabled" };

export const ReportsMode_name = {
  0: "Disabled",
  1: "End of day and end of month",
  2: "End of day only",
  3: "End of month only",
};

export const TransferMode_name = {
  0: "Disabled",
  1: "By matching name",
  2: "Within the group",
  3: "Matching name within the group",
};

export const FreeMarginMode_name = {
  0: "Do not use unrealized profit/loss",
  1: "Use unrealized profit/loss",
  2: "Use unrealized profit only",
  3: "Use unrealized loss only",
};

export const StopOutMode_name = { 0: "In percent", 1: "In money" };

export const MarginFreeProfitMode_name = {
  0: "Use daily fixed profit/loss",
  1: "Use daily fixed loss only",
};

export const HistoryLimit_name = {
  0: "All history",
  1: "1 month",
  2: "3 months",
  3: "6 months",
  4: "1 year",
  5: "2 years",
  6: "3 years",
};

export const PermissionFlag_labels = [
  { bit: 2, label: "Enable connections" },
  { bit: 1, label: "Enable certificate confirmation" },
  { bit: 4, label: "Change password at first login" },
  { bit: 16, label: "Show the risk warning window after connection" },
  { bit: 32, label: "Enforce country-specific regulatory restrictions for retail clients" },
];

export const PermissionFlag_forceOtp = 8;
export const PermissionFlag_notify = [
  { bit: 64, label: "Deals" },
  { bit: 128, label: "Orders" },
  { bit: 256, label: "Balance operations" },
];

export const TradeFlag_labels = [
  { bit: 4, label: "Allow Expert Advisors" },
  { bit: 2, label: "Allow trailing stops" },
  { bit: 1, label: "Charge swaps" },
  { bit: 256, label: "Close positions in FIFO order" },
  { bit: 512, label: "Prohibit hedge positions" },
  { bit: 1024, label: "Charge deal cost" },
];

export const TradeFlag_soFullyHedged = 128;
export const TradeFlag_soCompensation = 64;
export const TradeFlag_soCompensationCredit = 2048;

export const ReportsFlag_labels = [
  { bit: 4, label: "Generate statements" },
  { bit: 1, label: "Send reports by email" },
  { bit: 2, label: "Send copies to the support email" },
];

/** The platform derives the group type from a case-sensitive substring of the path. */
export function groupKind(path) {
  for (const kind of ["demo", "manager", "contest", "coverage", "preliminary"]) {
    if (path.includes(kind)) return kind;
  }
  return "real";
}
