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
    ├── pkg/             db, http, logger, nats
    ├── migrations/      <epoch-millis>_<name>.{up,down}.sql
    └── docs/            MT5 data model and design docs
```

All commands run from `hst-server/`.

## First time

```bash
cd hst-server
make tools     # migrate, staticcheck, errcheck, gosec, govulncheck
cp .env.example .env
```

## Start the server

```bash
make up            # postgres, nats, redis
make migrate-up    # apply migrations
make run           # http://localhost:8080
```

Check it: `curl localhost:8080/api/v1/system/monitor/health`

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
make check    # fmt, vet, staticcheck, errcheck, gosec, govulncheck
make test
```

`make check` must be clean. gosec and govulncheck are expected to report zero.

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
- Comments are one or two lines. Explain why, not what.
- Log through `Log.Journal(type, code, msg, kv...)` using the MT5 codes in
  `pkg/logger` — type 1 Cfg, 3 Net, 5 User, 6 Trade; code 0 OK, 1 Warn, 2 Err,
  3 Critical, 4 Login.
- Audited actions must record `actor` (the manager) and `target` (the account).
- Never log passwords, tokens or full request bodies.

## Logs

Written to `hst-server/logs/YYYYMMDD.log`, one file per day, JSON. Console
output is human readable.
