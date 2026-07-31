# hst-news

News ingestion microservice for hst-v2. Polls enabled news datafeeds from hst-server config, normalizes items, caches them in Redis, and publishes on NATS.

## Phase 1 scope

- **RSSNewsFeeder** connector (HTTP RSS/XML poll)
- Config from hst-server internal API + NATS config snapshots (no Postgres)
- Dedup + hot cache in Redis
- NATS publish: `hstnews.item.{datafeed_id}`
- NATS subscribe: `hstserver.datafeed.{created,updated,deleted}` and `hstserver.datafeed.config.>`

## Run locally

```sh
cp .env.example .env
Health probes: `http://localhost:8082/healthz`, `http://localhost:8082/readyz` (8082 in `.env` so hst-quote can use 8081 on the same machine).

Shared infra: run `make up` in **hst-server** (postgres, nats, redis). hst-news does not need its own stack if those are already up.

## Configure a feed (via hst-server)

Supported **`module`** values depend on **`mode`**. List options:

- Quotes: `GET /api/v1/datafeeds/modules?mode=1` → `fix44`, `fix43`, `QuoteSimulator`
- News: `GET /api/v1/datafeeds/modules?mode=2` → `RSSNewsFeeder`

Create a datafeed with `mode: 2` (news), `module: "RSSNewsFeeder"`, and `feed_server` set to the RSS URL. Optional params:

| Param | Purpose |
|-------|---------|
| News Category | MT5-style category prefix |
| News Request Period | Poll interval in seconds (default 300) |
| Language | Item language tag |

Activate with `POST /api/v1/datafeeds/{id}/activate` or set `enable: 1` on create.

## NATS subjects

| Subject | Direction |
|---------|-----------|
| `hstserver.datafeed.created` | hst-server → hst-news |
| `hstserver.datafeed.updated` | hst-server → hst-news |
| `hstserver.datafeed.deleted` | hst-server → hst-news |
| `hstnews.item.{datafeed_id}` | hst-news → consumers |

## Redis keys

| Key | Purpose |
|-----|---------|
| `hstnews:seen:{datafeed_id}:{item_id}` | Dedup marker (7d TTL) |
| `hstnews:feed:{datafeed_id}` | Latest 100 items JSON (24h TTL) |
