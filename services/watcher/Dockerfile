FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/watcher ./cmd/watcher

FROM alpine:latest

RUN adduser -D -g "" watcher \
    && apk add --no-cache ca-certificates

USER watcher
WORKDIR /app

COPY --from=builder /app/bin/watcher ./watcher
COPY config ./config

EXPOSE 50052

ENTRYPOINT ["/app/watcher"]