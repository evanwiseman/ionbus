package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/cmd/gateway/bridge"
	"github.com/evanwiseman/ionbus/internal/models"
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
	err := godotenv.Load()
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
	// Commands
	// ========================
	// Open the channel
	commandsCh := openChannel(rmqConn)

	// ========================
	// Gateways
	// ========================
	// Gateway queue name and routing key
	gatewayQueueName := fmt.Sprintf("%v", cfg.ID) // gateway-id
	gatewayRoutingKey := fmt.Sprintf(             // commands.gateways.gateway-id
		"%s.%s.%s",
		pubsub.CommandsPrefix,
		pubsub.GatewaysPrefix,
		cfg.ID,
	)

	// Declare and bind the gateway queue
	_, err = pubsub.DeclareAndBindQueue(
		commandsCh,
		pubsub.ExchangeCommandsTopic,
		gatewayQueueName,
		pubsub.Transient,
		gatewayRoutingKey,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/bind gateway queue: %v\n", err)
	}
	log.Println("Successfully created/binded gateway queue:", gatewayQueueName)

	// ========================
	// Client
	// ========================
	// Clients queue name and routing key
	clientsQueueName := fmt.Sprintf("%s.%s", cfg.ID, pubsub.ClientsPrefix) // gateway-id.clients
	devicesRoutingKey := fmt.Sprintf(                                      // commands.gateways.gateway-id.clients.#
		"%s.%s.%s.%s.#",
		pubsub.CommandsPrefix,
		pubsub.GatewaysPrefix,
		cfg.ID,
		pubsub.ClientsPrefix,
	)

	// Declare and bind the clients
	_, err = pubsub.DeclareAndBindQueue(
		commandsCh,
		pubsub.ExchangeCommandsTopic,
		clientsQueueName,
		pubsub.Transient,
		devicesRoutingKey,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create/bind clients GET queue: %v\n", err)
	}
	log.Println("Successfully created/binded clients GET queue:", clientsQueueName)

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
		pubsub.RMQSubscribeOptions{QueueName: clientsQueueName},
		pubsub.MQTTPublishOptions{Topic: "clients/"},
		models.ContentJSON,
	)
	b.MQTTToRMQ(
		ctx,
		pubsub.MQTTSubscribeOptions{},
		pubsub.RMQPublishOptions{},
		models.ContentJSON,
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
