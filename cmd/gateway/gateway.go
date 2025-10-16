package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanwiseman/ionbus/internal/bridge"
	"github.com/evanwiseman/ionbus/internal/broker"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/evanwiseman/ionbus/internal/routing"
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
	rmqConn, err := broker.StartRMQ(cfg.RMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer rmqConn.Close()
	log.Println("Sucessfully connected to RabbitMQ")

	// ========================
	// Test Bridge
	// ========================
	rmqCh, err := rmqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v\n", err)
	}
	b := bridge.Bridge{
		RMQCh:      rmqCh,
		MQTTClient: mqttClient,
	}

	// 1. Declare RabbitMQ infrastructure first
	err = rmqCh.ExchangeDeclare("telemetry", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to create exchange: %v\n", err)
	}
	_, err = rmqCh.QueueDeclare("telemetry.inbound", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare inbound queue: %v\n", err)
	}
	_, err = rmqCh.QueueDeclare("telemetry.outbound", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare outbound queue: %v\n", err)
	}
	err = rmqCh.QueueBind("telemetry.inbound", "telemetry.#.inbound", "telemetry", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind inbound queue: %v\n", err)
	}
	err = rmqCh.QueueBind("telemetry.outbound", "telemetry.#.outbound", "telemetry", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind outbound queue: %v\n", err)
	}

	// 2. Create bridge configuration
	err = b.BidirectionalBridge(
		ctx,
		// Inbound (server → gateway → device)
		pubsub.RMQSubscribeOptions{QueueName: "telemetry.inbound"},
		pubsub.MQTTPublishOptions{Topic: "device/+/telemetry/inbound"},

		// Outbound (device → gateway → server)
		pubsub.MQTTSubscribeOptions{Topic: "device/+/telemetry/outbound"},
		pubsub.RMQPublishOptions{Exchange: "telemetry", Key: "telemetry.outbound"},
		routing.ContentJSON,
	)
	if err != nil {
		log.Fatalf("Failed to bridge: %v\n", err)
	}

	// ========================
	// Subscribe to Topics/Queues
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
