/**
 *  manager permissions tree + role presets (admin_managers.htm#permissions).
 */
(function (global) {
  var LIMIT_OPTIONS = [
    { v: 0, l: "All" }, { v: 1, l: "1 month" }, { v: 2, l: "3 months" },
    { v: 3, l: "6 months" }, { v: 4, l: "1 year" }, { v: 5, l: "2 years" }, { v: 6, l: "3 years" }
  ];

  var ROLE_PRESETS = {
    Administrator: {
      right_admin: true, right_manager: true,
      right_cfg_groups: true, right_cfg_symbols: true, right_cfg_managers: true,
      right_cfg_requests: true, right_cfg_datafeeds: true, right_cfg_holidays: true,
      right_cfg_time: true, right_srv_journals: true, right_srv_reports: true,
      right_acc_read: true, right_trades_read: true, right_trades_dealer: true,
      right_accountant: true, right_clients_access: true, right_clients_edit: true,
      right_reports: true, right_email: true, right_news: true, right_export: true
    },
    Manager: {
      right_manager: true, right_acc_read: true, right_trades_read: true,
      right_clients_access: true, right_clients_edit: true, right_reports: true,
      right_srv_journals: true, right_export: true
    },
    Dealer: {
      right_manager: true, right_acc_read: true, right_trades_read: true,
      right_trades_dealer: true, right_quotes: true, right_market: true,
      right_clients_access: true
    },
    Accountant: {
      right_manager: true, right_acc_read: true, right_accountant: true,
      right_clients_access: true, right_export: true
    }
  };

  var PERMISSION_TREE = [
    {
      label: "Connect using the Administrator", key: "right_admin"
    },
    {
      label: "Connect using the Manager", key: "right_manager"
    },
    {
      label: "Configuration setup",
      children: [
        { label: "Configure groups", key: "right_cfg_groups" },
        { label: "Configure symbols", key: "right_cfg_symbols" },
        { label: "Configure managers' permissions", key: "right_cfg_managers" },
        { label: "Configure request routing", key: "right_cfg_requests" },
        { label: "Configure datafeeds", key: "right_cfg_datafeeds" },
        { label: "Configure holidays", key: "right_cfg_holidays" },
        { label: "Configure server operation time", key: "right_cfg_time" },
        {
          label: "Configure integrations",
          children: [
            { label: "Configure web services", key: "right_cfg_web_services" },
            { label: "Configure KYC", key: "right_cfg_kyc" },
            { label: "Configure messengers", key: "right_cfg_messengers" }
          ]
        }
      ]
    },
    {
      label: "Administration",
      children: [
        { label: "Access server logs", key: "right_srv_journals" },
        { label: "Receive automatic server reports", key: "right_srv_reports" },
        { label: "Request reports", key: "right_reports" }
      ]
    },
    {
      label: "Accounts",
      children: [
        { label: "Access accounts", key: "right_acc_read" },
        { label: "Accountant (deposit/withdraw)", key: "right_accountant" }
      ]
    },
    {
      label: "Trading",
      children: [
        { label: "Access orders and positions", key: "right_trades_read" },
        { label: "Dealer", key: "right_trades_dealer" },
        { label: "Manage trades", key: "right_trades_manager" },
        { label: "View quotes", key: "right_quotes" },
        { label: "Market Watch", key: "right_market" }
      ]
    },
    {
      label: "Back office",
      children: [
        { label: "Access clients", key: "right_clients_access" },
        { label: "Edit clients", key: "right_clients_edit" },
        { label: "Create clients", key: "right_clients_create" }
      ]
    },
    {
      label: "Communications",
      children: [
        { label: "Send emails", key: "right_email" },
        { label: "Send news", key: "right_news" },
        { label: "Export data", key: "right_export" }
      ]
    }
  ];

  var CUSTOM_ROLES_KEY = "hst_manager_custom_roles";

  function loadCustomRoles() {
    try {
      return JSON.parse(localStorage.getItem(CUSTOM_ROLES_KEY) || "{}");
    } catch (_) {
      return {};
    }
  }

  function saveCustomRole(name, rights) {
    var all = loadCustomRoles();
    all[name] = rights;
    localStorage.setItem(CUSTOM_ROLES_KEY, JSON.stringify(all));
  }

  function deleteCustomRole(name) {
    var all = loadCustomRoles();
    delete all[name];
    localStorage.setItem(CUSTOM_ROLES_KEY, JSON.stringify(all));
  }

  function roleNames() {
    var names = Object.keys(ROLE_PRESETS);
    Object.keys(loadCustomRoles()).forEach(function (n) {
      if (names.indexOf(n) < 0) names.push(n);
    });
    return names;
  }

  function rightsForRole(name) {
    if (ROLE_PRESETS[name]) return Object.assign({}, ROLE_PRESETS[name]);
    var custom = loadCustomRoles()[name];
    return custom ? Object.assign({}, custom) : null;
  }

  function detectRole(rights) {
    if (!rights) return "";
    var names = roleNames();
    for (var i = 0; i < names.length; i++) {
      var preset = rightsForRole(names[i]);
      if (!preset) continue;
      var keys = Object.keys(preset);
      var match = keys.every(function (k) { return !!rights[k] === !!preset[k]; });
      var extra = Object.keys(rights).filter(function (k) { return rights[k] && !preset[k]; });
      if (match && extra.length === 0) return names[i];
    }
    return "";
  }

  function initTabs(root) {
    if (global.HSTModule && HSTModule.bindConfigTabs) {
      HSTModule.bindConfigTabs(root);
      return;
    }
    var tabBar = root.querySelector(".config-tabs");
    if (!tabBar || tabBar.dataset.tabsBound) return;
    tabBar.dataset.tabsBound = "1";
    tabBar.addEventListener("click", function (e) {
      var tab = e.target.closest("[data-config-tab]");
      if (!tab || !tabBar.contains(tab)) return;
      e.preventDefault();
      var id = tab.getAttribute("data-config-tab");
      root.querySelectorAll("[data-config-tab]").forEach(function (t) {
        t.classList.toggle("active", t === tab);
      });
      root.querySelectorAll("[data-config-panel]").forEach(function (p) {
        p.classList.toggle("active", p.getAttribute("data-config-panel") === id);
      });
    });
  }

  function fillLimitSelect(select, value) {
    if (!select) return;
    select.innerHTML = "";
    LIMIT_OPTIONS.forEach(function (o) {
      var opt = document.createElement("option");
      opt.value = String(o.v);
      opt.textContent = o.l;
      if (Number(value) === o.v) opt.selected = true;
      select.appendChild(opt);
    });
  }

  function renderTree(container, tree, rights, opts) {
    opts = opts || {};
    container._expanded = container._expanded || {};
    container.innerHTML = "";
    container._rights = rights;

    function isChecked(key) {
      return !!(rights && rights[key]);
    }

    function setRight(key, on) {
      if (!container._rights) container._rights = {};
      container._rights[key] = on;
      if (opts.onChange) opts.onChange(container._rights);
    }

    function renderNode(node, depth) {
      var hasKids = node.children && node.children.length;
      var path = node.label;
      var expanded = container._expanded[path] !== false;

      function toggleExpanded(e) {
        if (e) e.stopPropagation();
        container._expanded[path] = !expanded;
        renderTree(container, tree, container._rights, opts);
      }

      var row = document.createElement("div");
      row.className = "perm-tree-row";
      row.style.paddingLeft = (4 + depth * 14) + "px";

      if (hasKids) {
        var toggle = document.createElement("span");
        toggle.className = "perm-tree-toggle";
        toggle.textContent = expanded ? "▼" : "▶";
        toggle.addEventListener("click", toggleExpanded);
        row.appendChild(toggle);
      } else {
        var spacer = document.createElement("span");
        spacer.className = "perm-tree-toggle perm-tree-spacer";
        row.appendChild(spacer);
      }

      if (node.key) {
        var label = document.createElement("label");
        label.className = "perm-tree-check";
        var cb = document.createElement("input");
        cb.type = "checkbox";
        cb.checked = isChecked(node.key);
        cb.addEventListener("change", function () {
          setRight(node.key, cb.checked);
        });
        label.appendChild(cb);
        label.appendChild(document.createTextNode(" " + node.label));
        row.appendChild(label);
      } else {
        var folder = document.createElement("span");
        folder.className = "perm-tree-folder-label";
        folder.textContent = node.label;
        if (hasKids) {
          folder.addEventListener("click", toggleExpanded);
          row.addEventListener("click", function (e) {
            if (e.target.closest(".perm-tree-toggle")) return;
            toggleExpanded(e);
          });
        }
        row.appendChild(folder);
      }

      container.appendChild(row);

      if (hasKids && expanded) {
        node.children.forEach(function (child) {
          renderNode(child, depth + 1);
        });
      }
    }

    tree.forEach(function (node) { renderNode(node, 0); });
  }

  function initRoleBar(root, rights, treeContainer, onRightsChange) {
    var roleSelect = root.querySelector("[data-role-select]");
    var btnSave = root.querySelector("[data-role-save]");
    var btnDelete = root.querySelector("[data-role-delete]");
    if (!roleSelect) return;

    function refreshRoleList(selected) {
      roleSelect.innerHTML = "";
      var empty = document.createElement("option");
      empty.value = "";
      empty.textContent = "(Custom)";
      roleSelect.appendChild(empty);
      roleNames().forEach(function (name) {
        var opt = document.createElement("option");
        opt.value = name;
        opt.textContent = name;
        roleSelect.appendChild(opt);
      });
      if (selected) roleSelect.value = selected;
    }

    function applyRights(next) {
      onRightsChange(next);
      renderTree(treeContainer, PERMISSION_TREE, next, { onChange: onRightsChange });
      refreshRoleList(detectRole(next));
    }

    refreshRoleList(detectRole(rights));

    roleSelect.addEventListener("change", function () {
      var name = roleSelect.value;
      if (!name) return;
      var preset = rightsForRole(name);
      if (preset) applyRights(preset);
    });

    if (btnSave) {
      btnSave.addEventListener("click", function () {
        var name = prompt("Save permission set as:");
        if (!name || !name.trim()) return;
        saveCustomRole(name.trim(), treeContainer._rights || rights);
        refreshRoleList(name.trim());
        roleSelect.value = name.trim();
        if (global.HSTContextMenu) HSTContextMenu.toast("Role saved: " + name.trim());
      });
    }

    if (btnDelete) {
      btnDelete.addEventListener("click", function () {
        var name = roleSelect.value;
        if (!name || ROLE_PRESETS[name]) {
          if (global.HSTContextMenu) HSTContextMenu.toast("Built-in roles cannot be deleted");
          return;
        }
        deleteCustomRole(name);
        refreshRoleList("");
        if (global.HSTContextMenu) HSTContextMenu.toast("Role deleted: " + name);
      });
    }
  }

  global.HSTManagersUI = {
    initTabs: initTabs,
    fillLimitSelect: fillLimitSelect,
    renderTree: renderTree,
    initRoleBar: initRoleBar,
    detectRole: detectRole,
    roleNames: roleNames,
    PERMISSION_TREE: PERMISSION_TREE,
    ROLE_PRESETS: ROLE_PRESETS,
    LIMIT_OPTIONS: LIMIT_OPTIONS
  };
})(window);
