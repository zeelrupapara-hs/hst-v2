# hst-admin-web — Project Rules

The rules every line of this codebase follows. They exist so that a reader can predict
where a thing lives and what it is called before looking. When a rule and convenience
disagree, the rule wins; when a rule is genuinely wrong, change the rule first, here,
then the code everywhere it applies.

The product spec is `../ADMIN_FRONTEND_PROMPT.md`. This file is only about how the code
is written.

---

## 1. Language

- **JavaScript (ES2022) + JSX. No TypeScript.** The skin, the dialog components and the
  tree code are ported from the prototype, which is JS; one language, no `.ts` islands.
- Types are carried by **JSDoc on the module boundary**: every exported function and every
  definition-object shape gets a `@param`/`@returns`/`@typedef`. Internals don't.
- `const` by default; `let` only when reassigned; never `var`.
- No classes. Function components, hooks, plain objects.

## 2. Directory structure (fixed — new code fits it, it does not grow new top levels)

```
src/
  app/            entry wiring only: routes, providers. No business code.
  api/            the ONLY place fetch/WebSocket appear.
    client.js       token store, request(), refresh-on-401 — uniform {ok,status,data}
    socket.js       one connection, event fan-out, start/stop market feed
    endpoints/      one file per backend area (groups.js, symbols.js, …), thin wrappers
  constants/      generated + hand-kept enum label maps. Mirrors hst-server/model names
                  exactly (OrderType_name style). Never hand-edit the generated file.
  components/
    ui/             dumb building blocks: SettingsDialog, InputDialog, PropSelect,
                    ContextMenu, Icon, DataTable, PropertyTable … No fetching, no routes.
    layout/         the shell: TerminalShell, NavTree, Toolbox, StatusBar, menus, toolbar.
    module/         the module framework: ModuleView, ListModule, dialogs host.
  modules/        one folder per business module (groups/, symbols/, datafeeds/ …):
                  its definition object, its dialog tabs, its columns. A module imports
                  from ui/, api/, constants/, lib/ — NEVER from another module.
  hooks/          shared hooks (useListData, useDialogDrag, useNav …).
  lib/            pure functions only (path trees, formatters, masks). No React,
                  no fetch, no window. If it needs a mock to test, it is not lib/.
  assets/         terminal.css, icons. The skin is extended, never restyled.
```

Import direction is one-way: `app → modules → components → hooks/lib/api/constants`.
Anything importing "upward" or sideways between modules is wrong by definition.

## 3. Naming

- Files: components `PascalCase.jsx`, everything else `camelCase.js`. One exported
  component per `.jsx` file, named the same as the file.
- Components are named for **what they show**, never for a vendor or a technology:
  `SettingsDialog`, not `Mt5ConfigDialog`; no `MT5`/`MetaTrader`/`MetaQuotes` anywhere —
  code, CSS, labels, comments.
- CSS classes: kebab-case, prefixed by their component family (`settings-dialog`,
  `nav-item`, `prop-select`). New classes go in `terminal.css` next to their family.
- Enum constants mirror the backend exactly: `OrderType_name[3] === "sell_limit"`.
  Humanising a label (`sell_limit` → `Sell Limit`) happens only in the view layer via
  `lib/labels.js`.
- Booleans read as assertions (`isFolder`, `canDeal`, `hasQuote`). Handlers are
  `onVerb`/`handleVerb`.

## 4. Data rules (the contract in one place)

- Volume is **lots** on every wire. Never rescale.
- Timestamps are **unix nanoseconds**: `ms = ns / 1e6`, formatted through `lib/time.js`
  only — no `new Date(raw)` on a raw API value anywhere else.
- Identity of groups and symbols is the **backslash path**. Escape as `\\` in literals,
  `encodeURIComponent` in URLs; folder queries use `?folder=<path>`.
- Enums cross the wire as ints; labels come from `constants/`. Zero is per-enum
  ("unset" vs "gtc" vs "all") — check the constant file's comment, never assume.
- Prices format with the record's own `digits`/`currency_digits`, never a fixed 2.
- Rights: gate every button and menu item on `can["right_*"]` from `/api/v1/navigation`.
  Never decode the packed `rights` array.

## 5. API layer rules

- Components never call `fetch`. They call `api/endpoints/*`, which call `client.js`.
- `request()` returns `{ok, status, data, message}` and **never throws**; callers branch
  on `ok`. A failed write shows the server's `message` to the user — no silent anything.
- **No mock fallback.** Demo mode, if ever built, is an explicit switch, not a rescue
  path for a failed live call.
- 401 → one refresh attempt → on failure clear the session and route to `/login`.
- The socket is one shared connection in `api/socket.js`; consumers subscribe to event
  types, never to the raw socket. `session.revoked` logs out. Nav-count refreshes are
  debounced (~2s), not per-event.

## 6. Module pattern (how every screen is built)

A module is **data first**: `modules/<name>/definition.js` exports
`{ id, title, endpoint, query, idKey, pathKey?, columns, dialog }`. `ModuleView` renders
any definition as list + dialog; module folders add only what is genuinely specific
(dialog tab panels, special columns). If building a module needs a new screen component
rather than a new definition, the framework is being bypassed — stop and fix the
framework.

Every module ships all four or it isn't merged: nav entry, list, settings dialog,
live refresh. Verification is the Playwright loop in the prompt file, per screen,
against the named `<DOCS>` reference images.

## 7. Consistency mechanics

- **The linter is the law**: `npm run lint` (oxlint + `scripts/rules-check.mjs`) must
  pass clean before any commit; the check script enforces the import direction and
  no-vendor-naming, so breaking rule 2 or 3 fails the build the moment it is written.
- Formatting is whatever the repo's Prettier config says; no hand-formatting debates.
- Comments: one line or none, saying **why**, not what. No banners, no commented-out code.
- No new dependencies without a reason a reviewer can read in the commit message.
  Current allowance: react, react-dom, react-router-dom. State lives in React state,
  context, and the URL — no state library until a concrete screen proves the need.
- Dev server runs on **port 5175** (5173/5174 belong to the trader terminal and the old
  prototype). API base defaults to `http://localhost:8080`.
- Commits follow the repo style: `type: what changed, in plain words` with a body that
  explains why; never any AI attribution.
