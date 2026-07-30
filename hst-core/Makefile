BIN      := bin/hstcore
STACK    := stack.yaml
GOBIN    := $(shell go env GOPATH)/bin
REGISTRY := registry.gitlab.com/hscore/hstcore

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

DB_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSL_MODE)

.PHONY: run build test tidy fmt vet check tools up down logs psql docker \
        push-dev push-staging push-prod

run:
	go run ./cmd

build:
	go build -o $(BIN) ./cmd

test:
	go test ./... -race -cover

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

# security and quality gate
check: fmt vet
	$(GOBIN)/staticcheck ./...
	$(GOBIN)/errcheck ./...
	$(GOBIN)/gosec -quiet ./...
	$(GOBIN)/govulncheck ./...

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/kisielk/errcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

up:
	docker compose -f $(STACK) up -d

down:
	docker compose -f $(STACK) down

logs:
	docker compose -f $(STACK) logs -f

psql:
	psql "$(DB_URL)"

docker:
	docker build -t $(REGISTRY):dev .

# multi arch, because the build hosts and the nodes do not agree
push-dev:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(REGISTRY):dev --push .

push-staging:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(REGISTRY):staging --push .

push-prod:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(REGISTRY):prod --push .
