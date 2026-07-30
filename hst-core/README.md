# hstcore

The microservice template. An independent Go module following the same
conventions as `hst-server`. Copy this folder, rename the module, and start
writing the service.

Nothing domain specific is here. What is here is the plumbing every service
needs — config, logging, postgres, nats, redis, a worker pool, and a lifecycle
that shuts down cleanly — plus the build, the quality gate and the local stack.

## Layout

```
hst-core/
  cmd/          entry point, nothing but app.Start()
  app/          boots config, logger, postgres, nats, redis, starts the handler
  config/       every env var, read once and validated at boot
  handler/
    handler.go  the service: one struct holding every dependency, New/Start/Stop
    nats.go     every subject and queue group, as constants
  internal/
    health/     /healthz and /readyz, on their own port
    worker/     fixed goroutine pool over one job queue
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

`model/`, `migrations/`, `utils/`, `proto/` and the rest of `internal/` belong
to whatever service is built from this, not to the template. Follow
`hst-server` for the shape of each.

## The handler

`app.go` builds the world and hands it to `handler.New`. Everything the service
does hangs off that one struct, so a method never reaches for a global and a
test can build a `Handler` with fakes. Add a file per concern beside
`handler.go` and keep `handler.go` to the lifecycle.

```go
h := handler.New(cfg, log, database, natsClient, redisClient)
if err := h.Start(context.Background()); err != nil { ... }
defer h.Stop()
```

`Start` does three things in an order that matters: load state into memory,
start the workers, subscribe last — so no message arrives before the state it
reads. `Stop` reverses it and is safe to call twice.

Rules the lifecycle holds to, each one a failure mode worth avoiding:

- **A failed subscribe returns an error, it does not `Fatal`.** A service that
  cannot hear its own subject should refuse to boot, not run deaf, and the
  caller decides which.
- **Subscriptions are recorded and unsubscribed on `Stop`,** so shutdown does
  not leave a consumer attached to a connection that is about to drain.
- **`h.Go(f)` instead of a bare `go f()`,** so `Stop` can wait for it.
- **Nothing panics or calls `os.Exit` to shut down.** The signal is handled in
  `app.go` and every defer runs.

## Probes and shutdown

`internal/health` serves two endpoints on their own port (`HEALTH_PORT`, 8081
by default), so a probe never queues behind application traffic and the port
can stay off the public service.

- **`/healthz` checks nothing.** It answers as long as the process runs. Fail
  it and kubernetes restarts the pod — so wiring a dependency check in here is
  how one database blip restarts every replica at once.
- **`/readyz` checks everything:** boot finished, shutdown not started, and
  every registered dependency answering within 2 seconds. Fail it and traffic
  is routed away, the pod is left alone.

Shutdown fails readiness *first*, waits `HEALTH_DRAIN_WAIT`, and only then
unwinds. Without that pause the endpoints controller is still sending traffic
to a pod that has already stopped answering — the usual source of 502s during
a rolling deploy. Set the wait to at least twice the readiness probe period.

```yaml
livenessProbe:
  httpGet: { path: /healthz, port: 8081 }
  periodSeconds: 10
readinessProbe:
  httpGet: { path: /readyz, port: 8081 }
  periodSeconds: 2
```

## The worker pool

`internal/worker` is a fixed pool of goroutines over one job channel. A
goroutine per message is fine until traffic spikes and the process is holding a
hundred thousand of them; a fixed pool bounds that. The pool defaults to
`runtime.NumCPU()` workers with a queue 64 times deeper.

```go
h.Workers.Submit(func(ctx context.Context) {
    // the job gets the pool's context, so it can notice shutdown
})
```

- **`Submit` returns false on a full queue instead of blocking.** The caller is
  usually a nats callback that must not stall. A false return is a real signal
  — log it, count it, shed load. `Pending()` is the metric to export: it is the
  first number that moves when the service falls behind.
- **A panicking job is recovered,** so one bad message does not take the pool
  down.
- **`Stop` drains the queue.** An accepted job is a promise; queued work runs.

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
