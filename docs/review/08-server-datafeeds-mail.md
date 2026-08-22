# server-datafeeds-mail review

Scope: hst-server/internal/server/v1/{datafeeds.go, datafeeds_worker.go, mails.go, mail_macros.go, mail_recipients.go, news.go}, admin/{mail_servers.go, mail_templates.go}, internal/workerstatus/subscriber.go, pkg/{datafeedmodules, events, mailer}, model/{datafeed.go, mailserver.go, mailtemplate.go}, templates/; plus callers read for the traces (admin/routes.go, admin/reference.go ReorderDatafeed, paging.go, trader/register.go, trader/recover.go, middleware/service_auth.go, migrations/1785412449359_datafeeds.up.sql, hst-quote/internal/configclient, hst-news/internal/configclient, hst-quote/internal/status/publisher.go). 19 files, ~6,100 lines (assigned) + ~900 lines of callers. Uncommitted diff on model/mailtemplate.go reviewed.

## Summary
Datafeed CRUD is solid on auth (every route behind `MgrRightCfgDatafeeds`), id handling and 404 mapping, and the worker snapshot shape matches what hst-quote/hst-news decode byte for byte. Two real bugs: `workerstatus.Stop` deadlocks on itself (shutdown hangs, pending counters lost), and param delete compacts priorities with a single `priority = priority - 1` under a non-deferrable unique index, which fails intermittently. Enum fields (`enable`, `type`, `mode` on some paths, timeouts) are not range-checked. Mail: the HTML path is properly sanitised (bluemonday on store, macros escaped) and SMTP headers are safe (Q-encoding, `net/smtp` line validation); gaps are the full request body written into the journal on every broadcast, drafts stored unsanitised, no SMTP I/O timeout in the drainer, and `limit=0` meaning "no limit" on mails/news paging. Readability: datafeeds.go (1,845 lines) and mails.go (1,145 lines, trader mailbox + manager broadcast + attachments in one file) are the two files a junior will struggle with; the rest is clear.

## Findings
| # | Sev | Type | Where | What | Why it matters | Fix |
|---|---|---|---|---|---|---|
| 1 | P1 | concurrency | hst-server/internal/workerstatus/subscriber.go:130-139 | `Stop()` takes `s.mu.Lock(); defer s.mu.Unlock()` then calls `s.flushAll(...)`, which does `s.mu.Lock()` again at :371. | `sync.Mutex` is not reentrant: server shutdown blocks forever in `WorkerStatus.Stop()` (server.go:118) and the final counter flush never runs. | Unsubscribe under the lock, release it, then call `flushAll`. |
| 2 | P1 | bug | hst-server/internal/server/v1/datafeeds.go:1279-1281 | `UPDATE hst.datafeed_params SET priority = priority - 1 WHERE datafeed_id = $1 AND priority > $2` with unique index `datafeed_params_feed_priority_uidx (datafeed_id, priority)` (migration :46) that is not DEFERRABLE. | Postgres checks the unique index per row; if the row at priority 3 is visited before priority 2 the update raises duplicate key and the delete 500s, intermittently, depending on heap order. | Park first (`SET priority = -priority - 1 WHERE priority > $2`) then `SET priority = -priority - 2`, or make the index DEFERRABLE INITIALLY DEFERRED. |
| 3 | P2 | validation | hst-server/internal/server/v1/datafeeds.go:596,605,996,1113,601-604 | `Enable`, `AllowImportSymbols`, param `Type`, `Timeout*`/`AttemptsSleep` are written as received: `ptrOr(body.Enable, ...)`, `ptrOr(body.Type, model.FeederParamType_string)`, no range check. | `enable=7` is stored, shows as enabled-ish in the UI but never matches `WHERE enable = 1` (datafeeds_worker.go:312) so the feed silently never runs; `type=99` is meaningless to workers; negative timeouts reach hst-quote. | Check `enable IN (0,1)`, `type` in `FeederParamType_name`, timeouts `>= 0`, `allow_import_symbols IN (0,1)` before the write. |
| 4 | P2 | bug | hst-server/internal/server/v1/datafeeds.go:592-593, 988-993 | `feed_index` and param `priority` are assigned with `(SELECT COALESCE(MAX(...)+1,0))` outside any lock; both columns are unique. | Two concurrent creates collide; the caller gets 409 `ErrAlreadyExists`, which reads as "name taken" and is wrong. | Retry once on unique violation of the index column, or lock the parent row (`SELECT ... FOR UPDATE` on hst.datafeeds) inside a tx. |
| 5 | P2 | security | hst-server/internal/server/v1/mails.go:751-752, 924-925 | `s.JournalEntry(..., journal.MailBroadcastMsg(...), in)` stores the whole `BodySendMail` (64 KB body, To expression, logins) as journal detail. | Convention: never log request bodies; the journal becomes a copy of every broadcast, including attachments ids and mail text. | Pass a small struct (subject, recipient count, queued, internal/email flags). |
| 6 | P2 | security | hst-server/internal/server/v1/mails.go:544 | `UpdateMyDraft` stores `in.Body` raw: `v.Subject, v.Body, v.UpdatedAt = in.Subject, in.Body, ...`, while `SendMyMail` runs `mailer.SanitizeHTML` (:429). | Draft bodies bypass the one sanitiser the terminal relies on (pkg/mailer/sanitize.go:9-11 says the policy, not the editor, is the boundary). Today only the author reads a draft, but any future "send draft" path would ship raw HTML. | `Body: mailer.SanitizeHTML(in.Body)`; also reject an all-empty draft like SendMyMail does. |
| 7 | P2 | robustness | hst-server/pkg/mailer/mailer.go:292-302, 250-290 | `smtp.Dial` / `tls.Dial` have no timeout and no read/write deadline on the connection. | One hung SMTP host blocks `Drain` forever; the ticker goroutine never returns, the whole outbox stalls behind one row. | `net.Dialer{Timeout: 15s}` + `conn.SetDeadline`, then `smtp.NewClient(conn, host)`. |
| 8 | P2 | validation | hst-server/internal/server/v1/paging.go:18-20 (used by mails.go:324 and news.go:79) | `limit < 0 || limit > 500 → def`, so `?limit=0` passes and `tail` at :54 emits no LIMIT. | A trader can pull the whole `hst.news` table (unbounded) or every mail in a folder in one call. | `if limit <= 0 || limit > 500 { limit = def }`. |
| 9 | P2 | security | hst-server/pkg/mailer/template.go:82-93; trader/register.go:166-173 | `Expand` substitutes macro values verbatim into an HTML template; `NAME` comes from the registration form. | A registrant can put markup in their name and get it rendered in the welcome mail (sent to their own address, so low impact, but it is an unescaped HTML sink). | `html.EscapeString` values in `Expand`, or have callers escape. |
| 10 | P2 | security | hst-server/pkg/events/datafeed.go:83,118-130; datafeeds_worker.go:20-21 | `FeedPassword` travels in clear text in `system.datafeeds.config.<id>` NATS payloads and the internal HTTP snapshot. | Any NATS subscriber on the bus reads feed credentials; mail server passwords are deliberately `json:"-"` (model/mailserver.go:15) but feed passwords are not. | Accept as design if NATS is private, else have workers fetch the password over the service-token HTTP path only. |
| 11 | P2 | robustness | hst-server/internal/server/v1/datafeeds_worker.go:425-468, 257-306 | `NotifyDatafeedsForSymbolID` loads the full symbol catalog once, then per enabled feed calls `feedIncludesSymbol` (one query) and `publishWorkerConfigSnapshot`, which itself calls `loadCatalogSymbolRefs` again (:279). | Every symbol edit is O(feeds × catalog) queries on the request path (admin/symbols.go:1765). | Pass the catalog into `loadWorkerDatafeedConfig`, or batch with `listWorkerDatafeedConfigs`. |
| 12 | P2 | robustness | hst-server/internal/server/v1/mails.go:814-858 | One `pgx.Batch` with up to 2 rows per recipient and no cap on recipient count (`group:*` matches everyone). | A broadcast to 100k accounts builds a 200k-statement batch in one tx and then fires 100k websocket notifies inline. | Cap recipients per send (MT5 limits broadcasts) or chunk the batch. |
| 13 | P2 | robustness | hst-server/internal/server/v1/mails.go:1044-1101 | Staged attachments (`attach_id IS NULL`) are never purged; no other code deletes them (grep). | Uploads that are never sent accumulate forever as bytea. | Periodic delete of unclaimed rows older than a day. |
| 14 | P2 | bug | hst-server/internal/server/v1/admin/reference.go:168-181 | `ReorderDatafeed` parks every feed at `-feed_index-1` then reassigns only the ids in the body; missing ids stay negative, unknown ids are silently ignored, a duplicate id hits the unique index and 500s. | hst-quote uses `feed_index` for per-symbol source priority; a feed left negative jumps to the front of the order. | Verify `len(body) == COUNT(*)` and all ids exist before parking; map unique violation to 400. |
| 15 | P2 | convention | hst-server/internal/server/v1/datafeeds.go:803-841 | `DeleteDatafeed` publishes `SubjectDatafeedDeleted` but no `publishWorkerConfigSnapshot`; `UpdateDatafeed`/`Activate` publish both. | Consistent with hst-quote (feeds.go:484 stops on the deleted subject), so not a bug, but the asymmetry is undocumented. | One-line comment, or fold into `notifyDatafeedConfigChanged`. |
| 16 | P3 | readability | hst-server/internal/server/v1/datafeeds.go:700-701 | `if body.GatewayLogin != nil { }` empty branch. | Dead code; staticcheck SA9003 will flag it, `make check` fails. | Delete. |
| 17 | P3 | convention | hst-server/model/mailtemplate.go:14 (uncommitted) | Blank line replaced by a line holding a single space. | `gofmt -l` lists the file; `make check` fails. | Revert the whitespace change. |
| 18 | P3 | readability | hst-server/internal/server/v1/datafeeds.go:901-911 | `notifyDatafeedConfigChanged` swallows the scan error (`if err != nil { return }`) with no log. | A DB blip silently leaves workers on stale config. | Log at warn like `publishWorkerConfigSnapshot` does. |
| 19 | P3 | readability | hst-server/internal/server/v1/datafeeds.go:374, 384, 404, 424 | `datafeedExists(c, s, id)`, `loadDatafeedParams(ctx, s, id)` are free functions taking `*HttpServer` as the second argument while siblings are methods. | Two calling styles in one file. | Make them methods on `*HttpServer`. |
| 20 | P3 | readability | hst-server/internal/server/v1/datafeeds_worker.go:320-331 | `type feedRow struct{ cfg events.WorkerDatafeedConfig }` wraps one field and is unwrapped immediately. | Noise. | Use `[]events.WorkerDatafeedConfig` directly. |
| 21 | P3 | readability | hst-server/internal/server/v1/datafeeds.go:1285-1294 | `isParamPriorityError` + `paramPriorityErrorMessage` wrap one sentinel. | Two helpers for `errors.Is(err, errParamPriorityOutOfRange)`. | Inline. |
| 22 | P3 | readability | hst-server/internal/server/v1/datafeeds.go:252-282, 1471-1492 | `resolveTranslateSymbol` silently prefers `symbol_id` when both id and a different `symbol` are sent. | Mismatched body goes unnoticed. | Return 400 when both are given and disagree, or document. |
| 23 | P3 | convention | hst-server/internal/server/v1/admin/mail_servers.go:67-80 | Comment says "ports 25, 465 and 587"; code accepts 1..65535. | Stale comment. | Fix the comment or enforce the list. |
| 24 | P3 | readability | hst-server/pkg/mailer/mailer.go:94-120, 143-147 | `Drain` returns silently on query/scan error; `failed` logs but `sent` swallows the stats update error. | Silent failures in a retry loop. | Log at warn. |
| 25 | P3 | convention | hst-server/pkg/mailer/mailer.go:305-321 | Message has no `Date:` or `Message-ID:` header. | Spam filters score both; some MTAs add them, some reject. | Add `Date` and a uuid `Message-ID`. |
| 26 | P3 | structure | hst-server/internal/server/v1/mails.go (1,145 lines) | Trader mailbox CRUD, manager broadcast, attachments and recipient resolution in one file. | New developer cannot tell the trader path from the manager path without reading all of it. | Split: `mails_trader.go`, `mails_send.go`, `mail_attachments.go`. |
| 27 | P3 | structure | hst-server/internal/server/v1/datafeeds.go (1,845 lines) | Feed CRUD, params with priority logic, translates and symbol rows in one file. | Same as above. | Split params/translates/symbols into their own files. |
| 28 | P3 | readability | hst-server/internal/workerstatus/subscriber.go:125 | `go s.flushLoop()` bare goroutine; `loadBaselines` failure only warns. | Fine for hst-server (the `h.Go` rule is engine-only), noted for consistency. | None required. |
| 29 | P3 | validation | hst-server/internal/workerstatus/subscriber.go:190-207 | `onQuoteJournal` inserts every NATS journal line with unbounded `Message` and unchecked `Code`. | A chatty or buggy worker fills hst.journal. | Cap message length, clamp code to 0..2. |
| 30 | P3 | readability | hst-server/internal/server/v1/mails.go:1142 | `Content-Disposition` built by string concat, only `"` stripped. | fasthttp strips CR/LF so not injectable, but `mime.FormatMediaType` is the obvious tool. | `mime.FormatMediaType("attachment", map[string]string{"filename": name})`. |

## Flow traces

### 1. Datafeed CRUD → params/translates/symbols → snapshot publish → worker config endpoint
1. `POST /api/v1/datafeeds` admin/routes.go:255 → `CreateDatafeed` datafeeds.go:550: validator, name regex :284, mode bits :295, module vs mode via `datafeedmodules.Validate` :305, gateway password policy :340.
2. Insert with `feed_index = MAX+1` :584-606 (finding 4) → `selectDatafeedDetail` :614 → `publishDatafeedEvent(system.datafeeds.created)` :623 → `publishWorkerConfigSnapshot` :624 → `notifyDatafeedWS` :625.
3. Params: `CreateDatafeedParam` :962 (priority MAX+1, finding 4), `UpdateDatafeedParam` :1069 (tx: priority swap :1209, then COALESCE patch), `DeleteDatafeedParam` :1163 → `deleteDatafeedParamWithCompact` :1266 (finding 2). Each ends in `notifyDatafeedConfigChanged` :901 → event + snapshot + WS.
4. Translates :1345-1573, symbols :1667-1800: resolve symbol by id or name, insert, `notifyDatafeedConfigChanged`.
5. Snapshot build: `loadWorkerDatafeedConfig` datafeeds_worker.go:257 → params :267, catalog :279, db translates :284, rules :288, `mergeEffectiveTranslates` :598 (scope rules ∪ explicit translates; with no rules only explicit translates), sessions :295, settings :300 → `events.PublishConfigSnapshot` → `system.datafeeds.config.<id>` (events/datafeed.go:92).
6. Worker pull: `GET /internal/v1/datafeeds?mode=1|2` routes.go:65-67 behind `ServiceAuth` (constant-time token compare, service_auth.go:25) → `listWorkerDatafeedConfigs` :308 (`enable = 1 AND (mode & $1) = $1`).
7. hst-quote `configclient.Snapshot` (client.go:24-39) and hst-news `Snapshot` (client.go:24-35) field names/types match `events.WorkerDatafeedConfig` exactly; hst-news ignores the extra arrays. hst-quote subscribes `system.datafeeds.config.>`, `.updated`, `.deleted` (handler.go:116-122) — subjects match.
Verdict: OK, with gaps at step 3 (finding 2) and step 2 (finding 4, 3).

### 2. Param priority swap / compact
1. `PATCH .../params/{paramId}` body `priority` → `applyDatafeedParamPriority` :1209: lock row FOR UPDATE, count, `validateDatafeedParamPriority` (0..n-1, int64 compare) :1201, no-op if equal, find occupant, park at -1, move occupant, move self. Correct swap under the unique index.
2. `DELETE` → `deleteDatafeedParamWithCompact` :1266: delete, then single-statement shift.
Verdict: gap at step 2 (finding 2). Swap OK.

### 3. Activate
1. `POST /datafeeds/{id}/activate` :856 → `UPDATE enable = 1` :863 → detail → `system.datafeeds.updated` + snapshot + WS :883-885.
2. hst-quote `onDatafeedEvent` (feeds.go:484) restarts unless deleted/no quote flag; snapshot arrives on `.config.<id>` for hot apply.
Verdict: OK. Note no `deactivate` endpoint; `PATCH enable=0` serves (valid given finding 3 is fixed).

### 4. Mail send (manager broadcast)
1. `POST /api/v1/mails` behind `MgrRightEmail` admin/routes.go:176 → `SendMail` mails.go:707: channel check :723, mailbox name required for internal :729-738.
2. Recipients: `mailRecipientsWhere` :586 — always starts from the caller's group masks (`groupWhere` → `utils.GroupAccessFor`) and ANDs the To expression (`parseMailTo`/`mailToWhere` mail_recipients.go:41,128; LIKE metachars escaped :208). Cannot widen reach. OK.
3. Body: `mailer.SanitizeHTML(in.Body)` :785 once; per recipient `expandMailMacros(body, r, true)` escapes values (mail_macros.go:37-42); subject unescaped (plain text). OK.
4. Tx: claim attachments :795, batch inserts inbox rows + outbox email rows + one sender outbox row :814-894, commit :912, WS notify after commit :917-922 (correct order). Journal with full body :924 (finding 5).
5. Email leg: `hst.outbox` row → `mailer.Drain` (pkg/mailer/mailer.go:94) → `Send` :250: 465 implicit TLS, STARTTLS when offered, `PlainAuth` (net/smtp refuses plaintext auth off-localhost), `Rcpt` rejects CR/LF; `message` :306 Q-encodes From name and Subject (CR/LF forced into encoded form). Header injection: not possible. Timeout: none (finding 7).
6. Credentials: `smtp_password` stored plaintext in DB (accepted, MT5 parity), `json:"-"` on model, never logged (mail_servers.go:239-241 logs name/smtp_server only), NotifySystem payload omits it. OK.
7. Trader side `SendMyMail` :391: mailbox must be a staff mailbox whose masks reach the trader's group :409, body sanitised :429; `UpdateMyDraft` not sanitised (finding 6).
8. System templates: `Render` template.go:34 → `read` restricts catalog to a single path segment :73 → `Expand` unescaped (finding 9); callers register.go:166, recover.go:364 queue via default server.
Verdict: OK for injection; gaps at steps 4, 5, 7, 8 (findings 5, 7, 6, 9).

### 5. Worker status subscribe
1. server.go:92 `WorkerStatus.Start(nats)` → baselines from hst.datafeeds :141 → queue-subscribe `hstquote.status.>`, `hstnews.status.>`, `hstquote.journal.>` :107-116 (subjects match hst-quote model/subjects.go:25-30 and hst-news publisher.go:11; Event JSON tags match hst-quote/internal/status/publisher.go:17-22).
2. Stats events (any delta ≠ 0) → `applyStats` :274, pending counters under `mu`, WS throttled 1/s; connection events → `applyConnection` :209 immediate DB write + journal on flip :235.
3. `flushLoop` :327 flushes pending deltas every 5s with `ticks_count = ticks_count + $2` (additive, safe across restarts), status beat every 10s.
4. `Stop` :130 deadlocks (finding 1).
Verdict: gap at step 4.

## Readability notes for onboarding
- datafeeds.go — feed/param/translate/symbol CRUD; clear per handler, too long as one file; split by sub-resource first, then delete the empty `if` and make the free helpers methods.
- datafeeds_worker.go — snapshot assembly and scope merge; `mergeEffectiveTranslates` is the one function to read carefully; drop the `feedRow` wrapper.
- mails.go — trader mailbox + manager broadcast + attachments; SendMail's `refs`/`delivered` indexing is the hardest 80 lines; split the file.
- mail_macros.go — small, clear; note the `escape` flag.
- mail_recipients.go — To-expression parser; well commented, fine.
- news.go — tiny, clear; relies on `readPage` (finding 8).
- admin/mail_servers.go — clear; fix the ports comment.
- admin/mail_templates.go — clear.
- workerstatus/subscriber.go — clear except the Stop lock bug; counters model (base + pending) worth a one-line comment on `feedState`.
- pkg/datafeedmodules — clear alias table; obvious where to add a module.
- pkg/events/datafeed.go — clear; the struct is the contract with two other repos, say so in a comment.
- pkg/mailer — clear; add timeouts, log on silent returns.
- model/datafeed.go, mailserver.go, mailtemplate.go — clear; revert the whitespace diff.
- templates/ — two HTML templates, MT5 comment macros; fine.

## Open questions
- Is NATS considered a trusted network such that feed passwords in `system.datafeeds.config.<id>` are acceptable (finding 10)?
- Does the trader terminal render mail `subject` as text only? Subject macros are not escaped (mail_macros.go:29).
- Is a cap on broadcast recipients desired (MT5 behaviour) or is unbounded intentional (finding 12)?
- Should `ReorderDatafeed` live with the datafeed handlers instead of admin/reference.go?
