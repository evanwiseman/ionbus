package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanwiseman/ionbus/internal/broker"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/joho/godotenv"
)

const (
	envFile = "docker/gateway/.env" // Change to ".env" for docker deployment
)

func cleanup() {
	log.Print("Stopping ionbus gateway... ")
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
	log.Println("Starting ionbus gateway...")

	// Load from .env update
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	cfg, err := config.LoadGatewayConfig()
	if err != nil {
		log.Fatalf("Failed to get gateway config: %v\n", err)
	}

	// ========================
	// Start MQTT
	// ========================
	log.Println("Connecting to MQTT...")
	mqttClient, err := broker.StartMQTT(cfg.MQTT, cfg.ID)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", err)
	}
	defer mqttClient.Disconnect(250)
	log.Println("Successfully connected to MQTT")

	// ========================
	// Start RabbitMQ
	// ========================
	log.Println("Connecting to RabbitMQ...")
	rabbitConn, err := broker.StartRMQ(cfg.RMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer rabbitConn.Close()
	log.Println("Sucessfully connected to RabbitMQ")

	// ========================
	// Subscribe to MQTT Topics
	// ========================

	// ========================
	// Confirm gateway is started
	// ========================
	log.Println("Successfully started ionbus gateway")

	// ========================
	// Wait for cancellation
	// ========================
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
