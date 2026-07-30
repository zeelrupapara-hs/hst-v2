# hstcore

The microservice template. An independent Go module, laid out the same way
`vfxcore` is, on the same conventions as `hst-server`. Copy this folder,
rename the module, and start writing the service.

Nothing domain specific is here. What is here is the plumbing every service
needs — config, logging, postgres, nats, redis, graceful shutdown — plus the
build, the quality gate and the local stack.

## Layout

```
hst-core/
  cmd/          entry point, nothing but app.Start()
  app/          boots config, logger, postgres, nats, redis, then blocks
  config/       every env var, read once and validated at boot
  pkg/
    db/         pgx pool
    logger/     zap, one log file per day
    nats/       connection with reconnect and slow consumer handling
    redis/      client with optional tls
    errors/     the error vocabulary
  stack.yaml    local postgres, nats, redis
  Dockerfile    multi stage, non root
  Makefile
  .env.example
```

Everything else — `model/`, `internal/`, `migrations/`, `utils/`, `proto/` —
belongs to whatever service is built from this, not to the template. Follow
`hst-server` for the shape of each, so a change of service does not mean a
change of habits.

## Starting a new service from it

```sh
cp -R hst-core hst-<name>
cd hst-<name>
# rename the module and its imports
sed -i '' 's|hstcore|hst<name>|g' go.mod $(grep -rl hstcore --include='*.go' .)
cp .env.example .env
make up      # postgres, nats, redis
make run
```

Then drop whatever the service does not use: not every service needs all three
of postgres, nats and redis, and an unused connection is one more thing that
can fail at boot.

## Make targets

| Target | What it does |
| --- | --- |
| `run` / `build` | run from source / build `bin/hstcore` |
| `test` | the suite with the race detector and coverage |
| `fmt` `vet` `check` | `check` is the CI gate: gofmt, vet, staticcheck, errcheck, gosec, govulncheck |
| `tools` | installs what `check` needs |
| `up` `down` `logs` | the local stack in `stack.yaml` |
| `psql` | a shell on the configured database |
| `docker` `push-dev` `push-staging` `push-prod` | image build and multi arch push |

`app/app.go` has a marked spot where the server is built and started;
everything above it is plumbing that will not change.
