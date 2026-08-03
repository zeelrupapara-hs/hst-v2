FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /app

RUN apk --update --no-cache add ca-certificates git

# the build context is the repo root, because the wire contract is a sibling module
COPY hst-model/go.mod ./hst-model/
COPY hst-quote/go.mod hst-quote/go.sum ./hst-quote/
RUN go mod download -C hst-quote

COPY hst-model ./hst-model
COPY hst-quote ./hst-quote

RUN go build -C hst-quote -trimpath -ldflags="-s -w" -o /out/hstquote ./cmd

FROM alpine:latest

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /out/hstquote /hstquote

# never run the service as root
RUN adduser -D -u 10001 hstquote
USER hstquote

ENTRYPOINT ["/hstquote"]
CMD []
