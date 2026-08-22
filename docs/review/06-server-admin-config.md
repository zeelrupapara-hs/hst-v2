# server-admin-config review

Scope: hst-server/internal/server/v1/admin/{symbols,symbol_folders,symbol_currency,group_symbols,groups,leverages,commissions,holidays,routing,routing_dealers,reference,end_of_day,market_watch,routes}.go, model/{symbol,group,commission,routing,leverage,holiday}.go, pkg/symbolpath/match.go; cross-read utils/group_access.go, utils/db_errors.go, migrations (groups, symbols, commissions, holidays, routing, leverages, trading, symbol_folders), hst-core/handler/nats.go, hst-core/handler/load.go, hst-core/internal/settings/settings.go. 21 files, ~10,700 lines.

## Summary
Transaction discipline is good almost everywhere: every multi-statement write uses begin/defer rollback/commit, and `NotifySystem` is called after commit on a `system.<table>.<verb>` subject that hst-core subscribes to with `system.<table>.>` (nats.go:22-36). The three biggest risks are (1) a group rename that leaves every account and manager mask pointing at the old path, so hst-core's `Store.For` cannot find the group and the accounts stop trading; (2) single-row group/group-symbol/commission handlers that skip the manager group-mask check the list/create handlers apply, so a manager can read and edit groups outside their masks by id; (3) `CloneSymbol` matching nothing for folder clones because `LIKE 'x\%'` escapes the `%`, and publishing no event when it does copy. `UptSymbol` has no validation tags at all, so PATCH accepts what POST refuses. Readability: the files are long but linear; a junior would get lost in `UpdateSymbol` (124 positional params) and `symbols.go` (1812 lines) first.

## Findings
| # | Sev | Type | Where | What | Why it matters | Fix |
|---|---|---|---|---|---|---|
| 1 | P1 | bug | groups.go:679, 698 | `"group" = COALESCE($41, "group")` renames the group row only; `hst.users."group"` (text, no FK), `hst.managers.groups` masks and open positions keep the old path. | hst-core keys groups by path (`settings.go:164 s.groups[groupPath]`), so every account in the renamed group fails `Store.For` and cannot trade, and managers lose access. | Either refuse rename when users exist, or cascade `UPDATE hst.users SET "group"=...` and mask rewrite in one tx and publish `system.accounts.updated` per login. |
| 2 | P1 | security | groups.go:354, 611, 639, 737; group_symbols.go:334, 568, 708; commissions.go:214, 370, 458 | GetGroup/UpdateGroup/DeleteGroup/CreateGroup and Get/Update/Delete of group symbols and commissions query by id without `utils.GroupAccessFor`; only ListGroups (309), ListGroupSymbols/CreateGroupSymbol, ListGroupCommissions/CreateGroupCommission call `groupExists`. | A manager with mask `demo\*` can read, edit, delete any group or its overrides by guessing the id; `CreateGroup` also lets them create a group under a path they do not own. | Call `groupExists(c, s, id)` (or an equivalent predicate) in every group-scoped handler before the write; in CreateGroup check `model.MasksCover(snap.ManagerGroups, []string{path})`. |
| 3 | P1 | bug | reference.go:77 | `where, args = "path LIKE $1", []any{body.Path + `\%`}` | Postgres LIKE treats `\` as the escape char, so `Forex\%` matches the literal string `Forex%`; a folder clone copies nothing (symbol_folders.go:103-105 documents this very trap). | Use `starts_with(path, $1 || E'\\')` as `sqlPathUnderFolder` does. |
| 4 | P1 | bug | reference.go:100-130 | CloneSymbol inserts rows and returns; no `NotifySystem(SubjectSystemSymbolCreated)`, no `NotifyWS`, no `JournalEntry`; `cloneColumns` (36-45) omits sessions, volume_*_ext, margin_initial_buy/sell..., margin_rate_*, filter_*, ie_*, re_*, price_* so the copy falls back to column defaults. | hst-core does not learn the new symbols until an unrelated config event; the clone trades with different margin/volume rules than its source, which is the opposite of what "clone" promises. | Copy every column except ids/names and sessions in the same tx, then publish one `system.symbols.created` per row (or one event) and journal. |
| 5 | P1 | bug | symbol_folders.go:262-268, 283-297 | RenameSymbolFolder rewrites `hst.symbols.path` for the subtree but publishes nothing on `system.symbols.>`. | Group overrides match on path prefix (`settings.go:199 matches`); after a rename the engine keeps the old paths in memory and group rules (`Forex\*`) stop or start matching wrongly until another reload. | Publish `SubjectSystemSymbolUpdated` after commit (and a WS event for the panel). |
| 6 | P1 | validation | symbols.go:164-286, 1397 | `UptSymbol` has no `validate` tags; `s.Validate.Struct(body)` checks nothing. Only digits/volume_step/volume_min_max are hand-checked (1405-1413). | PATCH accepts `trade_mode: 99`, `swap_mode: -1`, negative `volume_min`, `currency_base` > 16 chars (DB error → 500), sessions with `type: 7, day: 9` (validateSessions only checks open/close). | Mirror CrtSymbol tags with `omitempty` (`validate:"omitempty,gte=0,lte=4"` etc.) and `dive` on Sessions. |
| 7 | P1 | bug | symbols.go:1792-1793; symbol_folders.go:353-354 | `DELETE FROM hst.symbols WHERE symbol_id = $1` with no check for open positions/orders (positions.symbol is a bare VARCHAR(32), trading.up.sql:60). | The engine drops the instrument from `Settings` while positions remain; margin, stop-out and close on those positions fail. | Refuse (409) when `hst.positions`/`hst.orders` reference the symbol name, or require an explicit force flag. |
| 8 | P1 | security | reference.go:324-372 | ResetUserPassword writes the new hash and returns; no `s.OAuth2.InvalidateLogin` although the doc comment says "A new master password ends the account's sessions". | House rule: any password change invalidates sessions; a stolen session survives a staff reset. | `InvalidateLogin(ctx, login, ...)` after the UPDATE when `body.Kind == "main"` (as users.go:360 does). |
| 9 | P2 | bug | groups.go:609-619 | `if err := ...Scan(&call,&stop); err == nil { validate }` | A DB error or missing row silently skips the margin-call/stop-out sanity check; the later UPDATE then stores an inverted pair. | Return 500/404 on the error instead of swallowing it. |
| 10 | P2 | bug | groups.go:699-704 | UPDATE with rename maps only `ErrNoRows`; a `groups_group_uidx` violation or a bad `margin_leverage_id` FK returns 500. | Wrong status for a user error. | Add `utils.IsUniqueViolation` → 409 and `IsForeignKeyViolation` → 400/404 (CreateGroup:544 already maps unique). |
| 11 | P2 | validation | groups.go:21-131 | No range/enum checks on `AuthPasswordMin`, `CurrencyDigits`, `DemoLeverage`, `Limit*`, `TradeInterestRate`, `MarginCall/StopOut` upper bound, and the enum pointers (`AuthMode`, `ReportsMode`, `MarginSOMode`...) accept any int. | `limit_orders: -5`, `auth_password_min: 0`, `margin_so_mode: 9` are stored and hst-core reads them as-is. | Add `validate:"omitempty,gte=0,..."` and `oneof` tags. |
| 12 | P2 | validation | commissions.go:20-29, 44, 60, 277, 370 | `CrtCommissionTier` Mode/Type/Value/RangeFrom/RangeTo unchecked; header mode pointers unchecked; Update does not trim/refuse empty `name` (Create does at 264) so `name: ""` hits `commissions_name_set` CHECK → 500. | Tier `range_to < range_from` or `mode: 42` reaches the engine's commission math. | Validate tiers (enum, `range_from <= range_to`, value >= 0) and trim name on update. |
| 13 | P2 | validation | group_symbols.go:94-99, 370, 486 | Path mask is stored untrimmed/unchecked; `config_index` duplicates allowed (no unique index); no `volume_min <= volume_max` check; `IsUniqueViolation` branch at 509 is dead code. | Duplicate `config_index` makes "first match wins" (`settings.go:192`) depend on physical row order. | Validate the mask (symbolpath style), enforce a unique `(group_id, config_index)` or compute next index in a tx. |
| 14 | P2 | bug | routing.go:141-145, 330-340, 606-628 | CreateRouting honours any `routing_index` (gap beyond max+1) and `moveRoutingIndex` to `> max`; `moveRoutingOneStep` then looks up `routing_index = target` and returns 500 "neighbor not found" when the list is not dense. | Move up/down breaks after one out-of-range insert. | Clamp `routing_index` to `[0, count]` on create/update, or compute neighbour as the next row by `ORDER BY routing_index`. |
| 15 | P2 | validation | routing.go:34-57, 745-778 | `Action` accepts any int (only known actions get value checks); `Request`/`Type`/`Flags` bitmasks unchecked; condition `Condition` id unchecked beyond rule pairing. | An unknown action makes the engine's `canExecute`/switch fall through silently. | `oneof` over `model.RouteAction_name` keys; mask-domain checks like `checkFlagDomains`. |
| 16 | P2 | bug | end_of_day.go:48-53 | `QueryRow(... WHERE key = $1).Scan` returns 500 on `pgx.ErrNoRows`. | Fresh DB with no `end_of_day_at` row makes GET 500. | Map `ErrNoRows` to the default hour (core's `LoadSystemConfig` has one). |
| 17 | P2 | bug | reference.go:284-297 | UpdateTimeSettings runs up to three upserts outside a tx and publishes no event. | A failure after the first upsert leaves a half-applied clock; nothing tells hst-core (open question whether core reads `time_zone`). | One tx; publish a system subject if any consumer exists. |
| 18 | P2 | bug | reference.go:165-177 | ReorderDatafeed parks all `feed_index` negative then writes only the ids in the body; ids missing from the body stay negative, and ids not in the table are ignored. | A partial list silently corrupts feed priority. | Require `len(ids) == count(*)` and all owned, as ReorderRouting (497-513) does. |
| 19 | P2 | bug | symbols.go:39; trading.up.sql:14,60 | `Symbol validate:"max=64"` and `symbols.symbol VARCHAR(64)`, but `orders.symbol`/`positions.symbol` are `VARCHAR(32)`. | A 33-64 char symbol can be created and quoted but every order insert fails. | Cap symbol at 32 (or widen the trading columns). |
| 20 | P2 | bug | symbol_folders.go:270-281 | The two `UPDATE hst.symbol_folders` execs are not checked for unique violation; only `tx.Commit` is (284). | Renaming onto a path whose child folder already exists returns 500 instead of 409. | Check `IsUniqueViolation` on each Exec. |
| 21 | P2 | bug | symbols.go:961-962, 1765, 1808; reference.go:100 | `NotifyDatafeedsForSymbolID` runs only on UpdateSymbol; create, delete, clone and folder rename/cascade skip it. | Feed symbol selection by path mask (datafeeds_worker.go:69) is stale for new or removed symbols. | Call it (or a bulk variant) on every catalog change. |
| 22 | P2 | bug | symbols.go:468-481, 903-921 | Zero means "use default": `digits 0 → 5`, `trade_mode 0 (disabled) → 4 (full)`, `exec_mode 0 (request) → 2`; and the volume_max/min check is duplicated (903 and 919). | A symbol cannot be created disabled, with 0 digits, or with request execution; callers get full trading on what they asked to be disabled. | Use pointer fields for these or an explicit "defaults" marker; drop the duplicate check. |
| 23 | P2 | validation | holidays.go:27, 952; holidays.up.sql:19 | `Year validate:"gte=0,lte=9999"` but the CHECK is `year = 0 OR year BETWEEN 1970 AND 9999`. | `year: 1900` passes validation and returns 500 from the constraint. | `validate:"eq=0|gte=1970,lte=9999"` or check in `validateHoliday`. |
| 24 | P2 | bug | symbol_currency.go:111-131, 635 | On create `applyDerivedCurrencies` overwrites the caller's `currency_base/profit/margin/digits` whenever calc_mode is forex/cfd and the name parses as a pair; on patch (135-185) the derived values only fill nil fields. | POST and PATCH disagree: the same explicit `currency_margin: "USD"` is kept on patch and discarded on create, with no error. | Apply derivation only to fields the caller left empty, on both paths. |
| 25 | P3 | convention | routing.go:183, 389, 454, 536, 659; routing_dealers.go:137, 172, 295 | No `NotifyWS` and no `JournalEntry` on any routing or dealer change; ReorderRouting publishes `ViewRoutingRef{}` (id 0). | Every other config area journals and pushes to the panel; routing edits are invisible in the journal and the panel does not live-update. | Add `s.NotifyWS(model.SubjectRouting, ...)` and `s.JournalEntry(..., JournalType_routing, ...)`. |
| 26 | P3 | readability | symbols.go:1385-1771 | UpdateSymbol is ~390 lines with 124 positional `$n` args and `body.Digits` passed four times (1727-1730). | Adding a column means counting placeholders by hand; an off-by-one silently writes the wrong column. | Build the SET list from a `[]struct{col string; val any}` table shared with `symbolInsertArgs`. |
| 27 | P3 | structure | symbols.go (1812 lines), model/symbol.go (1243 lines) | Request types, column lists, session helpers, diff logging and five handlers in one file; 150-line industry enum table in the model. | Over the ~800 line bar; a new developer cannot find where a symbol field is added (4 places: CrtSymbol, UptSymbol, columns, args). | Split: symbols_types.go, symbols_sessions.go, symbols_handlers.go; move industry table to its own file. |
| 28 | P3 | readability | model/group.go:373-374, 279-297 vs 558-564 | `Root bool db:"root"` and `ParentID db:"parent_id"` map columns that do not exist in groups.up.sql; `MarginFreeProfitMode` and `FreeMarginProfitMode` are two enums for one column. | Stale fields mislead; two names for the same thing. | Delete `Root/ParentID`; keep one enum. |
| 29 | P3 | bug | model/symbol.go:480-484 | `case industry >= 84 && industry <= 95` then `case industry >= 95 && industry <= 119`; 95 belongs to the first. | Industry 95 maps to sector 7, likely meant 8. | Make the ranges disjoint (`<= 94`). |
| 30 | P3 | convention | model/holiday.go:63, 67; holidays.go:22 | Truncated comments: `the platform's "leave zero values in these.`, `which is every mask the terminal`, `the table is tiny and edits`. | Unfinished sentences read as bugs. | Finish or delete them. |
| 31 | P3 | readability | group_symbols.go:519 | `GroupPath := s.GroupPath(c, groupID)` — a capitalised local shadowing the method name. | Reads like a type or exported symbol. | `path := s.GroupPath(...)` as at 677, 722. |
| 32 | P3 | readability | leverages.go:319-320 | Journal message `LeverageUpdatedMsg(v1.PtrOr(body.Name, ""))` prints an empty name when the patch does not rename. | Journal line says "leverage profile '' updated". | Use the loaded profile name from `getLeverageProfile`. |
| 33 | P3 | convention | market_watch.go:77-125; end_of_day.go:113 | UpdateDefaultMarketWatch has no `Log.Log` actor line and no journal; RunEndOfDay journals but never logs the actor. | Audited config action without actor in the server log. | Add the `Log.Log(..., "actor", snap.Login)` line. |
| 34 | P3 | structure | routes.go:206-215 | Four system routes repeat `s.Middleware.Protect, s.Middleware.RequireManager` inline instead of a `system` group. | Easy to forget one middleware on the next system route. | `system := v1.Group("/system", Protect, RequireManager)`. |
| 35 | P3 | readability | groups.go:785-816 | `grantCreatorAccess` writes `path\*` and refreshes the session, swallowing every error into a warn log. | Fine as intent, but a junior will not expect a group POST to mutate `hst.managers`. | Mention it in the handler doc comment; keep. |

## Flow traces

**Symbol create/update/delete → engine**
1. routes.go:94,103,104 `MgrRightCfgSymbols` → symbols.go:893 CreateSymbol / 1385 UpdateSymbol / 1785 DeleteSymbol.
2. Validation: create 897-921 (tags + sessions + path length); update 1394-1413 (no tags, see #6); delete none.
3. tx: create 923-954 (insert + sessions, commit); update 1415-1753 (FOR UPDATE, update, sessions, commit); delete single statement, sessions cascade (symbols.up.sql:188).
4. Event after commit: 962 `system.symbols.created`, 1767 `.updated`, 1808 `.deleted` → hst-core nats.go:24 `system.symbols.>` → ConfigSystemEventHandler nats.go:138 → LoadSettings load.go:71 → `Store.Load` settings.go:103 → ResettleAccounts.
Verdict: OK for the three handlers; gap at CloneSymbol (#3, #4) and RenameSymbolFolder (#5) which change the catalog without an event; DeleteSymbol with open positions (#7).

**Group create/update/delete → engine**
1. routes.go:125-128 `MgrRightCfgGroups` → groups.go:448 / 589 / 729.
2. Scope: ListGroups 309 only (#2).
3. tx: create 492-559 (group + default `*` override); update single statement; delete single statement after children/users checks (747-763).
4. Event: 569 `system.groups.created`, 710 `.updated`, 779 `.deleted` → nats.go:23 → LoadSettings (groups keyed by path, load.go:97).
Verdict: gap at step 3 on rename (#1) and step 2 (#2).

**Group symbol / commission → engine**
1. routes.go:131-142 (`MgrRightCfgGroups` / `MgrRightGroupCommission`) → group_symbols.go:411/542/698, commissions.go:245/344/448.
2. Scope via `groupExists` only on list/create (#2).
3. tx: commission create/update 270-309, 363-415; group symbol single statement.
4. Event: `system.group_symbols.*` 521/679/725, `system.commissions.*` 323/429/475 → nats.go:24,26 → LoadSettings + LoadCommission.
Verdict: OK at steps 3-4; gap at 2.

**Leverage profile → engine**
1. routes.go:56-67 `MgrRightCfgGroups` → leverages.go:113/256/336/380/461/572/641.
2. Validation: validateRules/validateTiers 858-899; rule/tier ownership checked in tx (495-501, 591-597, 682-701).
3. tx with `FOR UPDATE` on the parent for append/reorder (407, 673); delete compacts index (603-606).
4. Event after commit via `NotifyLeverage` 902-913 on `system.leverages.*` → nats.go:28 → LoadSettings → loadLeverageProfiles load.go:228.
Verdict: OK.

**Holiday → engine**
1. routes.go:82-89 `MgrRightCfgHolidays` → holidays.go:93/249/361/427.
2. Validation tags + `validateHoliday` 641-657 (#23).
3. tx with advisory lock 114, 375, 453; compaction on delete 390-392.
4. Event `system.holidays.*` 147/341/406/499 → nats.go:36 `system.holidays.>` → CalendarSystemEventHandler nats.go:124 → LoadCalendar load.go:294.
Verdict: OK.

**Routing rule / dealers → engine**
1. routes.go:108-120 `MgrRightCfgRequests` → routing.go:111/276/406/472/580; routing_dealers.go:93/153/216.
2. Validation: action/condition pairs 745-810 (#15); dealer must hold trades_read+dealer rights 1030-1051.
3. tx for rule writes; index parking 434-444, 515-518, 632-645; dealer insert single statement with MAX+1 subselect (unique index catches races → 409).
4. Event `system.routing.*` 183/389/454/536/659 and dealers 137/172/295 → nats.go:27 `SubjectSystemRules = system.routing.>` → LoadRoutingRules load.go:350 (rules, conds, dealers by dealer_index).
Verdict: OK at 3-4; gap at 2 (#15) and the move-step with gaps (#14); no journal/WS (#25).

**End of day / market watch / time**
1. routes.go:200-215 `MgrRightCfgTime` / `MgrRightCfgSymbols`.
2. UpdateEndOfDay end_of_day.go:70-102 upsert then `system.core.endofday.time` → nats.go:48 EndOfDayTimeSystemEventHandler; RunEndOfDay 114 `system.core.endofday` → nats.go:44. OK.
3. UpdateDefaultMarketWatch market_watch.go:77-125 upsert, no event (trader API reads on demand). OK.
4. UpdateTimeSettings reference.go:266-303 three upserts, no tx, no event (#17).
Verdict: OK except #16, #17.

## Readability notes for onboarding
- symbols.go — symbol CRUD + sessions + change-diff logging; clear per function but 1812 lines and a 124-arg UPDATE; split first, then table-drive the column list.
- symbol_folders.go — folder tree over `symbols.path` + `symbol_folders`; clear; the `sqlPathUnderFolder` comment is the one to read first.
- symbol_currency.go — forex/cfd currency auto-fill; small and clear; name why create overwrites and patch fills (#24).
- group_symbols.go — override rows with use_default_* clusters; the 73-param UPDATE is readable because each cluster is labelled; add `groupExists` everywhere.
- groups.go — group CRUD + tree; `ViewGroupTree` and `grantCreatorAccess` need their comments; rename cascade missing.
- leverages.go — profile/rule/tier tree; best-structured file here (validate → lock → write → notify via one responder).
- commissions.go — header + tiers; clear; tier validation missing.
- holidays.go — calendar + resolver; clear; advisory-lock comment truncated.
- routing.go — rule CRUD + three ordering strategies (shift, move, swap, reorder); pick one parking scheme and one helper.
- routing_dealers.go — dealer list per rule; clear; no journal.
- reference.go — a grab bag (clone, datafeed reorder, time, password reset); move each to its own area.
- end_of_day.go, market_watch.go — tiny settings handlers; clear.
- routes.go — flat and readable; group the /system routes.
- model/symbol.go — enums + struct; split the 150-line industry table out; fix the 95 overlap.
- model/group.go — enums + struct; drop `Root/ParentID`, one free-profit enum.
- model/commission.go, model/routing.go, model/leverage.go, model/holiday.go — enum tables; fine; finish the two truncated comments.
- pkg/symbolpath/match.go — glob → regexp helper used only by datafeeds; clear; note it differs from hst-core's prefix-only `matches` (settings.go:199) and holiday `matchMask`, so three mask dialects exist.

## Open questions
- Does hst-core (or hst-quote) read `time_zone`/`time_dst` from `hst.settings`? If yes, UpdateTimeSettings needs an event; if no, the setting is decorative.
- Is leverage tier `margin_rate_initial: 0` (no margin) intentional, or should tiers require `> 0`?
- Should `CreateGroup` by a manager without the `*` mask be restricted to paths under their own masks (MT5 behaviour), or is "create anything, then get granted" the intended model (groups.go:565)?
- Three mask dialects (symbolpath glob, core prefix-star, holiday prefix-star) — is a group-symbol path like `Forex\*\Majors` expected to work anywhere?
