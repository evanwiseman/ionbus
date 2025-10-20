package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	cfg, err := LoadServerConfig()
	if err != nil {
		log.Fatalf("Failed to get server config: %v\n", err)
	}

	// ========================
	// Start RabbitMQ
	// ========================
	rmqConn, err := amqp.Dial(cfg.RMQ.GetUrl())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer rmqConn.Close()
	log.Println("Sucessfully connected to RabbitMQ")

	// ========================
	// Setup RMQ Commands
	// ========================
	commandCh, err := rmqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v\n", err)
	}

	err = pubsub.DeclareCommandExchange(commandCh)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ command exchange: %v\n", err)
	}

	// ========================
	// Open Postgres Database
	// ========================
	db, err := sql.Open("postgres", cfg.DB.GetUrl())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer db.Close()
	log.Println("Successfully connected to database")

	// ========================
	// Wait for cancellation
	// ========================
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
