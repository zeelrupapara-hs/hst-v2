FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /app

RUN apk --update --no-cache add ca-certificates git

# the module files alone, so the dependency layer survives a source change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -trimpath -ldflags="-s -w" -o /out/hstnews ./cmd

FROM alpine:latest

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /out/hstnews /hstnews

# never run the service as root
RUN adduser -D -u 10001 hstnews
USER hstnews

ENTRYPOINT ["/hstnews"]
CMD []
