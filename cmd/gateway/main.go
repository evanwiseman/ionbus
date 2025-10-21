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

	// Load env
	cfg := LoadGatewayEnv()

	// Start MQTT
	mqttClient := ConnectMQTT(cfg)
	defer mqttClient.Disconnect(250)

	// Start RMQ
	rmqConn := ConnectRMQ(cfg)
	defer rmqConn.Close()

	// ========================
	// Setup RabbitMQ
	// ========================
	// Topic Exchange
	topicCh, err := rmqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel on RabbitMQ: %v", err)
	}
	defer topicCh.Close()

	if err := pubsub.DeclareIonbusTopic(topicCh); err != nil {
		log.Fatalf("Failed to create/open %s: %v", pubsub.ExchangeIonbusTopic, err)
	}

	// Open channel for commands
	commandsCh := OpenChannel(rmqConn)
	defer commandsCh.Close()

	// Declare and Bind commands
	commandsExchange := pubsub.ExchangeIonbusTopic
	commandsQueueName := pubsub.CommandsPrefix
	commandsQueueType := pubsub.Durable
	commandsRoutingKey := fmt.Sprintf("%s.%s.%s", pubsub.CommandsPrefix, pubsub.GatewaysPrefix, cfg.ID)
	_, err = pubsub.DeclareAndBindQueue(
		commandsCh,
		commandsExchange,
		commandsQueueName,
		commandsQueueType,
		commandsRoutingKey,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/open %s: %v", cfg.ID, err)
	}

	// Open channel for gateway
	gatewayCh := OpenChannel(rmqConn)
	defer gatewayCh.Close()

	// Bind to gateway-id.#
	gatewayQueueName := cfg.ID
	gatewayRoutingKey := fmt.Sprintf("%s.#", cfg.ID)
	_, err = pubsub.DeclareAndBindQueue(
		gatewayCh,
		pubsub.ExchangeIonbusTopic,
		cfg.ID,
		pubsub.Durable,
		gatewayRoutingKey,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/open %s: %v", cfg.ID, err)
	}

	// Bind to gateway-id.clients.#
	clientsRoutingKey := fmt.Sprintf("%s.%s.#", gatewayQueueName, pubsub.ClientsPrefix)
	_, err = pubsub.DeclareAndBindQueue(
		gatewayCh,
		pubsub.ExchangeIonbusTopic,
		gatewayQueueName,
		pubsub.Durable,
		clientsRoutingKey,
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

// Grabs the Gateway .env, loads it, and returns the Gateway Config ptr
func LoadGatewayEnv() *GatewayConfig {
	if err := godotenv.Load("./cmd/gateway/.env"); err != nil {
		log.Fatalf("Error loading .env: %v", err)
	}
	cfg, err := LoadGatewayConfig()
	if err != nil {
		log.Fatalf("Failed to load gateway config: %v", err)
	}
	return cfg
}

// Connects to the MQTT broker using the Gateway Config's Parameters
func ConnectMQTT(cfg *GatewayConfig) mqtt.Client {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.GetUrl()).
		SetKeepAlive(cfg.MQTT.KeepAlive).
		SetCleanSession(cfg.MQTT.CleanSession).
		SetClientID(cfg.ID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", token.Error())
	}
	log.Println("Successfully connected to MQTT")
	return client
}

// Connects to RabbitMQ using the Gateway Config's Parameters
func ConnectRMQ(cfg *GatewayConfig) *amqp.Connection {
	conn, err := amqp.Dial(cfg.RMQ.GetUrl())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	log.Println("Sucessfully connected to RabbitMQ")

	return conn
}

// Opens a channel on RabbitMQ, returns the channel
func OpenChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	return ch
}
