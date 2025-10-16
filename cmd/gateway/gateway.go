package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanwiseman/ionbus/internal/broker"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/routing"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
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
	mqttClient, err := broker.StartMQTT(cfg.MQTT, cfg.ID)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", err)
	}
	defer mqttClient.Disconnect(250)
	log.Println("Successfully connected to MQTT")

	// ========================
	// Start RabbitMQ
	// ========================
	rmqConn, err := broker.StartRMQ(cfg.RMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer rmqConn.Close()
	log.Println("Sucessfully connected to RabbitMQ")

	// ========================
	// Commands
	// ========================
	commandsCh := openChannel(rmqConn)
	declareCommandsExchange(commandsCh)
	_ = declareAndBindCommandsQueue(commandsCh, cfg.ID)

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

func openChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v\n", err)
	}
	log.Println("Successfully opened channel for 'commands' on RabbitMQ")
	return ch
}

func declareCommandsExchange(ch *amqp.Channel) {
	err := broker.DeclareCommandExchange(ch)
	if err != nil {
		log.Fatalf("Failed to create commands exchange: %v\n", err)
	}
	log.Println("Successfully created commands exchange")
}

func declareAndBindCommandsQueue(ch *amqp.Channel, gatewayID string) amqp.Queue {
	queueName := fmt.Sprintf("%v.%v", routing.CommandsKey, gatewayID)
	routingKey := fmt.Sprintf("%v.%v.#", routing.CommandsKey, gatewayID)
	q, err := broker.DeclareAndBindQueue(
		ch,
		routing.ExchangeCommandsTopic,
		queueName,
		routing.Transient,
		routingKey,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/bind commands queue: %v\n", err)
	}
	log.Println("Successfully created/binded commands queue:", queueName)
	return q
}
