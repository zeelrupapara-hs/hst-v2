FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /app

RUN apk --update --no-cache add ca-certificates git

# the build context is the repo root, because the wire contract is a sibling module
COPY hst-model/go.mod ./hst-model/
COPY hst-news/go.mod hst-news/go.sum ./hst-news/
RUN go mod download -C hst-news

COPY hst-model ./hst-model
COPY hst-news ./hst-news

RUN go build -C hst-news -trimpath -ldflags="-s -w" -o /out/hstnews ./cmd

FROM alpine:latest

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /out/hstnews /hstnews

# never run the service as root
RUN adduser -D -u 10001 hstnews
USER hstnews

ENTRYPOINT ["/hstnews"]
CMD []
