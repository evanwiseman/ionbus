FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/server .
COPY docker/server/.env .env
CMD ["./server"]
