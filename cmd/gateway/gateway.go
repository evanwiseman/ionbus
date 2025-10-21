package main

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/evanwiseman/ionbus/internal/util"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Gateway struct {
	Ctx        context.Context
	Cfg        *GatewayConfig
	MQTTClient mqtt.Client
	RMQConn    *amqp.Connection
}

func NewGateway(ctx context.Context, cfg *GatewayConfig) (*Gateway, error) {
	mqttClient, err := util.ConnectMQTT(cfg.MQTT, cfg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to mqtt: %w", err)
	}
	rmqConn, err := util.ConnectRMQ(cfg.RMQ)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to rabbitmq: %w", err)
	}

	gateway := &Gateway{
		Ctx:        ctx,
		Cfg:        cfg,
		MQTTClient: mqttClient,
		RMQConn:    rmqConn,
	}

	if err := gateway.setupMQTT(); err != nil {
		return nil, fmt.Errorf("failed to setup mqtt: %w", err)
	}

	if err := gateway.setupRabbitMQ(); err != nil {
		return nil, fmt.Errorf("failed to setup rabbitmq: %w", err)
	}

	return nil, nil
}

func (g *Gateway) Start() error {
	// TODO
	return nil
}

func (g *Gateway) Close() {
	if g.MQTTClient != nil {
		g.MQTTClient.Disconnect(250)
	}
	if g.RMQConn != nil {
		g.RMQConn.Close()
	}
}

func (g *Gateway) setupMQTT() error {
	// TODO
	return nil
}

// Setup RabbitMQ topic exchange and dead letter exchange
func (g *Gateway) setupRabbitMQ() error {
	ch, err := util.OpenChannel(g.RMQConn)
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Declare exchanges
	if err := pubsub.DeclareIonbusTopic(ch); err != nil {
		return fmt.Errorf("failed to declare ionbus_topic exchange: %w", err)
	}
	if err := pubsub.DeclareIonbusDirect(ch); err != nil {
		return fmt.Errorf("failed to declare ionbus_direct exchange: %w", err)
	}
	if err := pubsub.DeclareIonbusBroadcast(ch); err != nil {
		return fmt.Errorf("failed to declare ionbus_broadcast exchange: %w", err)
	}

	// Declare Dead Lettering
	if err := pubsub.DeclareDLX(ch); err != nil {
		return fmt.Errorf("failed to declare ionbus_dlx: %w", err)
	}
	if err := pubsub.DeclareDLQ(ch); err != nil {
		return fmt.Errorf("failed to declare ionbus_dlq: %w", err)
	}

	return nil
}

// Queue name = gateway-id.commands,
// Routing key = gateways.gateway-id.commands.#
func (g *Gateway) setupCommands() (*amqp.Channel, amqp.Queue, error) {
	ch, err := util.OpenChannel(g.RMQConn)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to open channel: %w", err)
	}

	name := fmt.Sprintf("%s.%s", g.Cfg.ID, pubsub.CommandsPrefix)
	key := fmt.Sprintf("%s.%s.%s.#", pubsub.GatewaysPrefix, g.Cfg.ID, pubsub.CommandsPrefix)

	// Declare and Bind to topic queue refering to gateway-id
	q, err := pubsub.DeclareAndBindQueue(
		ch,
		pubsub.ExchangeIonbusTopic,
		name,
		pubsub.Durable,
		key,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to declare and bind topic exchange: %w", err)
	}

	// ALSO bind to the broadcast exchange for commands to all gateways
	err = ch.QueueBind(
		name,
		"",
		pubsub.ExchangeIonbusBroadcast,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind broadcast exchange: %w", err)
	}

	// Subscribe to commands sent by the server
	if err := pubsub.SubscribeRMQ(
		g.Ctx,
		ch,
		pubsub.RMQSubscribeOptions{
			QueueName: name,
		},
		models.ContentJSON,
		HandlerCommands,
	); err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to subscribe to queue: %w", err)
	}

	return ch, q, nil
}
