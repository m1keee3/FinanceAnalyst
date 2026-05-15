FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/gateway ./cmd/gateway

FROM alpine:latest

RUN adduser -D -g "" gateway \
    && apk add --no-cache ca-certificates

USER gateway
WORKDIR /app

COPY --from=builder /app/bin/gateway ./gateway
COPY config ./config

EXPOSE 8090

ENTRYPOINT ["/app/gateway"]
