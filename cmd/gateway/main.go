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

type Gateway struct {
	Ctx        context.Context
	Cfg        *GatewayConfig
	MQTTClient mqtt.Client
	RMQConn    *amqp.Connection
}

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
	cfg := mustLoadGatewayEnv()

	mqttClient := mustConnectMQTT(cfg)
	rmqConn := mustConnectRMQ(cfg)
	gateway := &Gateway{
		Ctx:        ctx,
		Cfg:        cfg,
		MQTTClient: mqttClient,
		RMQConn:    rmqConn,
	}

	defer mqttClient.Disconnect(250)
	defer rmqConn.Close()

	gateway.setupMQTT()
	gateway.setupRabbitMQ()

	commandsCh, _ := gateway.setupCommandsQueue()
	defer commandsCh.Close()

	telemetryCh, _ := gateway.setupTelemetryQueue()
	defer telemetryCh.Close()

	log.Println("Successfully started ionbus gateway")

	// ========================
	// Wait for cancellation
	// ========================
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}

// Checks if an error occurs, if so fatals with message "Failed to $msg: $err"
func must(err error, msg string) {
	if err != nil {
		log.Fatalf("Failed to %s: %v", msg, err)
	}
}

// Grabs the Gateway .env, loads it, and returns the Gateway Config ptr
func mustLoadGatewayEnv() *GatewayConfig {
	must(godotenv.Load("./cmd/gateway/.env"), "load .env")

	cfg, err := LoadGatewayConfig()
	must(err, "load gateway config")

	return cfg
}

// Connects to the MQTT broker using the Gateway Config's Parameters
func mustConnectMQTT(cfg *GatewayConfig) mqtt.Client {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.GetUrl()).
		SetKeepAlive(cfg.MQTT.KeepAlive).
		SetCleanSession(cfg.MQTT.CleanSession).
		SetClientID(cfg.ID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT: %v", token.Error())
	}
	return client
}

// Connects to RabbitMQ using the Gateway Config's Parameters
func mustConnectRMQ(cfg *GatewayConfig) *amqp.Connection {
	conn, err := amqp.Dial(cfg.RMQ.GetUrl())
	must(err, "connect to RabbitMQ")

	return conn
}

// Opens a channel on RabbitMQ, returns the channel
func mustOpenChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	must(err, "open RabbitMQ channel")

	return ch
}

func (g *Gateway) setupMQTT() {

}

// Setup RabbitMQ topic exchange and dead letter exchange
func (g *Gateway) setupRabbitMQ() {
	ch := mustOpenChannel(g.RMQConn)
	defer ch.Close()
	must(pubsub.DeclareIonbusTopic(ch), "declare ionbus topic exchange")

	must(pubsub.DeclareDLX(ch), "declare ionbus dead letter exchange")
	must(pubsub.DeclareDLQ(ch), "declare ionbus dead letter queue")
}

// Open a new channel, then declare and bind the commands queue to it
//
// # Queue name = gateway-id.commands
//
// # Routing key = gateways.gateway-id.commands.#
func (g *Gateway) setupCommandsQueue() (*amqp.Channel, amqp.Queue) {
	ch := mustOpenChannel(g.RMQConn)

	name := fmt.Sprintf("%s.%s", g.Cfg.ID, pubsub.CommandsPrefix)
	key := fmt.Sprintf("%s.%s.%s.#", pubsub.GatewaysPrefix, g.Cfg.ID, pubsub.CommandsPrefix)

	q, err := pubsub.DeclareAndBindQueue(
		ch,
		pubsub.ExchangeIonbusTopic,
		name,
		pubsub.Durable,
		key,
	)
	must(err, "declare command queue")

	return ch, q
}

func (g *Gateway) setupTelemetryQueue() (*amqp.Channel, amqp.Queue) {
	ch := mustOpenChannel(g.RMQConn)

	name := fmt.Sprintf("%s.%s", g.Cfg.ID, pubsub.TelemetryPrefix)
	key := fmt.Sprintf("%s.%s.%s.#", pubsub.GatewaysPrefix, g.Cfg.ID, pubsub.TelemetryPrefix)

	q, err := pubsub.DeclareAndBindQueue(
		ch,
		pubsub.ExchangeIonbusTopic,
		name,
		pubsub.Durable,
		key,
	)
	must(err, "declare telemetry queue")

	return ch, q
}
