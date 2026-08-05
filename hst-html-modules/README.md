# HST HTML Modules (MT5-style prototype)

Static HTML frontend for **Administrator** and **Manager** panels, modeled on the MT5 Administrator help docs.

**Phase 2:** Modules load live data from the HST API when signed in. **Demo mode** (default) shows MT5-style sample data without a backend.

## Architecture (React 19 + Vite)

Feature-based structure with shared UI primitives:

```
src/
  app/routes.jsx           Lazy-loaded routes + code splitting
  components/
    ui/                    Icon, ApiBar, SearchField, ToolbarButton, CenteredCard
    data/                  DataTable, ModuleToolbar
    layout/                TerminalShell, NavTree, Toolbox, StatusBar
    module/                ModuleView, ListModule, LegacyModuleFrame
  features/
    modules/
      registry.js          Module lookup API
      definitions/         admin.js, manager.js (per-panel configs)
      columns/             shared.js, trade.js (reusable column defs)
    navigation/            adminNav.js, managerNav.js
  hooks/                   useDemo, useListData, useRowSelection, useFullscreen
  lib/                     api, data, formatters, legacy script loader
  pages/                   HomePage, LoginPage, TerminalPage
```

**React 19 notes:** This is a Vite SPA (client-rendered). Server Components apply when migrating to React Router 7 Framework Mode or Next.js — until then, keep interactive leaves small and lazy-load routes/pages for fast initial load.

## Quick start (React + hot reload)

```bash
cd hst-html-modules
npm install
npm run dev
```

Open http://localhost:5173 → pick **Administrator** or **Manager**. Edit files under `src/` and changes appear instantly (HMR).

### Static HTML (legacy)

The original static prototype is still available under `admin.html`, `manager.html`, and `modules/`:

```bash
python3 -m http.server 5500
```

Open http://localhost:5500

### Live API (optional)

1. Start `hst-server` on `:8080`
2. Add `http://localhost:5500` to `CORS_ORIGINS` in `.env`
3. Use **Connect API** (`login.html`) — demo mode switches off after sign-in

## Structure

```
index.html          Vite entry (React app)
src/                React components, routes, module configs
admin.html          Legacy static terminal shell (admin)
manager.html        Legacy static terminal shell (manager)
assets/
  css/terminal.css  MT5 chrome styling
  icons/svg/        MT5 SVG navigation & entity icons (see build-icons)
  icons/sprite.svg  Combined icon sprite
  js/terminal.js    Nav + toolbox tabs
  js/icons.js       HSTIcons helper for SVG icons
  js/mock-data.js   Demo datasets (default when not signed in)
  js/api.js         Bearer auth client
  js/module.js      List/config/detail boot helpers
modules/
  admin/            Platform setup (18 modules)
  manager/          Day-to-day operations (9 modules)
```

## Module → API mapping

| Module | API |
|--------|-----|
| admin/groups | `GET /api/v1/groups?flat=1` |
| admin/groups-config | `GET /api/v1/groups/{id}` |
| admin/symbols | `GET /api/v1/symbols` |
| admin/symbols-config | `GET /api/v1/symbols/{id}` |
| admin/routing | `GET /api/v1/routing` |
| admin/routing-config | `GET /api/v1/routing/{id}` + `/dealers` |
| admin/datafeeds | `GET /api/v1/datafeeds` |
| admin/leverages | `GET /api/v1/leverage-profiles` |
| admin/holidays | `GET /api/v1/holidays` |
| admin/time | React settings view (demo mock) — timezone, DST, weekly schedule |
| admin/clients | `GET /api/v1/clients` |
| admin/managers | `GET /api/v1/managers` |
| admin/managers-detail | `GET /api/v1/managers/{login}` + `/rights` |
| admin/users | `GET /api/v1/users` |
| admin/orders | `GET /api/v1/orders` |
| admin/positions | `GET /api/v1/positions` |
| admin/deals | `GET /api/v1/deals` |
| manager/* | Same endpoints where applicable |
| manager/dealing | `GET /api/v1/dealing` |
| manager/balance | `POST /api/v1/balance/{deposit\|withdrawal\|credit\|correction}` |
| manager/journal | `GET /api/v1/journal` |

Navigator items marked *disabled* in HTML are MT5 parity placeholders (future phases).

## Icons

MT5-style SVG icons live in `assets/icons/svg/`. Rebuild from MT5 help-doc sources:

```bash
# Requires Pillow; vtracer binary in scripts/.tools/; svgo via npm install
npm install
/tmp/mt5crop/bin/python3 scripts/build-icons.py   # or: npm run build-icons (with Pillow on PATH)
```

Sources are listed in `assets/icons/manifest.json` (default: `~/Documents/mt5administrator-help/`).

## What's next (phase 3)

- Network Cluster module (includes End of Day — separate from Time)
- Write/edit forms (PATCH/POST) for remaining config windows
- Dynamic navigator counts from `GET /api/v1/navigation`
- Pagination controls wired to `page` / `limit` query params
