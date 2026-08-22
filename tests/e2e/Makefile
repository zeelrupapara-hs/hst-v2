ROOT := $(abspath ../..)
LOGS := $(CURDIR)/logs
BIN  := $(CURDIR)/bin

.PHONY: up migrate run-services wait test stop test-e2e

up:
	$(MAKE) -C $(ROOT)/hst-server up

migrate:
	$(MAKE) -C $(ROOT)/hst-server migrate-up

# each service reads its own .env, then the overrides; pids land in bin/ so stop can find them
run-services:
	mkdir -p $(LOGS) $(BIN)
	cd $(ROOT)/hst-server && go build -o $(BIN)/hst-server ./cmd
	cd $(ROOT)/hst-core && go build -o $(BIN)/hst-core ./cmd
	cd $(ROOT)/hst-quote && go build -o $(BIN)/hst-quote ./cmd
	cd $(ROOT)/hst-news && go build -o $(BIN)/hst-news ./cmd
	cd $(ROOT)/hst-server && set -a && . ./.env && set +a && nohup $(BIN)/hst-server > $(LOGS)/hst-server.log 2>&1 & echo $$! > $(BIN)/hst-server.pid
	cd $(ROOT)/hst-core && set -a && . ./.env && set +a && POD_NAME=core-a nohup $(BIN)/hst-core > $(LOGS)/hst-core.log 2>&1 & echo $$! > $(BIN)/hst-core.pid
	cd $(ROOT)/hst-quote && set -a && . ./.env && set +a && nohup $(BIN)/hst-quote > $(LOGS)/hst-quote.log 2>&1 & echo $$! > $(BIN)/hst-quote.pid
	cd $(ROOT)/hst-news && set -a && . ./.env && set +a && nohup $(BIN)/hst-news > $(LOGS)/hst-news.log 2>&1 & echo $$! > $(BIN)/hst-news.pid

wait:
	@for i in $$(seq 1 60); do \
		curl -sf localhost:8080/api/v1/system/monitor/health > /dev/null && curl -sf localhost:8091/readyz > /dev/null && curl -sf localhost:8081/readyz > /dev/null && exit 0; \
		sleep 2; \
	done; echo "services not ready"; exit 1

test:
	go test -tags e2e -count=1 -p 1 -timeout 40m ./...

stop:
	-for p in $(BIN)/*.pid; do kill $$(cat $$p) 2>/dev/null; rm -f $$p; done

test-e2e: up migrate run-services wait test stop
