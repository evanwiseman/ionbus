package main

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Gateway struct {
	Ctx        context.Context
	Cfg        *GatewayConfig
	Client     mqtt.Client
	Conn       *amqp.Connection
	CommandCh  *amqp.Channel
	DeadCh     *amqp.Channel
	ResponseCh *amqp.Channel
}

func NewGateway(ctx context.Context, cfg *GatewayConfig) (*Gateway, error) {
	// MQTT
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.GetUrl()).
		SetKeepAlive(cfg.MQTT.KeepAlive).
		SetCleanSession(cfg.MQTT.CleanSession).
		SetClientID(cfg.ID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to establish connection to mqtt: %w", token.Error())
	}

	// RabbitMQ
	conn, err := amqp.Dial(cfg.RMQ.GetUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to rabbitmq: %w", err)
	}

	gateway := &Gateway{
		Ctx:    ctx,
		Cfg:    cfg,
		Client: client,
		Conn:   conn,
	}

	// Setup infrastructure
	if err := gateway.setupMQTT(); err != nil {
		return nil, fmt.Errorf("failed to setup mqtt: %w", err)
	}

	if err := gateway.setupRMQ(); err != nil {
		return nil, fmt.Errorf("failed to setup rabbitmq: %w", err)
	}

	return gateway, nil
}

func (g *Gateway) Start() error {
	// TODO
	return nil
}

func (g *Gateway) Close() {
	if g.Client != nil {
		g.Client.Disconnect(250)
	}
	if g.CommandCh != nil {
		g.CommandCh.Close()
	}
	if g.ResponseCh != nil {
		g.ResponseCh.Close()
	}
	if g.Conn != nil {
		g.Conn.Close()
	}
}

func (g *Gateway) setupMQTT() error {
	// TODO
	return nil
}

// Setup RabbitMQ topic exchange and dead letter exchange
func (g *Gateway) setupRMQ() error {
	// Setup dead lettering
	ch, err := g.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	g.DeadCh = ch
	if err := pubsub.DeclareDLX(ch); err != nil {
		return fmt.Errorf("failed to declare ionbus_dlx: %w", err)
	}
	if err := pubsub.DeclareAndBindDLQ(ch); err != nil {
		return fmt.Errorf("failed to declare ionbus_dlq: %w", err)
	}

	// Setup commands for gateway from the server
	if err := g.setupCommands(); err != nil {
		return fmt.Errorf("failed to setup commands: %w", err)
	}

	if err := g.setupResponses(); err != nil {
		return fmt.Errorf("failed to setup responses: %w", err)
	}

	return nil
}

// Creates a commands queue to listen for server commands
func (g *Gateway) setupCommands() error {
	// Declare Exchanges
	ch, err := g.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	g.CommandCh = ch
	if err := pubsub.DeclareGatewayCommandTopicX(ch); err != nil {
		return fmt.Errorf("failed to declare command topic: %w", err)
	}
	if err := pubsub.DeclareGatewayCommandBroadcastX(ch); err != nil {
		return fmt.Errorf("failed to declare command broadcast: %w", err)
	}

	queueName := pubsub.GetGatewayCommandQ(g.Cfg.ID)
	routingKey := pubsub.GetGatewayCommandRK(g.Cfg.ID, "#")

	// Declare queue for gateway commands
	_, err = ch.QueueDeclare(
		queueName,
		false,
		true,
		true,
		false,
		amqp.Table{
			"x-dead-letter-exchange": pubsub.XIonbusDlx,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare %s: %w", queueName, err)
	}

	// Bind queue to topic exchange
	if err := ch.QueueBind(
		queueName,
		routingKey,
		pubsub.GetGatewayCommandTopicX(),
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind %s to topic exchange: %w", queueName, err)
	}

	// Bind queue to broadcast exchange
	if err := ch.QueueBind(
		queueName,
		"",
		pubsub.GetGatewayCommandBroadcastX(),
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind %s to broadcast exchange: %w", queueName, err)
	}

	// Subscribe to gateway commands sent by server
	if err := pubsub.SubscribeRMQ(
		g.Ctx,
		ch,
		pubsub.RMQSubscribeOptions{
			QueueName: queueName,
		},
		models.ContentJSON,
		g.HandlerGatewayCommands,
	); err != nil {
		return fmt.Errorf("failed to subscribe to queue: %w", err)
	}

	return nil
}

// Creates a channel to send responses on
func (g *Gateway) setupResponses() error {
	ch, err := g.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	g.ResponseCh = ch

	return nil
}
