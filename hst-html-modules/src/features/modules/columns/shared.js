import {
  fmtAuthMode,
  fmtClientStatus,
  fmtKycStatus,
  fmtTs,
} from "../../../lib/formatters.js";
import { iconImgHtml, resolveEntityIcon } from "../../../lib/icons.js";
import { getMock } from "../../../lib/legacy.js";

export function personNameCol() {
  return {
    key: "person_name",
    label: "Name",
    render: (r) => [r.person_name, r.person_last_name].filter(Boolean).join(" "),
  };
}

export function clientStatusCol() {
  return {
    key: "client_status",
    label: "Status",
    render: (r) => fmtClientStatus(r.client_status),
  };
}

export function kycStatusCol() {
  return {
    key: "kyc_status",
    label: "KYC",
    render: (r) => fmtKycStatus(r.kyc_status),
  };
}

export function accountsCountCol() {
  return {
    key: "accounts",
    label: "Accounts",
    render: (r) => getMock().accountsForClient(r.client_id).length,
  };
}

export function tsCol(key, label) {
  return { key, label, render: (r) => fmtTs(r[key]) };
}

export function clientLinkCol(panel) {
  return {
    key: "client_id",
    label: "Client",
    link: (r) => `/${panel}/clients/${r.client_id}`,
  };
}

export function loginLinkCol(panel) {
  return {
    key: "login",
    label: "Login",
    link: (r) => `/${panel}/users/${r.login}`,
  };
}

export function groupColumns({ extended = false } = {}) {
  const base = [
    { key: "group", label: "Group" },
    {
      key: "auth_mode",
      label: "Authorization",
      render: (r) => fmtAuthMode(r.auth_mode),
    },
    { key: "currency", label: "Currency" },
    { key: "status", label: "Status" },
  ];
  if (!extended) return base;
  return [
    ...base,
    { key: "margin_call", label: "Margin call" },
    { key: "margin_stop_out", label: "Stop out" },
    { key: "limit_orders", label: "Limit orders" },
  ];
}

export function clientColumns(panel, { extended = false } = {}) {
  const cols = [
    { key: "client_id", label: "ID" },
    personNameCol(),
  ];
  if (extended) {
    cols.push(
      { key: "contact_email", label: "Email" },
      { key: "contact_phone", label: "Phone" },
      { key: "address_country", label: "Country" },
      clientStatusCol(),
      kycStatusCol(),
      accountsCountCol(),
      tsCol("date_created", "Created")
    );
  } else {
    cols.push(
      { key: "contact_email", label: "Email" },
      clientStatusCol(),
      accountsCountCol()
    );
  }
  return cols;
}

export function userColumns(panel, { extended = false } = {}) {
  const cols = [
    { key: "login", label: "Login" },
    clientLinkCol(panel),
    { key: "group", label: "Group" },
    { key: "name", label: "Name" },
  ];
  if (extended) {
    cols.push(
      { key: "email", label: "Email" },
      { key: "leverage", label: "Leverage" },
      { key: "balance", label: "Balance" },
      { key: "credit", label: "Credit" },
      tsCol("last_access", "Last access")
    );
  } else {
    cols.push(
      { key: "balance", label: "Balance" },
      { key: "leverage", label: "Leverage" },
      tsCol("last_access", "Last access")
    );
  }
  return cols;
}

export function managerRoleIconCol() {
  return {
    key: "icon",
    label: "",
    html: true,
    render: (r) => {
      const icon = resolveEntityIcon(
        r.rights?.right_admin ? "administrator" : "manager"
      );
      return iconImgHtml(icon);
    },
  };
}
