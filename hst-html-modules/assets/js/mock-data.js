/**
 * MT5-style demo datasets — used when not signed in or API unavailable.
 */
(function (global) {
  var NOW = Math.floor(Date.now() / 1000);
  var NOW_NS = Date.now() * 1000000;
  var sessionIdSeq = 100;

  function symbolPath(folder, name) {
    return folder + "\\" + name;
  }

  function assignSessionIds(symbolId, rows) {
    return (rows || []).map(function (r) {
      return {
        session_id: sessionIdSeq++,
        symbol_id: symbolId,
        type: r.type,
        day: r.day,
        open: r.open,
        close: r.close
      };
    });
  }

  function forexWeekSessions() {
    var s = [];
    for (var day = 1; day <= 4; day++) {
      s.push({ type: 0, day: day, open: 0, close: 1440 });
      s.push({ type: 1, day: day, open: 5, close: 1440 });
    }
    s.push({ type: 0, day: 5, open: 0, close: 1439 });
    s.push({ type: 1, day: 5, open: 5, close: 1439 });
    return s;
  }

  var groups = [
    { group_id: 1, group: "demo\\forex-usd", name: "forex-usd", exists: true, status: "Enabled", auth_mode: 0, currency: "USD", margin_call: 100, margin_stop_out: 50, limit_orders: 200, permission_flags: 0, auth_password_min: 8, currency_digits: 2, company: "Demo Broker Ltd", company_page: "https://demo.example.com", company_email: "info@demo.example.com", company_support_page: "https://support.demo.example.com", company_support_email: "support@demo.example.com", company_catalog: "Forex", reports_mode: 0, reports_flags: 0, reports_email: "reports@demo.example.com", reports_smtp: "smtp.demo.example.com", reports_smtp_login: "reports", news_mode: 0, news_category: "Forex", news_langs: [9, 12], mail_mode: 0, trade_flags: 0, trade_interest_rate: 0, trade_virtual_credit: 0, trade_transfer_mode: 0, demo_leverage: 100, demo_deposit: 10000, margin_free_mode: 0, margin_so_mode: 0, margin_mode: 0, margin_flags: 0, limit_symbols: 0, limit_positions: 1000, limit_positions_volume: 0, limit_history: 0 },
    { group_id: 2, group: "demo\\forex-eur", name: "forex-eur", exists: true, status: "Enabled", auth_mode: 0, currency: "EUR", margin_call: 100, margin_stop_out: 50, limit_orders: 200 },
    { group_id: 3, group: "real\\standard", name: "standard", exists: true, status: "Enabled", auth_mode: 0, currency: "USD", margin_call: 80, margin_stop_out: 30, limit_orders: 500 },
    { group_id: 4, group: "real\\vip", name: "vip", exists: true, status: "Enabled", auth_mode: 0, currency: "USD", margin_call: 50, margin_stop_out: 20, limit_orders: 1000 },
    { group_id: 5, group: "managers\\administrators", name: "administrators", exists: true, status: "Enabled", auth_mode: 0, currency: "USD", margin_call: 0, margin_stop_out: 0, limit_orders: 0 },
    { group_id: 6, group: "managers\\dealers", name: "dealers", exists: true, status: "Enabled", auth_mode: 0, currency: "USD", margin_call: 0, margin_stop_out: 0, limit_orders: 0 },
    { group_id: 7, group: "preliminary", name: "preliminary", exists: true, status: "Enabled", auth_mode: 0, currency: "USD", margin_call: 100, margin_stop_out: 50, limit_orders: 50 },
    { group_id: 8, group: "demo\\crypto", name: "crypto", exists: true, status: "Enabled", auth_mode: 0, currency: "USD", margin_call: 100, margin_stop_out: 50, limit_orders: 100 }
  ];

  var symbols = [
    { symbol_id: 1, symbol: "EURUSD", path: symbolPath("Forex\\Majors", "EURUSD"), description: "Euro vs US Dollar", digits: 5, exec_mode: 3, spread: 0, contract_size: 100000, trade_mode: 4, calc_mode: 0, date_modified: NOW_NS - 86400000000000 },
    { symbol_id: 2, symbol: "GBPUSD", path: symbolPath("Forex\\Majors", "GBPUSD"), description: "British Pound vs US Dollar", digits: 5, exec_mode: 3, spread: 0, contract_size: 100000, trade_mode: 4, calc_mode: 0, date_modified: NOW_NS - 86400000000000 },
    { symbol_id: 3, symbol: "USDJPY", path: symbolPath("Forex\\Majors", "USDJPY"), description: "US Dollar vs Yen", digits: 3, exec_mode: 3, spread: 0, contract_size: 100000, trade_mode: 1, calc_mode: 0, date_modified: NOW_NS - 172800000000000 },
    { symbol_id: 4, symbol: "XAUUSD", path: symbolPath("Metals\\Gold", "XAUUSD"), description: "Gold vs US Dollar", digits: 2, exec_mode: 1, spread: 30, contract_size: 100, trade_mode: 4, calc_mode: 1, date_modified: NOW_NS - 3600000000000 },
    { symbol_id: 5, symbol: "XAGUSD", path: symbolPath("Metals\\Silver", "XAGUSD"), description: "Silver vs US Dollar", digits: 3, exec_mode: 1, spread: 25, contract_size: 5000, trade_mode: 3, calc_mode: 1, date_modified: NOW_NS - 7200000000000 },
    { symbol_id: 6, symbol: "US500", path: symbolPath("Indices\\US", "US500"), description: "S&P 500 Index", digits: 2, exec_mode: 0, spread: 40, contract_size: 1, trade_mode: 4, calc_mode: 2, date_modified: NOW_NS - 43200000000000 },
    { symbol_id: 7, symbol: "BTCUSD", path: symbolPath("Crypto\\Major", "BTCUSD"), description: "Bitcoin vs US Dollar", digits: 2, exec_mode: 2, spread: 80, contract_size: 1, trade_mode: 4, calc_mode: 0, date_modified: NOW_NS - 900000000000 },
    { symbol_id: 8, symbol: "ETHUSD", path: symbolPath("Crypto\\Major", "ETHUSD"), description: "Ethereum vs US Dollar", digits: 2, exec_mode: 2, spread: 60, contract_size: 1, trade_mode: 0, calc_mode: 0, date_modified: NOW_NS - 900000000000 }
  ];

  function cryptoWeekSessions() {
    var s = [];
    for (var day = 0; day <= 6; day++) {
      s.push({ type: 0, day: day, open: 0, close: 1440 });
      s.push({ type: 1, day: day, open: 0, close: 1440 });
    }
    return s;
  }

  var symbolSessions = {
    1: assignSessionIds(1, forexWeekSessions()),
    2: assignSessionIds(2, forexWeekSessions()),
    3: assignSessionIds(3, forexWeekSessions()),
    4: assignSessionIds(4, (function () {
      var s = forexWeekSessions();
      s.push({ type: 0, day: 6, open: 0, close: 1440 }, { type: 1, day: 6, open: 0, close: 1200 });
      return s;
    })()),
    5: assignSessionIds(5, (function () {
      var s = [];
      for (var day = 1; day <= 5; day++) {
        s.push({ type: 0, day: day, open: 60, close: 1380 });
        s.push({ type: 1, day: day, open: 120, close: 1320 });
      }
      return s;
    })()),
    6: assignSessionIds(6, (function () {
      var s = [];
      for (var day = 1; day <= 5; day++) {
        s.push({ type: 0, day: day, open: 540, close: 1260 });
        s.push({ type: 1, day: day, open: 570, close: 1230 });
      }
      return s;
    })()),
    7: assignSessionIds(7, cryptoWeekSessions()),
    8: assignSessionIds(8, cryptoWeekSessions())
  };

  var symbolDetails = {
    1: {
      exchange: "XCCY", isin: "", international: "", category: "", cfi: "",
      sector: 12, industry: 0, country: "", basis: "", source: "", page: "",
      currency_base: "EUR", currency_base_digits: 2,
      currency_profit: "USD", currency_profit_digits: 2,
      currency_margin: "EUR", currency_margin_digits: 2,
      spread_balance: 0, tick_book_depth: 0, tick_flags: 1,
      filter_soft: 0, filter_soft_ticks: 5, filter_hard: 0, filter_hard_ticks: 5,
      filter_discard: 0, filter_spread_max: 0, filter_spread_min: 0, subscriptions_delay: 15,
      gtc_mode: 0, fill_flags: 3, expir_flags: 15, order_flags: 127,
      stops_level: 1, freeze_level: 0, quotes_timeout: 0,
      volume_min: 100, volume_step: 100, volume_max: 1000000, volume_limit: 0,
      margin_initial: 0, margin_hedged: 0, margin_maintenance: 0, margin_flags: 0,
      margin_rate_liquidity: 1, margin_rate_currency: 0,
      margin_initial_buy: 1, margin_initial_sell: 1,
      margin_maintenance_buy: 1, margin_maintenance_sell: 1,
      swap_mode: 1, swap_long: -4.231, swap_short: -5.206, swap_year_day: 360, swap_flags: 0,
      swap_rate_sunday: 0, swap_rate_monday: 1, swap_rate_tuesday: 1, swap_rate_wednesday: 3,
      swap_rate_thursday: 1, swap_rate_friday: 1, swap_rate_saturday: 0,
      time_start: 0, time_expiration: 0
    },
    2: {
      exchange: "XCCY", sector: 12, industry: 0,
      currency_base: "GBP", currency_base_digits: 2,
      currency_profit: "USD", currency_profit_digits: 2,
      currency_margin: "GBP", currency_margin_digits: 2,
      spread_balance: 0, tick_book_depth: 0, tick_flags: 1,
      filter_soft: 0, filter_soft_ticks: 5, filter_hard: 0, filter_hard_ticks: 5,
      filter_discard: 0, filter_spread_max: 0, filter_spread_min: 0, subscriptions_delay: 15,
      gtc_mode: 0, fill_flags: 3, expir_flags: 15, order_flags: 127,
      stops_level: 1, freeze_level: 0, quotes_timeout: 0,
      volume_min: 100, volume_step: 100, volume_max: 1000000, volume_limit: 0,
      margin_initial: 0, margin_hedged: 0, margin_maintenance: 0, margin_flags: 0,
      margin_rate_liquidity: 1, margin_rate_currency: 0,
      swap_mode: 1, swap_long: -4.5, swap_short: -5.1, swap_year_day: 360,
      swap_rate_monday: 1, swap_rate_tuesday: 1, swap_rate_wednesday: 3,
      swap_rate_thursday: 1, swap_rate_friday: 1,
      time_start: 0, time_expiration: 0
    },
    4: {
      exchange: "", sector: 0,
      currency_base: "XAU", currency_profit: "USD", currency_margin: "XAU",
      currency_base_digits: 2, currency_profit_digits: 2, currency_margin_digits: 2,
      spread_balance: 0, tick_flags: 1,
      filter_soft: 0, filter_soft_ticks: 5, filter_hard: 0, filter_hard_ticks: 5,
      gtc_mode: 0, fill_flags: 1, expir_flags: 15, order_flags: 127,
      stops_level: 5, freeze_level: 0, quotes_timeout: 0,
      volume_min: 100, volume_step: 100, volume_max: 100000, volume_limit: 0,
      margin_initial: 0, margin_hedged: 0, margin_maintenance: 0,
      swap_mode: 1, swap_long: -2.1, swap_short: -1.8, swap_year_day: 360,
      swap_rate_monday: 1, swap_rate_wednesday: 3,
      time_start: NOW - 86400 * 30, time_expiration: NOW + 86400 * 365
    }
  };

  function symbolDetail(id) {
    var base = findById(symbols, "symbol_id", id) || symbols[0];
    if (!base) return null;
    var detail = Object.assign({}, symbolDetails[base.symbol_id] || {}, base);
    detail.sessions = (symbolSessions[base.symbol_id] || []).slice();
    if (detail.time_start == null) detail.time_start = 0;
    if (detail.time_expiration == null) detail.time_expiration = 0;
    return detail;
  }

  function getSymbol(id) {
    return symbolDetail(id);
  }

  function getSymbolSessions(symbolId) {
    return (symbolSessions[symbolId] || []).slice();
  }

  function updateSymbol(id, patch) {
    if (!patch) return symbolDetail(id);
    var num = Number(id);
    var base = findById(symbols, "symbol_id", id);
    if (!base) return null;

    var listKeys = ["symbol", "path", "description", "digits", "trade_mode", "calc_mode", "exec_mode", "spread", "contract_size"];
    listKeys.forEach(function (k) {
      if (patch[k] !== undefined) base[k] = patch[k];
    });
    if (patch.date_modified !== undefined) base.date_modified = patch.date_modified;

    if (!symbolDetails[num]) symbolDetails[num] = {};
    Object.keys(patch).forEach(function (k) {
      if (k === "sessions") return;
      symbolDetails[num][k] = patch[k];
    });

    if (patch.sessions) {
      symbolSessions[num] = patch.sessions.map(function (s) {
        return {
          session_id: s.session_id || sessionIdSeq++,
          symbol_id: num,
          type: s.type,
          day: s.day,
          open: s.open,
          close: s.close
        };
      });
    }
    if (patch.time_start !== undefined) {
      if (!symbolDetails[num]) symbolDetails[num] = {};
      symbolDetails[num].time_start = patch.time_start;
    }
    if (patch.time_expiration !== undefined) {
      if (!symbolDetails[num]) symbolDetails[num] = {};
      symbolDetails[num].time_expiration = patch.time_expiration;
    }

    base.date_modified = Date.now() * 1000000;
    return symbolDetail(id);
  }

  var routing = [
    { routing_id: 1, name: "Auto execution — majors", mode: 0, action: 0, routing_index: 0, date_modified: NOW - 86400, request: 0, type: 0, flags: 0, action_value: 0, conditions: [{ condition: 0, rule: 0, value: "Forex\\Majors\\*" }] },
    { routing_id: 2, name: "Dealer — gold", mode: 1, action: 1, routing_index: 1, date_modified: NOW - 43200, request: 0, type: 0, flags: 0, action_value: 0, conditions: [{ condition: 1, rule: 0, value: "XAUUSD" }] },
    { routing_id: 3, name: "Reject — crypto weekend", mode: 0, action: 2, routing_index: 2, date_modified: NOW - 21600, request: 0, type: 0, flags: 0, action_value: 0, conditions: [{ condition: 1, rule: 0, value: "Crypto\\*" }] }
  ];

  var routingDealers = {
    1: [{ login: 1001, name: "Dealer A", dealer_index: 0 }, { login: 1002, name: "Dealer B", dealer_index: 1 }],
    2: [{ login: 1001, name: "Dealer A", dealer_index: 0 }],
    3: []
  };

  var datafeedModuleCatalog = {
    1: [
      { module: "fix44", label: "FIX 4.4", description: "QuickFIX initiator, incremental refresh by default" },
      { module: "fix43", label: "FIX 4.3", description: "QuickFIX initiator, full refresh by default" }
    ],
    2: [
      { module: "RSSNewsFeeder", label: "RSS News", description: "Poll an HTTP RSS or Atom feed" }
    ]
  };

  var datafeeds = [
    {
      datafeed_id: 1, name: "FIX 4.4 Bridge", module: "fix44", enable: 1, mode: 1,
      feed_server: "fix.broker.example:9876", gateway_server: "127.0.0.1:16387",
      feed_login: 1001, gateway_login: 1, timeout: 60, timeout_reconnect: 5, timeout_sleep: 600, attempts_sleep: 3,
      ticks_count: 1845230, tick_stats_count: 1845230, books_count: 42000, news_count: 0,
      sys_connection: 1, sys_last_time: NOW - 60, updated_at: NOW - 60, symbol_count: 3
    },
    {
      datafeed_id: 2, name: "FIX 4.3 Bridge", module: "fix43", enable: 1, mode: 1,
      feed_server: "fix43.broker.example:9876", gateway_server: "",
      feed_login: 1002, gateway_login: 0, timeout: 60, timeout_reconnect: 5, timeout_sleep: 600, attempts_sleep: 3,
      ticks_count: 512000, tick_stats_count: 512000, books_count: 0, news_count: 0,
      sys_connection: 1, sys_last_time: NOW - 120, updated_at: NOW - 120, symbol_count: 2
    },
    {
      datafeed_id: 3, name: "Market News RSS", module: "RSSNewsFeeder", enable: 1, mode: 2,
      feed_server: "https://news.example.com/rss.xml", gateway_server: "",
      feed_login: 0, gateway_login: 0, timeout: 60, timeout_reconnect: 5, timeout_sleep: 600, attempts_sleep: 3,
      ticks_count: 0, tick_stats_count: 0, books_count: 0, news_count: 842,
      sys_connection: 1, sys_last_time: NOW - 45, updated_at: NOW - 45, symbol_count: 0
    }
  ];

  var datafeedDetails = {
    1: {
      feed_password: "", gateway_password: "", company: "Broker LP", issuer: "FIX",
      allow_import_symbols: 0,
      params: [
        { param_id: 1, param_key: "Quotes Delay", type: 0, value: "0", priority: 0 },
        { param_id: 2, param_key: "SenderCompID", type: 0, value: "HSTBRIDGE", priority: 1 },
        { param_id: 3, param_key: "TargetCompID", type: 0, value: "LPFIX", priority: 2 }
      ],
      feed_symbols: [
        { feed_symbol_id: 1, symbol: "Forex\\*", exclude: 0 },
        { feed_symbol_id: 2, symbol: "Metals\\*", exclude: 0 }
      ],
      translates: [
        { translate_id: 1, symbol: "EURUSD", source: "EUR/USD", bid_markup: 0, ask_markup: 0, digits: 5 }
      ]
    },
    2: {
      feed_password: "", gateway_password: "", company: "Broker LP", issuer: "FIX",
      allow_import_symbols: 1,
      params: [
        { param_id: 4, param_key: "Quotes Delay", type: 0, value: "0", priority: 0 },
        { param_id: 5, param_key: "SenderCompID", type: 0, value: "HSTBRIDGE43", priority: 1 },
        { param_id: 6, param_key: "TargetCompID", type: 0, value: "LPFIX43", priority: 2 }
      ],
      feed_symbols: [
        { feed_symbol_id: 3, symbol: "Forex\\Majors\\*", exclude: 0 },
        { feed_symbol_id: 4, symbol: "Crypto\\*", exclude: 0 }
      ],
      translates: []
    },
    3: {
      feed_password: "", gateway_password: "", company: "Example News", issuer: "RSS",
      allow_import_symbols: 0,
      params: [
        { param_id: 5, param_key: "News Category", type: 0, value: "Market", priority: 0 },
        { param_id: 6, param_key: "News Request Period", type: 1, value: "300", priority: 1 },
        { param_id: 7, param_key: "Language", type: 0, value: "en", priority: 2 }
      ],
      feed_symbols: [],
      translates: []
    }
  };

  function listDatafeedModules(mode) {
    var m = Number(mode) || 1;
    return (datafeedModuleCatalog[m] || datafeedModuleCatalog[1]).slice();
  }

  function datafeedDetail(id) {
    var num = Number(id);
    var base = datafeeds.find(function (d) { return d.datafeed_id === num; });
    if (!base) return null;
    var extra = datafeedDetails[num] || {
      feed_password: "", gateway_password: "", company: "", issuer: "",
      allow_import_symbols: 0, params: [], feed_symbols: [], translates: []
    };
    return Object.assign({}, base, extra, {
      params: (extra.params || []).slice(),
      feed_symbols: (extra.feed_symbols || []).slice(),
      translates: (extra.translates || []).slice()
    });
  }

  function getDatafeed(id) {
    return datafeedDetail(id);
  }

  function createDatafeed(data) {
    var nextId = datafeeds.reduce(function (m, d) {
      return Math.max(m, Number(d.datafeed_id) || 0);
    }, 0) + 1;
    var row = {
      datafeed_id: nextId,
      name: String(data.name || "New Feed").trim(),
      module: data.module || "fix44",
      enable: data.enable != null ? data.enable : 1,
      mode: data.mode != null ? data.mode : 1,
      feed_server: data.feed_server || "",
      gateway_server: data.gateway_server || "",
      feed_login: data.feed_login || 0,
      gateway_login: data.gateway_login || 0,
      timeout: data.timeout != null ? data.timeout : 60,
      timeout_reconnect: data.timeout_reconnect != null ? data.timeout_reconnect : 5,
      timeout_sleep: data.timeout_sleep != null ? data.timeout_sleep : 600,
      attempts_sleep: data.attempts_sleep != null ? data.attempts_sleep : 3,
      ticks_count: 0, tick_stats_count: 0, books_count: 0, news_count: 0,
      sys_connection: 0, sys_last_time: 0, updated_at: NOW, symbol_count: 0
    };
    datafeeds.push(row);
    datafeedDetails[nextId] = {
      feed_password: data.feed_password || "",
      gateway_password: data.gateway_password || "",
      company: data.company || "",
      issuer: data.issuer || "",
      allow_import_symbols: data.allow_import_symbols || 0,
      params: (data.params || []).slice(),
      feed_symbols: (data.feed_symbols || []).slice(),
      translates: (data.translates || []).slice()
    };
    return row;
  }

  function updateDatafeed(id, patch) {
    var num = Number(id);
    var idx = datafeeds.findIndex(function (d) { return d.datafeed_id === num; });
    if (idx < 0) return null;
    Object.assign(datafeeds[idx], patch, { updated_at: NOW });
    if (patch.params || patch.feed_symbols || patch.translates || patch.feed_password != null) {
      if (!datafeedDetails[num]) {
        datafeedDetails[num] = { params: [], feed_symbols: [], translates: [] };
      }
      Object.assign(datafeedDetails[num], patch);
    }
    if (patch.feed_symbols) {
      datafeeds[idx].symbol_count = patch.feed_symbols.length;
    }
    return datafeedDetail(num);
  }

  function deleteDatafeeds(ids) {
    var set = {};
    (ids || []).forEach(function (id) { set[Number(id)] = true; });
    var before = datafeeds.length;
    datafeeds = datafeeds.filter(function (d) { return !set[d.datafeed_id]; });
    return before - datafeeds.length;
  }

  function moveDatafeedInList(id, direction) {
    var num = Number(id);
    var idx = datafeeds.findIndex(function (d) { return d.datafeed_id === num; });
    if (idx < 0) return false;
    var swap = direction === "up" ? idx - 1 : idx + 1;
    if (swap < 0 || swap >= datafeeds.length) return false;
    var tmp = datafeeds[idx];
    datafeeds[idx] = datafeeds[swap];
    datafeeds[swap] = tmp;
    return true;
  }

  function setDatafeedEnable(id, enable) {
    return updateDatafeed(id, { enable: enable ? 1 : 0 });
  }

  var leverages = [
    { leverage_id: 1, name: "Standard FX", flags: 0, timestamp: NOW - 864000 },
    { leverage_id: 2, name: "VIP Metals", flags: 0, timestamp: NOW - 432000 },
    { leverage_id: 3, name: "Crypto Tiered", flags: 1, timestamp: NOW - 86400 }
  ];

  var holidays = [
    { holiday_id: 1, year: 2026, month: 1, day: 1, from: 0, to: 1439, mode: 1, description: "New Year", symbols: ["*"] },
    { holiday_id: 2, year: 0, month: 12, day: 25, from: 0, to: 1439, mode: 1, description: "Christmas", symbols: ["Forex\\*", "Metals\\*"] },
    { holiday_id: 3, year: 2026, month: 7, day: 4, from: 720, to: 900, mode: 0, description: "Independence Day (partial)", symbols: ["Indices\\US\\*"] }
  ];

  var clients = [
    { client_id: 10001, person_name: "John", person_last_name: "Smith", contact_email: "john.smith@mail.com", contact_phone: "+1 555 0101", address_country: "US", address_city: "New York", client_status: 1, kyc_status: 2, assigned_manager: 1000, comment: "VIP prospect", date_created: NOW - 8640000, date_modified: NOW - 86400 },
    { client_id: 10002, person_name: "Maria", person_last_name: "Garcia", contact_email: "maria.g@mail.com", contact_phone: "+34 600 123456", address_country: "ES", address_city: "Madrid", client_status: 1, kyc_status: 1, assigned_manager: 1001, comment: "", date_created: NOW - 4320000, date_modified: NOW - 172800 },
    { client_id: 10003, person_name: "Ahmed", person_last_name: "Hassan", contact_email: "ahmed.h@mail.com", contact_phone: "+971 50 1234567", address_country: "AE", address_city: "Dubai", client_status: 0, kyc_status: 0, assigned_manager: 1000, comment: "Pending KYC", date_created: NOW - 864000, date_modified: NOW - 43200 },
    { client_id: 10004, person_name: "Yuki", person_last_name: "Tanaka", contact_email: "yuki.t@mail.com", contact_phone: "+81 90 1234 5678", address_country: "JP", address_city: "Tokyo", client_status: 1, kyc_status: 2, assigned_manager: 1000, comment: "Institutional", date_created: NOW - 17280000, date_modified: NOW - 3600 }
  ];

  var managers = [
    {
      login: 1000, name: "Chief Dealer", mailbox: "chief@broker.com", server: 0,
      request_limit_logs: 3, request_limit_reports: 4,
      groups: ["real\\*", "demo\\*"], access: ["10.0.0.0/8", "192.168.0.0/16"],
      rights: {
        right_admin: true, right_manager: true, right_cfg_groups: true, right_cfg_symbols: true,
        right_cfg_managers: true, right_cfg_requests: true, right_cfg_datafeeds: true,
        right_acc_read: true, right_trades_read: true, right_trades_dealer: true,
        right_accountant: true, right_srv_journals: true, right_srv_reports: true,
        right_reports: true, right_clients_access: true, right_clients_edit: true,
        right_email: true, right_news: true, right_export: true
      }
    },
    {
      login: 1001, name: "Dealer A", mailbox: "dealer.a@broker.com", server: 0,
      request_limit_logs: 2, request_limit_reports: 2,
      groups: ["real\\standard", "demo\\forex-usd"], access: [],
      rights: {
        right_manager: true, right_acc_read: true, right_trades_read: true,
        right_trades_dealer: true, right_quotes: true, right_market: true,
        right_clients_access: true
      }
    },
    {
      login: 1002, name: "Back Office", mailbox: "backoffice@broker.com", server: 0,
      request_limit_logs: 4, request_limit_reports: 3,
      groups: ["real\\*"], access: ["203.0.113.0/24"],
      rights: {
        right_manager: true, right_acc_read: true, right_accountant: true,
        right_srv_journals: true, right_clients_access: true, right_clients_edit: true,
        right_export: true
      }
    }
  ];

  var managerReportLimits = {
    1000: [
      { report: "Accounts \\ Accounts Groups", limit: "year_1" },
      { report: "Trade \\ Trade Transactions", limit: "months_6" },
      { report: "Trade \\ Trade Modification", limit: "months_3" },
      { report: "Clients \\ Clients", limit: "all" }
    ],
    1001: [
      { report: "Accounts \\ Accounts Groups", limit: "months_3" },
      { report: "Trade \\ Trade Transactions", limit: "months_1" }
    ],
    1002: [
      { report: "Accounts \\ Accounts Groups", limit: "months_6" },
      { report: "Clients \\ Clients", limit: "year_1" }
    ]
  };

  var MANAGER_LIMITS = { 0: "All", 1: "1 month", 2: "3 months", 3: "6 months", 4: "1 year", 5: "2 years", 6: "3 years" };

  var RIGHTS_LABELS = {
    right_admin: "Connect using Administrator",
    right_manager: "Connect using Manager",
    right_cfg_groups: "Configure groups",
    right_cfg_symbols: "Configure symbols",
    right_cfg_managers: "Configure managers",
    right_cfg_requests: "Configure request routing",
    right_cfg_datafeeds: "Configure datafeeds",
    right_cfg_holidays: "Configure holidays",
    right_cfg_time: "Configure server time",
    right_srv_journals: "Access server logs",
    right_srv_reports: "Receive automatic server reports",
    right_acc_read: "Access accounts",
    right_accountant: "Accountant (deposit/withdraw)",
    right_trades_read: "View orders and positions",
    right_trades_dealer: "Dealer",
    right_trades_manager: "Manage trades",
    right_clients_access: "Access clients",
    right_clients_edit: "Edit clients",
    right_clients_create: "Create clients",
    right_reports: "Request reports",
    right_email: "Send emails",
    right_news: "Send news",
    right_export: "Export data",
    right_quotes: "View quotes",
    right_market: "Market Watch"
  };

  function managerForLogin(login) {
    return findById(managers, "login", login);
  }

  function managerDetailRecord(login) {
    var m = managerForLogin(login) || managers[0];
    if (!m) return null;
    var out = {
      login: m.login, name: m.name, mailbox: m.mailbox, server: m.server,
      request_limit_logs: m.request_limit_logs, request_limit_reports: m.request_limit_reports,
      groups: m.groups, access: m.access
    };
    Object.keys(m.rights || {}).forEach(function (k) {
      out[k] = m.rights[k] ? 1 : 0;
    });
    return out;
  }

  function managerRightsView(login) {
    var m = managerForLogin(login) || managers[0];
    return m ? { login: m.login, name: m.name, groups: m.groups, rights: m.rights } : null;
  }

  function managerReports(login) {
    return managerReportLimits[login] || managerReportLimits[1000] || [];
  }

  var users = [
    { login: 20001, client_id: 10001, group: "real\\standard", name: "John Smith", email: "john.smith@mail.com", phone: "+1 555 0101", country: "US", city: "New York", leverage: 100, balance: 10250.75, credit: 0, is_manager: false, last_access: NOW - 3600, updated_at: NOW - 3600 },
    { login: 20002, client_id: 10002, group: "demo\\forex-eur", name: "Maria Garcia", email: "maria.g@mail.com", phone: "+34 600 123456", country: "ES", city: "Madrid", leverage: 200, balance: 50000.00, credit: 500, is_manager: false, last_access: NOW - 7200, updated_at: NOW - 7200 },
    { login: 20003, client_id: 10003, group: "preliminary", name: "Ahmed Hassan", email: "ahmed.h@mail.com", phone: "+971 50 1234567", country: "AE", city: "Dubai", leverage: 50, balance: 1000.00, credit: 0, is_manager: false, last_access: NOW - 86400, updated_at: NOW - 86400 },
    { login: 20004, client_id: 10004, group: "real\\vip", name: "Yuki Tanaka", email: "yuki.t@mail.com", phone: "+81 90 1234567", country: "JP", city: "Tokyo", leverage: 500, balance: 250000.00, credit: 0, is_manager: false, last_access: NOW - 1800, updated_at: NOW - 1800 },
    { login: 20005, client_id: 10001, group: "demo\\forex-usd", name: "John Smith (Demo)", email: "john.smith@mail.com", phone: "+1 555 0101", country: "US", city: "New York", leverage: 100, balance: 100000.00, credit: 0, is_manager: false, last_access: NOW - 7200, updated_at: NOW - 7200 }
  ];

  var orders = [
    { order_id: 900001, login: 20001, symbol: "EURUSD", type: 0, state: 4, volume: 1.0, volume_initial: 1.0, price_order: 1.08542, price_sl: 1.08000, price_tp: 1.09000, time_setup: NOW - 86400, time_done: NOW - 86400, position_id: 800001, reason: 0, type_fill: 1, time_type: 0, digits: 5, rate_margin: 1.08542 },
    { order_id: 900002, login: 20002, symbol: "GBPUSD", type: 1, state: 4, volume: 0.5, volume_initial: 0.5, price_order: 1.26410, price_sl: 0, price_tp: 0, time_setup: NOW - 43200, time_done: NOW - 43200, position_id: 800002, reason: 0, type_fill: 1, time_type: 0, digits: 5 },
    { order_id: 900003, login: 20004, symbol: "XAUUSD", type: 2, state: 1, volume: 0.1, volume_initial: 0.1, price_order: 2345.50, price_sl: 2330, price_tp: 2360, time_setup: NOW - 600, reason: 0, type_fill: 1, time_type: 0, digits: 2 },
    { order_id: 900004, login: 20001, symbol: "USDJPY", type: 3, state: 2, volume: 2.0, volume_initial: 2.0, price_order: 149.850, price_sl: 0, price_tp: 0, time_setup: NOW - 7200, time_done: NOW - 7000, reason: 0, type_fill: 1, time_type: 0, digits: 3 },
    {
      order_id: 900010, login: 20005, symbol: "AUDJPY", symbol_description: "Australian Dollar vs Japanese Yen",
      type: 1, state: 4, volume: 0.04, volume_initial: 0.04, volume_remain: 0, price_order: 111.704, price_current: 111.704,
      price_sl: 0, price_tp: 0, time_setup: NOW - 120, time_setup_ms: 166, time_done: NOW - 120, time_done_ms: 166,
      position_id: 800010, reason: 5, type_fill: 1, time_type: 0, digits: 3, rate_margin: 0.70432,
      comment: "[so -29.73%/-41.83/140.71]", dealer: 0,
      chain: [
        { kind: "order", ticket: 900010, time: NOW - 120, time_ms: 166, type_label: "sell", volume: 0.04, volume_initial: 0.04, price: 111.704, reason: 5, digits: 3 },
        { kind: "deal", ticket: 700010, entry: 0, time: NOW - 119, time_ms: 789, action_label: "sell", volume: 0.04, price: 111.704, reason: 5, profit: -0.81, digits: 3 }
      ]
    }
  ];

  var positions = [
    { position_id: 800001, login: 20001, symbol: "EURUSD", action: 0, volume: 1.0, price_open: 1.08420, price_current: 1.08542, price_sl: 1.08000, price_tp: 1.09000, profit: 122.00, storage: -1.20, time_create: NOW - 86400, reason: 0, digits: 5, rate_margin: 1.08542 },
    { position_id: 800002, login: 20002, symbol: "GBPUSD", action: 1, volume: 0.5, price_open: 1.26600, price_current: 1.26410, profit: 95.00, storage: 0, time_create: NOW - 43200, reason: 0, digits: 5 },
    { position_id: 800003, login: 20004, symbol: "XAUUSD", action: 0, volume: 0.5, price_open: 2330.00, price_current: 2345.50, profit: 775.00, storage: -2.50, time_create: NOW - 172800, reason: 0, digits: 2 },
    {
      position_id: 800010, login: 20005, symbol: "AUDJPY", symbol_description: "Australian Dollar vs Japanese Yen",
      action: 1, volume: 0.04, price_open: 111.704, price_current: 111.720, price_sl: 0, price_tp: 0,
      profit: -0.81, storage: 0, time_create: NOW - 3600, time_update: NOW - 119, reason: 5, digits: 3,
      comment: "[so -29.73%/-41.83/140.71]",
      chain: [
        { kind: "order", ticket: 900010, time: NOW - 120, time_ms: 166, type_label: "sell", volume: 0.04, volume_initial: 0.04, price: 111.704, reason: 5, digits: 3 },
        { kind: "deal", ticket: 700010, entry: 0, time: NOW - 119, time_ms: 789, action_label: "sell", volume: 0.04, price: 111.704, reason: 5, profit: -0.81, digits: 3 }
      ]
    }
  ];

  var deals = [
    { deal_id: 700001, login: 20001, order_id: 900001, position_id: 800001, symbol: "EURUSD", action: 0, entry: 0, volume: 1.0, price: 1.08420, profit: 0, storage: 0, commission: 0, time: NOW - 86400, reason: 0, digits: 5 },
    { deal_id: 700002, login: 20002, order_id: 900002, position_id: 800002, symbol: "GBPUSD", action: 1, entry: 0, volume: 0.5, price: 1.26600, profit: 0, time: NOW - 43200, reason: 0, digits: 5 },
    { deal_id: 700003, login: 20004, order_id: 900003, position_id: 800003, symbol: "XAUUSD", action: 0, entry: 0, volume: 0.5, price: 2330.00, profit: 0, time: NOW - 172800, reason: 0, digits: 2 },
    { deal_id: 700004, login: 20001, order_id: 900001, position_id: 800001, symbol: "EURUSD", action: 0, entry: 0, volume: 0.5, price: 1.08500, profit: 40.00, time: NOW - 3600, reason: 0, digits: 5 },
    {
      deal_id: 700010, login: 20005, order_id: 900010, position_id: 800010, symbol: "AUDJPY", symbol_description: "Australian Dollar vs Japanese Yen",
      action: 1, entry: 0, volume: 0.04, volume_closed: 0.04, volume_closed_remain: 0,
      price: 111.704, price_sl: 0, price_tp: 0, profit: -0.81, raw_profit: -0.81, storage: 0, commission: 0, fee: 0,
      time: NOW - 119, time_ms: 789, reason: 5, digits: 3, rate_margin: 0, rate_profit: 0,
      market_bid: 111.704, market_ask: 111.736, market_last: 0, gateway_action: "BUY",
      comment: "[so -29.73%/-41.83/140.71]", dealer: 0,
      chain: [
        { kind: "order", ticket: 900010, time: NOW - 120, time_ms: 166, type_label: "sell", volume: 0.04, volume_initial: 0.04, price: 111.704, reason: 5, digits: 3 },
        { kind: "deal", ticket: 700010, entry: 0, time: NOW - 119, time_ms: 789, action_label: "sell", volume: 0.04, price: 111.704, reason: 5, profit: -0.81, digits: 3 }
      ]
    }
  ];

  var journal = [
    { journal_id: 1, created_at: NOW - 120, channel: "Trade", ip: "127.0.0.1", message: "Configuration synchronized" },
    { journal_id: 2, created_at: NOW - 300, channel: "Trade", ip: "127.0.0.1", message: "Manager login 1000 connected" },
    { journal_id: 3, created_at: NOW - 600, channel: "Trade", ip: "10.0.0.5", message: "20001: order #900001 placed buy 1.00 EURUSD at 1.08542" },
    { journal_id: 4, created_at: NOW - 900, channel: "Trade", ip: "10.0.0.5", message: "20004: position #800003 opened buy 0.50 XAUUSD at 2330.00" },
    { journal_id: 5, created_at: NOW - 1800, channel: "Gateway", ip: "127.0.0.1", message: "Data feed MetaQuotes Demo connected" }
  ];

  var endOfDay = { at: "23:59", updated_at: NOW - 604800 };

  function allHoursActive() {
    var h = [];
    for (var i = 0; i < 24; i++) h.push(true);
    return h;
  }

  var timeSettings = {
    time_zone: "UTC+05:30",
    daylight_saving: true,
    sync_servers: "time.cloudflare.com, pool.ntp.org",
    schedule: (function () {
      var day = allHoursActive();
      return [day.slice(), day.slice(), day.slice(), day.slice(), day.slice(), day.slice(), day.slice()];
    })()
  };

  function formatDayRange(hours) {
    if (!hours || !hours.length) return "";
    if (hours.every(Boolean) && hours.length === 24) return "00:00-24:00";
    if (!hours.some(Boolean)) return "";
    var ranges = [];
    var start = null;
    for (var i = 0; i < 24; i++) {
      if (hours[i]) {
        if (start === null) start = i;
      } else if (start !== null) {
        ranges.push([start, i]);
        start = null;
      }
    }
    if (start !== null) ranges.push([start, 24]);
    return ranges.map(function (r) {
      var pad = function (n) { return String(n).padStart(2, "0"); };
      return pad(r[0]) + ":00-" + pad(r[1]) + ":00";
    }).join(", ");
  }

  function cloneTimeSettings() {
    return {
      time_zone: timeSettings.time_zone,
      daylight_saving: timeSettings.daylight_saving,
      sync_servers: timeSettings.sync_servers,
      schedule: timeSettings.schedule.map(function (day) { return day.slice(); })
    };
  }

  function getTimeSettings() {
    return cloneTimeSettings();
  }

  function updateTimeSettings(patch) {
    if (!patch) return cloneTimeSettings();
    if (patch.time_zone != null) timeSettings.time_zone = patch.time_zone;
    if (patch.daylight_saving != null) timeSettings.daylight_saving = !!patch.daylight_saving;
    if (patch.sync_servers != null) timeSettings.sync_servers = patch.sync_servers;
    if (patch.schedule) {
      patch.schedule.forEach(function (day, i) {
        if (day && timeSettings.schedule[i]) {
          timeSettings.schedule[i] = day.slice();
        }
      });
    }
    if (typeof patch.dayIndex === "number" && patch.hours) {
      timeSettings.schedule[patch.dayIndex] = patch.hours.slice();
    }
    return cloneTimeSettings();
  }

  var lists = {
    "/api/v1/groups": groups,
    "/api/v1/symbols": symbols,
    "/api/v1/routing": routing,
    "/api/v1/datafeeds": datafeeds,
    "/api/v1/leverage-profiles": leverages,
    "/api/v1/holidays": holidays,
    "/api/v1/clients": clients,
    "/api/v1/managers": managers.map(function (m) {
      return { login: m.login, name: m.name, groups: m.groups, rights: m.rights, mailbox: m.mailbox };
    }),
    "/api/v1/users": users,
    "/api/v1/orders": orders,
    "/api/v1/positions": positions,
    "/api/v1/deals": deals,
    "/api/v1/dealing": orders.filter(function (o) { return o.state === 0; }),
    "/api/v1/journal": journal
  };

  var details = {
    "/api/v1/groups": groups,
    "/api/v1/symbols": symbols,
    "/api/v1/routing": routing,
    "/api/v1/clients": clients,
    "/api/v1/users": users,
    "/api/v1/managers": managers,
    "/api/v1/system/end-of-day": endOfDay
  };

  function findById(pool, key, id) {
    var num = Number(id);
    return pool.find(function (r) { return String(r[key]) === String(id) || r[key] === num; });
  }

  function accountsForClient(clientId) {
    return users.filter(function (u) { return u.client_id === Number(clientId); });
  }

  function ordersForLogin(login) {
    return orders.filter(function (o) { return o.login === Number(login); });
  }

  function positionsForLogin(login) {
    return positions.filter(function (p) { return p.login === Number(login); });
  }

  function dealsForLogin(login) {
    return deals.filter(function (d) { return d.login === Number(login); });
  }

  function clientForId(id) {
    return findById(clients, "client_id", id);
  }

  function userForLogin(login) {
    return findById(users, "login", login);
  }

  function groupForPath(path) {
    return groups.find(function (g) { return g.group === path; });
  }

  function normEndpoint(path) {
    return path.split("?")[0].replace(/\/+$/, "");
  }

  function list(endpoint, query) {
    var key = normEndpoint(endpoint);
    var rows = (lists[key] || []).slice();
    var search = query && query.search;
    if (search) {
      var s = search.toLowerCase();
      rows = rows.filter(function (r) {
        return JSON.stringify(r).toLowerCase().indexOf(s) >= 0;
      });
    }
    return rows;
  }

  function get(endpoint, id) {
    var key = normEndpoint(endpoint);
    if (key === "/api/v1/system/end-of-day") return endOfDay;
    if (key === "/api/v1/clients") return clientForId(id) || clients[0];
    if (key === "/api/v1/users") return userForLogin(id) || users[0];
    if (key === "/api/v1/managers") return managerDetailRecord(id);
    var pool = details[key] || lists[key] || [];
    if (key.indexOf("groups") >= 0) return findById(pool, "group_id", id) || pool[0];
    if (key.indexOf("symbols") >= 0) return symbolDetail(id);
    if (key.indexOf("routing") >= 0) return findById(pool, "routing_id", id) || pool[0];
    if (key.indexOf("datafeeds") >= 0) return getDatafeed(id);
    if (key.indexOf("deals") >= 0) return findById(deals, "deal_id", id) || deals[0];
    if (key.indexOf("orders") >= 0) return findById(orders, "order_id", id) || orders[0];
    if (key.indexOf("positions") >= 0) return findById(positions, "position_id", id) || positions[0];
    return findById(pool, "id", id) || pool[0] || null;
  }

  function routingDealersList(routingId) {
    return routingDealers[routingId] || routingDealers[1] || [];
  }

  function defaultId(endpoint) {
    var row = list(endpoint)[0];
    if (!row) return "1";
    if (row.group_id) return String(row.group_id);
    if (row.symbol_id) return String(row.symbol_id);
    if (row.routing_id) return String(row.routing_id);
    if (row.login) return String(row.login);
    if (row.client_id) return String(row.client_id);
    if (row.order_id) return String(row.order_id);
    if (row.deal_id) return String(row.deal_id);
    if (row.position_id) return String(row.position_id);
    return "1";
  }

  global.HSTMock = {
    list: list,
    get: get,
    routingDealers: routingDealersList,
    defaultId: defaultId,
    endOfDay: endOfDay,
    accountsForClient: accountsForClient,
    ordersForLogin: ordersForLogin,
    positionsForLogin: positionsForLogin,
    dealsForLogin: dealsForLogin,
    clientForId: clientForId,
    userForLogin: userForLogin,
    groupForPath: groupForPath,
    managerForLogin: managerForLogin,
    managerRightsView: managerRightsView,
    managerReports: managerReports,
    managerDetailRecord: managerDetailRecord,
    fmtManagerLimit: function (n) { return MANAGER_LIMITS[n] != null ? MANAGER_LIMITS[n] : String(n); },
    rightsLabels: RIGHTS_LABELS,
    getTimeSettings: getTimeSettings,
    updateTimeSettings: updateTimeSettings,
    formatDayRange: formatDayRange,
    getSymbol: getSymbol,
    getSymbolSessions: getSymbolSessions,
    updateSymbol: updateSymbol,
    symbolDetail: symbolDetail,
    getDatafeed: getDatafeed,
    createDatafeed: createDatafeed,
    updateDatafeed: updateDatafeed,
    deleteDatafeeds: deleteDatafeeds,
    moveDatafeedInList: moveDatafeedInList,
    setDatafeedEnable: setDatafeedEnable,
    listDatafeedModules: listDatafeedModules,
    managers: managers,
    clients: clients,
    users: users,
    groups: groups
  };
})(window);
