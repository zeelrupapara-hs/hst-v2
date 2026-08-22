# server-trading-api review

Scope: hst-server/internal/server/v1: orders.go, orders_http.go, orders_ws.go, positions.go, positions_http.go, positions_ws.go, positions_net.go, dealing.go, dealing_http.go, dealing_ws.go, balance.go, balance_http.go, balance_ws.go, deals.go, history.go, history_http.go, history_positions.go, exposure.go, live.go, quotes.go, market_ws.go, routes_ws.go, ws.go, routes.go, paging.go; pkg/ws/{hub,client,ctx}.go; model/{order,position,deal,balance,query,events,subjects}.go; cross-checked against hst-core/handler/{nats,dealing,accounts,positions,publish}.go and hst-core/model/{events,balance,query}.go, admin/routes.go, trader/routes.go, middleware/authorization.go, utils/group_access.go, pkg/influxdb/reader.go. 35 files, ~8,100 lines. `go build ./...` and `go vet` on the package are clean.

## Summary
The trading surface is thin and consistent: every handler does shape validation only, binds tickets to the login in SQL before publishing (orders, positions), and hands off to the engine on the right subject with the right queue group; payload shapes match hst-core field for field. The two real risks are in the dealing path and in the synchronous `request` callers: (1) dealer and requote handlers check group reach on a caller-supplied `login`, but hst-core resolves the account from `request_id` alone and ignores that login, so the reach check binds nothing and a trader can accept a requote on another account; (2) `FixPosition`, `DeletePosition`, `FixBalance`, `BulkBalance` and `CreateBalanceWS` discard the engine retcode and report success on refusal. Paging has a `limit=0` hole and the closed-positions tab bypasses the group history limit. The WS hub is sound (non-blocking send, bounded egress, single close), but the market feed keys watchers by session so a second tab silences the first. Readability for a junior is good: files are short, sibling handlers look alike, and one `answer`/`accepted` pair carries the journal; the things that will trip a newcomer are the duplicated `volume * 10000` literal, the `_ = c.BodyParser` one-offs, the HTTP-vs-WS rights mismatch, and the package-level globals in market_ws.go.

## Findings
| # | Sev | Type | Where | What | Why it matters | Fix |
|---|---|---|---|---|---|---|
| 1 | P1 | security | hst-server/internal/server/v1/dealing_http.go:37,69,101,133,215,270; dealing_ws.go:21,43,65,87; hst-core/handler/dealing.go:139-145,299-310 | Reach is checked on `body.Login` (`s.inReach(c, snap, body.Login)`) but the engine finds the request by id only (`p := h.takeRequest(ev.RequestId)` / `h.dealing[h.dealingKey(ev.RequestId)]`) and overwrites `res.Login = p.Request.Login`; `ev.Login` is never compared to the owner. | A dealer names any in-reach login and acts on a request of an account outside their masks; a trader calling `AcceptMyRequote` (dealing_http.go:240 sets `Login: snap.Login`) with another account's `request_id` accepts that account's requote. Request ids are wall-clock strings (see #7), so guessable. | Bind id to login in the server (`SELECT 1 FROM hst.orders WHERE request id/order_id = $1 AND login = $2 AND state IN (7,8,9)`) before sending, and have hst-core refuse when `ev.Login != p.Request.Login`. |
| 2 | P1 | security | hst-server/internal/server/v1/dealing_http.go:234-243 | `AcceptMyRequote` has no `isReadOnlyScope(snap.Scope)` guard; every other trader write (orders_http.go:210, positions_http.go:209,278,345) has one. | An investor-password session can complete a trade by accepting a requote. | Add the read-only check like its siblings. |
| 3 | P1 | bug | hst-server/internal/server/v1/positions_http.go:507-516,569-585; balance.go:215-226 | `if _, _, err := s.request(...)` (FixPosition, FixBalance ×2) and `res, _, err := s.request(...)` (DeletePosition) drop the status; `request` returns `(&res, 400, nil)` on `RetCode != 0` (orders.go:347-348). | A refused fix/delete is journaled as "position fixed"/"position deleted"/"balance fixed" and answered 200; DeletePosition returns the refusal body with 200. | Route these through `s.answer(c, res, status, err)` or check `res.RetCode` before logging/journaling. |
| 4 | P1 | bug | hst-server/internal/server/v1/balance.go:350-360; balance_ws.go:24-29 | BulkBalance sets `row.Ok = true` whenever `err == nil`; `makeBalance` returns `err == nil` with `res.RetCode != 0` on refusal. CreateBalanceWS sends `WSResponseOK(c.Type, res)` for the same case. | Insufficient-funds withdrawals and other engine refusals are reported as done to the accountant; bulk journal counts them as `done`. | Treat `res != nil && res.RetCode != 0` as refused (`row.Message = res.Message`); WS should send a bad_request event carrying `res`. |
| 5 | P1 | bug | hst-server/internal/server/v1/positions_http.go:511; model/events.go:143; hst-core/handler/positions.go:814-815 | Fix sends `Volume: check.validVolumeExt / model.ExtPerUnit` (integer division to legacy units) while `TradeRequest` documents "Volume is in extended units"; core's FixPosition reads it as legacy (`p.VolumeExt = model.FromLegacy(req.Volume)`). | Any position whose deals sum to an ext volume not divisible by 10000 is truncated on fix and never matches the check again (`v.volumeExt == v.validVolumeExt` stays false); the contract comment misleads the next caller. | Send ext units and have core's fix take ext (`p.VolumeExt = req.Volume; p.Volume = req.Volume/ExtPerUnit`), or document the legacy unit on the fix verb. |
| 6 | P2 | robustness | hst-server/internal/server/v1/orders.go:318 | `s.Nats.NC.Flush()` on every trade publish uses the nats.go default 60 s timeout and ignores the request context. | A NATS stall holds the HTTP/WS handler goroutine for up to a minute per request. | `FlushWithContext(ctx)` with `engineTimeout`, or `FlushTimeout`. |
| 7 | P2 | bug | hst-server/internal/server/v1/orders.go:389-391 | `newRequestId` is `time.Now().Format("20060102150405.000000000")`. | Not unique across hst-server pods or two requests in one nanosecond; hst-core keys its dealing map by this id (hst-core/handler/dealing.go:587-620), so a collision overwrites a queued request; it is also the guessable handle behind #1. | Random id (uuid) or pod prefix + counter. |
| 8 | P2 | validation | hst-server/internal/server/v1/positions.go:138-141 | `if closing <= 0 \|\| closing > volume { closing = volume }` clamps an over-volume partial close to a full close. | A client asking to close 0.5 of a 0.3 lot position closes everything silently; MT5 refuses with invalid volume. | Return 400 when `closing > volume`; keep zero meaning "all". |
| 9 | P2 | robustness | hst-server/internal/server/v1/positions_http.go:716-741; balance.go:346-362 | BulkClose has no row cap and does publish+Flush per row inside one request; BulkBalance runs up to 500 sequential `request` calls each with a 10 s timeout. | One request can run for minutes (bulk close over a large mask) or ~80 minutes worst case (bulk balance), inside a Fiber handler. | Cap rows / bounded concurrency, or hand the batch to a worker and return 202 with a job id. |
| 10 | P2 | bug | hst-server/internal/server/v1/market_ws.go:296-297,317-321; ws.go:71 | `feed.clients[c.Client.SessionId] = ...` keys watchers by session, while the hub allows several sockets per session (pkg/ws/hub.go:118). `dropMarketWatcher(client.SessionId)` runs when any socket of the session closes. | Two tabs on one session: the second start_market_feed replaces the first; closing either tab stops prices for the other. | Key by `c.Client.Id`. |
| 11 | P2 | robustness | hst-server/pkg/ws/client.go:113-124 | `dropped` is cumulative for the socket's life; `n >= maxDrops` (64) closes it. | A long-lived terminal that had a few brief stalls is disconnected hours later on its 64th drop, even if it is keeping up now. | Reset the counter on a successful send, or count drops in a sliding window. |
| 12 | P2 | validation | hst-server/internal/server/v1/paging.go:18-21,30 | `limit=0` passes the `< 0 \|\| > 500` check and `tail` treats 0 as "no LIMIT"; `offset: page * limit` has no cap on `page`. | A trader can dump their whole deal/order/closed-position history with `?limit=0`; a huge `page` overflows to a negative OFFSET and a 500. | Clamp `limit < 1` to `def` when `def > 0`; cap `page`. |
| 13 | P2 | validation | hst-server/internal/server/v1/history_positions.go:68 vs deals.go:60-72 | `GetMyClosedPositions` uses `readPage(c, 100)`; `GetMyDeals` uses `historyPage` which clamps `from` to the group's `limit_history`. | The closed-positions tab shows history the group setting is meant to hide. | Use `s.historyPage(c, 100)` there too. |
| 14 | P2 | robustness | hst-server/internal/server/v1/history.go:45-46,100-101,123-125; pkg/influxdb/reader.go:81-92 | `maxBars` is applied after `s.History.Candles(ctx, symbolId, res, q.From, q.To)` returns; the query range itself is unbounded. | A 1-minute request over years pulls the whole series from Influx before truncation; the comment at history.go:45 claims the opposite. | Clamp `from` to `to - maxBars*resolution` (or to `countback`) before querying. |
| 15 | P2 | robustness | hst-server/internal/server/v1/live.go:93-103; positions_http.go:49 | `overlayLiveAll` issues one `askEngine` (2 s timeout) per distinct login, sequentially, for the whole positions list. | A manager's GET /positions over N logins costs up to 2·N seconds when the engine is slow. | One batched query to the engine, or bounded parallelism with a total deadline. |
| 16 | P2 | convention | hst-server/internal/server/v1/orders_ws.go:62,105,147; positions_ws.go:14,57,100; dealing_ws.go:12,34,56,78 vs admin/routes.go:149-151,162-164,229-234 | WS manager routes require one right (`wsDealer(c, model.MgrRightTradesManager)` / `MgrRightTradesDealer`); the HTTP twins require `AccRead + TradesRead + TradesManager` (Authorization is all-of, internal/middleware/authorization.go:26-30). | The same action has two different rights rules depending on transport; a manager refused over HTTP can do it over the socket. | One helper naming the right set per action, used by both. |
| 17 | P2 | convention | hst-server/internal/server/v1/positions_http.go:687-690; admin/routes.go:159,163 | BulkClose is gated by `MgrRightTradesDealer` and by terminal type (`terminalOf(snap.ConnectionType) == TerminalAdmin`), while single close needs `MgrRightTradesManager`. | A dealer without TradesManager can close every position under a mask but not one; an admin-terminal manager with every right is refused by connection type, which is not a right. | Gate bulk on the same rights as the single-row verb it fans out to. |
| 18 | P2 | robustness | hst-server/internal/server/v1/orders_ws.go:48; market_ws.go:403; ws.go:102,469-476; balance_ws.go:24; pkg/ws/client.go:210 | WS handlers run inline in `readPump` and use `context.Background()` for DB and the 10 s engine request. | A slow query or engine stalls that socket's inbound frames (and its pong handling) for the duration; no per-handler deadline. | `context.WithTimeout(queryTimeout)` in WS helpers; consider dispatching handlers off the read goroutine. |
| 19 | P3 | convention | hst-server/internal/server/v1/balance_http.go:35-39,163-167 | `if reach, err := s.inReach(...); err != nil {500} else if !reach {403 ErrUnauthorizedToAccessResource}` — `inReach` already wrote 400/403/500 and returns `(false, <write error>)`. | The refusal is written twice with two different messages; siblings use `if ok, err := s.inReach(...); !ok { return err }`. | Use the sibling form. |
| 20 | P3 | convention | hst-server/internal/server/v1/balance_http.go:46-48,174-176 | `s.journalBalance(c, &body)` then `s.answer(...)` which calls `journalTrade`. | Two journal lines per balance operation; the other trading routes journal once in `answer`/`accepted`. | Drop `journalBalance`. |
| 21 | P3 | convention | hst-server/internal/server/v1/orders_http.go:349; positions_http.go:283 | `_ = c.BodyParser(&body)` while every sibling returns 400 on a parse error. | A malformed body silently cancels/closes with defaults (comment lost, volume 0 = close all). | Return 400 like the siblings, or comment why the body is optional. |
| 22 | P3 | readability | hst-server/internal/server/v1/balance_ws.go:17 | `s.wsFail(c, 400, err)` | Magic number; siblings use `nethttp.StatusBadRequest`. | Use the constant. |
| 23 | P3 | readability | hst-server/internal/server/v1/positions.go:216; positions_http.go:385-390,625,631; exposure.go:83; history_positions.go:33 | `GREATEST(volume_ext, volume * 10000)` / `volume * 10000` repeated in six queries. | `model.ExtPerUnit` exists (model/order.go:57); a change to the unit misses one. | One SQL fragment constant, or a view/generated column. |
| 24 | P3 | convention | hst-server/internal/server/v1/orders_ws.go:94-98 vs 136-140,178-183; positions_ws.go; dealing_ws.go; balance_ws.go | `CreateMyOrderWS` journals success and failure; the other trader WS writes journal only success; no manager/dealer WS handler calls `JournalWS` at all (HTTP twins journal through `accepted`/`answer`). | Dealer confirms/rejects and manager trades over the socket leave no audit line. | Journal in `wsFail` and after each manager WS call, as `accepted` does for HTTP. |
| 25 | P3 | readability | hst-server/internal/server/v1/orders_http.go:11-13 | `liveStates = "o.state IN (0, 1, 3, 7, 8, 9)"` | Duplicates `model.OrderState.IsLive` (model/order.go:161) as magic numbers. | Build the list from the named constants. |
| 26 | P3 | readability | hst-server/model/events.go:102-106; model/balance.go:39,47-66 | `PositionEvent_value` lacks fix/delete; `IsBalanceAction` casts `DealAction(action)` on a `DealAction`; `DealActionName` returns "balance" for `so_compensation_credit`. | Name maps that disagree with their enum mislead logs and journals. | Derive `_value` with `valuesOf`; add the missing case. |
| 27 | P3 | readability | hst-server/internal/server/v1/ws.go:185-187,359 | Doc comment "eventFromMsg turns a nats message..." sits above `scrubGroupForTrader`; "the event type comes from a header" sits above `ViewUserRef`. | Stale comments cost a newcomer a minute each. | Move them. |
| 28 | P3 | readability | hst-server/internal/server/v1/history.go:111 | `if first.After(q.To) \|\| first.After(q.From)` | `To > From` is enforced at :86, so the first test is redundant. | Keep `first.After(q.From)`. |
| 29 | P3 | structure | hst-server/internal/server/v1/market_ws.go:36,53,131 | `feed`, `spreads`, `symLive` are package globals. | Hidden state on a type that otherwise hangs everything off `*HttpServer`; untestable in isolation. | Fields on `HttpServer`. |
| 30 | P3 | bug | hst-server/internal/server/v1/market_ws.go:286-297 | `grantedSymbols` is snapshotted at start_market_feed; only `spreads` reload on config events (routes_ws.go:55). | A symbol added to or removed from the group is not reflected on the stream until the terminal restarts the feed. | Recompute grants on `system.group_symbols.>` / `system.groups.>` as the spread book does. |
| 31 | P3 | bug | hst-server/pkg/ws/hub.go:188-193 | `CloseSession` queues `session.revoked` and calls `Remove` immediately. | `close()` closes the socket before `writePump` drains, so the revoke event usually never reaches the client. | Write synchronously with a deadline, or delay Remove until the egress is flushed. |
| 32 | P3 | readability | hst-server/internal/server/v1/quotes.go:27-35,83-91 | `wireTick` omits `digits` and `last`; `TickLine` (market_ws.go:366-368) falls back to 5 digits for a thrown quote. | A thrown quote on a 2- or 3-digit symbol renders with 5 decimals on the terminal. | Fill `Digits` (already read at :68) and `Last`. |

## Flow traces

### 1. Place order, trader HTTP
1. trader/routes.go:37 `POST /api/trader/v1/orders` → Protect + RequireTrader.
2. orders_http.go:205 `CreateMyOrder`: read-only scope check :210, body parse, `crtFromMy` fills login from session.
3. orders.go:153 `makeOrder`: struct validate, enum membership :158, pending needs price :162, expiry rule :166, lots→ext :175.
4. orders.go:275 `sendOrder` → :306 `publish(system.orders)`: Publish + Flush, returns 202.
5. hst-core/handler/nats.go:54 QueueSubscribe `system.orders` / `engine_orders` → :284 `OrderSystemEventHandler` → `forwarded` to `system.owner.<shard>.orders` when not held → `NewOrder`.
6. nats.go:345 `reply`: no Reply subject on a publish, so only `PublishRejected` (publish.go:63) on refusal → `websocket.accounts.<login>.orders` event `order_rejected`; a fill arrives as `order_create`/`position_create` from the order path.
7. ws.go:101 trader socket subscribed to `SubjectTrader(login)` = `websocket.accounts.<login>.>` → :135 `eventFromMsg` → `scrubGroupForTrader` → `c.Send`.
8. orders.go:431 `accepted` journals "trade request on account #n: ...".
Verdict: OK. Gaps: Flush has no deadline (#6); request id not unique (#7).

### 2. Modify / cancel order, HTTP and WS
1. orders_http.go:271/339 or orders_ws.go:126/168 → `updateOrder`/`cancelOrder` (orders.go:194/243).
2. orders.go:355 `orderState` binds `order_id AND login` in SQL → 404 on `ErrNoRows`; state/type checks :215-220.
3. Same publish path as flow 1; engine `UpdateOrder`/`CancelOrder`.
Verdict: OK (IDOR closed by the SQL bind). Manager WS right set differs from HTTP (#16).

### 3. Position update / close / close-by
1. positions_http.go:204/273/340, positions_ws.go:35/78/121 → positions.go:94/124/161.
2. positions.go:214 `positionState` binds `position_id AND login`; close volume clamp :138 (#8).
3. positions.go:195 `sendPosition` → `system.positions` → hst-core nats.go:318 `PositionSystemEventHandler` → Update/Close/CloseBy → reply/PublishRejected → `websocket.accounts.<login>.positions`.
Verdict: OK; gap at step 2 for over-volume close (#8).

### 4. Position fix / delete, balance fix (sync)
1. positions_http.go:479/547, balance.go:190 → `checkOnePosition`/`checkOne` → `inReach`.
2. orders.go:326 `request`: `RequestWithContext` with 10 s; `RetCode != 0` → `(res, 400, nil)`.
3. hst-core nats.go:339-342 fix/delete → positions.go:794/859; accounts.go:14 balance → `NewBalance` → `msg.Respond`.
4. Server ignores status/retcode (#3), logs and journals success.
Verdict: gap at step 4 (#3); volume unit mismatch at step 3 (#5).

### 5. Dealer actions
1. admin/routes.go:229-234 (three rights) or routes_ws.go:32-35 → dealing_http.go / dealing_ws.go.
2. `inReach(body.Login)` / `wsInReach(payload.Login)`; `body.RequestId = c.Params("request_id", ...)`.
3. dealing.go:52 `sendDealing` → `request(system.dealing)` → hst-core dealing.go:55 → `takeRequest(ev.RequestId)` / `h.dealing[dealingKey]`, login not compared (#1) → Respond.
4. orders.go:470 `answer`: refusal body with 400, journal via `journalTrade`.
Verdict: gap at step 3 (#1); trader accept lacks read-only guard (#2).

### 6. Balance / credit operations
1. admin/routes.go:238-245 (Accountant) → balance_http.go → `balanceAs` → `inReach` (double write, #19) → balance.go:45 `makeBalance` (action membership, non-zero amount, `AllowNegative` only for correction) → `request(system.balance)`.
2. hst-core accounts.go:14 → `NewBalance` → Respond.
3. `answer` → journal; plus `journalBalance` (#20). Bulk (#4, #9); WS (#4).
Verdict: OK for single ops; gaps in bulk/WS result mapping.

### 7. History paging
1. deals.go:169 `GetMyDeals` → `historyPage` (floor from `hst.groups.limit_history`) → `readDeals` → `p.bound("d.time")` + `p.tail("d.deal_id")`.
2. orders_http.go:75 `GetMyOrders` → `orderPage` (no limit for active, 100 for history).
3. history_positions.go:62 closed positions: `readPage` only (#13).
4. paging.go:18 `limit=0` → unbounded (#12).
Verdict: gap at steps 3-4.

### 8. Market feed over WS
1. routes_ws.go:41 one `Subscribe(hstquote.tick.*)` → market_ws.go:324 `MarketFeedHandler`.
2. routes_ws.go:13 `start_market_feed` → market_ws.go:285: `grantedSymbols` (trader only), `groupOf`, insert into `feed.clients[SessionId]` (#10).
3. Handler: `AlertsOnTick`, `symLive.Touch`, per-group spread line via `spreads.forGroup`, `client.Send` binary under `feed.mu.RLock`; `Send` is non-blocking (egress 256, drop counted).
4. ws.go:71 `dropMarketWatcher(SessionId)` on socket exit.
Verdict: OK for one socket per session; gap at step 2/4 (#10); grants not refreshed (#30).

### 9. Account event fan-out (engine → WS)
1. ws.go:89 `subjectsFor`: session, broadcast, holiday; trader: `websocket.accounts.<login>.>`, symbols, own group + group symbols; manager: journal, dealer queue when TradesDealer, right-gated broker subjects, `model.Subscriptions(rights, masks)` (reduced masks, subjects.go:247).
2. ws.go:135 one NATS subscription per subject per socket → `eventFromMsg` → `Send`; unsubscribed in `Client.close` (pkg/ws/client.go:131).
3. ws.go:158 `RefreshLogin` rebuilds on rights/group change; old subs closed after new opened.
Verdict: OK.

## Readability notes for onboarding
- orders.go: request types, validation and the `publish`/`request`/`answer`/`accepted` core; clear; first simplify: one `requestId` helper that is random, and a one-line note that `request` returns 400+res on refusal so callers must check.
- orders_http.go: read + write handlers; clear; replace `liveStates` numbers with constants, drop the `_ = c.BodyParser`.
- orders_ws.go: WS twins + `wsFail/wsDealer/wsWritable/wsInReach`; clear; align rights and journaling with HTTP.
- positions.go: verbs + `positionState`; clear; refuse over-volume close.
- positions_http.go (766 lines): reads, writes, check/fix/delete, bulk; the check/fix/bulk half deserves its own file; check retcodes.
- positions_ws.go, positions_net.go: clear.
- dealing.go / dealing_http.go / dealing_ws.go: clear shape; needs the request_id→login bind in one helper used by all seven handlers.
- balance.go / balance_http.go / balance_ws.go: clear; collapse the duplicate reach/journal writes; treat retcode as refusal in bulk/WS.
- deals.go: reads plus raw deal edit/delete; clear; note that edits publish nothing.
- history.go / history_http.go: clear; clamp the range before the query.
- history_positions.go: one SQL fold; clear; use `historyPage`.
- exposure.go: clear; `coverage%` group convention is a hidden rule worth a constant and a line in the docs.
- live.go: clear; batch `overlayLiveAll`.
- quotes.go: clear; add digits/last to the tick.
- market_ws.go (430 lines): feed, spread book, liveness and the handler in one file with three globals; split liveness out, move globals onto `HttpServer`.
- routes_ws.go, routes.go: clear.
- ws.go: subscription rules, journal helpers, notify helpers and View*Ref types mixed; move the View*Ref types and Notify* out; fix the two misplaced comments.
- pkg/ws: hub/client/ctx are small and well commented; reset the drop counter; make CloseSession deliver the revoke.
- model/*: enums with `_name`/`_value`; derive `PositionEvent_value` with `valuesOf`; fix `DealActionName`.
- paging.go: clear; clamp `limit=0` and `page`.

## Open questions
- Should a trader's accept of a requote be allowed at all over HTTP for a read-only (investor) session? Assumed no (#2).
- Is a thrown quote (quotes.go:91, `datafeed_id: 0`, `source: <symbol>`) accepted by hst-core's tick path / hst-quote's per-symbol priority arbitration, or dropped as a lower-priority source?
- Does the engine validate `ExpiryAt` in the past and `PriceTrigger` for stop-limit types, or should the server refuse them in `makeOrder` (orders.go:162-168)?
- Are multiple sockets per session (two tabs) a supported case? If yes, #10 is a bug; if not, the hub should refuse the second upgrade.
- Should `UpdateDeal`/`DeleteDeal` (deals.go:292/358) publish a `deal_update` WS event so open blotters and the trader's history refresh?
