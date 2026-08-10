# Running the platform on a development machine.
#
# The services are started detached and write to $(LOGS). Infrastructure is left to brew, which
# already restarts it at login; only the platform's own services are ours to start.

BIN     := $(HOME)/.hstdev/bin
LOGS    := $(HOME)/.hstdev/logs
INFLUX  := $(HOME)/.hstdev/influx
SERVICES := hst-server hst-core hst-quote

.PHONY: dev up down status build influx

## dev: build everything and start it
dev: build up status

## build: compile every service
build:
	@mkdir -p $(BIN) $(LOGS)
	@for s in $(SERVICES); do \
		(cd $$s && go build -o $(BIN)/$$s ./cmd) && echo "  built $$s"; \
	done

## up: start the tick store and the services
up: influx
	@pgrep -f "$(BIN)/hst-server" >/dev/null 2>&1 || { \
		(cd hst-server && set -a && . ./.env && set +a && nohup $(BIN)/hst-server > $(LOGS)/hst-server.log 2>&1 &); \
		echo "  started hst-server"; \
	}
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
		curl -sf http://127.0.0.1:8080/auth/trader/v1/login -u '1:x' -H 'Content-Type: application/json' -d '{"connection_type":1}' >/dev/null 2>&1 && break; \
		sleep 1; \
	done
	@for s in hst-core hst-quote; do \
		pgrep -f "$(BIN)/$$s" >/dev/null 2>&1 && echo "  $$s already running" && continue; \
		(cd $$s && set -a && . ./.env && set +a && nohup $(BIN)/$$s > $(LOGS)/$$s.log 2>&1 &) ; \
		echo "  started $$s"; \
	done
	@sleep 4

## influx: the tick store, which the charts read and hst-quote writes
influx:
	@curl -sf http://localhost:8086/health >/dev/null 2>&1 && echo "  influx already running" || { \
		mkdir -p $(INFLUX); \
		(nohup $(shell brew --prefix influxdb@2)/bin/influxd --bolt-path $(INFLUX)/influxd.bolt \
			--engine-path $(INFLUX)/engine > $(LOGS)/influxd.log 2>&1 &) ; \
		echo "  started influx"; sleep 5; }

## down: stop the services, leaving the data where it is
down:
	@for s in $(SERVICES); do pkill -f "$(BIN)/$$s" 2>/dev/null && echo "  stopped $$s" || true; done

## status: what is up, and whether prices are actually moving
status:
	@echo "infrastructure:"
	@pg_isready -h localhost -p 5432 >/dev/null 2>&1 && echo "  postgres  up" || echo "  postgres  DOWN"
	@redis-cli -p 6379 ping >/dev/null 2>&1 && echo "  redis     up" || echo "  redis     DOWN"
	@nc -z localhost 4222 >/dev/null 2>&1 && echo "  nats      up" || echo "  nats      DOWN"
	@curl -sf http://localhost:8086/health >/dev/null 2>&1 && echo "  influx    up" || echo "  influx    DOWN"
	@echo "services:"
	@for s in $(SERVICES); do \
		pgrep -f "$(BIN)/$$s" >/dev/null 2>&1 && echo "  $$s up" || echo "  $$s DOWN"; \
	done
	@echo "prices:"
	@n=$$(timeout 4 nats sub 'hstquote.tick.*' 2>/dev/null | grep -c Received); \
	 [ "$$n" -gt 0 ] && echo "  $$n ticks in 4s, the feed is live" || echo "  NO TICKS, the feed is dead"
