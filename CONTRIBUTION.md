# Contributing

## Repo layout

```
hst-v2/
└── hst-server/          API server (Fiber + pgx + NATS)
    ├── cmd/             entry point
    ├── app/             wiring and startup
    ├── config/          env config
    ├── internal/
    │   ├── middleware/  request logger, auth
    │   └── server/v1/   handlers and routes
    ├── pkg/             db, http, logger, nats, cache, crypto, jwt, oauth2
    ├── migrations/      <epoch-millis>_<name>.{up,down}.sql
    ├── seed/            starting rows for a fresh install
    ├── swagger/         generated, do not edit by hand
    └── docs/            MT5 data model (gitignored, except docs/infra which is committed)
```

All commands run from `hst-server/`.

## First time

```bash
cd hst-server
make tools     # migrate, staticcheck, errcheck, gosec, govulncheck
cp .env.example .env
make gen-keys  # copy the printed seed into AUTH_JWT_PRIVATE_KEY in .env
```

`make` loads `.env`, so `make run` picks it up. Values must not contain a `#`.

`FIRST_MANAGER_PASSWORD` is already set in the example. On first boot, while `hst.managers`
is empty, the server creates the administrator defined in `seed/manager.go`
with that password: login **1000**, **First Admin**.

The account is ready to use straight away. Change the password when you want
to, it is not forced.

To add a seed: write `seed/<subject>.go` holding the values as a Go struct and
a `Seed<Subject>` function, then call it from `Seed()` in `seed/seed.go`. Each
one checks for itself whether it is needed, so a rerun is a no-op.

## Start the server

```bash
make up            # postgres, nats, redis
make migrate-up    # apply migrations
make run           # http://localhost:8080
```

Check it: `curl localhost:8080/api/v1/system/monitor/health`

## API docs

```bash
make swagger    # regenerate from the handler annotations
```

Then open http://localhost:8080/swagger/index.html

Two security schemes are defined: `BasicAuth` on `/auth/v1/oauth2/login`, and
`BearerAuth` everywhere else. In the UI click **Authorize**, and for
`BearerAuth` enter `Bearer <access_token>` including the word Bearer.

`swagger/` is generated output and committed, because `cmd/main.go` imports it.
Never edit it by hand; change the annotations and re-run `make swagger`.

## Migrations

```bash
make migrate-create name=symbols   # new up/down pair
make migrate-up                    # apply pending
make migrate-down                  # roll back 1  (n=3 for more)
make migrate-version               # current version
make migrate-force v=<version>     # clear dirty flag after a failure
make migrate-drop                  # drop everything, destructive
```

Versions are epoch milliseconds, not sequential, so parallel branches never
collide. Always write the `down` file — an unreversible migration blocks
rollback.

## Before you push

```bash
make check      # fmt, vet, staticcheck, errcheck, gosec, govulncheck
make swagger    # if you touched a handler annotation
```

`make check` must be clean. gosec and govulncheck are expected to report zero.
CI runs it on every push.

This project carries no test files of any kind. Verify by running the server
and exercising it by hand.

Deploying to Kubernetes: see `hst-server/docs/infra/server/README.md`.

## Other commands

```bash
make build    # binary into bin/
make logs     # follow container logs
make psql     # open a psql shell
make down     # stop containers
```

## Conventions

- Handlers are methods on `*HttpServer` in `internal/server/v1/`, routes
  registered in `routes.go`.
- Every handler carries the full swagger block: `@Id`, `@Tags`, `@Accept`,
  `@Produce`, `@Success`, one `@Failure` per status it can return, `@Security`
  and `@Router`. Failure bodies are `ErrorResponse`.
- Request and response types follow `CrtX`, `UptX`, `ViewX`.
- Credentials are parsed by `BasicAuthParser` into Locals; handlers never read
  the Authorization header directly.
- **The auth middleware never reads Postgres.** A session miss goes to Redis and
  then to 401 (refresh) or 503 (store down) — never to SQL. Only `Login` and
  `RefreshToken` may read the database for auth. Adding a query to `Protect`
  breaks the guarantee that a request costs zero database round trips.
- Route rights come from the packed manager bitset:
  `s.Middleware.Authorization(model.MgrRightClientsCreate)`. The bit constants in
  `model/manager.go` follow the column order of the managers migration.
- Any handler that changes `rights`, `group` or a password must call
  `s.OAuth2.InvalidateLogin` after commit, or the change won't reach live sessions.
- Comments are one line, one sentence, simple words. Never a paragraph. If the
  reasoning needs more room it belongs in the commit message, not the source.
- Log through `Log.Journal(type, code, msg, kv...)` using the MT5 codes in
  `pkg/logger` — type 1 Cfg, 3 Net, 5 User, 6 Trade; code 0 OK, 1 Warn, 2 Err,
  3 Critical, 4 Login.
- Audited actions must record `actor` (the manager) and `target` (the account).
- Never log passwords, tokens or full request bodies.

## Logs

Written to `hst-server/logs/YYYYMMDD.log`, one file per day, JSON. Console
output is human readable.
