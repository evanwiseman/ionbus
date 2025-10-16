package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanwiseman/ionbus/internal/broker"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/evanwiseman/ionbus/internal/routing"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	envFile = "docker/server/.env" // Change to ".env" for docker deployment
)

func cleanup() {
	log.Println("Stopping ionbus server...")
}

func main() {
	// Listen for OS signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %s, shutting down...\n", sig)
		cancel()
	}()

	// Run main program logic
	run(ctx)

	// Perform cleanup after run returns
	cleanup()
}

func run(ctx context.Context) {
	log.Println("Starting ionbus server...")

	// Load from .env update
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("Failed to get server config: %v\n", err)
	}

	// ========================
	// Start RabbitMQ
	// ========================
	log.Println("Connecting to RabbitMQ...")
	rmqConn, err := broker.StartRMQ(cfg.RMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer rmqConn.Close()
	log.Println("Sucessfully connected to RabbitMQ")

	// ========================
	// Open Postgres Database
	// ========================
	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", cfg.DB.GetUrl())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer db.Close()
	log.Println("Successfully connected to database")

	// ========================
	// Test Bridge
	// ========================
	rmqCh, err := rmqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v\n", err)
	}

	rmqCh.ExchangeDeclare("telemetry", "topic", true, false, false, false, nil)
	rmqCh.QueueDeclare("telemetry.outbound", true, false, false, false, nil)
	rmqCh.QueueBind("telemetry.outbound", "telemetry.#.outbound", "telemetry", false, nil)

	pubsub.SubscribeRMQ(
		ctx,
		rmqCh,
		pubsub.RMQSubscribeOptions{QueueName: "telemetry.outbound"},
		routing.ContentJSON,
		func(msg any) routing.AckType {
			log.Printf("Received telemetry: %s\n", msg)
			return routing.Ack
		},
	)

	// ========================
	// Wait for cancellation
	// ========================
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
