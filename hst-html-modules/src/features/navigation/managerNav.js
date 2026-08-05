/** Manager navigator tree — mirrors manager.html */
export const managerNav = [
  {
    label: "Servers",
    icon: "servers",
    indent: 0,
    open: true,
    children: [
      {
        label: "MetaTrader Server",
        icon: "server",
        indent: 1,
        open: true,
        children: [
          {
            label: "Clients & Orders",
            icon: "orders-deals",
            indent: 2,
            open: true,
            children: [
              { moduleId: "users", label: "Trading Accounts", icon: "accounts-tree", indent: 3, count: "(5)" },
              { moduleId: "clients", label: "Clients", icon: "clients-tree", indent: 3, count: "(4)" },
              { moduleId: "positions", label: "Positions", icon: "positions", indent: 3, count: "(3)" },
              { moduleId: "orders", label: "Orders", icon: "orders-tree", indent: 3, count: "(4)" },
              { moduleId: "deals", label: "Deals", icon: "deals-tree", indent: 3, count: "(4)" },
            ],
          },
          { moduleId: "dealing", label: "Dealing", icon: "dealing", indent: 2 },
          { moduleId: "balance", label: "Balance Operations", icon: "balance", indent: 2 },
          { moduleId: "groups", label: "Groups", icon: "groups", indent: 2, count: "(39)" },
          { moduleId: "journal", label: "Journal", icon: "journal", indent: 2 },
        ],
      },
    ],
  },
  { label: "Support Center", icon: "support-center", indent: 0, disabled: true },
];
