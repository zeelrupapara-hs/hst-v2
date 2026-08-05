// Mirrors hst-server/model/user.go UsersRights bits; trade_disabled is inverted (set = OFF).

export const UsersRights_value = {
  enabled: 0x0001,
  password: 0x0002,
  trade_disabled: 0x0004,
  investor: 0x0008,
  confirmed: 0x0010,
  trailing: 0x0020,
  expert: 0x0040,
  reports: 0x0100,
  readonly: 0x0200,
  reset_pass: 0x0400,
  otp_enabled: 0x0800,
  sponsored_hosting: 0x2000,
  api_enabled: 0x4000,
  push_notification: 0x8000,
  technical: 0x10000,
  exclude_reports: 0x20000,
};

export const AccountRight_checks = [
  { bit: UsersRights_value.enabled, label: "Enable this account" },
  { bit: UsersRights_value.password, label: "Enable password change" },
  { bit: UsersRights_value.otp_enabled, label: "Enable one-time password" },
  { bit: UsersRights_value.reset_pass, label: "Change password at next login" },
];

export const LimitRight_checks = [
  { bit: UsersRights_value.trade_disabled, label: "Enable trading", inverted: true },
  { bit: UsersRights_value.expert, label: "Enable algo trading by Expert Advisors" },
  { bit: UsersRights_value.trailing, label: "Enable trailing stops" },
  { bit: UsersRights_value.reports, label: "Enable daily reports" },
  { bit: UsersRights_value.exclude_reports, label: "Exclude from server reports", },
  { bit: UsersRights_value.api_enabled, label: "Enable API connections" },
  { bit: UsersRights_value.sponsored_hosting, label: "Enable sponsored VPS hosting" },
  { bit: UsersRights_value.push_notification, label: "Enable push notifications" },
  { bit: UsersRights_value.technical, label: "Technical account (hide from regular managers)" },
];

export const ClientStatus_name = {
  0: "Not registered",
  100: "Registered",
  200: "Not interested",
  300: "Application incomplete",
  400: "Application completed",
  500: "Information requested",
  600: "Application rejected",
  700: "Approved",
  800: "Funded",
  900: "Active",
  1000: "Inactive",
  1100: "Suspended",
  1200: "Closed",
  1300: "Terminated",
};

export const ClientType_name = { 0: "Undefined", 1: "Individual", 2: "Corporate", 3: "Fund" };

export const KycStatus_name = { 0: "Undefined", 1: "Approved", 2: "Declined" };
