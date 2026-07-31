BIN        := bin/hstserver
MIGRATIONS := migrations
STACK      := stack.yaml
GOBIN      := $(shell go env GOPATH)/bin

# load .env when it is there, so make run works after cp .env.example .env
ifneq (,$(wildcard .env))
include .env
export
endif

# override from the environment or .env
POSTGRES_USER     ?= hst
POSTGRES_PASSWORD ?= hst
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
POSTGRES_DB       ?= hst
POSTGRES_SSL_MODE ?= disable
# search_path is pinned: the role is also named hst, so once the hst schema
# exists "$$user" resolves to it and migrate would look for schema_migrations
# in the wrong schema, losing all version state.
DB_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSL_MODE)&search_path=public

# psql rejects search_path, which only golang-migrate understands
PSQL_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSL_MODE)

MIGRATE := $(GOBIN)/migrate

.PHONY: run build test tidy fmt vet check up down logs psql tools gen-keys swagger \
        migrate-create migrate-up migrate-down migrate-version migrate-force migrate-drop

run: swagger
	go run ./cmd

build: swagger
	go build -o $(BIN) ./cmd

test: swagger
	go test ./... -race -cover

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

# security and quality gate
check: swagger fmt vet
	$(GOBIN)/staticcheck ./...
	$(GOBIN)/errcheck ./...
	$(GOBIN)/gosec -quiet ./...
	$(GOBIN)/govulncheck ./...

# regenerates swagger/ from handler annotations (gitignored; run before build/check).
# one spec per panel, so neither audience is shown the other's API
swagger:
	@$(GOBIN)/swag fmt -g cmd/main.go
	@$(GOBIN)/swag init -g cmd/main.go -o swagger/admin --instanceName admin \
		--packageName swaggeradmin --tags '!Trader' --parseDependency --parseInternal
	@$(GOBIN)/swag init -g cmd/main.go -o swagger/trader --instanceName trader \
		--packageName swaggertrader --tags 'Trader,Auth' --parseDependency --parseInternal

# prints an ed25519 key pair; the seed goes into AUTH_JWT_PRIVATE_KEY
gen-keys:
	@go run ./cmd/genkeys

tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/kisielk/errcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/swaggo/swag/cmd/swag@v1.16.6

up:
	docker compose -f $(STACK) up -d

down:
	docker compose -f $(STACK) down

logs:
	docker compose -f $(STACK) logs -f

psql:
	psql "$(DB_URL)"

# make migrate-create name=symbols
# version is epoch milliseconds so parallel branches never collide
migrate-create:
ifndef name
	$(error usage: make migrate-create name=your_migration_name)
endif
	@mkdir -p $(MIGRATIONS)
	@ts=$$(python3 -c 'import time;print(int(time.time()*1000))'); \
	for d in up down; do \
		f=$(MIGRATIONS)/$${ts}_$(name).$$d.sql; \
		: > $$f; echo "$$f"; \
	done

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS) -database "$(DB_URL)" up

# one step back, or make migrate-down n=3
migrate-down:
	$(MIGRATE) -path $(MIGRATIONS) -database "$(DB_URL)" down $(or $(n),1)

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS) -database "$(DB_URL)" version

# clears the dirty flag after a failed migration: make migrate-force v=2
migrate-force:
ifndef v
	$(error usage: make migrate-force v=version)
endif
	$(MIGRATE) -path $(MIGRATIONS) -database "$(DB_URL)" force $(v)

# migrate's own drop only clears the search_path, so the hst schema survives it
migrate-drop:
	@psql "$(PSQL_URL)" -q -c 'DROP SCHEMA IF EXISTS hst CASCADE'
	@psql "$(PSQL_URL)" -q -c 'DROP TABLE IF EXISTS public.schema_migrations'
	@echo "dropped the hst schema and the migration state"
