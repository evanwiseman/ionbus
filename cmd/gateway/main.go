package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
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
	err := godotenv.Load("./cmd/gateway/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	cfg, err := LoadGatewayConfig()
	if err != nil {
		log.Fatalf("Failed to get gateway config: %v\n", err)
	}

	// ========================
	// Start MQTT
	// ========================
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTT.GetUrl())
	opts.SetKeepAlive(cfg.MQTT.KeepAlive)
	opts.SetCleanSession(cfg.MQTT.CleanSession)
	opts.SetClientID(cfg.ID)
	opts.SetUsername(cfg.MQTT.Username)
	opts.SetPassword(cfg.MQTT.Password)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", token.Error())
	}
	defer mqttClient.Disconnect(250)
	log.Println("Successfully connected to MQTT")

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
	// Setup RabbitMQ
	// ========================
	topicCh, err := rmqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel on RabbitMQ: %v", err)
	}
	defer topicCh.Close()

	if err := pubsub.DeclareIonbusTopic(topicCh); err != nil {
		log.Fatalf("Failed to create/open %s: %v", pubsub.ExchangeIonbusTopic, err)
	}

	gatewayCh, err := rmqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel on RabbitMQ: %v", err)
	}
	defer gatewayCh.Close()

	_, err = pubsub.DeclareAndBindQueue(
		gatewayCh,
		pubsub.ExchangeIonbusTopic,
		cfg.ID,
		pubsub.Durable,
		fmt.Sprintf("gateways.%s.#", cfg.ID),
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/open %s: %v", cfg.ID, err)
	}

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
