# Build Prompt — HST Administrator & Manager Web Frontends

The single source-of-truth prompt for building two web frontends against the HST backend:
an **Administrator** panel and a **Manager** panel, pixel-faithful to the MT5 Administrator
application. Give this file to the builder (human or agent). Everything needed is either in
this file or at a path this file names. All research below was extracted from the actual
MT5 docs, the actual swagger specs, and the actual code — with source references.

---

## 1. Mission and non-negotiables

- The design is not ours to improve: **exact** MT5 Administrator layout, icons, dialogs,
  table conventions, interaction rules. Where this file cites a doc image, the image IS the
  spec. `<DOCS>` = `/Users/jeelrupapara/Documents/MetaTrader5/Meta Document - SQL/`.
- **No vendor naming anywhere**: no `MT5`, `MetaTrader`, `MetaQuotes`, `mql5` in components,
  classes, labels, comments. Components are named for what they show (`SettingsDialog`).
  Product name: **HST Administrator** / **HST Manager**; the server is **Trade Server**.
- Data comes from our backend (Section 9–11 are the contract). The backend MAY be changed
  to serve the frontend better (Section 13 lists known gaps); the design may NOT be changed
  to suit the backend.
- Well-structured React: modules are data (definition objects), one shared component set,
  no copy-pasted screens. The prototype's good patterns (Section 5) are the base.
- Every screen is verified with **Playwright MCP** against the MT5 reference image before
  it is called done (Section 12).

## 2. Sources of truth

| What | Where |
|---|---|
| MT5 Administrator docs (HTML + 1285 PNGs) | `<DOCS>` |
| MT5 Manager/Admin PDFs | `~/Documents/MetaTrader5/Meta Document - Admin/`, `.../Meta Document - Manager/` |
| Admin API (91 paths / 141 ops) | `hst-server/swagger/admin/admin_swagger.json` |
| Trader API (35 paths / 44 ops) | `hst-server/swagger/trader/trader_swagger.json` |
| Backend models & enums | `hst-server/model/` |
| Navigation endpoint | `hst-server/internal/server/v1/navigation.go` |
| Prototype (patterns + CSS skin to keep) | `hst-v2/hst-html-modules/` |
| Socket contract | `hst-server/model/subjects.go`, `events.go`, `internal/server/v1/ws.go`, `hst-core/model/wire.go` |

## 3. Scope for v1, in build order

1. **Shell** — chrome, menus, toolbar, Navigator, Toolbox (Journal + Search), status bar, login.
2. **Symbols** — nav folder tree from `path`, list, full settings dialog.
3. **Data Feeds** — list + status dot, dialog: creds, symbol patterns, Translations, Parameters.
4. **Groups** — nav folder tree from `group` paths, list, full 8-tab dialog.
5. **Group Symbols** — the Symbols tab inside the group dialog (scope rows + overrides).
6. **Accounts & Clients** — lists + dialogs.
7. **Leverages** — profiles, rules, tiers.
8. **Routing** — rules, conditions, actions, dealers, ordering.
9. **Journal & Search** — Toolbox tabs fed by `/api/v1/journal` + live `journal` events.

A module is done only when all four exist: nav entry → list → settings dialog → live refresh.
Post-v1 (specs included so nothing is designed twice): Managers, Holidays, Mail Servers,
Positions/Orders/Deals admin views, Balance operations, Dealing desk.

---

## 4. What exists — keep, replace, delete

The prototype at `hst-html-modules/` is a single Vite SPA (`index.html` → `src/main.jsx` →
`/admin|/manager/:moduleId/:recordId?`). Full audit findings:

**KEEP (the architecture of the real build):**
- `src/components/ui/SettingsDialog.jsx` — dialog shell `{title, tabs, children, footer,
  width=656, mode:"inline"|"overlay", draggable, onTitlePointerDown}`. Every settings
  dialog uses it, dragged via `useDialogDrag`.
- Module **definition objects** (`src/features/modules/definitions/admin.js`, `manager.js`,
  merged by `registry.js` into `"panel.moduleId"` keys). Shape: `{type: list|settings|split,
  title, endpoint, query, idKey, pathKey, columns, editLabel, detailKind, note,
  searchPlaceholder}`. New module = new definition, not new screen.
- `ListModule` (generic list: ApiBar, toolbar, search, DataTable, `?folder=` prefix filter
  via `pathKey`), `SplitModule` (symbols workspace), `DataTable`, `PropertyTable`,
  `PropSelect`, `ContextMenu` + builders, `InputDialog`, `Icon`/`NavIcon`.
- `NavTree.jsx` — folder rows (symbol/group/datafeed) are `div`s with custom active checks,
  NOT `NavLink`s: folders are addressed by `?folder=` query and a NavLink matches pathname
  only, which highlights the whole tree (this bug was fixed once; do not reintroduce).
- `src/lib/groupNavTree.js` / `symbolNavTree.js` / `datafeedNavTree.js` — path→tree
  injectors (algorithm in Section 8).
- `assets/css/terminal.css` (2597 lines) — the faithful skin. Tokens: `--chrome-bg #ece9d8`,
  `--select-bg #316ac5`, `--nav-width 220px`, heights `--title-h 22/--menu-h 24/--toolbar-h
  28/--toolbox-h 140/--status-h 22`, Tahoma 11px. Classes `.config-window .settings-dialog
  .config-title .config-tabs .config-body .prop-table .prop-select .nav-item .journal-table`.
  Extend, never restyle.
- Icons: `assets/icons/svg/` (51 SVGs) + `sprite.svg` — committed artifacts cropped from the
  MT5 doc screenshots by `scripts/build-icons.py`. Treat as final; new icons follow the same
  pipeline.
- Legacy pages as **spec only** (field layouts to port, then delete): `modules/admin/
  groups-config.html`, `managers-detail.html` + `managers-ui.js` (rights matrix),
  `routing-config.html`, `modules/manager/{clients,users}-detail.html`, trade `*-detail.html`
  + `trade-ui.js` (chain view), `assets/js/context-menus.js` (21 registered menus = the
  context-menu content spec, with right-gating semantics).

**REPLACE:** every legacy iframe detail (`legacyDetail`) with a React `SettingsDialog`
implementation. Only symbols, datafeeds and time have real React detail today.

**DELETE:** root `admin.html`/`manager.html`/`login.html` + `assets/js/{terminal,context-menu*,
managers-ui,symbols-ui,module,icons,api}.js` (pre-React prototype), `modules/admin/
{clients-accounts,orders-deals,symbols,symbols-config,time}.html` (redirect/stub),
`src/config/*` (deprecated re-export shims with zero importers).

**Known debts to fix in the real build** (found in audit):
- `lib/data.js` silently falls back to mock on live failure and **all writes are mock-only**
  — real build: writes go to the API, failures are shown, demo mode is explicit.
- No token refresh / 401 handling / route guard in `lib/api.js` — add refresh + redirect.
- Menu bar, toolbar buttons, Toolbox rows are decorative — wire them (Section 6).
- `getLegacyDetailSrc` builds broken paths for admin clients/users details (files live under
  `modules/manager/`) — moot once details are React.
- `note` renders raw HTML via `dangerouslySetInnerHTML` — sanitize or drop.
- `useAdminNavTree` refetches symbols+groups+datafeeds on every folder click — fetch once,
  key on `refreshKey`, compute open state locally.

---

## 5. Shell specification

Reference images: `<DOCS>/start_page_platform.png` (whole window), `tree.png` (navigator),
`groups.png` (list view + folder counts), `toolbox_journal.png` (bottom panel),
`status_bar.png`, `groups_common.png` (canonical dialog), `toolbar_settings.png`,
`search_window.png`, `group_work_symbols.png` (bulk edit).

### Regions
- **Title bar**: `HST Administrator — Trade Server (Live)`.
- **Menu bar** — File, Edit, View, Services, Help (contents per `<DOCS>/menu_*.htm`,
  adapted to what we support; omit, don't stub):
  - File: Export (JSON, exports selected tree section if one selected), Import, Exit.
  - Edit: Apply Changes · Add · Edit · Delete · Move Up · Move Down · Sort Alphabetically
    (server-side) · Find (`Ctrl+F`) · Find Next (`F3`).
  - View: Color Themes (light/dark later), Toolbar, Status Bar, Toolbox.
  - Services: Refresh Configuration.
  - Help: About.
- **Toolbar** (order from `start_page_platform.png`): Connect/Disconnect · Refresh · Apply ·
  Add · Edit · Delete · Sort · Move Up · Move Down · Export · Import · Toolbox · Help.
  Greyed when inapplicable. Acts on the active list's selection. Icons:
  `<DOCS>/toolbar_*_button.png`.
- **Navigator** (left, 220px, draggable split): Section 8.
- **Content** (right): active module.
- **Toolbox** (bottom, hideable): tabs `Journal | Search` (tab strip at the BOTTOM of the
  panel, close ✕, per `toolbox_journal.png`). Journal columns `Time | Server | Message`,
  severity row icons (`journal_info_icon.png` / `journal_warning_icon.png` /
  `journal_error_icon.png`), context menu Copy (`Ctrl+C`) / Auto Arrange / Grid.
- **Status bar**: hint text left ("For Help, press F1" style); right: connection dot
  (connected/disconnected per `status_bar_connected.png`) + feed-alive indicator.

### Interaction conventions (every list, every module)
- **Double-click row → settings dialog.** `Enter` on selection does the same.
- Add/Edit/Delete identical in toolbar, Edit menu, and row context menu.
- Multi-select (Ctrl/Shift) + Edit → **bulk edit dialog**: non-bulk fields greyed, only
  touched fields written (`group_work_symbols.png`).
- Save-as-copy: change the key field in an open dialog, OK → new record. Key fields:
  Groups=Name, Symbols=Symbol, Managers=Login, Routing/Datafeeds=Name,
  Holidays=Date+Description.
- Canonical list context menu (`admin_groups.htm`): Add · Edit · Delete · Move Up · Move
  Down · Sort · Export/Import · Request ▸ (Journal/Orders/Deals/Positions) · Find · Auto
  Arrange (auto column sizing) · Grid (separators). Port the 21 menus in
  `assets/js/context-menus.js` with their right-gating.
- Config edits commit on **Apply**; dialogs commit on OK.

### Dialog conventions
- `SettingsDialog`, fixed 656px, draggable by title, never maximizable.
- Title `Entity: key` (`Group: demo\forex-usd`).
- Horizontal tab strip under the title; body opens with the module icon + 2–3 grey
  description lines (exactly as `groups_common.png`).
- Right-aligned labels / left controls; two columns where MT5 uses two; checkbox stacks
  indented. Inapplicable controls **greyed, not hidden**.
- Footer `OK | Cancel` bottom-right, OK default.

---

## 6. Data conventions (memorize)

- **Volume is lots** (`0.01`) everywhere on the HTTP/WS wire. Never rescale. (Internally the
  server uses integer units; conversion is at the handler boundary — trust the DTO.)
- **Timestamps are unix nanoseconds** → `ms = ns / 1e6`. `Event.at` too.
- **Identity is the backslash path**: groups `demo\forex\usd`, symbols carry `symbol` +
  `path` `Forex\Majors\EURUSD`. No numeric symbol id in lists. JS literals need `\\`;
  URL-encode path segments.
- **Enums are ints on the wire; labels come from `_name` maps** (Section 10 = the generated
  constants file). `EventType` is the exception — a string enum, the wire value IS the label.
- **Zero often means unset**, but per-enum: routing `TypeFlags` 0 = ALL types; `OrderTime`
  0 = gtc (real default); `UsersRights` 0 = nothing; event verb enums are 1-based. Nullable
  `GroupSymbol` fields mean "inherit" — a PATCH must distinguish omitted from 0.
- Prices/money: float64 + a digits field alongside (`digits`, `currency_digits`) — format
  with those, never fixed 2.
- Envelope: every REST response is `{success, code, data, error, message}`; lists put the
  array directly in `data`, **no total** (gap #5 in Section 13).
- Paging: `limit` (≤500), `page` (0-based), `from`/`to` unix **seconds**. Config lists also
  take `search`, `sort_by` (allowlisted), `order`.
- Auth: `POST /auth/v1/login` — HTTP **Basic** (login:password) + JSON body
  `{connection_type: 32}` admin / `33` manager. Returns `{access_token, refresh_token,
  expires_in, login, session_id, connection_type}`. Bearer token thereafter; refresh via
  `/auth/v1/refresh`. Staff may hold parallel sessions (traders are single-session).

---

## 7. Navigation contract (backend-driven)

`GET /api/v1/navigation` (Bearer; panel comes from the session's connection_type — a client
cannot request the other tree). Response `data`:

```json
{ "login": 1000, "terminal": "administrator", "groups": ["demo\\*"],
  "nodes": [ { "key","label","icon","route","section","count?","children?" } ],
  "can": { "right_admin": true, "...77 keys always present...": false },
  "rights": [/* [2]uint64 — DO NOT USE, loses precision in JS */],
  "visible": 17, "total": 17, "hidden": 0 }
```

**adminTree** (exact): heading `clients_accounts` {accounts→/users, clients→/clients,
managers→/managers} · heading `orders_deals` {positions, orders, deals} · `groups`→/groups ·
`symbols`→/symbols · `leverages`→/leverage-profiles · `routing`→/routing ·
`datafeeds`→/datafeeds · `mail_servers`→/mail-servers · `holidays`→/holidays ·
`end_of_day`→/system/end-of-day · `journal`→/journal.
**managerTree**: heading `clients_orders` {accounts, clients, positions, orders, deals} ·
`dealing` · `balance` · `groups` · `journal`.

Rules the frontend must follow:
- Use `can{}` for every button/menu gate. Never decode `rights`. Derived helpers to mirror:
  deal = `trades_read && trades_dealer`; manage trades = `trades_read && trades_manager`.
- `route: ""` = heading (always has children, else the server dropped it).
- Counts: render `(n)` when present. Count 0 can mean "count query failed" — never treat as
  authoritative empty. Orders count = open states only {0,1,3,7,8,9}.
- `groups` node arrives with **flat** children (`key:"group:demo\forex-usd"`, label = leaf,
  count = accounts in that exact group). Rebuild folders client-side (Section 8). Note the
  backend count is *accounts*; the MT5 tree count is *groups beneath* — MT5 wins: show
  groups-beneath on folders (compute client-side).
- `symbols`/`datafeeds` have no children server-side — enrich client-side (Section 8),
  lazily on first expand, and never refetch on folder click.
- Panel prefix: backend routes are bare (`/users`); links are `/{panel}` + route.
- Live refresh: map socket events → refetch `/api/v1/navigation` **debounced ~2s** (one
  coalesced call recomputes all counts): group_*/symbol_*/user_*/client_*/manager_*/
  holiday_*/leverage_*/mail_server_* events, order/position events for those counts.
  `session.revoked` → drop cache, back to login.
- MT5 wrapper: render a synthetic `Servers > Trade Server` spine above the sections, and
  group section `feeds` under an `Integrations` heading — until the backend adds these
  nodes (gap #A in Section 13).

## 8. Path → folder tree algorithm (groups, symbols)

From `src/lib/groupNavTree.js` (reuse verbatim):
1. Root `{name:"", path:"", children:{}, exists:false, count:0}`.
2. Per record, split path on `\`, walk/create per segment accumulating the path; mark the
   final node `exists=true`. A node can be **both** folder and record (`demo` is a group
   AND holds 29). Folder-only nodes exist (`exists:false`) — MT5's `a`/`demo2`/`demo3` rows.
3. Post-order count = self(1 if exists) + descendants. Render `(n)`.
4. Sorted children (`localeCompare`); nav node `{moduleId, groupPath|folderPath, label,
   icon, indent, count, children?}`.
5. `openAlongPath(active)` expands ancestors; selection = exact path match only.
6. Selecting a folder navigates `?folder=<path>` and the list filters
   `path === folder || path.startsWith(folder + "\\")`.
Symbols: identical, but drop the last path segment first (it's the symbol, not a folder);
folder icon = module glyph with folder overlay, record = plain glyph (see `groups.png`).

---

## 9. Module specs

Format: reference images → list → dialog tabs/fields → endpoints. Field tables come from
the MT5 docs; endpoint/model names from swagger. Grey any field the backend lacks (design
stays; backend catches up).

### 9.1 Symbols
Images: `symbols.png`, `symbols_common.png`, `symbols_trade.png`, `symbols_sessions.png`,
`symbols_filtration.png`, `symbols_sessions_adjust.png`.

- List columns: `Symbol | Type | Execution | Digits` (Type = top folder). Bottom filter tab
  accepting comma masks (`EUR*, !GBPUSD`) for cross-folder batch edit.
- Endpoints: `GET/POST /api/v1/symbols`, `GET/PATCH/DELETE /api/v1/symbols/{id}`. List model
  `ViewSymbol{symbol_id, symbol, path, description, digits, spread, contract_size,
  calc_mode, exec_mode, trade_mode, date_modified}`. Detail `ViewSymbolDetail` (~150 fields)
  fills every tab in one call, `sessions[]` embedded `{day 0=Sun..6, type 0=quote|1=trade,
  open, close}` (minutes from 00:00). Create requires `symbol, path, currency_base,
  currency_profit, currency_margin`.
- Dialog tabs (each with icon+description header):
  **Common**: Symbol, Description, ISIN, International name, Exchange, Category, CFI,
  Sector, Industry, Country, Basis, Page, Source, Digits, Background color, Market depth
  (off/1–32), Spread (0=floating), Spread balance (bid/ask shifts), Chart mode (Bid/Last).
  **Currency**: Base / Profit / Margin.
  **Quotes (Filtration)**: realtime-quotes ✓, negative prices ✓, raw prices ✓, market
  stats ✓, soft/hard/discard levels + counts, min/max spread, gap level, disable-gap-after.
  **Trade**: Contract size, Tick size/value, Calculation (CalcMode labels, Section 10),
  Trade (disabled/long_only/short_only/close_only/full), GTC mode, Limit & stop level,
  Freeze level, Max quote delay, Filling ✓✓ (FOK/IOC/BOC; Return implicit), Expiration
  ✓✓✓✓, Orders ✓✓✓✓✓✓✓ (market/limit/stop/stop-limit/SL/TP/close-by), Volumes
  (min/max/step/limit).
  **Execution**: Mode (Instant/Request/Market/Exchange) + per-mode: time/profit/losing
  deviations, max volume before Request, fast-confirmation ✓, timeout, confirm orders ✓.
  **Margin**: Initial/Maintenance/Hedged, larger-leg ✓, exclude-long-PnL ✓, EOD rate
  recalc ✓, additional checks; **Margin Rates** grid: Initial+Maintenance × Buy/Sell ×
  Market/Limit/Stop/Stop-Limit (+ liquidity & currency rates).
  **Swaps**: enable, Type (SwapMode labels), Long, Short, Days-in-year (360/365/366),
  weekday multiplier grid with helper buttons `Forex` / `All week` / `From symbol`,
  consider-holidays ✓.
  **Sessions**: 7-day grid, Quotes + Trade rows per day, drag-to-create windows, Ctrl
  multi-day, Shift for minute precision; plus Use-time-limits + From/To.

### 9.2 Data Feeds
Images: `data_feeds.png`, `data_feeds_common.png`, `data_feeds_symbols.png`,
`data_feeds_translation.png`, `data_feeds_parameters.png`.

- Nav: one child per feed with green/red state dot (`enable` + `sys_connection`).
- List: `Name | Source (Q/N) | Server | Symbols | Last active | State`. Row order =
  priority; Up/Down reorder. Red row in Symbols tab = symbol no longer exists.
- Endpoints: `/api/v1/datafeeds` CRUD, `/modules?mode=` (module dropdown), `/{id}/activate`,
  `/{id}/params` CRUD, `/{id}/translates` CRUD, `/{id}/symbols` (GET/POST/DELETE +
  `/resolve` preview; **no PATCH** — gap #9). `ViewDatafeedDetail` = one call fills all
  tabs (`params[]`, `feed_symbols[]`, `translates[]`). Passwords never returned.
- Dialog tabs:
  **Common**: Enable, Name, Module (from /modules), Feed server `host:port`, Feed login,
  Password; advanced: Gateway server/login/password.
  **Symbols**: scope rows — symbol or path mask (`Forex\*`, `!GBPUSD`), exclude flag,
  Add/Delete, resolve preview; "Allow importing symbol settings" ✓ (imports land disabled
  in `Preliminary`).
  **Translations**: `Symbol (ours) | Source (theirs) | Bid shift | Ask shift` — first match
  wins; this is where unmapped source names are translated.
  **Parameters**: typed key/value rows (`FeederParamType` 0 string…9 color), Add/Edit/
  Delete/Default. FIX identities (`SenderCompID`/`TargetCompID`) live here.
  **Timeouts**: reconnect interval / attempts / series interval (backend stores
  `timeout, timeout_reconnect, timeout_sleep, attempts_sleep`).

### 9.3 Groups (the flagship module)
Images: `groups.png`, `groups_common.png`, `groups_company.png`, `groups_news_mail.png`,
`groups_permissions.png`, `groups_margin.png`, `groups_symbols.png`,
`groups_comissions.png`, `groups_comissions_settings.png`, `groups_reports.png`,
`groups_symbols_settings_*.png`, `group_section_add.png`.

- **List columns** (per `groups.png` + `leverages_group_list.png`): `Group | Server |
  Company | Type (Netting/Hedged) | Authentication | Risk Management | Margin ("50 / 30 %")
  | Currency`. NO "Floating Leverage Profile" column (screenshots prove it — it's a Margin
  tab field only). Folder rows appear only in the left tree; the list shows real groups of
  the selected node. Sortable; query bar with comma group masks + Request + result tabs
  `Groups (N)` / `Groups <mask> (n of N)`.
- Rules: implicit folder creation (`real\IB\realUSD` in Name creates sections), section
  dies with its last group, group with accounts cannot be deleted, name charset
  letters/digits/`_`/`-`, group type by case-sensitive substring (demo/manager/contest/
  coverage/preliminary else Real).
- Endpoints: `GET /api/v1/groups` (tree; `?flat=1` list) — `ViewGroup{group, name, exists,
  groups[], group_id, + 41 columns}`; CRUD `/{id}`; `/{id}/symbols` CRUD (`ViewGroupSymbol`
  with per-order-type margin rates, swaps, `use_default_*` inherit toggles);
  `/{id}/commissions` CRUD (`ViewCommission` + `tiers[]`). **`UptGroup` cannot rename** —
  gap #3.
- **Dialog — 8 tabs** `Common | Company | News & Mail | Permissions | Margin | Symbols |
  Commissions | Reports`, footer OK/Cancel:
  **Common** (two-column): Name (path), Currency (editable combo), Trade server (fixed),
  Digits (greyed for standard ccy), Authentication (Normal / 1024-bit RSA / 2048-bit RSA /
  Custom SSL), Min password length, One-time password (Disabled / TOTP SHA-256 / TOTP for
  Web only), Force OTP ✓, Push notifications (deals/orders/balances; real groups only);
  checkboxes: Enable connections, Enable certificate confirmation, Change password at
  first login, Show risk warning, Enforce country-specific restrictions.
  Backing: `auth_mode, auth_password_min, currency, currency_digits, permission_flags`.
  **Company**: Company (mandatory select), Company site, Company email, Deposit URL,
  Withdrawal URL, Support site (macros `[lang:…]`, `[login]`), Support email, Templates
  folder (`company_catalog`).
  **News & Mail**: News (disabled/headers/full package), News categories (comma list),
  News languages (text + Change), Enable internal mail ✓.
  **Permissions**: Maximum symbols/positions/orders (`unlimited` combos → `limit_*`),
  Available history (All/1m/3m/6m/1y/2y/3y), Deposit by default (demo), Leverage by
  default, Annual interest rate %, Trading Signals (disabled/all/own), Transfer of funds
  (4 modes), checkboxes from `trade_flags`: Expert Advisors, FIFO close (hedging only),
  Trailing stops, Prohibit hedge (hedging only), Charge swaps, Deal cost, Inactivity
  period (demo, days).
  **Margin**: Risk management (`for Retail Forex, CFD, Futures` = netting / `…with
  hedging` / `for Stock Exchange, based on margin discount rates`) — server restart noted;
  Margin call + Stop out + in (%/money); Stop-out-fully-hedged ✓, Compensate negative
  balance ✓ → Withdraw credit after ✓ (dependent); Floating leverage profile (from
  Leverages); group box "Profit/loss in free margin": Unrealized profit (4 modes), Daily
  fixed profit (2 modes) → Release fixed profit at EOD ✓ (dependent); Virtual credit.
  **Symbols**: inner grid `Symbol | Spread | Trade` with `*` default row, rails **Up/Down**
  (top) + **Add/Edit/Delete** (bottom), red row = dead path, top-down application, masks
  `Forex\*,!EURUSD` (≤128 chars, not exclusions-only). Row editor = 6-tab dialog
  (Common/Trade/Execution/Margin/Margin Rates/Swaps) where every section has a
  `Use default …` checkbox that greys its fields (→ `use_default_*` + nullable columns).
  **Commissions**: inner grid `Name | Symbol | Type | Range | Charge | summary`; entry
  dialog: Name, Description (=deal comment, ≤31), Symbol mask, Commission type radio
  (Standard/Agent/Fee), Range (Volume/Turnover money/Turnover volume/Notional/Profit),
  Charge (Instant/Daily/Monthly), Turnover currency, Deal entry/action/profit/reason
  filters (Instant only); **levels grid** `From | To | Commission | Minimal | Maximal |
  Mode | Currency | Type` (modes: deposit/base/profit/margin ccy, points, percents,
  specified ccy; type per-trade/per-volume; ranges non-overlapping, first match).
  **Reports**: Generate report data (Disabled / EOD+EOM / EOD only / EOM only) — controls
  below grey when disabled; Generate statements ✓, Send by email ✓, Mail server (Default +
  configured list from `/api/v1/mail-servers`), Copies to support ✓.

### 9.4 Accounts & Clients
Images: `accounts.png`, `account_overview.png`, `account_limits.png`,
`account_security.png`, `clients.png`, `clients_general.png`, `clients_accounts.png`.

- **Accounts list**: columns are user-selectable (docs name Balance/Checked,
  Credit/Checked); sensible default: `Login | Name | Group | Leverage | Balance | Credit |
  Email | Last access`. Type icons demo/real/manager/preliminary/technical/disabled
  (`<DOCS>/{demo,real,manager,preliminary,technica}_account_icon.png`,
  `account_disabled_icon.png`). Request bar: logins/`*`/group + Request.
- Endpoints: `/api/v1/users` CRUD (`ViewUser{login, name, group, email, phone, city,
  country, leverage, balance, credit, rights, client_id, is_manager, last_access}`;
  `CrtUser` req `email, group, name, password_main, password_investor`).
- **Account dialog tabs**: Overview (positions grid + status bar Balance/Credit/Equity/
  Margin/Free/Level + pending orders), Personal (name/contact/address/comment), Account
  (Group, Color, Leverage, checkboxes Enable account / Enable password change / OTP /
  Change password at next login), Limits (rights checkboxes from `UsersRights`: trading,
  expert advisors, trailing, API, reports, sponsored VPS, show-to-managers, daily reports;
  numeric position-value + orders limits), Security (Master/Investor/API password Change/
  Generate, OTP secret, certificate). `rights` bits in Section 10 — note `trade_disabled`
  is inverted (set = trading OFF).
- **Clients list**: default `ID | Name | Email | Phone | Status | KYC | Assigned manager |
  Modified`. Endpoints `/api/v1/clients` CRUD (list `ViewClient` 15 fields; detail
  `model.Client` ~65: person_*, company_*, compliance_*, experience_*, contact_*, lead_*,
  address_*, kyc_status). Delete takes `?force=`.
- **Client card** tabs (from `clients.htm`): General (type, ClientStatus pipeline,
  KYC, assigned manager, preferred group, lead source/campaign), Personal data, Address,
  Company (corporate), Regulation (employment/education/wealth/experience), Documents,
  Comments, History, Trading accounts (link/unlink, New Account).
- Balance operations (manager): dialog Operation (Balance/Credit/Charge/Correction/Bonus/
  Commission/Dividend/Franked/Tax) + Amount + Comment; blue button = deposit, red =
  withdrawal; endpoints `POST /api/v1/balance{,/deposit,/withdrawal,/credit,/correction}`;
  history grid = `/api/v1/deals` filtered by balance actions (`IsBalanceAction` set
  {2,3,4,5,6,7,12,15,16,17,18,19}).

### 9.5 Leverages
Images: `leverages.png`, `leverages_rule.png`.
- Profiles list (Name) → ordered rules, first-match-wins, up to 1024×1024.
- Rule dialog: Name, Description, Symbol mask, Range (volume / volume per symbol /
  notional value / notional value per symbol — currency field shows only for notional,
  `NeedsCurrency()`), tiers grid `To | Initial margin ratio | Maintenance margin ratio`
  (From auto-derived; maintenance 0 = no margin charged).
- Endpoints: `/api/v1/leverage-profiles` CRUD (PUT for update), `/{id}/rules` POST,
  `/{id}/rules/reorder` PUT `{rule_ids}`, `/{id}/rules/{ruleId}` PUT/DELETE.

### 9.6 Routing
Images: `routing.png`, `routing_common.png`, `routing_dealers.png`.
- List `Name | Dealers`, evaluated top→bottom, first match processes; changes restart
  in-flight requests. Up/Down + bulk reorder.
- Rule dialog Common: Enable, Name, **Perform action** (Delay ms / Delay ticks ≤60 /
  Clear TP / Clear SL / Clear SLTP / Process to dealers / Process to online dealers /
  Reject+reason ≤31 / Requote / Confirm by request price / Confirm by market price /
  Cancel order — RouteAction 0..4, 1001..1007), Where request is (RouteFlags bits),
  Where order is (TypeFlags bits; **0 = ALL**), condition rows `Type | Condition | Value`
  AND-ed — full RouteCondition catalogue in Section 10 (bands: request 0–16, account
  1000–1010, money 2000–2005, daily 3000–3002, book 4000–4015), comparators
  ConditionRule 0..5 (strings: `=` exact, `>`/`≥` substring-in-value, `<`/`≤` inverse).
- Dealers tab: rows of manager logins with dealing rights, Add/Edit/Delete, ordered.
- Endpoints: `/api/v1/routing` CRUD, `/order` bulk reorder, `/{id}/move-up|down`,
  `/{id}/dealers` POST/DELETE + `/{login}/move`.

### 9.7 Journal & Search (Toolbox)
- Journal: `GET /api/v1/journal` — `{journal_id, created_at(ns), type, code, channel, os,
  ip, message, login, detail}`; query `page, limit, from, to (s), type, mode
  (full|without_logins|errors_only), search (case-sensitive), sort_by, order`. Live append
  via `journal` socket events, newest first. Row icon by `code` (0 ok, 1 warn, 2 error,
  3 critical, 4 login). `type` labels: 0 all, 1 cfg, 2 sys, 3 net, 4 hst, 5 user, 6 trade,
  7 api, 8 notify.
- Search: client-side over loaded module lists (MT5 scopes to loaded rows); results
  `Server | Item`; Enter/double-click opens the record.

### 9.8 Post-v1 (specs ready)
- **Managers** (`managers.png`, `managers_permissions.png`): list `Login | Name | Mailbox |
  Groups`; dialog Common (login, mailbox, group masks `demo*`, `!managers*,*` — no
  prohibition-only), Permissions (the 77-right tree grouped as in Section 10, roles
  save/load), Reports (per-report depth), IP Access (From/To rows). Endpoints
  `/api/v1/managers` CRUD + `/{login}/rights`. Note gap #10: verify the write shape of
  `rights` before building the matrix.
- **Holidays** (`holidays.png`): ordered list; dialog Enable, Every year, Date, Work time
  From/To (both 0 = closed all day), Description, Symbols tab (masks). Top-down: first rule
  that ALLOWS trading wins. `/api/v1/holidays` CRUD + `/reorder` + `/check?symbol&date`.
- **Mail Servers** (`mail_servers.png`): list Name + Sent/Errors/Queue/TimeMin-Avg-Max;
  dialog Enable, Name, Sender email/name, SMTP `host:port` (25/465/587), login, password
  (write-only), Default server (single default enforced server-side).
- **Positions/Orders/Deals (admin)**: list columns per Manager docs — positions `Login |
  Position | ID | Symbol | Time | Type | Volume | Price | S/L | T/P | Current | Reason |
  Swap | Profit`; orders add `Order price | Trigger | State | Expiration`; deals `Time |
  Deal | Order | ID | Symbol | Type | Direction | Volume | Price | Commission | Profit |
  Swap | Comment | Reason`. Red rows = admin-modified. Endpoints exist (Section 13 notes
  the missing filters). Dealing desk drives `/api/v1/dealing/*` + dealer socket events.

---

## 10. Constants file (generate from `hst-server/model/`, do not hand-type)

The naming law (de facto, consistent across all 23 model files): constants are
`TypeName_lower_snake_value`; every UI enum ships `TypeName_name` (int→label) and
`TypeName_value` (label→int); flags are hex powers of two with decimal keys in `_name`.
Mirror the label strings exactly — they are API contract; humanize only in the view.

Key tables (full catalogue in `hst-server/model/`, file:line refs in the research):

- **OrderType** 0 buy · 1 sell · 2 buy_limit · 3 sell_limit · 4 buy_stop · 5 sell_stop ·
  6 buy_stop_limit · 7 sell_stop_limit · 8 close_by. Pending = 2..7.
- **OrderState** 0 started · 1 placed · 2 canceled · 3 partial · 4 filled · 5 rejected ·
  6 expired · 7 request_add · 8 request_modify · 9 request_cancel. Live = {0,1,3,7,8,9};
  awaiting dealer = {7,8,9} (drives Dealing).
- **OrderFilling** 0 fok · 1 ioc · 2 return · 3 boc. **OrderTime** 0 gtc · 1 day ·
  2 specified · 3 specified_day (2,3 require expiry).
- **OrderReason** 0 client · 1 expert · 2 dealer · 3 sl · 4 tp · 5 so · 6 rollover ·
  7 external_client · 8 vmargin · 9 gateway · 10 signal · 11 settlement · 12 transfer ·
  13 sync · 14 external_service · 15 migration · 16 mobile · 17 web · 18 split ·
  19 corporate_action · 20 ultency · 21 coverage.
- **PositionAction** 0 buy · 1 sell. **DealEntry** 0 in · 1 out · 2 inout · 3 out_by.
- **DealAction** 0 buy · 1 sell · 2 balance · 3 credit · 4 charge · 5 correction ·
  6 bonus · 7 commission · 8 commission_daily · 9 commission_monthly · 10 agent_daily ·
  11 agent_monthly · 12 interest · 13 buy_canceled · 14 sell_canceled · 15 dividend ·
  16 dividend_franked · 17 tax · 18 agent · 19 so_compensation · 20 so_compensation_credit.
- **MarginMode** 0 retail_netting · 1 exchange · 2 retail_hedging (UI: Netting/Hedged).
- Group: **AuthMode** 0 standard · 1 rsa1024 · 2 rsa2048; **NewsMode** 0/1/2
  disabled/headers/full; **MailMode** 0/1; **ReportsMode** 0 disabled · 1 full · 2 day_only
  · 3 month_only; **ReportsFlags** 1 email · 2 support · 4 statements;
  **TransferMode** 0 disabled · 1 by_name · 2 group · 3 name_group;
  **HistoryLimit** 0 all · 1..6 (1m/3m/6m/1y/2y/3y);
  **PermissionsFlags** 1 cert_confirm · 2 enable_connection · 4 reset_password ·
  8 forced_otp_usage · 16 risk_warning · 32 regulation_protect · 64 notify_deals ·
  128 notify_orders · 256 notify_balances;
  **GroupTradeFlags** 1 swaps · 2 trailing · 4 experts · 8 expiration · 16 signals_all ·
  32 signals_own · 64 so_compensation · 128 so_fully_hedged · 256 fifo_close ·
  512 hedge_prohibit · 1024 deal_cost · 2048 so_compensation_credit;
  **FreeMarginMode** 0 not_use_pl · 1 use_pl · 2 profit · 3 loss; **StopOutMode** 0 percent
  · 1 money; **MarginFreeProfitMode** 0 pl · 1 loss; **GroupMarginFlags** 1 clear_acc.
- Symbol: **CalcMode** (sparse!) 0 forex · 1 futures · 2 cfd · 3 cfd_index · 4 cfd_leverage
  · 5 forex_no_leverage · 32 exch_stocks · 33 exch_futures · 34 exch_forts ·
  35 exch_options · 36 exch_options_margin · 37 exch_bonds · 64 serv_collateral;
  **TradeMode** 0 disabled · 1 long_only · 2 short_only · 3 close_only · 4 full;
  **ExecMode** 0 request · 1 instant · 2 market · 3 exchange; **GTCMode** 0 gtc · 1 daily ·
  2 daily_no_stops; **FillingFlags** 1 fok · 2 ioc; **ExpirationFlags** 1 gtc · 2 day ·
  4 specified · 8 specified_day; **OrderFlags** 1 market · 2 limit · 4 stop · 8 stop_limit
  · 16 sl · 32 tp · 64 closeby · 127 all; **SwapMode** 0 disabled · 1 by_points ·
  2 by_symbol_currency · 3 by_margin_currency · 4 by_group_currency · 5 by_interest_current
  · 6 by_interest_open · 7 reopen_by_close_price · 8 reopen_by_bid · 9 by_profit_currency;
  **SymbolMarginFlags** 1 check_process · 2 check_sltp · 4 hedge_large_leg · 8 exclude_pl ·
  16 recalc_rates; **ChartMode** 0 bid_price · 1 last_price; **SymbolSessionType** 0 quote
  · 1 trade; SymbolSector/Industry per model.
- **UsersRights** (bitmask, hex): 0x1 enabled · 0x2 password · 0x4 **trade_disabled
  (inverted!)** · 0x8 investor · 0x10 confirmed · 0x20 trailing · 0x40 expert · 0x100
  reports · 0x200 readonly · 0x400 reset_pass · 0x800 otp_enabled · 0x2000
  sponsored_hosting · 0x4000 api_enabled · 0x8000 push_notification · 0x10000 technical ·
  0x20000 exclude_reports.
- **UsersConnectionTypes**: 0 client(desktop) · 1/2/4/5/6 mobile · 3 client_api_web(api) ·
  11 client_web(web) · 32 admin · 33 manager · 34 manager_api · 36 admin_api ·
  37 manager_api_web. `IsStaff` = ≥32.
- Datafeed: **FeederFlags** 1 quotes · 2 news · 8 remote; **FeederParamType** 0 string ·
  1 int · 2 float · 3 time · 4 date · 5 datetime · 6 groups · 7 symbols · 8 bool · 9 color.
- Commission: Mode 0 standard·1 agent; RangeMode 0 volume·1 turnover_money·2
  turnover_volume; ChargeMode 0 daily·1 monthly·2 instant; EntryMode 0 all·1 in·2 out;
  ActionMode 0 all·1 buy·2 sell; ProfitMode 0 all·1 profit·2 loss; TierMode 0
  deposit_currency·1 base·2 profit·3 margin·4 points·5 percent·6 specified; TierType 0
  per_trade·1 per_volume; ReasonFlags 1 client·2 expert·4 dealer·8 external·16 mobile·32
  web·64 signal·128 gateway·256 ultency.
- Routing: **RouteAction** 0 delay_time·1 delay_tick·2 clear_tp·3 clear_sl·4 clear_sltp·
  1001 dealer·1002 dealer_online·1003 reject·1004 requote·1005 confirm_client·1006
  confirm_market·1007 cancel_order; **ConditionRule** 0 equal·1 not_equal·2 greater·3
  not_less·4 less·5 not_greater; **RouteCondition** bands — request: 0 datetime·1 symbol·2
  volume·3 deviation·4 time·5 weekday·6 comment·7 expert·8 signal·9 dealer_login·10
  source_login·11 deviation_spread·12 gap·13 reason·14 request_price·15 value·16
  current_spread; account: 1000 login·1001 group·1002 country·1003 city·1004 color·1005
  leverage·1006 comment_client·1007 zip·1008 status·1009 client_id·1010 party_id; money:
  2000 margin·2001 margin_level·2002 margin_free·2003 equity·2004 balance·2005 profit;
  daily: 3000 daily_deals·3001 daily_deals_period·3002 daily_profit; book: 4000
  position_volume·4001 position_profit·4002 position_age·4003 position_modify_time·4004
  position_average_time·4005 position_total·4006 position_total_symbol·4007 order_total·
  4008 order_total_symbol·4009 position_sl_touched·4010 position_tp_touched·4011
  order_sl_touched·4012 order_tp_touched·4013 position_value·4014 order_in·4015 order_out.
  RouteFlags/TypeFlags have `_name` only — build the inverse client-side.
- Client: **ClientStatus** step-100 pipeline 0 unregistered · 100 registered · 200
  notinterested · 300 application_incompleted · 400 application_completed · 500
  application_information · 600 application_rejected · 700 approved · 800 funded · 900
  active · 1000 inactive · 1100 suspended · 1200 closed · 1300 terminated; KycStatus 0
  undefined · 1 approved · 2 declined; plus Gender/Employment/Industry/Education/Wealth/
  Communication/Experience/Origin per model.
- Leverage **RangeMode** 0 volume · 1 volume_per_symbol · 2 notional_value · 3
  notional_value_per_symbol (2,3 need currency).
- **RetCodes** (`hst-core/model/retcode.go`): 0 done · 10001 trading disabled · 10002
  market closed · 10003 not enough money · 10004 price off market · 10005 volume limit ·
  10006 stops too close · 10007 too many orders · 10008 invalid volume · 10009 order
  exists · 10010 unknown symbol · 10011 no price · 10012 rejected by routing · **10013
  requote (open requote dialog)** · 10014 timeout · 10015 hedging not allowed · 10016
  closing only · 10017 fill policy · 10018 expiry not allowed · 10019 frozen · 10020 no
  routing rule · 10021 position volume limit · 10022 account not found · **10023 dealer
  queued (move to Dealing)** · 10024 dealers returned · 10025 close-by needs older
  position. `TradeResult` carries `retcode` + server `message` — display the message,
  branch on the code.
- **Manager rights**: 77 bits, names `right_admin … right_comments_delete` (full ordered
  list in `model/manager.go:192`); wire form for UI = `can{}` map. Group-access masks:
  empty list = NO access; `*` = all; `demo` = exact; `demo\*` = subtree incl. itself;
  star-in-middle = one segment per star. Manager detail returns 77 flat 0/1 ints.
- Enums with constants but **no `_name` map** (frontend defines labels): AlertKind/Cond,
  ActivationFlags, ModifyFlags, FeederFieldFlags, DatafeedEnable, DatafeedSysConnection.

## 11. Socket contract

- **URL** `GET /ws` (root, not under /api). Auth: `Authorization: Bearer`, or
  `Sec-WebSocket-Protocol: bearer, <token>`, or `?access_token=`. First frame
  `{"type":"welcome","at":<ns>}`. Native ping every 25s (browser auto-pongs); app-level
  `{"type":"ping"}`→`pong` available. Inbound frames ≤4096 bytes. 64 dropped events ⇒
  server closes the socket. `{"type":"session.revoked"}` then close = go to login.
- **Envelope**: JSON text frames `{"type", "group"?, "at", "payload"}` (`group` = backslash
  path on group-scoped events). Binary frames = bare `market_feed` CSV:
  `symbol,bid,ask,last,volume,ts(s),open,high,low,close,change,change_percent`. Text
  frames = bare `account_summary` CSV: `summary,login,balance,credit,equity,margin,
  margin_free,margin_level%,profit[,position_id,profit]…` (margin_level carries a literal
  `%`).
- **A manager session is auto-subscribed** (no subscribe message needed) to: its session +
  broadcast subjects, its own journal (`…accounts.<login>.operations`), dealer approvals
  if `right_trades_dealer`, right-gated flat subjects (symbols/holidays/leverages/managers
  update feeds), and group-scoped families crossed with its masks
  (`websocket.groups.<family>.<token>`): groups/users/clients/accounts/group_symbols/
  group_commissions. Rights changes re-derive subscriptions live.
- **Market feed**: send `{"type":"start_market_feed"}` / `stop_market_feed`.
- **Config events → live refresh** (the payload is the full `View*` record on create/
  update, a small `*Ref` on delete): `group_*`, `group_symbol_*`, `group_commission_*`,
  `user_*` (+`user_moved` fires on the OLD group's subject), `client_*`, `account_updated`,
  `symbol_*`, `holiday_*` (+reordered), `leverage_*` (+rules, reordered), `manager_*`,
  `mail_server_*`. Refresh the open module's table on its events; bump nav counts
  debounced (Section 7).
- **Trading events** (admin sees its own; dealer desk sees `dealer_request`/
  `dealer_request_done`): `order_create/update/cancel/rejected/expired` (WireOrder /
  TradeResult), `position_create/update/close/close_by` (WirePosition),
  `deal_create` (WireDeal), `money_change`, `margin_call`, `alert_triggered`, `journal`.
  Wire shapes (volume in lots): WirePosition `{position_id, login, symbol, action, reason,
  volume, price_open, price_current, price_sl, price_tp, profit, storage, time_create,
  time_update, comment}`; WireOrder `{order_id, …, type, state, price_order, price_trigger,
  time_setup, time_expiration}`; WireDeal `{deal_id, order_id, position_id, action, entry,
  price, profit, storage, commission, fee, time}`.
- **Client may send** (dealer/manager gated): `order_dealer_create/update/cancel`,
  `position_dealer_update/close/close_by`, `balance_create` (echoes result),
  `dealer_confirm/requote/reject/cancel` `{request_id, login, price|reason}`. Refusals come
  back as `bad_request`/`forbidden` events with `{message, reason}`; trade outcomes arrive
  asynchronously on the account subjects, never as a reply.

## 12. Verification loop (mandatory per screen)

1. Open the module's reference images from `<DOCS>` (named in Section 9).
2. Build. `npm run dev` for the frontend; `make dev` at `hst-v2/` when live data is needed
   (`make status` proves the feed is alive). Demo mode must be explicit, never a silent
   fallback.
3. Drive with **Playwright MCP**: navigate → screenshot → compare with the reference —
   regions, tab order, column order, control types, icon placement, greyed states. Fix,
   re-screenshot, repeat until it matches.
4. Exercise behaviour: double-click opens the dialog; folder click narrows the list and
   highlights only itself; Add/Edit/Delete act on selection; OK persists against the live
   backend and the change survives reload; the module's socket events refresh the open
   table; `can{}` gating hides what the manager may not do.
5. Only then the next module.

## 13. Known backend gaps (fix in backend, not by bending the design)

From the swagger/code audit — each is sanctioned backend work when its module lands:
A. Navigation lacks the `Servers > Trade Server` spine and an `Integrations` heading;
   `groups` children are flat (no folders), `symbols`/`datafeeds` have no children; group
   route embeds an unencoded backslash. Nesting server-side would delete client tree code.
B. No datafeed enable/disable socket event — nav dots can go stale.
C. `websocket.broker.mail_servers_update` is published but nothing subscribes to it —
   wire it into `rightSubjects` (right_cfg_mails).
1. No group **folder** counts; no per-folder/`path` filtering or paging on `GET /groups`.
2. `GET /symbols` capped at 500 with no tree endpoint — the whole table pages down to
   build the tree.
3. `UptGroup` cannot rename/move a group (no `group` field) — MT5 allows it.
4. No total/meta in list envelopes → no "N of M" or last-page jump anywhere.
5. No balance-operations list endpoint (`/deals` has no `action` filter).
6. `/positions` has no paging/filters at all; `/orders`, `/deals`, `/dealing` lack
   symbol/group/login filters — the MT5 request bar has nothing behind it.
7. Datafeed symbols rows have no PATCH (delete+create today).
8. `CrtManager`/`UptManager` expose a `rights` field of unverified shape vs the 77 flat
   ints on the read side — check the handler before building the rights matrix.
9. No mail-server test-send; `/mail-servers` has no paging/search; `/managers` has no
   search/sort.
10. No journal export; no enumeration endpoint for journal `type`/`code` filter dropdowns.
11. No symbol batch-edit / copy-settings endpoints (MT5 masks-based bulk edit).

Backend-change rules: MT5 semantics; the repo naming law (`TypeName_lower_snake_value`,
`_name` + `_value` maps); swagger annotations on every new route; `make swagger` after.
