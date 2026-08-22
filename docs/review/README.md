# Backend review — merged index

Ten reviewer agents, one per area of `docs/review-plan.md`, each using `.claude/skills/hst-go-review/SKILL.md`.
Every finding carries `path:line` in its report. Nothing was fixed. Build and vet are clean in all four modules.

| Report | P0 | P1 | P2 | P3 |
|---|---|---|---|---|
| [01 core sharding](01-core-sharding.md) | 3 | 5 | 6 | 4 |
| [02 core orders / routing / dealing](02-core-orders.md) | 6 | 9 | 13 | 5 |
| [03 core positions / margin / swaps](03-core-positions.md) | 2 | 9 | 11 | 8 |
| [04 server auth](04-server-auth.md) | 3 | 5 | 8 | 5 |
| [05 server trading API](05-server-trading-api.md) | 0 | 5 | 13 | 14 |
| [06 server admin config](06-server-admin-config.md) | 0 | 8 | 16 | 11 |
| [07 server admin accounts / trader](07-server-admin-accounts.md) | 2 | 5 | 8 | 6 |
| [08 server datafeeds / mail](08-server-datafeeds-mail.md) | 0 | 2 | 13 | 15 |
| [09 hst-quote / hst-news](09-quote-news.md) | 0 | 4 | 9 | 16 |
| [10 platform / quality / onboarding](10-platform-quality.md) | 0 | 1 | 10 | 19 |
| **Total** | **16** | **53** | **107** | **103** |

## P0 — fix before anything else

Sharding (report 01)
1. `hst-core/handler/shards.go:89-117` — no handoff: gainer serves a shard up to 5s before the loser flushes; two pods execute, tick, stop-out and write the same accounts.
2. `hst-core/handler/load.go:49-54` + `shards.go:171-193` — every boot loads all accounts under an all-shards map, then the first reassign `SaveAccount`s every "lost" account, overwriting real owners' writes.
3. `hst-core/handler/shards.go:66-73` — Redis `KEYS` error → `livePods` returns `[me]` → pod claims all 1024 shards and fights the others.

Orders / dealing (report 02)
4. `hst-core/handler/store.go:199-206` — `updateOrder` writes extended volume into `volume_current`, 0 into `_ext`; modified orders reload at 10 000× volume.
5. `hst-core/handler/dealing.go:180-184`, `orders.go:423-438` — dealer confirm re-INSERTs the order row.
6. `hst-core/handler/dealing.go:98-101`, `orders.go:595` — `refuseByRule` state overwritten; cancel branch unreachable.
7. `hst-core/handler/orders.go:194,400`, `dealing.go:393-630` — reject/timeout of a modify deletes the live order.
8. `hst-core/handler/money.go:34`, `floating_margin.go:279`, symbols migration `:65` — margin rate defaults 0 → no margin reserved, `checkMoney` skipped.
9. `hst-core/handler/orders.go:92-96`, `validate.go:200-204` — at-market price handling (see report).

Positions / money (report 03)
10. `hst-core/handler/end_of_day.go:156` → `positions.go:635` — nightly swaps never persisted; lost on restart/shard move.
11. `hst-core/handler/commissions_accrual.go:166,193` — turnover reads legacy `volume`, converts with `Lots()` → period commissions 10 000× too small.

Auth / accounts (reports 04, 07)
12. `hst-server/internal/server/v1/oauth.go:190` — access+refresh token written to journal detail, readable by any journals manager.
13. `hst-server/pkg/oauth2/sessions.go:146-215` — revocation leaves refresh token alive; logout/password/rights change do not end sessions.
14. `hst-server/internal/server/v1/admin/reference.go:324` — `ResetUserPassword`: no reach check, no `InvalidateLogin`; any manager can take over an admin.
15. `hst-server/internal/server/v1/admin/users.go:350` — `DeleteUser` hard-deletes with no balance/position guard; cascades destroy `accounts`/`managers` rows.
16. (same as 14, also reported by 07)

## P1 themes (53) — see each report's table
- Group rename not cascaded to `hst.users."group"` / masks / engine keys (04, 06).
- Manager group-mask not checked on Get/Update/Delete by id and on create/move user (06, 07).
- Engine retcode dropped → success reported on refusal (05: Fix/Delete position, balances, WS).
- Dealing resolved by guessable `request_id` only; reach check on caller-supplied login is meaningless (05).
- Memory mutated before DB write, no rollback on error (02, 03).
- Floating P/L uses open-time rate; hedged margin ignores leverage/price (03).
- hst-news subscribes the wrong subjects; tick subject breaks on dotted symbols; dead feeds never restart (09).
- `workerstatus.Stop` deadlocks; param priority compaction hits the unique index (08).
- No rate limit on register/recover/verify (04). `time.Sleep` inside a NATS callback (02).

## Readability / onboarding verdict (for the junior joining)
Conventions are good and mostly followed; the problems are size and duplication, not style. Start with report 10's
"Onboarding guide gaps". Biggest readability costs: `datafeeds.go` (1.8k lines), `admin/symbols.go` (1.8k),
`mails.go`, `positions.go` in core (940 lines, many >80-line functions), two error vocabularies, `pkg/` copied
into four modules and drifting, zero tests. Assign the junior to modules 08/09 first (lowest P0 risk, most P3 cleanup).

## Suggested order of work
1. P0 1–3 (sharding handoff) and 10–11 (money persistence) — engine correctness.
2. P0 12–15 — auth/session.
3. P0 4–9 — order/dealing store bugs.
4. P1 by theme above, then P2.
5. Hand P3 + report 10 gaps to the junior as starter tasks.
