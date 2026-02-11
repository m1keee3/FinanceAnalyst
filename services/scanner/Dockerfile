FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/scanner ./cmd/scanner

FROM alpine:latest

RUN adduser -D -g "" scanner \
    && apk add --no-cache ca-certificates

USER scanner
WORKDIR /app

COPY --from=builder /app/bin/scanner ./scanner
COPY config ./config

EXPOSE 50050

ENTRYPOINT ["/app/scanner"]
