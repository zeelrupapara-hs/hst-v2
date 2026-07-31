# Auth walkthrough

Every request and response below is real output from a running server.

## The model in one line

A **user** is an identity. A **manager** is a row in `hst.managers` pointing at
that user's login. No manager row, no panel access.

```
hst.clients   1 ─── N   hst.users   1 ─── 1   hst.accounts
                            │
                            └── 0..1   hst.managers   ← this row is what lets you log in
```

Connection types: `32` admin panel, `33` manager panel.

---

## 0. The seed admin

On an empty database the server creates login **1000** from `seed/manager.go`,
using `FIRST_MANAGER_PASSWORD` from `.env`. It gets all 77 manager rights, so
every endpoint works immediately.

```bash
curl -u '1000:Bootstrap-Admin-2026!' -X POST localhost:8080/auth/v1/login \
  -H 'Content-Type: application/json' -d '{"connection_type":33}'
```

```json
{
  "success": true,
  "code": 0,
  "data": {
    "login": 1000,
    "access_token": "eyJhbGciOiJFZERTQSIsImtpZCI6IjY5MTNkMjk5NWM2N2RiMDgiLCJ0eXAi...",
    "refresh_token": "CrFY1IZRvGuhnBxvTKM1camwLr7N9UQ_gDh8z8i9WBo",
    "session_id": "d521f178-d9cd-461c-ae3d-eeebce218e67",
    "expires_in": 7200,
    "connection_type": 33,
    "code": 0
  }
}
```

`code 0` means the session is ready. Changing the password is your choice, not
forced:

```bash
curl -X POST localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"old_password":"Bootstrap-Admin-2026!","new_password":"Admin-Pass-2026!"}'
# HTTP 204
```

That kills every other session of this login, so log in again afterwards.

---

## 1. Create the client

The KYC person. Optional, but a trading account normally belongs to one.

```bash
curl -X POST localhost:8080/api/v1/clients \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "person_name": "Sara",
    "person_last_name": "Khan",
    "contact_email": "sara.khan@example.com",
    "client_type": 1,
    "address_country": "AE"
  }'
```

```json
{
  "client_id": 1,
  "client_type": 1,
  "person_name": "Sara",
  "person_last_name": "Khan",
  "contact_email": "sara.khan@example.com"
}
```

---

## 2. Create the user

This writes `hst.users` **and** `hst.accounts` in one transaction. `login` is
assigned by the server.

```bash
curl -X POST localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "client_id": 1,
    "group": "managers\\admin",
    "rights": 3,
    "name": "Sara Khan",
    "email": "sara.khan@example.com",
    "leverage": 100,
    "password_main": "Sara-Main-2026!",
    "password_investor": "Sara-Inv-2026!"
  }'
```

```json
{
  "login": 1001,
  "client_id": 1,
  "group": "managers\\admin",
  "rights": 3,
  "name": "Sara Khan",
  "email": "sara.khan@example.com",
  "is_manager": false,
  "balance": 0
}
```

`rights: 3` = `enabled` (1) + `password` (2). `is_manager: false` — she is not
staff yet.

---

## 3. She cannot log in yet

```bash
curl -u '1001:Sara-Main-2026!' -X POST localhost:8080/auth/v1/login \
  -H 'Content-Type: application/json' -d '{"connection_type":33}'
```

```json
{"success":false,"code":1011,"error":"Forbidden","message":"the login has no manager rights"}
```

`HTTP 403`, `code 1011`. The password was correct. There is no manager row.

---

## 4. Promote her to manager

```bash
curl -X POST localhost:8080/api/v1/managers \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "login": 1001,
    "name": "Sara Khan",
    "groups": ["real\\*"],
    "rights": ["right_manager", "right_clients_access", "right_trades_read"]
  }'
```

```json
{
  "login": 1001,
  "name": "Sara Khan",
  "groups": ["real\\*"],
  "right_manager": 1,
  "right_clients_access": 1,
  "right_trades_read": 1,
  "right_admin": 0
}
```

`right_manager: 1` opens the manager panel. `right_admin: 0` keeps the admin
panel shut.

---

## 5. Now she logs in

```bash
curl -u '1001:Sara-Main-2026!' -X POST localhost:8080/auth/v1/login \
  -H 'Content-Type: application/json' -d '{"connection_type":33}'
```

```
code 0 | session 62610352-f24f-410b-b09c-fe0aa736e6d0 | expires_in 7200
```

Admin panel with the same credentials:

```bash
curl -u '1001:Sara-Main-2026!' -X POST localhost:8080/auth/v1/login \
  -H 'Content-Type: application/json' -d '{"connection_type":32}'
```

```json
{"success":false,"code":1024,"message":"this terminal type is not permitted for the login"}
```

`code 1024` — she has no `right_admin`.

---

## 6. Who am I

```bash
curl -H "Authorization: Bearer $SARA_TOKEN" localhost:8080/api/v1/auth/me
```

```json
{
  "login": 1001,
  "is_manager": true,
  "granted": ["right_clients_access", "right_manager", "right_trades_read"]
}
```

---

## 7. Rights decide each route

```
GET /api/v1/clients   has right_clients_access   -> 200
GET /api/v1/users     no right_acc_read          -> 403
```

Grant more with `PATCH /api/v1/managers/1001`. Note it **replaces** the whole
right list, so always send the complete set.

---

## 8. Refresh

The access token lasts 2 hours; the refresh token lasts 7 days.

```bash
curl -X POST localhost:8080/auth/v1/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"HlE1AnPMjJU1rzX7cBtZtdCka4VcGBZH6aoIXk4toPQ"}'
```

```
new session e07d859c-5c0a-43f0-a406-1d526eb2ec96
```

Refresh tokens are single use. Send the same one twice:

```json
{"success":false,"code":60001,"message":"expired or invalid session"}
```

`HTTP 401`. A replay is treated as theft, so **every** session of that family is
killed, including the one just issued. Both the thief and the real user are
logged out.

---

## 9. Demote

```bash
curl -X DELETE localhost:8080/api/v1/managers/1001 -H "Authorization: Bearer $TOKEN"
# HTTP 204
```

Her live token stops working on the next request, and login returns `1011`
again. The user and the account survive untouched.

---

# What the server does inside

Same person, Sara, login **1001**, password `Sara-Main-2026!`. Everything below
is captured from a real run.

## Before she logs in

Postgres holds her; Redis holds nothing.

```
hst.users     login 1001   rights 3   password_main $argon2id$v=19$m=65536,t=3,p=2$+BJeh3qJuJ6Jw...
                           locked_until 0
hst.managers  login 1001   right_manager 1   right_admin 0   access {}
redis         0 keys
```

`rights 3` = `enabled`(1) + `password`(2). `access {}` = any IP allowed.

## The login, check by check

```
POST /auth/v1/login
Authorization: Basic MTAwMTpTYXJhLU1haW4tMjAyNiE=
{"connection_type": 33}
```

| # | Check | On this request | If it fails |
|---|---|---|---|
| 1 | decode Basic header | `1001` / `Sara-Main-2026!` | 401 |
| 2 | login is a number > 0 | 1001 ✓ | 401 · 1020 |
| 3 | ip not throttled | `GET fail:ip:127.0.0.1` → nil ✓ | 429 · 1018 |
| 4 | **DB READ** `SELECT ... FROM hst.users` | 1 row | 401 · 1020 + a dummy hash so the timing matches |
| 5 | `locked_until` in the past | 0 ✓ | 403 · 60002 |
| 6 | **argon2 verify**, ~60 ms | match ✓ | 401 · 1001, failure counted |
| 7 | `rights` has `enabled` | 3 ✓ | 403 · 1002 |
| 8 | connection_type ≥ 32 | 33 ✓ | 403 · 1024 |
| 9 | **DB READ** `SELECT ... FROM hst.managers` | 1 row | 403 · 1011 |
| 10 | that row permits terminal 33 | `right_manager=1` ✓ | 403 · 1024 |
| 11 | caller ip in `access` | list empty ✓ | 403 · 1012 |

Checks 7–11 run **after** the password on purpose. Answering "disabled" or "not
a manager" first would tell an attacker the login exists.

## Then the writes

**Postgres — one row:**

```
session_id   b9a05cfe-cd4d-42a1-9cc3-1778d995b238
login        1001
family_id    b9a05cfe-cd4d-42a1-9cc3-1778d995b238   ← same as session_id on a first login
parent_id    (null)                                 ← set only by a refresh
token_hash   4582ac97a54e5b2a48cb15ed2f7b8093ea29c70267c333f21dcecb199ee92f6f
```

The refresh token itself is never stored, only its SHA-256. A database dump
cannot be replayed.

**Redis — four keys:**

```
sess:b9a05cfe-…    the snapshot, TTL 604800s (7 days)
rt:4582ac97…       refresh token hash → session
fam:b9a05cfe-…     the rotation family, so all of it can be killed at once
user:1001          every session of this login, for "rights changed"
```

**The snapshot** — everything needed to authorise, so nothing has to be looked
up again:

```json
{
  "sid": "b9a05cfe-cd4d-42a1-9cc3-1778d995b238",
  "login": 1001,
  "cid": 1,
  "grp": "managers\\admin",
  "rights": 3,
  "ct": 33,
  "rst": false,
  "mgr": true,
  "mrights": [72075186223972354, 0],
  "ver": 1,
  "expires_at": 1785908719286990000
}
```

`mrights` is the 77 rights packed into two numbers:

```
72075186223972354  → bits 1, 44, 56
                   → right_manager, right_trades_read, right_clients_access
```

Checking a right is a bit test, not a query.

**Memory** — the same snapshot is put in the in-process cache with a 30 second
TTL.

## The token she gets back

```json
header  {"alg": "EdDSA", "kid": "b6f341d6bdb8d1b1", "typ": "JWT"}
claims  {"sid": "b9a05cfe-…", "cid": 1, "ct": 33, "rst": false,
         "ver": 1, "iss": "hstserver", "sub": "1001",
         "exp": 1785311119, "iat": 1785303919}
```

**No rights in the token.** Only `sid`. Rights live in the snapshot, so
revoking them takes effect without waiting for the token to expire.

## Every request after that

```
GET /api/v1/clients
Authorization: Bearer eyJhbGciOiJFZERTQSIsImtpZCI6…
```

```
1. verify signature       Ed25519, ~25 µs   ← no I/O, a forged token dies here
2. read the snapshot
     memory hit?          ~60 ns            → go to 3
     memory miss?         Redis GET ~200 µs → cache it 30s, go to 3
     not in Redis?        401 · 60001       ← "refresh", never a DB lookup
     Redis down?          503 · 1018        ← our fault, not hers
3. snapshot not expired
4. rights has enabled
5. rst false, or the path is change-password
6. bit 56 right_clients_access set?   ~1 ns
7. handler runs
```

Measured: **10 requests to `/auth/me` → 0 Postgres statements.**

The database is read on exactly three paths: login, refresh, and whatever the
handler itself needs. Never for authentication.

## Where each thing lives

| Thing | Postgres | Redis | Memory |
|---|---|---|---|
| password hash | yes | no | no |
| session row, audit trail | yes | no | no |
| live session, rights snapshot | no | yes, 7d | yes, 30s |
| refresh token hash | yes | yes | no |
| the tokens themselves | never | never | never |

## When rights change

```
PATCH /api/v1/managers/1001   →  UPDATE hst.managers
                              →  delete every sess: key listed in user:1001
                              →  drop them from local memory
                              →  publish on hst:auth:invalidate so other pods drop them too
```

Next request rebuilds from Redis and sees the new rights. Change the rights
straight in SQL instead and nothing is notified, so it takes up to the 30
second cache TTL.

---

## Return codes

| Code | HTTP | Meaning |
|---|---|---|
| 0 | 200 | fine |
| 1001 | 401 | wrong password |
| 1002 | 403 | account disabled, `rights` has no `enabled` bit |
| 1011 | 403 | no manager row |
| 1012 | 403 | ip not in the manager's allowlist |
| 1018 | 429/503 | throttled or the hasher is saturated, see `Retry-After` |
| 1020 | 401 | no such login |
| 1024 | 403 | terminal type not permitted |
| 1026 | 200/403 | reset_pass is set on the account: 200 from login, 403 elsewhere until changed |
| 60001 | 401 | session gone, refresh or log in again |
| 60002 | 403 | locked after 10 failed attempts, 15 minutes |

`1000`–`1034` are MT5. `60001`–`60002` are ours, because MT5 has no expiring
bearer session.

---

## Rules worth knowing

- **Rights are cached for 30 seconds.** Changing them through the API drops the
  sessions immediately; changing them straight in SQL takes up to 30 seconds.
- **Password change kills every other session** of that login.
- **10 wrong passwords locks the account** for 15 minutes.
- **The password checks come first.** "Disabled" and "no manager row" are only
  reported after the password is verified, so failures cannot be used to find
  out which logins exist.
