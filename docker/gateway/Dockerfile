FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gateway ./cmd/gateway

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/gateway .
COPY cmd/gateway/.env .env
CMD ["./gateway"]
