# quote-news review

Scope: hst-quote (app, config, handler, internal/{arbiter,configclient,filter,fixconfig,health,provider,session,status,tickcache,translate,worker}, model, repository, pkg/influxdb, README) and hst-news (app, config, handler, connector, internal/{configclient,health,newscache,normalize,status,worker}, model, README); cross-checked against hst-server `pkg/events/datafeed.go`, `internal/server/v1/datafeeds_worker.go`, `internal/workerstatus/subscriber.go`, `model/tick.go`, `internal/server/v1/market_ws.go` and hst-core `handler/nats.go`, `model/tick.go`, `model/subjects.go`. 45 files, ~6100 lines (+ shared pkg). `go build ./... && go vet ./...` clean in both modules.

## Summary
hst-quote is in reasonable shape: the hot path (connector → translate → arbiter → filter → spread → cache/influx/NATS) is single-goroutine per feed, locks are small, shutdown is ordered and idempotent. The real risks are at module boundaries: the tick subject breaks for any symbol containing `.` or a space (hst-core/hst-server subscribe `hstquote.tick.*`), the tick payload has drifted from what hst-server decodes (`volume` int vs float, no `digits`), and a feed whose connector died is never retried. hst-news is weaker: it subscribes to `hstserver.datafeed.*` while hst-server publishes `system.datafeeds.*`, so no config event ever reaches it, and its periodic reload ignores changed feeds — in practice a news feed edit never applies until the process restarts. Nothing consumes `hstnews.item.*` yet. Readability is good for a junior: files are short, names are plain; the things to simplify first are dead code (repository, worker pool, unused provider types) and the two stale READMEs.

## Findings
| # | Sev | Type | Where | What | Why it matters | Fix |
|---|-----|------|-------|------|----------------|-----|
| 1 | P1 | bug | hst-news/handler/nats.go:11-14 vs hst-server/pkg/events/datafeed.go:12-15 | hst-news subscribes `hstserver.datafeed.{created,updated,deleted}` / `hstserver.datafeed.config.>`; hst-server publishes `system.datafeeds.*` / `system.datafeeds.config.<id>` | No create/update/delete/snapshot ever reaches hst-news; only the 60s `reload` runs | Use the same constants as hst-quote/model/subjects.go (`system.datafeeds.*`) |
| 2 | P1 | bug | hst-news/handler/feeds.go:89-91 `if _, running := f.runners[id]; running { continue }` | `reload` never compares the running feed with the new config, so a changed `feed_server`, login, or `News Request Period` is ignored | Combined with #1, a news feed edit never applies until the process restarts or the feed is disabled and re-enabled | Keep the last `model.NewsFeed` per runner and restart when `!reflect.DeepEqual(old, new)` as hst-quote does (`quoteFeedNeedsRestart`) |
| 3 | P1 | bug | hst-quote/model/subjects.go:21 `fmt.Sprintf("hstquote.tick.%s", symbol)`; hst-core/model/subjects.go:146 and hst-server/model/tick.go:27 `hstquote.tick.*`; hst-server/internal/server/v1/admin/symbols.go:39 `validate:"required,max=64"` | A symbol with a dot (`EURUSD.m`) becomes a 4-token subject the `*` wildcard does not match; a space makes `Publish` fail with invalid subject | Those symbols are quoted, cached and stored but never reach the engine or the websocket; nothing warns | Publish on `hstquote.tick.<symbol_id>` (or escape `.`/space) and subscribe `hstquote.tick.>` on both sides; or restrict the symbol charset in hst-server |
| 4 | P1 | bug | hst-quote/handler/feeds.go:183-191 and :150-155 | When `conn.Run` returns an error (FIX cfg unreadable, `ParseSettings`/`NewInitiator` failure, `startInitiator` error) the goroutine exits but the entry stays in `f.runners`; `reload` then sees it as running and `continue`s | A dead feed stays dead until its config changes or the process restarts, while the admin sees it enabled; `status.Disconnected` is the only trace | On exit, remove the runner under `f.mu` (or mark it dead) so the next `reload` restarts it; add a backoff |
| 5 | P2 | bug | hst-quote/model/tick.go:17 `Volume float64 json:"volume,omitempty"`, no `digits`; hst-server/model/tick.go:20 `Volume int64`, :12 `Digits int32`; hst-server/internal/server/v1/market_ws.go:326-328 returns silently on decode error, :366-369 defaults digits to 5 | Payload drifted from the consumer: a fractional volume (FIX `MDEntrySize` 0.5, DDE `VolumeIndex`) fails `json.Unmarshal` into `int64` and the tick is dropped; `digits` is never sent so every JPY/metal/index line is formatted with 5 decimals | Wrong or missing prices on the trader websocket for some symbols | Add `Digits` (from `set.Digits`/`tr.Digits`) to hst-quote `model.Tick`, make hst-server `Volume` float64 or send `volume_real`; one shared wire struct in `pkg/` |
| 6 | P2 | bug | hst-news/handler/feeds.go:203-223 | After a poll that ingests items no status event is published (no `NewsDelta`, no `Connected`); `Connected` is only sent when `len(newItems)==0` | hst-server `news_count`/`bytes_received` never move (workerstatus/subscriber.go:177-187 only counts deltas) and a feed that errored once stays `down` until an empty poll | Publish `Event{NewsDelta: len(newItems), BytesReceivedDelta: bytes, Connected: true}` after a successful poll; use the `bytes` return of `Fetch` that is currently discarded (`raw, _, err`) |
| 7 | P2 | concurrency | hst-quote/handler/handler.go:113-123 and hst-news/handler/handler.go:97-107 `QueueSubscribe(..., GroupDatafeedConfig, ...)` | Config events and snapshots are queue-grouped, but feeds are sharded by `QUOTE_INSTANCE_INDEX/COUNT` (`ownsFeed`); with >1 instance the event is delivered to one instance that may not own the feed, which then only calls `status.Disconnected` | A disable/delete/param change takes up to `ReloadInterval` (60s) to apply on the owning pod; a disabled FIX feed keeps logging on to the LP meanwhile | Plain `Subscribe` for config subjects (every instance filters by `ownsFeed`), or put the instance index in the group name |
| 8 | P2 | bug | hst-quote/internal/translate/markup.go:24-25 `bid := round(raw.Bid+float64(tr.BidMarkup)*point, digits)` with hst-quote/internal/provider/fixquotes/connector.go:263 `if tick.bid > 0 \|\| tick.ask > 0` | A one-sided book (incremental before snapshot, or an LP that sends only BID) reaches translate with the other side 0; a non-zero markup turns 0 into `markup*point`, which then passes `filter.Apply` (filter.go:84 only rejects `<= 0`) | A fake ask/bid of a few points is cached, stored and published | Skip markup on a zero side and let the filter reject it, or require both sides in `ApplyMarkup` for non-exchange symbols |
| 9 | P2 | robustness | hst-quote/internal/provider/ddequotes/connector.go:107 `conn.Write([]byte(c.cfg.Token))` | No write deadline; the read deadline is set only afterwards (:118) | A stalled server socket blocks `session` forever; `Run` never reconnects and `Close` is the only exit | `conn.SetWriteDeadline(time.Now().Add(c.cfg.DialTimeout))` before the write |
| 10 | P2 | validation | hst-news/connector/rss.go:77 `io.ReadAll(resp.Body)` | Unbounded body read from an external URL the admin configures | A feed that returns hundreds of MB (or a misconfigured URL pointing at a binary) takes the pod down on memory | `io.LimitReader(resp.Body, maxFeedBytes)` and error past the cap |
| 11 | P2 | bug | hst-news/connector/rss.go:17 "RSS/Atom", :24-39 parses `rss/channel/item` only | An Atom feed (`<feed><entry>`) unmarshals to zero items with no error | The feed shows connected and quiet forever; admin has no signal it is the wrong format | Either parse `<entry>` too or return an error when `channel` is absent |
| 12 | P2 | security | hst-quote/internal/fixconfig/builder.go:48-65 generated cfg has no `SocketUseSSL`/TLS settings; hst-quote/internal/provider/fixquotes/connector.go:237-243 sets Username/Password on logon | FIX logon credentials go over plain TCP unless the LP terminates TLS elsewhere; no param exposes `SocketUseSSL`, `SocketCAFile` etc. | Credentials on the wire to a real LP | Pass through a `SocketUseSSL` (and cert) param set into the cfg; document it |
| 13 | P2 | bug | hst-quote/internal/fixconfig/settings.go:80-85 `HeartBtInt`, `ResetOnLogon`, `ReconnectInterval` copied raw from params | Not validated (`ResetOnLogon=yes`, `HeartBtInt=abc`) → quickfix `ParseSettings`/`NewInitiator` fails → runner dies (#4) and is never retried | One typo in the admin panel silently kills a feed | Validate as int / Y,N in `FromFeed` and return an error the journal shows |
| 14 | P3 | structure | hst-quote/repository/datafeed.go (whole file), hst-quote/config/config.go:24-32,127-137,208-216 Postgres config | Postgres repository and config exist but `app.go` never opens a DB; README says "no Postgres" | Dead code with SQL that includes `feed_password`; a junior will think the service reads the DB | Delete `repository/` and the Postgres/GRPC config blocks |
| 15 | P3 | structure | hst-quote/handler/handler.go:50,65,91 and hst-news/handler/handler.go:44,59,85 `worker.Pool` | Pool is started and stopped but nothing calls `Submit` (hst-news copy does not even have `Submit`) | NumCPU idle goroutines and a misleading "workers" log line | Remove the pool from both services until something needs it |
| 16 | P3 | structure | hst-quote/internal/provider/provider.go:30-45 `FixConfig`, `StreamConnector`; hst-quote/model/subjects.go:15 `SubjectSnapshot` | Unused types and constants | Two `StreamConnector` shapes (one here, one in handler/feeds.go:30) confuse where a new protocol plugs in | Delete the unused ones; keep `handler.streamInner` as the single contract |
| 17 | P3 | readability | hst-quote/handler/feeds.go:446-461 `isQuoteSessionOpen` | Builds a `[]session.Window` by scanning every session row of the feed on every tick | O(sessions) alloc per tick on the hot path for a check that is per symbol | Index sessions by `SymbolID` once in `configclient.toQuoteFeed` (map[int64][]Window) |
| 18 | P3 | robustness | hst-quote/internal/tickcache/cache.go:49-53 two sequential `Set`; hst-quote/handler/feeds.go:442 `f.status.Tick` NATS publish per tick | Three network round trips plus two NATS publishes per accepted tick, all synchronous in the runner goroutine; channel is 256 deep (feeds.go:264) so a slow Redis drops ticks ("tick channel full") | Throughput ceiling per feed is Redis latency bound | Pipeline the two `Set`s; batch the status delta (hst-server already aggregates in memory) |
| 19 | P3 | readability | hst-quote/handler/feeds.go:46-60 `stopRunnerLocked` waits up to 10s (+5s in `stopInitiator`) holding `f.mu` | NATS callbacks (`onDatafeedEvent`, `onConfigSnapshot`) and the periodic reload serialize behind it | A slow LP logout stalls every other feed's config change | Collect runners to stop under the lock, wait outside it |
| 20 | P3 | convention | hst-quote/internal/provider/fixquotes/connector.go:205 `time.Sleep(500 * time.Millisecond)` inside `OnLogon` | Blocks the quickfix session goroutine; magic number without a named constant | A junior will not know why it is there (likely "let the LP settle") | Name it and comment the reason, or move subscription to a goroutine |
| 21 | P3 | bug | hst-quote/internal/arbiter/arbiter.go:122-126 `Forget` only drops the local map | When a feed is stopped on purpose (disable, reload) its Redis claims live until `DatafeedsTimeout` | The symbol is silent for up to 10s before the next feed is accepted even though the owner is known to be gone | Track owned symbols per feed (already in `active`) and `DEL` the keys whose value is this owner in `Forget` |
| 22 | P3 | bug | hst-quote/handler/feeds.go:476-494 with hst-server/internal/server/v1/admin/reference.go:187 `NotifySystem(events.SubjectDatafeedUpdated, body)` | The feed-reorder event body is `{datafeed_ids:[...]}`; it decodes to `DatafeedEvent{DatafeedID:0}` → `!HasQuoteFlag()` → `status.Disconnected(0)` published | Noise status row for feed 0 in hst-server; priority change itself only lands on the 60s reload (README admits this) | Publish one `DatafeedEvent` per reordered id (or a snapshot) from hst-server; ignore `DatafeedID==0` in hst-quote |
| 23 | P3 | readability | hst-quote/internal/filter/filter.go:99-102 | A tick equal to the last one in bid/ask/volume and in the same minute is dropped | Time is receive time (connector.go:286), so two identical LP quotes 30s apart are dropped; MT5 counts them; is this intended? | One-line comment stating the rule, or drop only exact-time duplicates |
| 24 | P3 | readability | hst-news/internal/normalize/normalize.go:38 `time.ParseDuration(raw + "s")` | `"1m"` becomes `"1ms"` and silently falls back to 300s | Surprising; `strconv.Atoi` is what the README documents (seconds) | `strconv.Atoi` and log the fallback |
| 25 | P3 | robustness | hst-news/handler/feeds.go:208-220 `MarkSeen` before `PrependFeed`/`publishItem`; hst-news/internal/newscache/cache.go:44-54 one `EXISTS` per item | Items are marked seen even when publish fails, so they are never re-sent; N round trips per poll | Lost news on a NATS hiccup; slow polls on a 100-item feed | Publish first, then `MarkSeen`; use `MGET`/pipeline |
| 26 | P3 | structure | hst-news/connector/connector.go:39 `return &RSS{}, true` and rss.go:47-50 new `http.Client` per call | A new client per poll: no keep-alive, no shared timeouts | Extra TLS handshakes every poll for every feed | One `RSS{Client: ...}` built in `newFeeds` |
| 27 | P3 | convention | hst-quote/README.md:13,209 (`hstserver.datafeed.*`), :139 ("ask_markup Subtracted" — markup.go:25 adds), :154 ("always writes Redis cache and Influx" — feeds.go:418-422 skips both when the session is closed), :221 (`INFLUX_ENABLED` does not exist, config.go:281 requires the token), Redis `hstquote:last:` undocumented; hst-news/README.md:11,42-44 wrong subjects, :15-17 unclosed code fence | READMEs contradict the code in the places a new developer reads first | Onboarding cost | Fix the four statements and the fence |
| 28 | P3 | readability | hst-quote/handler/feeds.go:461 `time.Now().UTC()` in session check; hst-quote/internal/session/session.go:15-17 | Quote sessions are compared against UTC wall clock; hst-core/hst-server use server time with a configurable offset | Sessions open/close at the wrong minute if the platform time zone is not UTC | Take the platform time zone from config or the snapshot |
| 29 | P3 | readability | hst-quote/internal/provider/fixquotes/connector.go:237-243 `msg.Header.Set(field.NewUsername(...))` | FIX 553/554 are body fields of Logon; set on the header here | Works with lenient LPs; a strict one rejects the logon and the reason is not obvious | `msg.Body.Set(...)` |

## Flow traces

**Config load (boot) and periodic reload**
1. hst-quote/app/app.go:127-131 `handler.New` → `h.Start` → hst-quote/handler/handler.go:60 `h.load` → hst-quote/handler/feeds.go:109 `reload`.
2. feeds.go:110 `configclient.ListQuoteFeeds` → GET `<HST_SERVER_URL>/internal/v1/datafeeds?mode=1` with `X-Service-Token` (client.go:103-110,139) → hst-server routes.go:65 group `/internal/v1` behind `ServiceAuth` (service_auth.go:14-28, constant-time compare) → datafeeds_worker.go:498-514 `ListInternalDatafeeds` → `listWorkerDatafeedConfigs` → `[]WorkerDatafeedConfig` → envelope `{success,data}` decoded at client.go:158-170.
3. feeds.go:115-134 filter enabled/owned/supported/has-translates → lock → stop unwanted (:139-147) → start or hot-swap wanted (:149-160); hot swap when `quoteFeedNeedsRestart` false (:194-208) via `atomic.Pointer` (:98-107).
4. handler.go:128-148 ticks every `QUOTE_RELOAD_INTERVAL` (60s) → `reload(context.Background())`.
Verdict: OK, with gap at step 3 for a dead runner (#4). hst-news same shape (hst-news/handler/feeds.go:55-101) — gap at step 3: running feeds never compared (#2).

**NATS config events and snapshot**
1. hst-server datafeeds.go:885-890 `publishDatafeedEvent` → `events.PublishDatafeed` (datafeed.go:97-115) on `system.datafeeds.{created,updated,deleted}` payload `{datafeed_id,mode,enable}`; `publishWorkerConfigSnapshot` (datafeeds_worker.go:394-407) on `system.datafeeds.config.<id>` payload `WorkerDatafeedConfig`.
2. hst-quote handler.go:113-123 queue-subscribes the same four subjects (model/subjects.go:10-13 match) → feeds.go:476-500 `onDatafeedEvent` (stop or `reloadOne` → GET `/internal/v1/datafeeds/<id>`, 404 → nil → stop) and :502-536 `onConfigSnapshot` (`configclient.FromSnapshot`, hot swap or restart). Field names of `Snapshot` (client.go:24-39) match `WorkerDatafeedConfig` (events/datafeed.go:74-89).
3. hst-news handler.go:97-107 subscribes `hstserver.datafeed.*` — never published.
Verdict: hst-quote OK (gaps #7 queue group when sharded, #22 reorder body); hst-news gap at step 3 (#1).

**FIX session build and reconnect**
1. feeds.go:269-290 `fixconfig.FromFeed` (settings.go:37-102: module defaults, `feed_server` host:port, params, CompIDs required) → `ResolveConfigPath` (builder.go:18-29: `FixConfigPath` param, `.cfg` feed_server, else `WriteConfig` to `FIX_CONFIG_DIR/feed_<id>/session.cfg`, :32-73).
2. `fixquotes.NewConnector` (connector.go:55-76) routes 4.3/4.4 snapshot+incremental → `streamRunner.Run` (feeds.go:345-375) starts `inner.Run` → `startInitiator` (connector.go:116-162: parse cfg, file store/log, `quickfix.NewInitiator`, `Start`).
3. quickfix reconnects itself on `ReconnectInterval` (cfg line builder.go:52 from `timeout_reconnect`/param); `OnLogon` (connector.go:192-215) → `OnSession(true)` → `status.Connected` + journal, `logons++`, subscribe every translate source (sender.go:23-44, MarketDepth/MDUpdateType from settings); `OnLogout` → `Disconnected`.
4. `streamRunner.Run` resets the filter state when `Logons()` changes (feeds.go:368-371).
5. Stop: `stopRunnerLocked` cancel → `Close` (connector.go:107-114 shutdown chan + `stopInitiator` with 5s cap, lock released before `initiator.Stop` to avoid the OnLogout deadlock) → wait `done` ≤10s.
Verdict: OK for reconnect of a live session; gap: a failed `startInitiator` is final (#4, #13); no TLS (#12).

**DDE auth, parse, idle timeouts**
1. feeds.go:291-306 `ddequotes.FromFeed` (config.go:27-101: host:port required, `Token` param or `generateToken(login,password)` with `)`,`_`,`G` escaped, defaults dial 5s / first frame 30s / idle 90s / reconnect 5s).
2. `Connector.Run` (connector.go:50-71) loops `session` with `ReconnectInt` backoff until ctx/shutdown.
3. `session` (:88-149): dial with ctx and timeout → write token (no deadline, #9) → scanner with 64KB/1MB cap and `!` split (:168-176) → first-frame read deadline → on first frame `logons++`, `notify(true)` → per frame reset idle deadline → `parseFrame` (parser.go:33-75: header check, 16-field chunks, arity error drops the frame, bad numbers become 0 and are rejected later by filter.go:84) → `emitTick` non-blocking (drop + warn on full channel).
4. On scanner end `notify(false, detail)`; detail distinguishes "no data after auth" from "connection lost".
Verdict: OK except #9; malformed frames are dropped per frame, never panic (slice bounds guarded by the arity check at parser.go:40).

**Translate / markup / spread**
1. feeds.go:379-383 `translate.ApplyMarkup` (markup.go:11-43): first translate whose `ExternalSymbol()` equals the source; `digits` default 5; `bid+bid_markup*point`, `ask+ask_markup*point`, rounded.
2. feeds.go:399-427: realtime flag, quote session, `filter.Apply` (filter.go:80-138: zero side reject for non-exchange, min/max spread for floating-spread OTC, same-minute duplicate drop, soft/hard/discard channel with break handling, gap flag).
3. feeds.go:429 `translate.ApplySpread` (spread.go:10-39): skip DOM; `Spread==0` → shift both by balance; spread already equal → rebase ask; else centre on mid. `PointValue()` (model/datafeed.go:110-117) never 0, so no division by zero.
Verdict: OK; gap at step 1 for a zero side (#8); README sign text wrong (#27).

**Arbiter: per-symbol priority and active-source switch**
1. feeds.go:387 `arbiter.Claim(ctx, datafeedID, feed.Datafeed.FeedIndex, symbolID)` reads the live `FeedIndex` from the atomic pointer, so a reorder applies on the next reload without restart.
2. arbiter.go:39-62 Lua on `hstquote:src:<symbol_id>` value `<feed_index>:<datafeed_id>`, TTL `QUOTE_DATAFEEDS_TIMEOUT`: owner refresh (1), free claim (2), strictly higher priority or lower id on tie takes over (2), else ignored (0); atomic across instances.
3. arbiter.go:84-96 Redis error → accept all, log once a minute; :98-118 local `active` map turns result 2 into `activated` only on a transition → feeds.go:391-397 `MarkBreak` + journal "<SYMBOL> activation".
4. Silent owner: key expires after timeout → the next lower-priority tick claims (2) → activation. `Forget` on stop drops only the local view (#21).
Verdict: OK; matches README §Feed priority.

**Tick cache / Influx / NATS publish → hst-core**
1. feeds.go:431-434 `tickcache.Put` → `hstquote:tick:<feed>:<symbol>` 5m TTL and `hstquote:last:<symbol>` no TTL (cache.go:43-54).
2. feeds.go:436-438 `influx.WriteTick` → async batched `WriteAPI`, measurement `<symbol_id>`, fields bid/ask/high/low/open/close/volume, time = tick time (pkg/influxdb/influxdb.go:65-111); discards go to `raw_<symbol_id>` only when `TickFlagCollectRaw` (feeds.go:401-405). Retention: `EnsureRetention` on boot (app.go:92-98) keeps the raw bucket at `INFLUX_TICK_RETENTION_DAYS` (7d) and the candle bucket forever, installs/updates the 1m rollup task (retention.go:30-143); failure is logged, not fatal.
3. feeds.go:463-474 `json.Marshal(model.Tick)` → `Publish("hstquote.tick.<symbol>")`; feeds.go:440-443 `status.Tick` → `hstquote.status.<id>` (hst-server workerstatus/subscriber.go:165-175 aggregates deltas in memory, flushes every 5s).
4. hst-core handler/nats.go:18 subscribes `SubjectSystemMarketFeedAll` = `hstquote.tick.*` (model/subjects.go:20,146) → `MarketSystemEventHandler` (:271-286) decodes `model.Tick` (model/tick.go:10-26: `datafeed_id, symbol_id, symbol, source, digits, bid, ask, last, high, low, open, close, volume, gap, time`) — field names match hst-quote's; `time` RFC3339 parsed by the custom `UnmarshalJSON` (:29-64); `Ok()` requires both sides >0; `digits`/`last` arrive zero. hst-server market_ws.go:324-328 decodes its own `model.Tick` with `Volume int64`.
Verdict: gap at step 3/4 for symbols with `.` or space (#3) and payload drift (#5); otherwise OK.

**RSS poll → normalise → dedup → publish**
1. hst-news/handler/feeds.go:157-178 `runFeed`: `normalize.PollInterval` (default 300s), poll once then on ticker.
2. :180-192 `connector.ForModule` → `RSS.Fetch` (rss.go:41-109): GET `feed_server`, basic auth from `Feed login`/`feed_login` + `feed_password`, 2xx only, XML → `RawItem` with `ExternalID` = guid|link|title.
3. :194-195 `normalize.Items` (normalize.go:46-86): drop empty, category `<prefix>\<category>`, language, `pubDate` layouts with receive-time fallback, id = sha256(ext,subject,link,pubdate,feed)[:32].
4. :196-216 `FilterNew` (EXISTS per item) → `MarkSeen` (7d) → `PrependFeed` (`hstnews:feed:<id>` JSON, 100 items, 24h).
5. :218-220 `publishItem` → `hstnews.item.<datafeed_id>`; no consumer found in hst-server or hst-core.
Verdict: gap at step 2 (#10, #11), step 4/5 ordering (#25), step 5 status (#6); subscriber unknown (open question).

**Health / readiness**
1. app.go:101-126 (quote) / :76-88 (news): `/healthz` always ok; `/readyz` 503 until `Ready()` after `h.Start`, 503 while `Draining()`, then runs redis/nats/(influx) checks with a 2s cap (health.go:111-136).
2. On SIGTERM: `Draining()` → sleep `HEALTH_DRAIN_WAIT` (5s) → deferred stops.
Verdict: OK. Both services are outbound-only, so readiness mostly gates rollouts.

**Shutdown ordering**
1. app.go:142-163 signal → `Draining` → sleep → return → defers in reverse: `h.Stop` → `probes.Stop` → influx `Close` (flush) → redis `Close` → nats `Close` (drain) → log sync.
2. handler.go:78-106 `Stop` (once): unsubscribe config subjects → cancel ctx → `Feeds.stopAll` (each runner cancel → `Close` → wait ≤10s, status Disconnected) → `Workers.Stop` → wait all `h.Go` goroutines ≤15s.
3. `Start` order: load → workers → subscribe → reload loop, matching the convention; no `os.Exit` inside handler.
Verdict: OK; note a runner that overruns the 15s cap may publish on a drained NATS connection (logged only).

## Readability notes for onboarding
- hst-quote/app/app.go — boot/shutdown wiring; clear; drop Postgres/GRPC config noise.
- hst-quote/config/config.go — env → struct with collected errors; clear; remove unused Postgres/GRPC blocks.
- hst-quote/handler/handler.go — start/stop/subscribe skeleton; clear; worker pool is unused, delete.
- hst-quote/handler/nats.go — subject aliases; clear.
- hst-quote/handler/feeds.go — runner lifecycle + tick pipeline in one 536-line file; readable but `handleRawTick` mixes gating, filtering, storage and publishing; split `newConnector` and the tick pipeline into their own files, fix the dead-runner case first.
- hst-quote/internal/arbiter/arbiter.go — Lua claim + local transition map; clear; the script comment says what each return means, good.
- hst-quote/internal/configclient/client.go — HTTP client + snapshot → model mapping; clear; `FromSnapshot` is just `toQuoteFeed`.
- hst-quote/internal/filter/filter.go — MT5 soft/hard/discard/gap channel; dense but commented; the same-minute duplicate rule needs one line of why.
- hst-quote/internal/fixconfig/{builder,settings}.go — params → quickfix cfg; clear; validate the copied-through values.
- hst-quote/internal/health/health.go — probe server; clear.
- hst-quote/internal/provider/provider.go — types; two unused, delete.
- hst-quote/internal/provider/ddequotes/* — small, clear; add write deadline.
- hst-quote/internal/provider/fixquotes/connector.go — quickfix app; 4.3/4.4 handlers are copy-pasted four times; one generic helper over the entry getters would halve it.
- hst-quote/internal/provider/fixquotes/sender.go — clear.
- hst-quote/internal/session/session.go — clear; time zone question.
- hst-quote/internal/status/publisher.go — clear.
- hst-quote/internal/tickcache/cache.go — clear; pipeline the two sets.
- hst-quote/internal/translate/{markup,spread}.go — clear; zero-side guard.
- hst-quote/internal/worker/worker.go — good code, unused here.
- hst-quote/model/* — clear; `Tick` needs `digits`.
- hst-quote/repository/datafeed.go — dead; delete.
- hst-quote/README.md — useful but four statements are wrong (#27).
- hst-news/app/app.go, config/config.go, handler/handler.go — same skeleton as quote; clear.
- hst-news/handler/nats.go — wrong subjects; should import one shared constant set.
- hst-news/handler/feeds.go — clear; needs change detection and status deltas.
- hst-news/connector/{connector,rss}.go — clear; size cap, Atom, shared client.
- hst-news/internal/configclient/client.go — copy of quote's minus translates; fine.
- hst-news/internal/newscache/cache.go — clear; ordering and pipelining.
- hst-news/internal/normalize/normalize.go — clear; `ParseDuration` trick.
- hst-news/internal/status/publisher.go — clear; never called with a delta.
- hst-news/model/* — clear.
- hst-news/README.md — wrong subjects, broken fence.

## Open questions
- Who consumes `hstnews.item.<id>`? Nothing in hst-server or hst-core subscribes; is news delivery to the terminal planned elsewhere?
- Is the same-minute duplicate drop in filter.go:99-102 intended MT5 behaviour, or should only exact duplicates be dropped?
- Markup sign: README says ask markup is subtracted, code adds both; which is the MT5 semantics you want?
- Should quote sessions use the platform time zone rather than UTC (session.go:15-17)?
- Do any target LPs require TLS on the FIX socket? The generated cfg cannot express it.
- Is `hstquote:last:<symbol>` (no TTL) read by hst-core on boot as the comment claims (cache.go:40-42)? No reader was found in this pass.
