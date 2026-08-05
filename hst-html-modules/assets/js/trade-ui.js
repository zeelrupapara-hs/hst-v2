/**
 * MT5-style trading operation dialog (admin_deals.htm#view, admin_orders.htm#view, admin_positions.htm#view).
 */
(function (global) {
  var ORDER_TYPE = {
    0: "Buy", 1: "Sell", 2: "Buy Limit", 3: "Sell Limit", 4: "Buy Stop", 5: "Sell Stop",
    6: "Buy Stop Limit", 7: "Sell Stop Limit", 8: "Close By"
  };
  var ORDER_STATE = {
    0: "STARTED", 1: "PLACED", 2: "CANCELED", 3: "PARTIAL", 4: "FILLED",
    5: "REJECTED", 6: "EXPIRED", 7: "REQUEST ADD", 8: "REQUEST MODIFY", 9: "REQUEST CANCEL"
  };
  var ORDER_FILL = { 0: "ALL OR NONE", 1: "IMMEDIATE OR CANCEL", 2: "RETURN", 3: "BOOK OR CANCEL" };
  var ORDER_TIME = { 0: "GTC", 1: "DAY", 2: "SPECIFIED", 3: "SPECIFIED DAY" };
  var DEAL_ACTION = {
    0: "Buy", 1: "Sell", 2: "Balance", 3: "Credit", 4: "Charge", 5: "Correction",
    6: "Bonus", 7: "Commission", 19: "SO Compensation"
  };
  var DEAL_ENTRY = { 0: "IN", 1: "OUT", 2: "IN/OUT", 3: "OUT BY" };
  var REASON = {
    0: "Client", 1: "Expert", 2: "Dealer", 3: "Stop loss", 4: "Take profit", 5: "Stop out",
    6: "Rollover", 7: "External Client", 8: "Variation margin", 9: "Gateway", 10: "Signal",
    13: "Synchronization", 16: "Mobile", 17: "Web"
  };
  var POS_ACTION = { 0: "Buy", 1: "Sell" };

  function esc(s) {
    if (s == null || s === "") return "";
    return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/"/g, "&quot;");
  }

  function fmtTs(sec, ms) {
    if (sec == null || sec === 0) return "1970.01.01 00:00:00";
    var n = Number(sec);
    var d = new Date(n * (String(sec).length > 10 ? 1 : 1000));
    if (isNaN(d.getTime())) return String(sec);
    var pad = function (x, w) { return String(x).padStart(w, "0"); };
    var s = d.getFullYear() + "." + pad(d.getMonth() + 1, 2) + "." + pad(d.getDate(), 2) + " " +
      pad(d.getHours(), 2) + ":" + pad(d.getMinutes(), 2) + ":" + pad(d.getSeconds(), 2);
    if (ms != null) s += "." + pad(ms, 3);
    return s;
  }

  function fmtNum(v, digits) {
    if (v == null || v === "") return "0";
    var n = Number(v);
    if (isNaN(n)) return String(v);
    if (digits != null) return n.toFixed(digits);
    return String(n);
  }

  function fmtVol(v) {
    if (v == null) return "0.00";
    return fmtNum(v, 2);
  }

  function queryParam(name) {
    var m = location.search.match(new RegExp("[?&]" + name + "=([^&]+)"));
    return m ? decodeURIComponent(m[1]) : "";
  }

  function field(label, value, opts) {
    opts = opts || {};
    var ro = opts.readonly !== false ? " readonly" : "";
    var cls = opts.wide ? " trade-field-wide" : "";
    var input = opts.link
      ? '<a class="trade-link" href="' + esc(opts.link) + '">' + esc(value) + "</a>"
      : '<input class="trade-input" value="' + esc(value == null ? "" : value) + '"' + ro + ">";
    return '<div class="trade-field' + cls + '"><label>' + esc(label) + "</label>" + input + "</div>";
  }

  function selectField(label, value, options) {
    var html = '<div class="trade-field"><label>' + esc(label) + '</label><select class="trade-input">';
    (options || []).forEach(function (o) {
      var sel = String(o).toUpperCase() === String(value).toUpperCase() ? " selected" : "";
      html += "<option" + sel + ">" + esc(o) + "</option>";
    });
    html += "</select></div>";
    return html;
  }

  function buildChain(type, record, chain) {
    return (chain || []).map(function (item) {
      var kind = item.kind || "deal";
      var icon = kind === "order"
        ? (global.HSTIcons ? global.HSTIcons.img("order") : "📄")
        : (item.entry === 1 || item.entry === "OUT"
          ? (global.HSTIcons ? global.HSTIcons.img("deal-out") : "↩")
          : (global.HSTIcons ? global.HSTIcons.img("deal") : "📄"));
      var vol = item.volume_initial != null
        ? fmtVol(item.volume) + " / " + fmtVol(item.volume_initial)
        : fmtVol(item.volume);
      return {
        kind: kind,
        ticket: item.ticket || item.deal_id || item.order_id || item.position_id,
        icon: icon,
        time: fmtTs(item.time || item.time_setup || item.time_create, item.time_ms),
        type: item.type_label || item.action_label || item.type || "",
        volume: vol,
        price: fmtNum(item.price || item.price_order || item.price_open, item.digits || 5),
        reason: REASON[item.reason] || item.reason_label || REASON[5] || "—",
        profit: item.profit != null ? fmtNum(item.profit, 2) : "",
        selected: (function () {
          var tid = String(item.ticket || item.deal_id || item.order_id || item.position_id);
          if (type === "deal") return kind === "deal" && tid === String(r.deal_id);
          if (type === "order") return kind === "order" && tid === String(r.order_id);
          return kind === "position" && tid === String(r.position_id);
        })()
      };
    });
  }

  function defaultChain(type, r) {
    var items = [];
    if (r.order_id) {
      items.push({
        kind: "order", ticket: r.order_id, time: r.time_setup || r.time, time_ms: r.time_setup_ms,
        type_label: ORDER_TYPE[r.type] || "Sell", volume: r.volume, volume_initial: r.volume_initial || r.volume,
        price: r.price_order || r.price, reason: r.reason, digits: r.digits || 5
      });
    }
    if (type === "deal" || r.deal_id) {
      items.push({
        kind: "deal", ticket: r.deal_id, entry: r.entry, time: r.time, time_ms: r.time_ms,
        action_label: (DEAL_ACTION[r.action] || "Sell").toLowerCase(), volume: r.volume,
        price: r.price, reason: r.reason, profit: r.profit, digits: r.digits || 5
      });
    }
    if (r.position_id && type !== "order") {
      items.push({
        kind: "position", ticket: r.position_id, time: r.time_create,
        action_label: (POS_ACTION[r.action] || "buy").toLowerCase(), volume: r.volume,
        price: r.price_open || r.price, reason: r.reason, digits: r.digits || 5
      });
    }
    return items;
  }

  function titleFor(type, r) {
    if (type === "deal") {
      return "Deal #" + r.deal_id + " " + (DEAL_ACTION[r.action] || "sell").toLowerCase() + " " +
        fmtVol(r.volume) + " " + (r.symbol || "") + " at " + fmtNum(r.price, r.digits || 5);
    }
    if (type === "order") {
      var vol = r.volume_initial != null ? fmtVol(r.volume) + " / " + fmtVol(r.volume_initial) : fmtVol(r.volume);
      return "Order #" + r.order_id + " " + (ORDER_TYPE[r.type] || "sell").toLowerCase() + " " +
        vol + " " + (r.symbol || "") + " at " + fmtNum(r.price_order || r.price, r.digits || 5);
    }
    return "Position #" + r.position_id + " " + (POS_ACTION[r.action] || "buy").toLowerCase() + " " +
      fmtVol(r.volume) + " " + (r.symbol || "") + " at " + fmtNum(r.price_open, r.digits || 5);
  }

  function renderAccount(user) {
    if (!user) return "Account —";
    return "👤 " + [user.name, user.login, user.group, "1 : " + (user.leverage || "—")].join(", ");
  }

  function renderDealDetails(r) {
    var sym = r.symbol + (r.symbol_description ? ", " + r.symbol_description : "");
    return '<div class="trade-form-row">' +
      field("Deal", r.deal_id) +
      field("Position", r.position_id, { link: r.position_id ? "positions-detail.html?id=" + r.position_id : null }) +
      field("Order", r.order_id, { link: r.order_id ? "orders-detail.html?id=" + r.order_id : null }) +
      "</div>" +
      '<div class="trade-form-row">' +
      selectField("Action", DEAL_ACTION[r.action] || "Sell", ["Buy", "Sell"]) +
      selectField("Entry", DEAL_ENTRY[r.entry] || "IN", ["IN", "OUT", "IN/OUT", "OUT BY"]) +
      field("Volume", fmtVol(r.volume)) +
      selectField("Symbol", sym, [sym]) +
      "</div>" +
      '<div class="trade-form-row">' +
      selectField("Gateway", r.gateway_action || "BUY", ["BUY", "SELL"]) +
      field("Closed volume", fmtVol(r.volume_closed || r.volume) + "  " + fmtVol(r.volume_closed_remain || 0)) +
      selectField("Reason", REASON[r.reason] || "Stop out", Object.values(REASON)) +
      field("Create time", fmtTs(r.time, r.time_ms)) +
      "</div>" +
      '<div class="trade-form-grid">' +
      field("Dealer ID", r.dealer || 0) + field("Expert ID", r.expert_id || "") +
      field("Party ID", r.party_id || "") + field("External ID", r.external_id || "") +
      field("Market Bid", fmtNum(r.market_bid, r.digits || 5)) +
      field("Market Ask", fmtNum(r.market_ask, r.digits || 5)) +
      field("Market Last", fmtNum(r.market_last || 0, r.digits || 5)) +
      field("Comment", r.comment || "", { wide: true }) +
      field("Position price", fmtNum(r.position_price || r.price, r.digits || 5)) +
      field("Gateway price", fmtNum(r.gateway_price || 0, r.digits || 5)) +
      field("Price", fmtNum(r.price, r.digits || 5)) +
      field("Stop loss", fmtNum(r.price_sl || 0, r.digits || 5)) +
      field("Take profit", fmtNum(r.price_tp || 0, r.digits || 5)) +
      field("Commission", fmtNum(r.commission || 0, 2)) +
      field("Fee", fmtNum(r.fee || 0, 2)) +
      field("Swap", fmtNum(r.storage || 0, 2)) +
      field("Profit", fmtNum(r.profit || 0, 2)) +
      field("Raw profit", fmtNum(r.raw_profit || r.profit || 0, 2)) +
      field("Profit rate", fmtNum(r.rate_profit || 0, 2)) +
      field("Margin rate", fmtNum(r.rate_margin || 0, 8)) +
      field("Modifications", r.modifications || "", { wide: true }) +
      "</div>";
  }

  function renderOrderDetails(r) {
    var sym = r.symbol + (r.symbol_description ? ", " + r.symbol_description : "");
    return '<div class="trade-form-row">' +
      field("Order", r.order_id) +
      field("Position", r.position_id, { link: r.position_id ? "positions-detail.html?id=" + r.position_id : null }) +
      selectField("Type", ORDER_TYPE[r.type] || "SELL", Object.values(ORDER_TYPE)) +
      field("Volume", fmtVol(r.volume)) +
      selectField("Symbol", sym, [sym]) +
      field("Remained volume", fmtVol(r.volume_remain || 0)) +
      "</div>" +
      '<div class="trade-form-grid">' +
      selectField("Reason", REASON[r.reason] || "Stop out", Object.values(REASON)) +
      selectField("State", ORDER_STATE[r.state] || "FILLED", Object.values(ORDER_STATE)) +
      selectField("Expiration", ORDER_TIME[r.time_type] || "GTC", Object.values(ORDER_TIME)) +
      selectField("Filling", ORDER_FILL[r.type_fill] || "IMMEDIATE OR CANCEL", Object.values(ORDER_FILL)) +
      field("Dealer ID", r.dealer || 0) + field("Expert ID", r.expert_id || "") +
      field("Party ID", r.party_id || "") + field("External ID", r.external_id || "") +
      field("Setup time", fmtTs(r.time_setup, r.time_setup_ms)) +
      field("Done time", fmtTs(r.time_done || r.time_setup, r.time_done_ms)) +
      field("Expiration time", fmtTs(r.time_expiration || 0)) +
      field("Order price", fmtNum(r.price_order || r.price, r.digits || 5)) +
      field("Current price", fmtNum(r.price_current || r.price_order, r.digits || 5)) +
      field("Trigger price", fmtNum(r.price_trigger || 0, r.digits || 5)) +
      field("Stop loss", fmtNum(r.price_sl || 0, r.digits || 5)) +
      field("Take profit", fmtNum(r.price_tp || 0, r.digits || 5)) +
      field("Margin rate", fmtNum(r.rate_margin || 0, 8)) +
      field("Comment", r.comment || "", { wide: true }) +
      field("Disabled activations", r.activation_flags_label || "", { wide: true }) +
      field("Modifications", r.modifications || "", { wide: true }) +
      "</div>";
  }

  function renderPositionDetails(r) {
    var sym = r.symbol + (r.symbol_description ? ", " + r.symbol_description : "");
    return '<div class="trade-form-row">' +
      field("Position", r.position_id) +
      selectField("Type", POS_ACTION[r.action] || "Buy", ["Buy", "Sell"]) +
      field("Volume", fmtVol(r.volume)) +
      selectField("Symbol", sym, [sym]) +
      "</div>" +
      '<div class="trade-form-grid">' +
      selectField("Reason", REASON[r.reason] || "Client", Object.values(REASON)) +
      field("Open Time", fmtTs(r.time_create, r.time_create_ms)) +
      field("Update Time", fmtTs(r.time_update || r.time_create, r.time_update_ms)) +
      field("Price", fmtNum(r.price_open, r.digits || 5)) +
      field("Current price", fmtNum(r.price_current, r.digits || 5)) +
      field("Stop loss", fmtNum(r.price_sl || 0, r.digits || 5)) +
      field("Take profit", fmtNum(r.price_tp || 0, r.digits || 5)) +
      field("Gateway price", fmtNum(r.gateway_price || 0, r.digits || 5)) +
      field("Swap", fmtNum(r.storage || 0, 2)) +
      field("Profit", fmtNum(r.profit || 0, 2)) +
      field("Margin rate", fmtNum(r.rate_margin || 0, 8)) +
      field("Comment", r.comment || "", { wide: true }) +
      field("Modifications", r.modifications || "", { wide: true }) +
      "</div>";
  }

  function renderChainTable(chain, type) {
    var html = "<thead><tr><th></th><th>Ticket</th><th>Time</th><th>Type</th><th>Volume</th><th>Price</th><th>Reason</th><th>Profit</th></tr></thead><tbody>";
    chain.forEach(function (row) {
      var href = row.kind === "order" ? "orders-detail.html?id=" + row.ticket
        : row.kind === "position" ? "positions-detail.html?id=" + row.ticket
          : "deals-detail.html?id=" + row.ticket;
      html += '<tr class="' + (row.selected ? "selected" : "") + '" data-href="' + esc(href) + '">' +
        "<td>" + row.icon + "</td>" +
        "<td>" + row.ticket + "</td>" +
        "<td>" + esc(row.time) + "</td>" +
        "<td>" + esc(row.type) + "</td>" +
        "<td>" + esc(row.volume) + "</td>" +
        "<td>" + esc(row.price) + "</td>" +
        "<td>" + esc(row.reason) + "</td>" +
        "<td>" + esc(row.profit) + "</td></tr>";
    });
    html += "</tbody>";
    return html;
  }

  function boot(cfg) {
    var root = document.querySelector("[data-trade-module]");
    if (!root) return;
    var id = queryParam("id") || queryParam(cfg.idParam || "id");
    var titleEl = root.querySelector("[data-trade-title]");
    var accountEl = root.querySelector("[data-trade-account]");
    var chainEl = root.querySelector("[data-trade-chain]");
    var detailsEl = root.querySelector("[data-trade-details]");
    var tabs = root.querySelectorAll("[data-trade-tab]");
    var panels = root.querySelectorAll("[data-trade-panel]");

    tabs.forEach(function (tab) {
      tab.addEventListener("click", function () {
        var key = tab.getAttribute("data-trade-tab");
        tabs.forEach(function (t) { t.classList.remove("active"); });
        panels.forEach(function (p) { p.classList.remove("active"); });
        tab.classList.add("active");
        var panel = root.querySelector('[data-trade-panel="' + key + '"]');
        if (panel) panel.classList.add("active");
      });
    });

    async function load() {
      HSTModule.setApiBar(root, { loading: true });
      var res = await HSTModule.fetchOne(cfg.endpoint, id);
      var r = res.data || {};
      var user = HSTMock.userForLogin(r.login);

      if (titleEl) titleEl.textContent = titleFor(cfg.type, r);
      if (accountEl) {
        accountEl.innerHTML = user
          ? '<a class="trade-account-link" href="users-detail.html?login=' + user.login + '">' + esc(renderAccount(user)) + "</a>"
          : esc(renderAccount(null));
      }

      var chain = buildChain(cfg.type, r, r.chain || defaultChain(cfg.type, r));
      if (chainEl) {
        chainEl.innerHTML = renderChainTable(chain, cfg.type);
        chainEl.querySelectorAll("tbody tr").forEach(function (tr) {
          tr.addEventListener("dblclick", function () {
            var href = tr.getAttribute("data-href");
            if (href) location.href = href;
          });
        });
      }

      if (detailsEl) {
        if (cfg.type === "deal") detailsEl.innerHTML = renderDealDetails(r);
        else if (cfg.type === "order") detailsEl.innerHTML = renderOrderDetails(r);
        else detailsEl.innerHTML = renderPositionDetails(r);
      }

      HSTModule.setApiBar(root, {
        demo: res.demo,
        message: (cfg.title || cfg.type) + " #" + id + (res.demo ? " — demo data" : " — " + res.elapsed + " ms")
      });
    }

    var reloadBtn = root.querySelector("[data-reload]");
    if (reloadBtn) reloadBtn.addEventListener("click", load);
    load();
  }

  global.HSTTradeUI = {
    boot: boot,
    fmtTs: fmtTs,
    fmtVol: fmtVol,
    ORDER_TYPE: ORDER_TYPE,
    ORDER_STATE: ORDER_STATE,
    DEAL_ACTION: DEAL_ACTION,
    DEAL_ENTRY: DEAL_ENTRY,
    REASON: REASON,
    POS_ACTION: POS_ACTION,
    actionIcon: function (action, entry) {
      if (entry === 1 || entry === "OUT") {
        return global.HSTIcons ? global.HSTIcons.img("deal-out") : '<span class="trade-ico sell">↩</span>';
      }
      if (action === 0 || action === "buy") return '<span class="trade-ico buy">▲</span>';
      if (action === 1 || action === "sell") return '<span class="trade-ico sell">▼</span>';
      return global.HSTIcons ? global.HSTIcons.img("deal") : '<span class="trade-ico">•</span>';
    },
    labelOrderType: function (t) { return ORDER_TYPE[t] || t; },
    labelOrderState: function (s) { return ORDER_STATE[s] || s; },
    labelDealAction: function (a) { return DEAL_ACTION[a] || a; },
    labelDealEntry: function (e) { return DEAL_ENTRY[e] || e; },
    labelReason: function (r) { return REASON[r] || r; },
    labelPosAction: function (a) { return POS_ACTION[a] || a; }
  };
})(window);
