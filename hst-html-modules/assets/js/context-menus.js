/**
 * Per-module context menu definitions (MT5 admin_managers.htm#context, etc.)
 */
(function () {
  if (!window.HSTContextMenu) return;

  var toast = HSTContextMenu.toast;
  var copyText = HSTContextMenu.copyText;

  function hasSelection(ctx) {
    return ctx.getSelectedRows().length > 0;
  }

  function singleSelection(ctx) {
    return ctx.getSelectedRows().length === 1;
  }

  function sep() { return { type: "separator" }; }

  function viewToggles(ctx) {
    return [
      {
        label: "Auto Arrange",
        check: true,
        checked: function (c) { return c.view.autoArrange; },
        action: function (c) { c.toggleView("autoArrange"); }
      },
      {
        label: "Grid",
        check: true,
        checked: function (c) { return c.view.grid; },
        action: function (c) { c.toggleView("grid"); }
      },
      {
        label: "Columns",
        children: function (c) {
          return c.getColumns().map(function (col) {
            return {
              label: col.label,
              check: true,
              checked: function (x) { return !x.view.hiddenCols[col.index]; },
              action: function (x) { x.toggleColumn(col.index); }
            };
          });
        }
      }
    ];
  }

  function copyAsMenu(ctx) {
    return {
      label: "Copy As",
      disabled: function (c) { return !hasSelection(c); },
      children: [
        {
          label: "Lines",
          icon: "⎘",
          action: function (c) {
            copyText(c.formatRowsAsLines());
            toast("Copied " + c.getSelectedRows().length + " row(s) to clipboard");
          }
        },
        {
          label: "List of Logins",
          when: function (c) { return c.cfg.idKey === "login" || c.rowsHaveKey(c, "login"); },
          action: function (c) {
            copyText(c.getSelectedRows().map(function (r) { return r.login; }).filter(Boolean).join("\n"));
            toast("Copied logins");
          }
        },
        {
          label: "List of Tickets",
          when: function (c) {
            return c.cfg.idKey === "position_id" || c.cfg.idKey === "order_id" ||
              c.cfg.idKey === "deal_id" || c.rowsHaveKey(c, "position_id") ||
              c.rowsHaveKey(c, "order_id") || c.rowsHaveKey(c, "deal_id");
          },
          action: function (c) {
            var key = c.cfg.idKey;
            copyText(c.getSelectedRows().map(function (r) {
              return r[key] || r.position_id || r.order_id || r.deal_id;
            }).filter(function (v) { return v != null; }).join("\n"));
            toast("Copied tickets");
          }
        }
      ]
    };
  }

  function crudItems(opts) {
    opts = opts || {};
    var items = [];
    if (opts.add !== false) {
      items.push({
        label: opts.addLabel || "Add",
        iconId: "add",
        shortcut: opts.addShortcut || "Ctrl+N",
        right: opts.addRight,
        action: function (c) {
          if (c.cfg.configPage && !opts.addOnlyToast) {
            toast("Add — not wired yet (use toolbar)");
          } else {
            toast("Add — not implemented");
          }
        }
      });
    }
    items.push({
      label: opts.editLabel || "Edit",
      iconId: "edit",
      shortcut: opts.editShortcut || "Ctrl+U",
      disabled: function (c) { return !hasSelection(c); },
      action: function (c) { c.openEdit(); }
    });
    if (opts.delete !== false) {
      items.push({
        label: "Delete",
        iconId: "delete",
        shortcut: opts.deleteShortcut || "Ctrl+D",
        right: opts.deleteRight,
        disabled: function (c) { return !hasSelection(c); },
        action: function () { toast("Delete — not implemented"); }
      });
    }
    return items;
  }

  function requestItem() {
    return {
      label: "Request",
      iconId: "refresh",
      action: function (c) { c.reload(); }
    };
  }

  function restoreItem() {
    return {
      label: "Restore",
      icon: "⎌",
      when: hasSelection,
      disabled: function (c) { return c.dbMode !== "backup"; },
      action: function () { toast("Restore from backup — not implemented"); }
    };
  }

  function exportItem(label) {
    return {
      label: label || "Export",
      icon: "⭳",
      disabled: function (c) { return !hasSelection(c); },
      action: function (c) {
        copyText(c.formatRowsAsCsv());
        toast("Exported selection as CSV to clipboard");
      }
    };
  }

  function importItem() {
    return {
      label: "Import from File",
      icon: "⭱",
      action: function () { toast("Import from File — not implemented"); }
    };
  }

  function findItem() {
    return {
      label: "Find",
      icon: "🔍",
      shortcut: "Ctrl+F",
      action: function (c) {
        var input = c.root.querySelector("[data-search]");
        if (input) input.focus();
        else toast("Find — use search bar");
      }
    };
  }

  function serversMenu() {
    return {
      label: "Servers",
      children: [
        {
          label: "All servers",
          check: true,
          checked: function (c) { return !c.view.serverFilter; },
          action: function (c) {
            c.view.serverFilter = "";
            toast("Showing all servers");
          }
        },
        {
          label: "MetaTrader Server",
          check: true,
          checked: function (c) { return c.view.serverFilter === "main"; },
          action: function (c) {
            c.view.serverFilter = "main";
            toast("Filter by MetaTrader Server — not implemented");
          }
        }
      ]
    };
  }

  function moveItems() {
    return [
      {
        label: "Move Up",
        icon: "↑",
        disabled: function (c) {
          var idx = c.getSelectedIndices();
          return idx.length !== 1 || idx[0] <= 0 || c.view.columnSort;
        },
        action: function () { toast("Move Up — not implemented"); }
      },
      {
        label: "Move Down",
        icon: "↓",
        disabled: function (c) {
          var idx = c.getSelectedIndices();
          return idx.length !== 1 || idx[0] >= c.getRows().length - 1 || c.view.columnSort;
        },
        action: function () { toast("Move Down — not implemented"); }
      }
    ];
  }

  function tradingMenu(extra) {
    return function () {
      return crudItems({ add: false, editLabel: "Edit" })
        .concat([
          sep(),
          requestItem(),
          restoreItem(),
          copyAsMenu(),
          exportItem(),
          sep(),
          findItem(),
          {
            label: "Show Milliseconds",
            check: true,
            checked: function (c) { return c.view.showMs; },
            action: function (c) { c.toggleView("showMs"); toast(c.view.showMs ? "Milliseconds on" : "Milliseconds off"); }
          }
        ])
        .concat(sep())
        .concat(viewToggles())
        .concat(extra || []);
    };
  }

  function configMenu(opts) {
    opts = opts || {};
    return function () {
      return crudItems({
        editLabel: opts.editLabel || "Edit",
        addRight: opts.addRight,
        deleteRight: opts.deleteRight
      }).concat([
        sep(),
        {
          label: "Move Up",
          icon: "↑",
          when: singleSelection,
          disabled: function (c) { return !singleSelection(c); },
          action: function () { toast("Move Up — server sort not implemented"); }
        },
        {
          label: "Move Down",
          icon: "↓",
          when: singleSelection,
          disabled: function (c) { return !singleSelection(c); },
          action: function () { toast("Move Down — not implemented"); }
        },
        {
          label: "Sort Alphabetically",
          icon: "A↓",
          action: function () { toast("Sort on server — not implemented"); }
        },
        sep(),
        exportItem(),
        findItem(),
        sep()
      ]).concat(viewToggles());
    };
  }

  HSTContextMenu.register("admin.groups", configMenu({ editLabel: "Edit", addRight: "right_cfg_groups", deleteRight: "right_cfg_groups" }));

  HSTContextMenu.register("admin.symbols.tree", function () {
    return [
      { label: "Add", icon: "＋", right: "right_cfg_symbols", action: function () { toast("Add symbol group — not implemented"); } },
      { label: "Edit", icon: "✎", action: function () { toast("Open group — not implemented"); } },
      { label: "Delete", icon: "✕", right: "right_cfg_symbols", action: function () { toast("Delete group — not implemented"); } },
      sep(),
      { label: "Sort Alphabetically", icon: "A↓", action: function () { toast("Sort on server — not implemented"); } },
      sep(),
      findItem()
    ];
  });

  HSTContextMenu.register("admin.symbols", function () {
    return crudItems({ editLabel: "Edit", addRight: "right_cfg_symbols" }).concat([
      sep(),
      {
        label: "Add Copy",
        when: hasSelection,
        action: function () { toast("Add Copy — not implemented"); }
      },
      {
        label: "Move Up",
        icon: "↑",
        when: singleSelection,
        disabled: function (c) { return c.view.columnSort; },
        action: function () { toast("Move Up — disabled while column sort active"); }
      },
      {
        label: "Move Down",
        icon: "↓",
        when: singleSelection,
        disabled: function (c) { return c.view.columnSort; },
        action: function () { toast("Move Down — disabled while column sort active"); }
      },
      { label: "Sort Alphabetically", icon: "A↓", action: function () { toast("Sort on server — not implemented"); } },
      sep(),
      { label: "Export to File", icon: "⭳", when: hasSelection, action: function () { toast("Export — not implemented"); } },
      { label: "Import from File", icon: "⭱", action: function () { toast("Import — not implemented"); } },
      sep(),
      findItem(),
      sep()
    ]).concat(viewToggles());
  });

  HSTContextMenu.register("admin.users", function () {
    return [
      { label: "New", iconId: "add", right: "right_acc_read", action: function () { toast("New account — not implemented"); } },
      { label: "Edit", iconId: "edit", when: hasSelection, action: function (c) { c.openEdit(); } },
      {
        label: "Edit group",
        iconId: "groups-folder",
        when: singleSelection,
        right: "right_cfg_groups",
        action: function (c) {
          var row = c.getSelectedRows()[0];
          if (row && row.group) location.href = "groups-config.html?group=" + encodeURIComponent(row.group);
          else toast("No group on account");
        }
      },
      {
        label: "Edit manager",
        iconId: "manager",
        when: singleSelection,
        right: "right_cfg_managers",
        action: function () { toast("Edit manager — open Managers if account is staff"); }
      },
      { label: "Delete", iconId: "delete", when: hasSelection, action: function () { toast("Delete — not implemented"); } },
      sep(),
      requestItem(),
      {
        label: "Move to Archive",
        icon: "📦",
        when: hasSelection,
        disabled: function (c) { return c.dbMode === "backup"; },
        action: function () { toast("Move to Archive — not implemented"); }
      },
      {
        label: "Restore",
        icon: "⎌",
        when: hasSelection,
        disabled: function (c) { return c.dbMode !== "backup"; },
        action: function () { toast("Restore account — not implemented"); }
      },
      {
        label: "Balance",
        when: singleSelection,
        right: "right_accountant",
        children: [
          { label: "Check Balance", action: function () { toast("Check Balance — not implemented"); } },
          { label: "Fix Balance", action: function () { toast("Fix Balance — not implemented"); } }
        ]
      },
      copyAsMenu(),
      exportItem(),
      sep(),
      { label: "E-Mail", iconId: "mailbox", when: hasSelection, action: function () { toast("E-Mail — not implemented"); } },
      { label: "Journal", iconId: "journal", when: singleSelection, action: function () { toast("Journal — not implemented"); } },
      findItem(),
      {
        label: "Enabled only",
        check: true,
        checked: function (c) { return c.view.enabledOnly; },
        action: function (c) { c.toggleView("enabledOnly"); toast(c.view.enabledOnly ? "Showing enabled only" : "Showing all accounts"); }
      },
      sep()
    ].concat(viewToggles());
  });

  HSTContextMenu.register("admin.clients", function () {
    return crudItems({ addLabel: "New Client", editLabel: "Edit", deleteRight: "right_clients_delete" }).concat([
      sep(),
      exportItem(),
      findItem(),
      sep()
    ]).concat(viewToggles());
  });

  HSTContextMenu.register("admin.managers", function () {
    return [
      serversMenu(),
      sep(),
      {
        label: "Add",
        icon: "＋",
        shortcut: "Ctrl+N",
        right: "right_cfg_managers",
        action: function () { toast("Add manager — not wired yet (use toolbar)"); }
      },
      {
        label: "Edit",
        icon: "✎",
        shortcut: "Ctrl+U",
        disabled: function (c) { return !hasSelection(c); },
        action: function (c) { c.openEdit(); }
      },
      {
        label: "Delete",
        icon: "✕",
        shortcut: "Ctrl+D",
        right: "right_cfg_managers",
        disabled: function (c) { return !hasSelection(c); },
        action: function () { toast("Delete manager — not implemented"); }
      },
      sep()
    ].concat(moveItems()).concat([
      { label: "Sort by Login", icon: "A↓", action: function () { toast("Sort on server — not implemented"); } },
      sep(),
      copyAsMenu(),
      exportItem("Export to File"),
      importItem(),
      sep(),
      findItem(),
      sep()
    ]).concat(viewToggles());
  });

  HSTContextMenu.register("admin.positions", tradingMenu());
  HSTContextMenu.register("admin.orders", tradingMenu());
  HSTContextMenu.register("admin.deals", tradingMenu());

  HSTContextMenu.register("admin.routing", configMenu({ addRight: "right_cfg_requests", deleteRight: "right_cfg_requests" }));
  HSTContextMenu.register("admin.datafeeds", configMenu({ addRight: "right_cfg_datafeeds", deleteRight: "right_cfg_datafeeds" }));
  HSTContextMenu.register("admin.holidays", configMenu({ addRight: "right_cfg_holidays", deleteRight: "right_cfg_holidays" }));
  HSTContextMenu.register("admin.leverages", configMenu({ addRight: "right_cfg_groups" }));

  HSTContextMenu.register("manager.clients", function () {
    return crudItems({ addLabel: "New Client", editLabel: "Edit" }).concat([
      sep(), exportItem(), findItem(), sep()
    ]).concat(viewToggles());
  });

  HSTContextMenu.register("manager.users", function () {
    return [
      { label: "Edit", icon: "✎", when: hasSelection, action: function (c) { c.openEdit(); } },
      sep(),
      requestItem(),
      copyAsMenu(),
      exportItem(),
      findItem(),
      sep()
    ].concat(viewToggles());
  });

  HSTContextMenu.register("manager.positions", tradingMenu());
  HSTContextMenu.register("manager.orders", tradingMenu());
  HSTContextMenu.register("manager.deals", tradingMenu());
  HSTContextMenu.register("manager.groups", configMenu({ editLabel: "Edit", addRight: "right_cfg_groups" }));

  HSTContextMenu.register("manager.journal", function () {
    return [
      { label: "Open", icon: "📂", action: function () { toast("Open log folder — not implemented"); } },
      { label: "Copy", icon: "⎘", when: hasSelection, action: function (c) {
        copyText(c.formatRowsAsLines());
        toast("Copied to clipboard");
      }},
      sep()
    ].concat(viewToggles());
  });

  HSTContextMenu.register("manager.dealing", function () {
    return [
      { label: "Request", iconId: "refresh", action: function (c) { c.reload(); } },
      findItem(),
      sep()
    ].concat(viewToggles());
  });
})();
