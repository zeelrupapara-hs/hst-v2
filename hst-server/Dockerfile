FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /app

RUN apk --update --no-cache add ca-certificates git

# the module files alone, so the dependency layer survives a source change
COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6 && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

COPY . .

RUN swag init -g cmd/main.go -o swagger/admin --instanceName admin \
        --packageName swaggeradmin --tags '!Trader' --parseDependency --parseInternal && \
    swag init -g cmd/main.go -o swagger/trader --instanceName trader \
        --packageName swaggertrader --tags 'Trader' --parseDependency --parseInternal

RUN go build -trimpath -ldflags="-s -w" -o /out/hstserver ./cmd

FROM alpine:latest

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /out/hstserver /hstserver
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/templates /templates
COPY entrypoint.sh /entrypoint.sh

# never run the service as root
RUN adduser -D -u 10001 hstserver && chmod +x /entrypoint.sh
USER hstserver

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
