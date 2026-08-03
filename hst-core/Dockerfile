FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /app

RUN apk --update --no-cache add ca-certificates git

# the build context is the repo root, because the wire contract is a sibling module
COPY hst-model/go.mod ./hst-model/
COPY hst-core/go.mod hst-core/go.sum ./hst-core/
RUN go mod download -C hst-core

COPY hst-model ./hst-model
COPY hst-core ./hst-core

RUN go build -C hst-core -trimpath -ldflags="-s -w" -o /out/hstcore ./cmd

FROM alpine:latest

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /out/hstcore /hstcore

# never run the service as root
RUN adduser -D -u 10001 hstcore
USER hstcore

ENTRYPOINT ["/hstcore"]
CMD []
