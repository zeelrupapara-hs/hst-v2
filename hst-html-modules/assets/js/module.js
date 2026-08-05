/**
 * Shared module helpers — list views, config tabs, API status bar.
 */
(function (global) {
  function esc(s) {
    if (s == null) return "";
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function fmtAuthMode(n) {
    var m = { 0: "Normal", 1: "RSA 1024", 2: "RSA 2048", 3: "Custom SSL" };
    return m[n] != null ? m[n] : String(n);
  }

  function fmtExecMode(n) {
    var m = { 0: "Request", 1: "Instant", 2: "Market", 3: "Exchange" };
    return m[n] != null ? m[n] : String(n);
  }

  function fmtClientStatus(n) {
    var m = { 0: "Registered", 1: "Active", 2: "Suspended", 3: "Closed" };
    return m[n] != null ? m[n] : String(n);
  }

  function fmtKycStatus(n) {
    var m = { 0: "None", 1: "Pending", 2: "Approved", 3: "Rejected" };
    return m[n] != null ? m[n] : String(n);
  }

  function renderTable(tbody, rows, columns, opts) {
    if (!tbody) return;
    opts = opts || {};
    tbody.innerHTML = "";
    if (!rows.length) {
      tbody.innerHTML = '<tr><td colspan="' + columns.length + '">No records</td></tr>';
      return;
    }
    rows.forEach(function (row, idx) {
      var tr = document.createElement("tr");
      if (opts.selectable !== false && idx === 0) tr.className = "selected";
      if (opts.rowAttrs) opts.rowAttrs(tr, row, idx);
      columns.forEach(function (col) {
        var td = document.createElement("td");
        var v = col.render ? col.render(row, idx) : row[col.key];
        td.innerHTML = col.html ? v : esc(v);
        tr.appendChild(td);
      });
      tbody.appendChild(tr);
    });
  }

  function fmtTs(sec) {
    if (sec == null || sec === 0) return "—";
    var d = new Date(Number(sec) * (String(sec).length > 10 ? 1 : 1000));
    if (isNaN(d.getTime())) return String(sec);
    return d.toISOString().replace("T", " ").slice(0, 19);
  }

  function fillFields(root, data) {
    if (!data) return;
    root.querySelectorAll("[data-field]").forEach(function (el) {
      var key = el.getAttribute("data-field");
      var val = data[key];
      if (Array.isArray(val)) val = val.join(", ");
      if (el.tagName === "INPUT" || el.tagName === "SELECT" || el.tagName === "TEXTAREA") {
        el.value = val != null ? val : "";
      } else {
        el.textContent = val != null ? val : "—";
      }
    });
  }

  function val(id, root) {
    var el = (root || document).querySelector("[data-field='" + id + "']");
    return el ? el.value : "";
  }

  function useDemo() {
    return !global.HSTApi || !HSTApi.isLoggedIn();
  }

  function setApiBar(root, state) {
    var bar = root.querySelector("[data-api-bar]");
    if (!bar) return;
    var dot = bar.querySelector("[data-api-dot]");
    var msg = bar.querySelector("[data-api-msg]");
    if (state.demo) {
      if (dot) dot.className = "api-dot warn";
      if (msg) msg.textContent = state.message || "Demo data — sign in via Connect API for live server";
      return;
    }
    if (!HSTApi.isLoggedIn()) {
      if (dot) dot.className = "api-dot warn";
      if (msg) msg.innerHTML = 'Demo mode — <a href="../../login.html" target="_top">Connect API</a> for live data';
      return;
    }
    if (state.loading) {
      if (dot) dot.className = "api-dot load";
      if (msg) msg.textContent = "Loading…";
    } else if (state.error) {
      if (dot) dot.className = "api-dot err";
      if (msg) msg.textContent = state.error;
    } else {
      if (dot) dot.className = "api-dot ok";
      if (msg) msg.textContent = state.message || "API OK";
    }
  }

  async function fetchList(cfg) {
    var q = Object.assign({}, cfg.query || {});
    if (useDemo()) {
      return { ok: true, data: HSTMock.list(cfg.endpoint, q), elapsed: 0, demo: true };
    }
    var path = cfg.endpoint + buildQuery(q);
    var res = await HSTApi.get(path);
    if (!res.ok) {
      return { ok: true, data: HSTMock.list(cfg.endpoint, q), elapsed: res.elapsed, demo: true, fallback: res.error || res.status };
    }
    return res;
  }

  async function fetchOne(endpoint, id) {
    if (useDemo()) {
      return { ok: true, data: HSTMock.get(endpoint, id), elapsed: 0, demo: true };
    }
    var res = await HSTApi.get(endpoint + "/" + encodeURIComponent(id));
    if (!res.ok) {
      return { ok: true, data: HSTMock.get(endpoint, id), elapsed: res.elapsed, demo: true, fallback: res.status };
    }
    return res;
  }

  function renderRows(tbody, rows, columns, selectedSet, idKey) {
    tbody.innerHTML = "";
    rows.forEach(function (row, idx) {
      var tr = document.createElement("tr");
      if (selectedSet && selectedSet[idx]) tr.className = "selected";
      tr.dataset.index = String(idx);
      if (idKey && row[idKey] != null) tr.dataset.id = String(row[idKey]);
      columns.forEach(function (col) {
        var td = document.createElement("td");
        var v = col.render ? col.render(row, idx) : row[col.key];
        td.innerHTML = col.html ? v : esc(v);
        tr.appendChild(td);
      });
      tbody.appendChild(tr);
    });
  }

  function bindRowSelect(tbody, controller) {
    var anchor = 0;
    tbody.addEventListener("click", function (e) {
      var tr = e.target.closest("tr");
      if (!tr || tr.parentElement !== tbody) return;
      var idx = Number(tr.dataset.index);
      if (isNaN(idx)) return;

      if (e.shiftKey) {
        var from = Math.min(anchor, idx);
        var to = Math.max(anchor, idx);
        controller.clearSelection();
        for (var i = from; i <= to; i++) controller.selectIndex(i, true);
      } else if (e.ctrlKey || e.metaKey) {
        controller.toggleIndex(idx);
        anchor = idx;
      } else {
        controller.clearSelection();
        controller.selectIndex(idx, true);
        anchor = idx;
      }
    });
  }

  function applyTableView(controller) {
    var table = controller.table;
    if (!table) return;
    table.classList.toggle("data-table-grid", !!controller.view.grid);
    table.classList.toggle("data-table-auto", !!controller.view.autoArrange);
    var hidden = controller.view.hiddenCols || {};
    var ths = table.querySelectorAll("thead th");
    ths.forEach(function (th, i) {
      th.style.display = hidden[i] ? "none" : "";
    });
    table.querySelectorAll("tbody tr").forEach(function (tr) {
      Array.from(tr.children).forEach(function (td, i) {
        td.style.display = hidden[i] ? "none" : "";
      });
    });
  }

  function createListController(root, tbody, cfg, state, loadFn) {
    var table = tbody ? tbody.closest("table") : null;
    var editBtn = root.querySelector("[data-edit]");
    var reloadBtn = root.querySelector("[data-reload]");

    var controller = {
      root: root,
      tbody: tbody,
      table: table,
      cfg: cfg,
      state: state,
      dbMode: "live",
      view: {
        grid: true,
        autoArrange: false,
        showMs: false,
        enabledOnly: false,
        columnSort: false,
        serverFilter: "",
        hiddenCols: {}
      },
      can: {},
      rights: {},

      getRows: function () { return state.rows; },

      getSelectedIndices: function () {
        if (!tbody) return [];
        var out = [];
        tbody.querySelectorAll("tr.selected").forEach(function (tr) {
          var i = Number(tr.dataset.index);
          if (!isNaN(i)) out.push(i);
        });
        out.sort(function (a, b) { return a - b; });
        return out;
      },

      getSelectedRows: function () {
        var self = this;
        return this.getSelectedIndices().map(function (i) { return self.state.rows[i]; }).filter(Boolean);
      },

      clearSelection: function () {
        if (!tbody) return;
        tbody.querySelectorAll("tr.selected").forEach(function (r) { r.classList.remove("selected"); });
        state.selected = -1;
      },

      selectIndex: function (idx, on) {
        if (!tbody) return;
        var tr = tbody.querySelector('tr[data-index="' + idx + '"]');
        if (!tr) return;
        if (on) tr.classList.add("selected");
        else tr.classList.remove("selected");
        state.selected = idx;
      },

      toggleIndex: function (idx) {
        if (!tbody) return;
        var tr = tbody.querySelector('tr[data-index="' + idx + '"]');
        if (!tr) return;
        tr.classList.toggle("selected");
        state.selected = idx;
      },

      syncSelectionFromDom: function () {
        var indices = this.getSelectedIndices();
        state.selected = indices.length ? indices[indices.length - 1] : -1;
      },

      redraw: function () {
        var sel = {};
        this.getSelectedIndices().forEach(function (i) { sel[i] = true; });
        renderRows(tbody, state.rows, cfg.columns, sel, cfg.idKey);
        applyTableView(this);
      },

      reload: function () { return loadFn(); },

      openEdit: function () {
        var rows = this.getSelectedRows();
        if (!rows.length || !cfg.configPage) return;
        var row = rows[rows.length - 1];
        var id = row[cfg.idKey];
        window.location.href = cfg.configPage + encodeURIComponent(id);
      },

      toggleView: function (key) {
        this.view[key] = !this.view[key];
        if (key === "enabledOnly") {
          loadFn();
          return;
        }
        applyTableView(this);
      },

      toggleColumn: function (index) {
        this.view.hiddenCols[index] = !this.view.hiddenCols[index];
        applyTableView(this);
      },

      getColumns: function () {
        if (!table) return [];
        var ths = table.querySelectorAll("thead th");
        var out = [];
        ths.forEach(function (th, i) {
          out.push({ index: i, label: (th.textContent || "").trim() || ("Column " + (i + 1)) });
        });
        return out;
      },

      rowsHaveKey: function (c, key) {
        return c.getSelectedRows().some(function (r) { return r[key] != null; });
      },

      formatRowsAsLines: function () {
        var cols = cfg.columns;
        return this.getSelectedRows().map(function (row) {
          return cols.map(function (col) {
            var v = col.render ? col.render(row, 0) : row[col.key];
            if (v == null) return "";
            return String(v).replace(/<[^>]+>/g, "");
          }).join("\t");
        }).join("\n");
      },

      formatRowsAsCsv: function () {
        var cols = cfg.columns;
        var header = this.getColumns().map(function (c) { return c.label; });
        var lines = [header.join(",")];
        this.getSelectedRows().forEach(function (row) {
          lines.push(cols.map(function (col) {
            var v = col.render ? col.render(row, 0) : row[col.key];
            if (v == null) v = "";
            v = String(v).replace(/<[^>]+>/g, "");
            if (/[",\n]/.test(v)) v = '"' + v.replace(/"/g, '""') + '"';
            return v;
          }).join(","));
        });
        return lines.join("\n");
      }
    };

    bindRowSelect(tbody, controller);

    if (reloadBtn) reloadBtn.addEventListener("click", function () { loadFn(); });
    if (editBtn && cfg.configPage) {
      editBtn.addEventListener("click", function () { controller.openEdit(); });
    }

    if (cfg.configPage && tbody) {
      tbody.addEventListener("dblclick", function (e) {
        var tr = e.target.closest("tr");
        if (!tr || tr.parentElement !== tbody) return;
        var idx = Number(tr.dataset.index);
        if (idx >= 0 && state.rows[idx]) {
          window.location.href = cfg.configPage + encodeURIComponent(state.rows[idx][cfg.idKey]);
        }
      });
    }

    if (cfg.menu && global.HSTContextMenu) {
      var menuTargets = [];
      function addMenuTarget(el) {
        if (el && menuTargets.indexOf(el) < 0) menuTargets.push(el);
      }
      addMenuTarget(tbody);
      addMenuTarget(root.querySelector(".table-wrap"));
      addMenuTarget(root.querySelector(".module-content"));
      addMenuTarget(root.querySelector(".module-split"));
      if (cfg.contextMenuRoot) addMenuTarget(root.querySelector(cfg.contextMenuRoot));
      if (!menuTargets.length) addMenuTarget(root);
      menuTargets.forEach(function (target) {
        target.addEventListener("contextmenu", function (e) {
          var tr = e.target.closest("tr");
          if (tr && tr.parentElement === tbody && !tr.classList.contains("selected")) {
            controller.clearSelection();
            controller.selectIndex(Number(tr.dataset.index), true);
          }
          e.preventDefault();
          var menuKey = cfg.menu;
          if (typeof cfg.resolveMenu === "function") {
            menuKey = cfg.resolveMenu(controller, e) || menuKey;
          }
          HSTContextMenu.show(menuKey, controller, e.clientX, e.clientY);
        });
      });
    }

    return controller;
  }

  function buildQuery(params) {
    var parts = [];
    Object.keys(params || {}).forEach(function (k) {
      if (params[k] != null && params[k] !== "") parts.push(encodeURIComponent(k) + "=" + encodeURIComponent(params[k]));
    });
    return parts.length ? "?" + parts.join("&") : "";
  }

  function statusMessage(cfg, res, count) {
    var base = cfg.title + ": " + count + " row(s)";
    if (res.demo) return base + " — demo data";
    return base + " — " + res.elapsed + " ms";
  }

  /** Flat list module: GET endpoint → table */
  function bootList(cfg) {
    var root = document.querySelector("[data-module]");
    if (!root) return;

    var tbody = root.querySelector("[data-tbody]");
    var searchInput = root.querySelector("[data-search]");
    var state = { rows: [], selected: -1 };
    var controller = null;

    async function load() {
      setApiBar(root, { loading: true });
      var q = Object.assign({}, cfg.query || {});
      if (searchInput && searchInput.value.trim()) q.search = searchInput.value.trim();
      var res = await fetchList(Object.assign({}, cfg, { query: q }));

      if (cfg.transform) {
        state.rows = cfg.transform(res.data) || [];
      } else {
        state.rows = Array.isArray(res.data) ? res.data : (res.data && res.data.items) || [];
      }
      if (controller && controller.view.enabledOnly) {
        state.rows = state.rows.filter(function (r) {
          return r.enable !== false && r.enabled !== false && r.status !== "disabled";
        });
      }
      state.selected = state.rows.length ? 0 : -1;
      controller.redraw();
      if (state.rows.length) {
        controller.clearSelection();
        controller.selectIndex(0, true);
      }
      setApiBar(root, {
        demo: res.demo,
        message: statusMessage(cfg, res, state.rows.length) + (res.fallback ? " (API fallback)" : "")
      });
    }

    controller = createListController(root, tbody, cfg, state, load);
    applyTableView(controller);
    if (searchInput) {
      searchInput.addEventListener("keydown", function (e) {
        if (e.key === "Enter") load();
      });
    }
    load();
    return controller;
  }

  /** Config window with tab panels */
  function bindConfigTabs(root) {
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

  function bootConfig(cfg) {
    var root = document.querySelector("[data-config-module]");
    if (!root) return;

    bindConfigTabs(root);

    async function loadDetail() {
      setApiBar(root, { loading: true });
      var id = cfg.getId();
      if (!id && useDemo()) id = HSTMock.defaultId(cfg.endpoint);
      if (!id) {
        setApiBar(root, { error: "No id in URL" });
        return;
      }
      var res = await fetchOne(cfg.endpoint, id);
      if (!res.data) {
        setApiBar(root, { error: "Record not found" });
        return;
      }
      if (cfg.fillForm) await cfg.fillForm(res.data, root, id, res);
      else fillFields(root, res.data);
      setApiBar(root, {
        demo: res.demo,
        message: cfg.title + " #" + id + (res.demo ? " — demo data" : " — " + res.elapsed + " ms")
      });
    }

    var reloadBtn = root.querySelector("[data-reload]");
    if (reloadBtn) reloadBtn.addEventListener("click", loadDetail);

    var saveBtn = root.querySelector("[data-save]");
    if (saveBtn && cfg.onSave) saveBtn.addEventListener("click", cfg.onSave);

    loadDetail();
  }

  /** Single-record GET → form fields (Time, End of Day, etc.) */
  function bootDetail(cfg) {
    var root = document.querySelector("[data-detail-module]");
    if (!root) return;

    async function load() {
      setApiBar(root, { loading: true });
      var res;
      if (useDemo()) {
        res = { ok: true, data: HSTMock.get(cfg.endpoint), elapsed: 0, demo: true };
      } else {
        res = await HSTApi.get(cfg.endpoint);
        if (!res.ok) {
          res = { ok: true, data: HSTMock.get(cfg.endpoint), elapsed: res.elapsed, demo: true, fallback: res.status };
        }
      }
      if (cfg.fillForm) cfg.fillForm(res.data, root);
      else fillFields(root, res.data);
      setApiBar(root, {
        demo: res.demo,
        message: cfg.title + (res.demo ? " — demo data" : " — " + res.elapsed + " ms")
      });
    }

    var reloadBtn = root.querySelector("[data-reload]");
    if (reloadBtn) reloadBtn.addEventListener("click", load);
    var saveBtn = root.querySelector("[data-save]");
    if (saveBtn && cfg.onSave) saveBtn.addEventListener("click", cfg.onSave);

    load();
  }

  /** Flatten group tree for table display */
  function flattenGroups(nodes, out) {
    out = out || [];
    (nodes || []).forEach(function (n) {
      if (n.exists !== false && n.group_id) {
        out.push(n);
      }
      if (n.groups) flattenGroups(n.groups, out);
    });
    return out;
  }

  global.HSTModule = {
    esc: esc,
    fmtAuthMode: fmtAuthMode,
    fmtExecMode: fmtExecMode,
    fmtClientStatus: fmtClientStatus,
    fmtKycStatus: fmtKycStatus,
    fmtTs: fmtTs,
    fillFields: fillFields,
    val: val,
    useDemo: useDemo,
    setApiBar: setApiBar,
    fetchList: fetchList,
    fetchOne: fetchOne,
    bootList: bootList,
    bootConfig: bootConfig,
    bindConfigTabs: bindConfigTabs,
    bootDetail: bootDetail,
    flattenGroups: flattenGroups,
    buildQuery: buildQuery,
    renderRows: renderRows,
    renderTable: renderTable,
    statusMessage: statusMessage,
    createListController: createListController,
    applyTableView: applyTableView
  };
})(window);
