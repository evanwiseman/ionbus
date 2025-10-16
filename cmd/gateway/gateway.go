package main

import (
	"context"
	"fmt"
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
	// Open the channel
	commandsCh := openChannel(rmqConn)

	// Declare the exchange
	err = broker.DeclareCommandExchange(commandsCh)
	if err != nil {
		log.Fatalf("Failed to create commands exchange: %v\n", err)
	}
	log.Println("Successfully created commands exchange")
	_ = declareAndBindCommandsQueue(commandsCh, cfg.ID)

	// ========================
	// Gateways
	// ========================
	// Gateway queue name and routing key
	gatewayQueueName := fmt.Sprintf("%v", cfg.ID)
	gatewayRoutingKey := fmt.Sprintf(
		"%s.%s.%s.#",
		routing.CommandsPrefix,
		routing.GatewaysPrefix,
		cfg.ID,
	)

	// Declare and bind the gateway queue
	_, err = broker.DeclareAndBindQueue(
		commandsCh,
		routing.ExchangeCommandsTopic,
		gatewayQueueName,
		routing.Transient,
		gatewayRoutingKey,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/bind gateway queue: %v\n", err)
	}
	log.Println("Successfully created/binded gateway queue:", routing.GatewaysPrefix)

	// ========================
	// Devices
	// ========================
	// Devices GET queue name and routing key
	devicesGetQueueName := fmt.Sprintf("%s.%s.*.%s", cfg.ID, routing.DevicesPrefix, routing.GetKey)
	devicesGetRoutingKey := fmt.Sprintf(
		"%s.%s.%s.%s.*.%s",
		routing.CommandsPrefix,
		routing.GatewaysPrefix,
		cfg.ID,
		routing.DevicesPrefix,
		routing.GetKey,
	)

	// Declare and bind the devices GET key
	_, err = broker.DeclareAndBindQueue(
		commandsCh,
		routing.ExchangeCommandsTopic,
		devicesGetQueueName,
		routing.Transient,
		devicesGetRoutingKey,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/bind devices GET queue: %v\n", err)
	}
	log.Println("Successfully created/binded devices GET queue:", routing.GatewaysPrefix)

	// ========================
	// Bridge Commands to MQTT
	// ========================
	b := bridge.Bridge{
		RMQCh:      commandsCh,
		MQTTClient: mqttClient,
	}
	// Bind Devices GET to MQTT
	b.RMQToMQTT(
		ctx,
		pubsub.RMQSubscribeOptions{QueueName: devicesGetQueueName},
		pubsub.MQTTPublishOptions{Topic: "devices/+/get"},
		routing.ContentJSON,
	)
	b.MQTTToRMQ(
		ctx,
		pubsub.MQTTSubscribeOptions{},
		pubsub.RMQPublishOptions{},
		routing.ContentJSON,
	)

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
func declareAndBindCommandsQueue(ch *amqp.Channel, gatewayID string) amqp.Queue {
	queueName := fmt.Sprintf("%v.%v", routing.CommandsPrefix, gatewayID)
	routingKey := fmt.Sprintf("%v.%v.#", routing.CommandsPrefix, gatewayID)
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
