/**
 * MT5-style Symbols tree, icons, path bar (admin_symbols.htm).
 */
(function (global) {
  function topType(path) {
    if (!path) return "—";
    return path.split("\\")[0] || "—";
  }

  function tradeModeLabel(mode) {
    var m = { 0: "Disabled", 1: "Long only", 2: "Short only", 3: "Close only", 4: "Full" };
    return m[mode] != null ? m[mode] : String(mode);
  }

  function tradeModeClass(mode) {
    if (mode === 0) return "sym-trade-disabled";
    if (mode === 1) return "sym-trade-long";
    if (mode === 2) return "sym-trade-short";
    if (mode === 3) return "sym-trade-close";
    return "sym-trade-full";
  }

  function sessionClass(row) {
    var trade = row.session_trade;
    var quote = row.session_quote;
    if (trade === undefined && quote === undefined) return "sym-session-open";
    if (trade === false && quote === false) return "sym-session-closed";
    if (trade === false) return "sym-session-quote";
    return "sym-session-open";
  }

  function iconCell(row) {
    var trade = tradeModeClass(row.trade_mode != null ? row.trade_mode : 4);
    var sess = sessionClass(row);
    var title = tradeModeLabel(row.trade_mode != null ? row.trade_mode : 4);
    if (row.session_trade === false) title += " — market closed";
    else if (row.session_quote === false) title += " — quote only";
    return '<span class="sym-icon-cell" title="' + title.replace(/"/g, "&quot;") + '">' +
      '<span class="sym-trade-icon ' + trade + '"></span>' +
      '<span class="sym-session-dot ' + sess + '"></span></span>';
  }

  function buildPathTree(rows) {
    var root = { name: "All symbols", path: "", children: {}, isRoot: true };
    (rows || []).forEach(function (r) {
      // Stored path is folder\symbol — tree shows folders only (admin MT5).
      var full = r.path || "";
      var folder = full.lastIndexOf("\\") >= 0 ? full.slice(0, full.lastIndexOf("\\")) : "";
      if (!folder) return;
      var parts = folder.split("\\").filter(Boolean);
      var node = root;
      var acc = "";
      parts.forEach(function (p) {
        acc = acc ? acc + "\\" + p : p;
        if (!node.children[p]) {
          node.children[p] = { name: p, path: acc, children: {} };
        }
        node = node.children[p];
      });
    });
    return root;
  }

  function renderPathBar(el, path) {
    if (!el) return;
    el.innerHTML = "";
    if (!path) {
      el.appendChild(document.createTextNode("All symbols"));
      return;
    }
    var parts = path.split("\\");
    parts.forEach(function (p, i) {
      if (i > 0) {
        var sep = document.createElement("span");
        sep.className = "sym-path-sep";
        sep.textContent = "\\";
        el.appendChild(sep);
      }
      var seg = document.createElement("span");
      seg.className = "sym-path-seg";
      seg.textContent = p;
      el.appendChild(seg);
    });
  }

  function renderTree(container, tree, opts) {
    opts = opts || {};
    container.innerHTML = "";
    container._expanded = container._expanded || { "": true };

    function renderNode(node, depth) {
      var hasKids = node.children && Object.keys(node.children).length > 0;
      var path = node.path || "";
      var expanded = container._expanded[path] !== false;

      var row = document.createElement("div");
      row.className = "sym-tree-row";
      row.dataset.path = path;
      row.style.paddingLeft = (4 + depth * 14) + "px";
      if (opts.selectedPath === path) row.classList.add("sel");

      if (hasKids) {
        var toggle = document.createElement("span");
        toggle.className = "sym-tree-toggle";
        toggle.textContent = expanded ? "▼" : "▶";
        toggle.addEventListener("click", function (e) {
          e.stopPropagation();
          container._expanded[path] = !expanded;
          renderTree(container, tree, opts);
        });
        row.appendChild(toggle);
      } else {
        var spacer = document.createElement("span");
        spacer.className = "sym-tree-toggle sym-tree-spacer";
        row.appendChild(spacer);
      }

      var icon = document.createElement("span");
      icon.className = node.isRoot ? "sym-tree-root-icon" : "sym-tree-folder";
      row.appendChild(icon);

      var label = document.createElement("span");
      label.className = "sym-tree-label";
      label.textContent = node.name;
      row.appendChild(label);

      row.addEventListener("click", function () {
        if (opts.onSelect) opts.onSelect(path);
      });

      row.addEventListener("contextmenu", function (e) {
        e.preventDefault();
        e.stopPropagation();
        if (opts.onContextMenu) opts.onContextMenu(e, path, "tree");
      });

      container.appendChild(row);

      if (hasKids && expanded) {
        Object.keys(node.children).sort().forEach(function (key) {
          renderNode(node.children[key], depth + 1);
        });
      }
    }

    renderNode(tree, 0);
  }

  /* ── Symbol Sessions tab (symbol_settings_sessions.htm) ── */

  var DAY_NAMES = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
  var SESSION_QUOTE = 0;
  var SESSION_TRADE = 1;

  function minutesToTime(m) {
    m = Number(m);
    if (m >= 1440) return "24:00";
    var h = Math.floor(m / 60);
    var min = m % 60;
    return (h < 10 ? "0" : "") + h + ":" + (min < 10 ? "0" : "") + min;
  }

  function fmtMt5DateTime(sec) {
    if (sec == null || sec === 0) return "1970.01.01 00:00";
    var d = new Date(Number(sec) * (String(sec).length > 10 ? 1 : 1000));
    if (isNaN(d.getTime())) return "1970.01.01 00:00";
    var y = d.getFullYear();
    var mo = d.getMonth() + 1;
    var da = d.getDate();
    var h = d.getHours();
    var mi = d.getMinutes();
    function pad(n) { return n < 10 ? "0" + n : String(n); }
    return y + "." + pad(mo) + "." + pad(da) + " " + pad(h) + ":" + pad(mi);
  }

  function windowsForDay(sessions, type, day) {
    return (sessions || [])
      .filter(function (s) { return s.type === type && s.day === day; })
      .sort(function (a, b) { return a.open - b.open; });
  }

  function formatDaySessions(sessions, type, day) {
    var wins = windowsForDay(sessions, type, day);
    if (!wins.length) return "";
    return wins.map(function (w) {
      return minutesToTime(w.open) + "-" + minutesToTime(w.close);
    }).join(", ");
  }

  function sessionsEqual(a, b) {
    if (a.length !== b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i].open !== b[i].open || a[i].close !== b[i].close) return false;
    }
    return true;
  }

  function hasSeparateTrade(sessions, day) {
    return !sessionsEqual(
      windowsForDay(sessions, SESSION_QUOTE, day),
      windowsForDay(sessions, SESSION_TRADE, day)
    );
  }

  function closeSessionDialog() {
    var dlg = document.getElementById("symSessionDialog");
    if (dlg) dlg.remove();
  }

  function buildTimelineRow(label, windows) {
    var row = document.createElement("div");
    row.className = "sym-timeline-row";

    var lbl = document.createElement("span");
    lbl.className = "sym-timeline-label";
    lbl.textContent = label;
    row.appendChild(lbl);

    var trackWrap = document.createElement("div");
    trackWrap.className = "sym-timeline-track-wrap";

    var scale = document.createElement("div");
    scale.className = "sym-timeline-scale";
    [0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24].forEach(function (h) {
      var tick = document.createElement("span");
      tick.className = "sym-timeline-tick" + (h === 24 ? " end" : "");
      tick.style.left = (h / 24 * 100) + "%";
      tick.textContent = h < 10 ? "0" + h : String(h);
      scale.appendChild(tick);
    });
    trackWrap.appendChild(scale);

    var track = document.createElement("div");
    track.className = "sym-timeline-track";
    windows.forEach(function (w) {
      var seg = document.createElement("div");
      seg.className = "sym-timeline-seg";
      seg.style.left = (w.open / 1440 * 100) + "%";
      seg.style.width = ((w.close - w.open) / 1440 * 100) + "%";
      track.appendChild(seg);

      var mOpen = document.createElement("div");
      mOpen.className = "sym-timeline-marker";
      mOpen.style.left = (w.open / 1440 * 100) + "%";
      mOpen.innerHTML = '<span class="sym-timeline-marker-label">' + minutesToTime(w.open) + "</span><span class=\"sym-timeline-marker-pin\"></span>";
      track.appendChild(mOpen);

      var mClose = document.createElement("div");
      mClose.className = "sym-timeline-marker";
      mClose.style.left = (w.close / 1440 * 100) + "%";
      mClose.innerHTML = '<span class="sym-timeline-marker-label">' + minutesToTime(w.close) + "</span><span class=\"sym-timeline-marker-pin\"></span>";
      track.appendChild(mClose);
    });
    trackWrap.appendChild(track);
    row.appendChild(trackWrap);
    return row;
  }

  function openSessionEditor(data, day) {
    closeSessionDialog();
    var symbol = data.symbol || "…";
    var dayName = DAY_NAMES[day] || "Day";
    var sessions = data.sessions || [];
    var quoteWins = windowsForDay(sessions, SESSION_QUOTE, day);
    var tradeWins = windowsForDay(sessions, SESSION_TRADE, day);
    if (!tradeWins.length && quoteWins.length) tradeWins = quoteWins.slice();
    var separate = hasSeparateTrade(sessions, day);

    var overlay = document.createElement("div");
    overlay.className = "sym-session-dialog-overlay";
    overlay.id = "symSessionDialog";

    var dlg = document.createElement("div");
    dlg.className = "sym-session-dialog";
    dlg.innerHTML =
      '<div class="sym-session-dialog-title">Sessions ' + symbol + ": " + dayName + "</div>" +
      '<div class="sym-session-dialog-body"></div>' +
      '<div class="sym-session-dialog-footer">' +
      '<label class="sym-session-separate"><input type="checkbox"' + (separate ? " checked" : "") + " disabled> Enable separate trading sessions</label>" +
      '<div class="sym-session-dialog-actions">' +
      '<button type="button" class="sym-session-ok">OK</button>' +
      '<button type="button" class="sym-session-cancel">Cancel</button>' +
      "</div></div>";

    var body = dlg.querySelector(".sym-session-dialog-body");
    var intro = document.createElement("div");
    intro.className = "sym-sessions-intro sym-sessions-intro-compact";
    intro.innerHTML =
      '<span class="sym-sessions-clock-icon" aria-hidden="true"></span>' +
      "<p>The setting up of sessions within the selected day. It is possible to determine several sessions of each type, " +
      "trading sessions must be within the quotation ones. If specific trade sessions are not determined they will coincide with quotation ones.</p>";
    body.appendChild(intro);
    body.appendChild(buildTimelineRow("Quotes:", quoteWins));
    body.appendChild(buildTimelineRow("Trade:", tradeWins));

    overlay.appendChild(dlg);
    document.body.appendChild(overlay);

    function close() { closeSessionDialog(); }
    overlay.addEventListener("mousedown", function (e) {
      if (e.target === overlay) close();
    });
    dlg.querySelector(".sym-session-ok").addEventListener("click", close);
    dlg.querySelector(".sym-session-cancel").addEventListener("click", close);
  }

  function renderSessionsPanel(root, data) {
    var panel = root.querySelector("[data-config-panel=\"sessions\"]");
    if (!panel) return;

    var sessions = data.sessions || [];
    var useLimits = !!(data.time_start || data.time_expiration);
    panel.innerHTML =
      '<div class="sym-sessions-intro">' +
      '<span class="sym-sessions-header-icon" aria-hidden="true"></span>' +
      "<p>The setting up of trade and quotation sessions of the symbol by days. During the quotation session it is possible to view the price dynamics but trading is prohibited.</p>" +
      "</div>" +
      '<div class="sym-sessions-table-wrap">' +
      '<table class="sym-sessions-table">' +
      "<thead><tr><th>Day</th><th>Quotes</th><th>Trade</th></tr></thead>" +
      '<tbody data-sym-sessions-tbody></tbody></table></div>' +
      '<div class="sym-sessions-toolbar"><button type="button" class="sym-sessions-edit" disabled>Edit</button></div>' +
      '<div class="sym-sessions-limits">' +
      '<label><input type="checkbox" data-sym-use-limits' + (useLimits ? " checked" : "") + " disabled> Use time limits</label>" +
      '<label>From:</label><input type="text" readonly value="' + fmtMt5DateTime(data.time_start) + '">' +
      '<label>To:</label><input type="text" readonly value="' + fmtMt5DateTime(data.time_expiration) + '">' +
      "</div>";

    var tbody = panel.querySelector("[data-sym-sessions-tbody]");
    var editBtn = panel.querySelector(".sym-sessions-edit");
    var selectedDay = null;

    DAY_NAMES.forEach(function (name, day) {
      var tr = document.createElement("tr");
      tr.dataset.day = String(day);
      tr.innerHTML =
        '<td><span class="sym-day-icon" aria-hidden="true"></span>' + name + "</td>" +
        "<td>" + (formatDaySessions(sessions, SESSION_QUOTE, day) || "&nbsp;") + "</td>" +
        "<td>" + (formatDaySessions(sessions, SESSION_TRADE, day) || "&nbsp;") + "</td>";
      tr.addEventListener("click", function () {
        tbody.querySelectorAll("tr").forEach(function (r) { r.classList.remove("selected"); });
        tr.classList.add("selected");
        selectedDay = day;
        editBtn.disabled = false;
      });
      tr.addEventListener("dblclick", function () {
        tr.click();
        openSessionEditor(data, day);
      });
      tbody.appendChild(tr);
    });

    editBtn.addEventListener("click", function () {
      if (selectedDay == null) return;
      openSessionEditor(data, selectedDay);
    });
  }

  global.HSTSymbolsUI = {
    topType: topType,
    tradeModeLabel: tradeModeLabel,
    iconCell: iconCell,
    buildPathTree: buildPathTree,
    renderPathBar: renderPathBar,
    renderTree: renderTree,
    DAY_NAMES: DAY_NAMES,
    fmtMt5DateTime: fmtMt5DateTime,
    formatDaySessions: formatDaySessions,
    renderSessionsPanel: renderSessionsPanel
  };
})(window);
