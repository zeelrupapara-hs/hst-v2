---
name: hst-go-review
description: Go review standard for hst-v2 backend (hst-server, hst-core, hst-quote, hst-news). Load before reviewing or writing any Go in this repo. Covers the house conventions, the architecture a reviewer must know, the bug/validation checklist, the readability bar for junior developers, and the report format.
---

# hst-go-review

You are reviewing a Go trading backend (MT5-style broker platform). Read this whole file, then the code you are assigned, file by file, top to bottom. No sampling. No guessing: every finding cites `path:line` and quotes or describes the exact code.

## 1. Architecture you must know

Four independent Go modules, no shared module, `pkg/` is copy-pasted into each:

| Module | Role | Stack |
|---|---|---|
| `hst-server` | HTTP/WS API (Fiber), auth, admin + trader panels, publishes commands and config events on NATS | pgx, NATS, Redis, Influx |
| `hst-core` | Trading engine. Holds accounts in memory, sharded across pods; validates, routes, fills, margin, stop-out, swaps, EOD | pgx, NATS, Redis |
| `hst-quote` | FIX/DDE price feeds → ticks on NATS `hstquote.tick.*`, Influx history | NATS, Redis, Influx |
| `hst-news` | RSS feeds → `hstnews.item.*` | NATS, Redis |

Flow of a trade: client → `hst-server` handler (shape validation only, no trading logic) → NATS `system.orders` (queue group `engine_orders`) → any `hst-core` pod → if it does not own the login, `forwarded()` to `system.owner.<shard>.<topic>` with the reply subject intact → owning pod validates (`validate.go`), routes (`routing.go`), executes, writes Postgres, publishes `websocket.accounts.<login>.*` → `hst-server` WS hub → client.

Sharding (`hst-core/internal/shardmap`, `handler/shards.go`): 1024 fixed shards, `ShardOf(login)` FNV hash, pods on a consistent-hash ring with 100 replicas; pods register in Redis `core:pods:<name>` TTL 15s, refresh 5s; `ReassignShards` diff → close lost inboxes → flush+release lost accounts → load gained → open gained inboxes. Invariants to check: exactly one pod acts for a login at any time; nothing is lost or doubled while shards move; in-flight commands during a move are answered once; account money is flushed before release; `hst-server` never needs to know the shard (it publishes to the queue, the engine forwards).

Config events: `hst-server` publishes `system.<table>.>` after commits; `hst-core` reloads settings wholesale. Auth: `hst-server` middleware never touches Postgres (Redis sessions only). Rights: packed manager bitset `model.MgrRight*`, checked by `s.Middleware.Authorization(...)`.

Docs: `CONTRIBUTION.md`, `docs/order-lifecycle.md`, `docs/manager-panel.md`, `hst-core/README.md`, `hst-quote/README.md`, `hst-news/README.md`.

## 2. House conventions (violations are findings)

- Handlers are methods on `*HttpServer`; routes in `routes.go`; full swagger block on every handler (`@Id @Tags @Accept @Produce @Success @Failure… @Security @Router`); failure body `ErrorResponse`.
- Request/response types `CrtX`, `UptX`, `ViewX`.
- Credentials only via `BasicAuthParser` → Locals. Handlers never read `Authorization` directly.
- `Protect` middleware: zero Postgres reads. Only `Login`/`RefreshToken` may hit the DB for auth.
- Any change to `rights`, `group`, or a password → `s.OAuth2.InvalidateLogin` after commit.
- Comments: one line, one sentence, plain words. Never a paragraph.
- Logging: `Log.Journal(type, code, msg, kv...)` with MT5 codes; audited actions record `actor` and `target`; never log passwords, tokens, or request bodies.
- Migrations: epoch-millis names, always a `down` file; `model/` constants follow migration column order.
- Engine: `h.Go(f)` not bare `go`; subscribe errors return, never `Fatal`; nothing calls `os.Exit`/panics for shutdown; `Workers.Submit` false return is handled; `Start` order = load state → workers → subscribe; `Stop` reverses and is idempotent.
- `make check` must be clean: gofmt, vet, staticcheck, errcheck, gosec, govulncheck.

## 3. Bug and validation checklist

Go through every item for every handler/function you read.

**Input validation (trust boundary = every HTTP handler, every NATS message, every feed frame)**
- Every field bound from the request validated: required, ranges, enum membership, string length, positive ids, volume step/min/max, price > 0, expiry rules, pagination bounds.
- Path/query ids parsed with error handled; mismatch between path id and body id.
- Validation happens before any DB write; partial writes on later failure are rolled back (tx).
- Tenant/scope: a manager only touches groups they may (`utils/group_access.go`); a trader only touches their own login; admin vs trader route separation.
- Rights bit on every mutating admin route; correct bit for the operation.

**Correctness**
- Error returned from DB/NATS/Redis checked; `pgx.ErrNoRows` mapped to 404 not 500; unique-violation mapped (`utils/db_errors.go`).
- Transactions: begin/commit/rollback paths complete; events published after commit, not before; no publish inside a tx that may roll back.
- Money math: float vs decimal, rounding to digits, sign of swaps/commissions, currency conversion direction, division by zero on missing quotes/contract size/leverage.
- Time: UTC, server time vs session time, EOD boundary, holidays, expiry comparisons.
- Nil maps/slices/pointers; index out of range; integer overflow/narrowing; string → number parsing unchecked.
- Off-by-one in paging, priority swaps, step rounding.

**Concurrency (engine especially)**
- Every shared map/slice/struct guarded; lock order consistent (`h.mu`, `dealingMu`, `book.Entry` lock); no lock held across NATS publish/DB call without reason; no copy of a struct containing a mutex; goroutine leaks on shutdown; channel close by the right side; `context` honoured.
- Sharding: reassign vs in-flight command; duplicate handling when two pods briefly hold one shard; Redis `KEYS` at scale; forwarded messages answered exactly once; tempId uniqueness across pods.

**Security**
- Auth on every route; IDOR on login/ticket/id params; rate limits on login/register/recover; password hashing (argon2), token entropy, JWT alg/exp; SQL built with parameters only; no secrets in logs; CORS config; WS auth on upgrade; mail templates sanitised.

**Resilience**
- Reconnect handling, slow consumers, timeouts on every outbound call, bounded queues, panics recovered in workers, readiness reflects dependencies.

## 4. Readability bar (junior developer joins next week)

Flag, per file, anything that would cost a new developer more than a minute to understand:
- Function > ~80 lines or doing more than one thing; file > ~800 lines mixing concerns; deep nesting (> 3 levels).
- Names that hide meaning (`tmp`, `d`, `x`, `data2`), booleans without a verb, abbreviations not used elsewhere.
- Magic numbers/strings without a named constant; retcodes without the `model.Ret*` name.
- Missing one-line doc comment on exported identifiers; stale or wrong comments.
- Duplicated blocks that should be one helper that already exists (look in `utils/`, `pkg/http`, `model/`).
- Inconsistent patterns between sibling handlers (one validates, the next does not; different error mapping for the same case).
- Control flow a junior could misread: early-return missing, error swallowed, `else` after return.
- Structure: where a new endpoint/rule/feed type should be added is not obvious.

## 5. How to work

1. `cat` every assigned file fully. Then read callers/callees you need (`grep -rn`).
2. Trace each assigned flow end to end across module boundaries (handler → NATS subject → engine handler → DB → event → WS). Check the subject strings match on both sides.
3. Try to refute each finding before writing it: re-read the code path, check for a guard elsewhere. Only keep what survives.
4. Do not fix anything. Do not run the services. `go build ./...` and `go vet ./...` in the module are allowed.

## 6. Report format

Write the report to the path you were given. Markdown, this shape:

```
# <area> review

Scope: <files / flows read>, <N> files, <N> lines.

## Summary
3–6 lines: overall health, the biggest risks, readability verdict for a junior.

## Findings
| # | Sev | Type | Where | What | Why it matters | Fix |
Sev: P0 data loss/money/security, P1 wrong behaviour, P2 robustness, P3 readability/structure.
Type: bug, validation, concurrency, security, readability, structure, convention.
Where: path:line. What: one sentence + exact code fragment. Fix: one sentence.

## Flow traces
For each assigned flow: steps as a numbered list with path:line at each hop, and a verdict (OK / gap at step N).

## Readability notes for onboarding
Per file: 1 line — what it does, is it clear, what to simplify first.

## Open questions
Things you could not decide from the code.
```

Be exact, be short, no praise paragraphs. A finding without `path:line` is not a finding.
