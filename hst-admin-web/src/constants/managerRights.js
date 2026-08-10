// The manager permission tree, sectioned the way the reference administrator draws it.
// Entries are the right column names the API speaks; children nest one level.

export const ManagerLimit_options = [
  { value: 0, label: "All" },
  { value: 1, label: "1 month" },
  { value: 2, label: "3 months" },
  { value: 3, label: "6 months" },
  { value: 4, label: "1 year" },
  { value: 5, label: "2 years" },
  { value: 6, label: "3 years" },
];

const r = (key, label, children) => ({ key, label, children });

export const ManagerRightsTree = [
  r("connection", "Connection type", [
    r("right_admin", "Connect using Administrator"),
    r("right_manager", "Connect using Manager"),
  ]),
  r("configuration", "Configuration setup", [
    r("right_cfg_time", "Configure server operation time"),
    r("right_cfg_holidays", "Configure holidays"),
    r("right_cfg_groups", "Configure groups"),
    r("right_cfg_managers", "Configure managers' permissions"),
    r("right_cfg_requests", "Configure request routing"),
    r("right_cfg_gateways", "Configure gateways"),
    r("right_cfg_datafeeds", "Configure datafeeds"),
    r("right_cfg_reports", "Configure reports"),
    r("right_cfg_symbols", "Configure symbols"),
    r("right_cfg_web_services", "Configure web services"),
    r("right_cfg_messengers", "Configure messengers"),
    r("right_cfg_kyc", "Configure KYC services"),
    r("right_cfg_automations", "Configure automations"),
    r("right_cfg_allocations", "Configure allocations"),
    r("right_cfg_corporate", "Configure corporate links"),
    r("right_cfg_payments", "Configure payments"),
    r("right_cfg_mails", "Configure mail servers"),
    r("right_cfg_streaming", "Configure streaming"),
  ]),
  r("administration", "Administration", [
    r("right_srv_journals", "Access server logs"),
    r("right_srv_reports", "Receive automatic server reports"),
    r("right_charts", "Edit charts"),
    r("right_email", "Send emails"),
    r("right_news", "Send news"),
    r("right_export", "Export data"),
    r("right_techsupport", "Technical support"),
    r("right_market", "Market"),
  ]),
  r("accounts", "Accounts", [
    r("right_accountant", "Accountant"),
    r("right_acc_read", "Access accounts"),
    r("right_acc_technical", "View technical accounts"),
    r("right_acc_tech_modify", "Manage technical accounts"),
    r("acc_details", "Access the account personal details", [
      r("right_acc_details_name", "Name"),
      r("right_acc_details_location", "Location"),
      r("right_acc_details_address", "Address"),
      r("right_acc_details_id", "Document number"),
      r("right_acc_details_email", "Email"),
      r("right_acc_details_phone", "Phone"),
      r("right_acc_details_general", "General information"),
    ]),
    r("right_acc_manager", "Edit accounts"),
    r("right_acc_delete", "Delete accounts"),
    r("right_acc_online", "View currently connected clients"),
    r("right_confirm_actions", "Confirm dangerous actions"),
    r("right_notifications", "Push notifications"),
  ]),
  r("dealing", "Dealing", [
    r("right_trades_read", "Access orders and positions"),
    r("right_trades_manager", "Edit orders, positions and deals"),
    r("right_trades_delete", "Delete orders, positions and deals"),
    r("right_trades_dealer", "Dealer"),
    r("right_trades_supervisor", "Supervisor"),
    r("right_quotes_raw", "Show raw quotes without spread difference"),
    r("right_quotes", "Throw in quotes"),
    r("right_symbol_details", "Modify spread and execution mode"),
    r("right_risk_manager", "Risk manager"),
    r("right_group_margin", "Edit groups (margin settings)"),
    r("right_group_commission", "Edit groups (commission settings)"),
    r("right_reports", "Receive reports"),
  ]),
  r("backoffice", "Back office", [
    r("right_clients_access", "Access clients"),
    r("clients_details", "Access the client personal details", [
      r("right_clients_details_name", "Name"),
      r("right_clients_details_location", "Location"),
      r("right_clients_details_address", "Address"),
      r("right_clients_details_id", "Document number"),
      r("right_clients_details_email", "Email"),
      r("right_clients_details_phone", "Phone"),
      r("right_clients_details_general", "General information"),
    ]),
    r("right_clients_create", "Create clients"),
    r("right_clients_edit", "Edit clients"),
    r("right_clients_delete", "Delete clients"),
    r("right_clients_kyc", "KYC check"),
    r("right_documents_access", "Access documents"),
    r("right_documents_create", "Create documents"),
    r("right_documents_edit", "Edit documents"),
    r("right_documents_delete", "Delete documents"),
    r("right_documents_files_add", "Add document files"),
    r("right_documents_files_delete", "Delete document files"),
    r("right_comments_access", "Read comments"),
    r("right_comments_create", "Write comments"),
    r("right_comments_delete", "Delete comments"),
  ]),
];

// A right that needs another right on first: unchecking the base clears its dependents.
export const ManagerRightDeps = {
  right_acc_delete: "right_acc_manager",
  right_trades_manager: "right_trades_read",
  right_trades_dealer: "right_trades_read",
  right_trades_delete: "right_trades_manager",
};

export function treeRightKeys(nodes = ManagerRightsTree) {
  return nodes.flatMap((n) => (n.children ? treeRightKeys(n.children) : [n.key]));
}

// Built-in role templates, kept in the frontend as the reference keeps them in the terminal.
// "*" grants the whole tree; the rest are the classic staff shapes.
export const ManagerRoleTemplates = {
  Administrator: "*",
  Manager: [
    "right_manager", "right_acc_read", "right_acc_manager", "right_acc_online",
    "right_acc_details_name", "right_acc_details_location", "right_acc_details_address",
    "right_acc_details_id", "right_acc_details_email", "right_acc_details_phone",
    "right_acc_details_general", "right_trades_read", "right_notifications",
    "right_email", "right_reports",
  ],
  Dealer: [
    "right_manager", "right_acc_read", "right_trades_read", "right_trades_manager",
    "right_trades_dealer", "right_quotes", "right_quotes_raw", "right_reports",
  ],
  "Risk Manager": [
    "right_manager", "right_risk_manager", "right_acc_read", "right_trades_read",
    "right_quotes_raw", "right_srv_journals", "right_reports",
  ],
  Accountant: [
    "right_manager", "right_accountant", "right_acc_read", "right_acc_details_name",
    "right_acc_details_id", "right_acc_details_general", "right_cfg_payments",
    "right_reports",
  ],
  "Back Office": [
    "right_manager", "right_acc_read", "right_acc_details_name", "right_acc_details_location",
    "right_acc_details_address", "right_acc_details_id", "right_acc_details_email",
    "right_acc_details_phone", "right_acc_details_general", "right_clients_access",
    "right_clients_create", "right_clients_edit", "right_clients_kyc",
    "right_clients_details_name", "right_clients_details_location",
    "right_clients_details_address", "right_clients_details_id",
    "right_clients_details_email", "right_clients_details_phone",
    "right_clients_details_general", "right_documents_access", "right_documents_create",
    "right_documents_edit", "right_documents_files_add", "right_comments_access",
    "right_comments_create",
  ],
  Support: ["right_manager", "right_acc_read", "right_acc_online", "right_trades_read"],
};
