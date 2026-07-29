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
using `FIRST_MANAGER_PASSWORD` from `.env`.

```bash
curl -u '1000:Bootstrap-Admin-2026!' -X POST localhost:8080/auth/v1/oauth2/login \
  -H 'Content-Type: application/json' -d '{"connection_type":33}'
```

```json
{
  "success": true,
  "code": 1026,
  "data": {
    "login": 1000,
    "access_token": "eyJhbGciOiJFZERTQSIsImtpZCI6IjA3YjQyM2VhYTc5YjI0ZjciLCJ0eXAiOiJKV1QifQ...",
    "refresh_token": "HlE1AnPMjJU1rzX7cBtZtdCka4VcGBZH6aoIXk4toPQ",
    "session_id": "fb7f3022-17e4-48cf-964a-1d57f6aa4ff8",
    "expires_in": 7200,
    "connection_type": 33,
    "code": 1026
  }
}
```

`code 1026` = must change the password. The token works, but only for one call.

```bash
curl -X POST localhost:8080/api/v1/auth/oauth2/change-password \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"old_password":"Bootstrap-Admin-2026!","new_password":"Admin-Pass-2026!"}'
# HTTP 204
```

Log in again and `code` is `0`.

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
curl -u '1001:Sara-Main-2026!' -X POST localhost:8080/auth/v1/oauth2/login \
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
curl -u '1001:Sara-Main-2026!' -X POST localhost:8080/auth/v1/oauth2/login \
  -H 'Content-Type: application/json' -d '{"connection_type":33}'
```

```
code 0 | session 62610352-f24f-410b-b09c-fe0aa736e6d0 | expires_in 7200
```

Admin panel with the same credentials:

```bash
curl -u '1001:Sara-Main-2026!' -X POST localhost:8080/auth/v1/oauth2/login \
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
curl -X POST localhost:8080/auth/v1/oauth2/refresh \
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
| 1026 | 200/403 | must change password: 200 from login, 403 from anywhere else |
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
