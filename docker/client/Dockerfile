FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o client ./cmd/client

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/client .
COPY cmd/client/.env .env
CMD ["./client"]
