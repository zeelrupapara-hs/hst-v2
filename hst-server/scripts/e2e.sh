#!/usr/bin/env bash
# End to end scenario check against a locally running hst-server.
# Every case asserts the http status and, where the api reports one, the MT5 retcode.
#
# THIS SUITE RUNS ONCE PER DATABASE. It changes the bootstrap admin password and
# writes clients, users and managers, so a second run against the same database
# fails. Start from a fresh database every time: drop it, migrate, boot the
# server so it seeds the first admin, then run this. The check below refuses to
# start on a database that has already been used.
#
# Environment:
#   E2E_BASE_URL      server under test, default http://localhost:8080
#   E2E_DATABASE_URL  postgres url for the assertions, default local hst database
#   E2E_ADMIN_PASSWORD      bootstrap admin password, must match FIRST_MANAGER_PASSWORD
#   E2E_NEW_ADMIN_PASSWORD  password the suite sets during the change password case
BASE=${E2E_BASE_URL:-http://localhost:8080}
DB_URL=${E2E_DATABASE_URL:-postgres://hst:hst@localhost:5432/hst?sslmode=disable}
PASS=0; FAIL=0; SECTION=""
PSQL="psql $DB_URL -tAc"

section() { SECTION="$1"; printf '\n\033[1m== %s\033[0m\n' "$1"; }

# t <name> <want_status> <want_code|-> <curl args...>
t() {
  local name="$1" want_st="$2" want_code="$3"; shift 3
  local out st body code
  out=$(curl -s -o /tmp/e2e_body -w '%{http_code}' "$@")
  st="$out"; body=$(cat /tmp/e2e_body)
  code=$(printf '%s' "$body" | python3 -c 'import sys,json
try: print(json.load(sys.stdin).get("code","-"))
except Exception: print("-")' 2>/dev/null)

  local ok=1
  [ "$st" = "$want_st" ] || ok=0
  [ "$want_code" = "-" ] || [ "$code" = "$want_code" ] || ok=0

  if [ $ok -eq 1 ]; then
    PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s %s/%s\n' "$name" "$st" "$code"
  else
    FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s got %s/%s want %s/%s\n' "$name" "$st" "$code" "$want_st" "$want_code"
    printf '        body: %.150s\n' "$body"
  fi
}

login() { # login <user> <pass> <conntype> -> prints json
  curl -s -u "$1:$2" -X POST "$BASE/auth/v1/oauth2/login" \
    -H 'Content-Type: application/json' -d "{\"connection_type\":$3}"
}
field() { python3 -c "import sys,json;d=json.load(sys.stdin);print(d['data']['$1'] if d.get('data') else '')"; }

ADMIN_PW=${E2E_ADMIN_PASSWORD:-Bootstrap-Admin-2026!}
NEW_ADMIN_PW=${E2E_NEW_ADMIN_PASSWORD:-NewAdminPass-2026!}
ADMIN=$($PSQL "SELECT min(login) FROM hst.managers")

# stop early on a database a previous run already dirtied
if [ -z "$ADMIN" ]; then
  echo "ERROR: hst.managers is empty, boot the server once so it seeds the first admin" >&2
  exit 2
fi
DIRT=$($PSQL "SELECT (SELECT count(*) FROM hst.managers) - 1 + (SELECT count(*) FROM hst.clients) + (SELECT count(*) FROM hst.sessions)")
if [ "$DIRT" != "0" ]; then
  echo "ERROR: this database has already been used, the suite is not idempotent" >&2
  echo "       expected only the seeded admin, found $DIRT extra rows in managers, clients or sessions" >&2
  echo "       drop the database, run make migrate-up, restart the server, then try again" >&2
  exit 2
fi
echo "bootstrap admin login: $ADMIN"

############################################################
section "1. Health and liveness"
t "health"                       200 0 "$BASE/api/v1/system/monitor/health"
t "live"                         200 0 "$BASE/api/v1/system/monitor/live"
t "unknown endpoint"             404 - "$BASE/api/v1/does-not-exist"

############################################################
section "2. Login, happy path"
RES=$(login $ADMIN "$ADMIN_PW" 33)
AT=$(printf '%s' "$RES" | field access_token)
RT=$(printf '%s' "$RES" | field refresh_token)
SID=$(printf '%s' "$RES" | field session_id)
t "manager login ct=33"          200 1026 -u "$ADMIN:$ADMIN_PW" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "admin login ct=32"            200 1026 -u "$ADMIN:$ADMIN_PW" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":32}'
[ -n "$AT" ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s token issued\n' "access token returned"; } \
             || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s no token\n' "access token returned"; }

############################################################
section "3. Login failures (MT5 retcodes)"
t "wrong password -> 1001"       401 1001 -u "$ADMIN:nope" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "unknown login -> 1020"        401 1020 -u "8888888:nope" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "non numeric login -> 1020"    401 1020 -u "abc:nope" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "no basic auth header"         401 -    -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "trader terminal ct=11 -> 1024" 403 1024 -u "$ADMIN:$ADMIN_PW" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":11}'
t "missing connection_type"      400 -    -u "$ADMIN:$ADMIN_PW" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{}'
t "malformed json body"          400 -    -u "$ADMIN:$ADMIN_PW" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{oops'

############################################################
section "4. Bearer token validation"
t "valid token (restricted, so 1026)" 403 1026 -H "Authorization: Bearer $AT" "$BASE/api/v1/auth/me"
t "no token -> 60001"            401 60001 "$BASE/api/v1/auth/me"
t "garbage token"                401 60001 -H "Authorization: Bearer not.a.jwt" "$BASE/api/v1/auth/me"
t "wrong scheme"                 401 60001 -H "Authorization: Basic $AT" "$BASE/api/v1/auth/me"
TAMPER=$(python3 -c "t='$AT';i=len(t)-20;print(t[:i]+('B' if t[i]!='B' else 'C')+t[i+1:])")
t "tampered signature"           401 60001 -H "Authorization: Bearer $TAMPER" "$BASE/api/v1/auth/me"
t "token of another server"      401 60001 -H "Authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzaWQiOiJ4In0.AAAA" "$BASE/api/v1/auth/me"

############################################################
section "5. Restricted session (MT5 reset_pass, 1026)"
t "restricted blocked from /clients" 403 1026 -H "Authorization: Bearer $AT" "$BASE/api/v1/clients"
t "restricted blocked from /users"   403 1026 -H "Authorization: Bearer $AT" "$BASE/api/v1/users"
t "restricted refused on /me too"    403 1026 -H "Authorization: Bearer $AT" "$BASE/api/v1/auth/me"
t "change password, wrong old"       401 1001 -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/auth/oauth2/change-password" -d '{"old_password":"WRONG","new_password":"NewAdminPass-2026!"}'
t "change password, too short"       400 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/auth/oauth2/change-password" -d "{\"old_password\":\"$ADMIN_PW\",\"new_password\":\"abc\"}"
t "change password ok"               204 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/auth/oauth2/change-password" -d "{\"old_password\":\"$ADMIN_PW\",\"new_password\":\"$NEW_ADMIN_PW\"}"
t "old session dead after change"    401 60001 -H "Authorization: Bearer $AT" "$BASE/api/v1/auth/me"
t "old password rejected"            401 1001 -u "$ADMIN:$ADMIN_PW" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'

ADMIN_PW="$NEW_ADMIN_PW"
RES=$(login $ADMIN "$ADMIN_PW" 33)
AT=$(printf '%s' "$RES" | field access_token)
RT=$(printf '%s' "$RES" | field refresh_token)
t "login after change -> code 0"     200 0 -u "$ADMIN:$ADMIN_PW" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "no longer restricted"             200 0 -H "Authorization: Bearer $AT" "$BASE/api/v1/auth/me"

############################################################
section "6. Refresh rotation and reuse detection"
R2=$(curl -s -X POST "$BASE/auth/v1/oauth2/refresh" -H 'Content-Type: application/json' -d "{\"refresh_token\":\"$RT\"}")
AT2=$(printf '%s' "$R2" | field access_token)
RT2=$(printf '%s' "$R2" | field refresh_token)
t "rotate refresh token"         200 0    -X POST "$BASE/auth/v1/oauth2/refresh" -H 'Content-Type: application/json' -d "{\"refresh_token\":\"$RT2\"}"
t "invalid refresh token"        401 60001 -X POST "$BASE/auth/v1/oauth2/refresh" -H 'Content-Type: application/json' -d '{"refresh_token":"garbage"}'
t "missing refresh token"        400 -    -X POST "$BASE/auth/v1/oauth2/refresh" -H 'Content-Type: application/json' -d '{}'
t "REPLAY spent token"           401 60001 -X POST "$BASE/auth/v1/oauth2/refresh" -H 'Content-Type: application/json' -d "{\"refresh_token\":\"$RT\"}"
t "family killed after replay"   401 60001 -H "Authorization: Bearer $AT2" "$BASE/api/v1/auth/me"
REUSE=$($PSQL "SELECT count(*) FROM hst.sessions WHERE revoked_reason='reuse_detected'")
[ "$REUSE" -ge 1 ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s %s rows\n' "reuse_detected recorded in postgres" "$REUSE"; } \
                   || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s 0 rows\n' "reuse_detected recorded in postgres"; }

############################################################
section "7. Logout"
RES=$(login $ADMIN "$ADMIN_PW" 33); AT=$(printf '%s' "$RES" | field access_token)
t "logout"                       204 -    -H "Authorization: Bearer $AT" -X POST "$BASE/api/v1/auth/logout"
t "token dead after logout"      401 60001 -H "Authorization: Bearer $AT" "$BASE/api/v1/auth/me"

############################################################
section "8. Authorization by manager right"
RES=$(login $ADMIN "$ADMIN_PW" 33); AT=$(printf '%s' "$RES" | field access_token)
t "no right_clients_access"      403 -    -H "Authorization: Bearer $AT" "$BASE/api/v1/clients"
$PSQL "UPDATE hst.managers SET right_clients_access=1,right_clients_create=1,right_clients_edit=1,right_clients_delete=1,right_acc_read=1,right_acc_manager=1,right_acc_delete=1,right_cfg_managers=1 WHERE login=$ADMIN" >/dev/null
RES=$(login $ADMIN "$ADMIN_PW" 33); AT=$(printf '%s' "$RES" | field access_token)
t "granted right_clients_access" 200 0    -H "Authorization: Bearer $AT" "$BASE/api/v1/clients"

############################################################
section "9. Clients CRUD"
CID=$(curl -s -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/clients" \
  -d '{"person_name":"Alice Tester","contact_email":"alice@example.com","client_type":1}' | field client_id)
t "create client"                201 0    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/clients" -d '{"person_name":"Bob Tester","contact_email":"bob@example.com","client_type":1}'
t "create, missing name"         400 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/clients" -d '{"contact_email":"x@example.com"}'
t "create, bad email"            400 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/clients" -d '{"person_name":"X","contact_email":"not-an-email"}'
t "list clients"                 200 0    -H "Authorization: Bearer $AT" "$BASE/api/v1/clients?limit=10"
t "get client"                   200 0    -H "Authorization: Bearer $AT" "$BASE/api/v1/clients/$CID"
t "get missing client"           404 -    -H "Authorization: Bearer $AT" "$BASE/api/v1/clients/$ADMIN"
t "patch client"                 200 0    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X PATCH "$BASE/api/v1/clients/$CID" -d '{"contact_phone":"+15550100","address_city":"Dubai"}'
t "patch missing client"         404 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X PATCH "$BASE/api/v1/clients/$ADMIN" -d '{"address_city":"X"}'

############################################################
section "10. Users, accounts and managers in one transaction"
LOGIN=$(curl -s -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" \
  -d "{\"client_id\":$CID,\"group\":\"real\\\\standard\",\"rights\":1,\"name\":\"Trader A\",\"email\":\"ta@example.com\",\"leverage\":100,\"password_main\":\"TraderAPass-2026!\",\"password_investor\":\"TraderAInv-2026!\"}" | field login)
t "create user"                  201 0    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" -d '{"group":"real\\standard","rights":1,"name":"Trader B","email":"tb@example.com","leverage":100,"password_main":"TraderBPass-2026!","password_investor":"TraderBInv-2026!"}'
t "create user (no manager block)" 201 0   -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" -d '{"group":"real\\standard","rights":1,"name":"Mgr C","email":"mc@example.com","leverage":100,"password_main":"MgrCPass-2026!","password_investor":"MgrCInv-2026!"}'
t "create user, no investor password"  400 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" -d '{"group":"real\\standard","name":"NoInv","email":"noinv@example.com","leverage":100,"password_main":"NoInvPass-2026!"}'
t "create user, short password"  400 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" -d '{"group":"real\\standard","name":"X","email":"x2@example.com","leverage":100,"password_main":"abc"}'
t "create user, missing group"   400 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" -d '{"name":"X","email":"x3@example.com","leverage":100,"password_main":"XPass-2026!!"}'
t "list users"                   200 0    -H "Authorization: Bearer $AT" "$BASE/api/v1/users?limit=10"
t "get user"                     200 0    -H "Authorization: Bearer $AT" "$BASE/api/v1/users/$LOGIN"
t "get missing user"             404 -    -H "Authorization: Bearer $AT" "$BASE/api/v1/users/12345678"
t "patch user"                   200 0    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X PATCH "$BASE/api/v1/users/$LOGIN" -d '{"city":"Dubai","leverage":200}'

ACC=$($PSQL "SELECT count(*) FROM hst.accounts WHERE login=$LOGIN")
[ "$ACC" = "1" ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s 1:1 account row\n' "user has exactly one accounts row"; } \
                 || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s got %s\n' "user has exactly one accounts row" "$ACC"; }

BEFORE=$($PSQL "SELECT count(*) FROM hst.users")
curl -s -o /dev/null -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" \
  -d '{"client_id":987654,"group":"real\\standard","name":"Ghost","email":"ghost@example.com","leverage":100,"password_main":"GhostPass-2026!","password_investor":"GhostInv-2026!"}'
AFTER=$($PSQL "SELECT count(*) FROM hst.users")
[ "$BEFORE" = "$AFTER" ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s rolled back cleanly\n' "failed create writes nothing"; } \
                         || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s %s -> %s\n' "failed create writes nothing" "$BEFORE" "$AFTER"; }

############################################################
section "11. Managers"
t "list managers"                200 0 -H "Authorization: Bearer $AT" "$BASE/api/v1/managers"
t "manager rights, 77 flags"     200 0 -H "Authorization: Bearer $AT" "$BASE/api/v1/managers/$ADMIN/rights"
NFLAGS=$(curl -s -H "Authorization: Bearer $AT" "$BASE/api/v1/managers/$ADMIN/rights" | python3 -c 'import sys,json;print(len(json.load(sys.stdin)["data"]["rights"]))')
[ "$NFLAGS" = "77" ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s 77 flags\n' "rights bitset decodes to 77"; } \
                     || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s got %s\n' "rights bitset decodes to 77" "$NFLAGS"; }
t "rights of a non manager"      404 - -H "Authorization: Bearer $AT" "$BASE/api/v1/managers/$LOGIN/rights"

############################################################
section "11b. Manager is a role attached to an existing login"
NEWU=$(curl -s -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" \
  -d '{"group":"real\\standard","rights":1,"name":"Promote Me","email":"promote@example.com","leverage":100,"password_main":"PromotePass-2026!","password_investor":"PromoteInv-2026!"}' | field login)
t "new user is not staff yet -> 1011"  403 1011 -u "$NEWU:PromotePass-2026!" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "promote to manager"                 201 0    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/managers" -d "{\"login\":$NEWU,\"name\":\"Promoted\",\"groups\":[\"real\\\\*\"],\"rights\":[\"right_manager\"]}"
t "same login can log in now"          200 0    -u "$NEWU:PromotePass-2026!" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
t "promote twice -> 409"               409 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/managers" -d "{\"login\":$NEWU,\"name\":\"Dup\",\"rights\":[\"right_manager\"]}"
t "promote a missing login -> 404"     404 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/managers" -d '{"login":98765432,"name":"Ghost","rights":["right_manager"]}'
t "promote with a bogus right -> 400"  400 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/managers" -d "{\"login\":$NEWU,\"name\":\"Evil\",\"rights\":[\"right_make_me_god\"]}"
t "update manager rights"              200 0    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X PATCH "$BASE/api/v1/managers/$NEWU" -d '{"name":"Promoted","rights":["right_manager","right_trades_read"]}'
t "update a missing manager -> 404"    404 -    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X PATCH "$BASE/api/v1/managers/98765432" -d '{"name":"X","rights":[]}'

# auth must react immediately when the role is revoked
PT=$(login $NEWU "PromotePass-2026!" 33 | field access_token)
t "promoted login has a working token" 200 0    -H "Authorization: Bearer $PT" "$BASE/api/v1/auth/me"
t "demote (delete manager row)"        204 -    -H "Authorization: Bearer $AT" -X DELETE "$BASE/api/v1/managers/$NEWU"
t "its session died at once"           401 60001 -H "Authorization: Bearer $PT" "$BASE/api/v1/auth/me"
t "and it can no longer log in"        403 1011 -u "$NEWU:PromotePass-2026!" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
USERSTILL=$($PSQL "SELECT count(*) FROM hst.users WHERE login=$NEWU")
[ "$USERSTILL" = "1" ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s user survived\n' "demote keeps the user and its account"; } \
                       || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s user gone\n' "demote keeps the user and its account"; }

############################################################
section "12. Query validation"
t "limit over the cap"           400 - -H "Authorization: Bearer $AT" "$BASE/api/v1/clients?limit=5000"
t "unknown sort_by"              400 - -H "Authorization: Bearer $AT" "$BASE/api/v1/clients?sort_by=created_at"
t "sql injection in sort_by"     400 - -H "Authorization: Bearer $AT" "$BASE/api/v1/clients?sort_by=1%3BDROP+TABLE+hst.clients"
t "valid sort_by"                200 0 -H "Authorization: Bearer $AT" "$BASE/api/v1/clients?sort_by=person_name&order=asc"
t "pagination page 2"            200 0 -H "Authorization: Bearer $AT" "$BASE/api/v1/clients?limit=1&page=2"
STILL=$($PSQL "SELECT count(*) FROM hst.clients")
[ "$STILL" -ge 1 ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s table intact\n' "clients table survived injection"; } \
                   || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s table gone\n' "clients table survived injection"; }

############################################################
section "13. Non manager login and rights revocation"
NM=$(curl -s -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -X POST "$BASE/api/v1/users" \
  -d '{"group":"real\\standard","rights":1,"name":"Plain","email":"plain@example.com","leverage":100,"password_main":"PlainPass-2026!","password_investor":"PlainInv-2026!"}' | field login)
t "user without managers row -> 1011" 403 1011 -u "$NM:PlainPass-2026!" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
$PSQL "UPDATE hst.users SET rights=0 WHERE login=$NM" >/dev/null
t "disabled account -> 1002"     403 1002 -u "$NM:PlainPass-2026!" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'

############################################################
section "14. Account lockout"
for i in 1 2 3 4 5 6 7 8 9 10 11; do
  curl -s -o /dev/null -u "$LOGIN:wrongpass" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'
done
t "locked after 10 failures"     403 60002 -u "$LOGIN:TraderAPass-2026!" -X POST "$BASE/auth/v1/oauth2/login" -H 'Content-Type: application/json' -d '{"connection_type":33}'

############################################################
section "15. Deletion rules"
t "delete client with users -> 409" 409 - -H "Authorization: Bearer $AT" -X DELETE "$BASE/api/v1/clients/$CID"
t "delete client force"          204 - -H "Authorization: Bearer $AT" -X DELETE "$BASE/api/v1/clients/$CID?force=true"
t "delete user"                  204 - -H "Authorization: Bearer $AT" -X DELETE "$BASE/api/v1/users/$LOGIN"
t "delete missing user"          404 - -H "Authorization: Bearer $AT" -X DELETE "$BASE/api/v1/users/$LOGIN"
ORPH=$($PSQL "SELECT count(*) FROM hst.accounts a WHERE NOT EXISTS (SELECT 1 FROM hst.users u WHERE u.login=a.login)")
[ "$ORPH" = "0" ] && { PASS=$((PASS+1)); printf '  \033[32mPASS\033[0m %-52s none\n' "no orphaned accounts after delete"; } \
                  || { FAIL=$((FAIL+1)); printf '  \033[31mFAIL\033[0m %-52s %s\n' "no orphaned accounts after delete" "$ORPH"; }

############################################################
section "16. Cache monitor"
t "cache stats"                  200 0 -H "Authorization: Bearer $AT" "$BASE/api/v1/system/monitor/cache"

printf '\n\033[1mTOTAL  pass %d  fail %d\033[0m\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
