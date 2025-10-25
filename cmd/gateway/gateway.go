package main

import (
	"context"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Gateway struct {
	Ctx        context.Context
	Cfg        *GatewayConfig
	MQTTClient mqtt.Client
	RMQConn    *amqp.Connection
	CommandCh  *amqp.Channel
	DeadCh     *amqp.Channel
	PublishCh  *amqp.Channel
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
		Ctx:        ctx,
		Cfg:        cfg,
		MQTTClient: client,
		RMQConn:    conn,
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
	g.RequestServerIdentifiers("*", nil, "device initialization")
	g.RequestClientIdentifiers("+", nil, "device initialization")
	return nil
}

func (g *Gateway) Close() {
	if g.MQTTClient != nil {
		g.MQTTClient.Disconnect(250)
	}
	if g.CommandCh != nil {
		g.CommandCh.Close()
	}
	if g.DeadCh != nil {
		g.DeadCh.Close()
	}
	if g.PublishCh != nil {
		g.PublishCh.Close()
	}
	if g.ResponseCh != nil {
		g.ResponseCh.Close()
	}
	if g.RMQConn != nil {
		g.RMQConn.Close()
	}
}

func (g *Gateway) setupMQTT() error {
	g.setupMQTTCommands()
	g.setupMQTTResponses()
	return nil
}

func (g *Gateway) setupMQTTCommands() error {
	topic := pubsub.GetMQTTGatewayCommandTopic(g.Cfg.ID, "#")
	qos := byte(1)
	err := pubsub.SubscribeMQTT(
		g.Ctx,
		g.MQTTClient,
		pubsub.MQTTSubOpts{
			Topic: topic,
			QoS:   qos,
		},
		models.ContentJSON,
		g.HandlerMQTTGatewayCommands,
	)
	if err != nil {
		return err
	}

	broadcast := pubsub.GetMQTTGatewayCommandBroadcast("#")
	err = pubsub.SubscribeMQTT(
		g.Ctx,
		g.MQTTClient,
		pubsub.MQTTSubOpts{
			Topic: broadcast,
			QoS:   qos,
		},
		models.ContentJSON,
		g.HandlerMQTTGatewayCommands,
	)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) setupMQTTResponses() error {
	topic := pubsub.GetMQTTGatewayResponseTopic(g.Cfg.ID, "#")
	qos := byte(1)
	err := pubsub.SubscribeMQTT(
		g.Ctx,
		g.MQTTClient,
		pubsub.MQTTSubOpts{
			Topic: topic,
			QoS:   qos,
		},
		models.ContentJSON,
		func(msg any) {
			log.Println(msg)
		},
	)

	if err != nil {
		return err
	}

	broadcast := pubsub.GetMQTTGatewayResponseBroadcast("#")
	err = pubsub.SubscribeMQTT(
		g.Ctx,
		g.MQTTClient,
		pubsub.MQTTSubOpts{
			Topic: broadcast,
			QoS:   qos,
		},
		models.ContentJSON,
		func(msg any) {
			log.Println(msg)
		},
	)

	if err != nil {
		return err
	}

	return nil
}

// Setup RabbitMQ topic exchange and dead letter exchange
func (g *Gateway) setupRMQ() error {
	// Setup dead lettering
	deadCh, err := pubsub.OpenChannel(g.RMQConn)
	if err != nil {
		return err
	}
	g.DeadCh = deadCh
	if err := pubsub.DeclareDLX(deadCh); err != nil {
		return err
	}
	if err := pubsub.DeclareAndBindDLQ(deadCh); err != nil {
		return err
	}

	// Setup publisher channel
	pubCh, err := pubsub.OpenChannel(g.RMQConn)
	if err != nil {
		return err
	}
	g.PublishCh = pubCh

	// Setup commands for gateway from the server
	if err := g.setupRMQCommands(); err != nil {
		return fmt.Errorf("failed to setup commands: %w", err)
	}

	if err := g.setupRMQResponses(); err != nil {
		return fmt.Errorf("failed to setup responses: %w", err)
	}

	return nil
}

// Creates a commands queue to listen for server commands
func (g *Gateway) setupRMQCommands() error {
	// Declare Exchanges
	ch, err := pubsub.OpenChannel(g.RMQConn)
	if err != nil {
		return err
	}
	g.CommandCh = ch
	if err := pubsub.DeclareGatewayCommandTopicX(ch); err != nil {
		return err
	}
	if err := pubsub.DeclareGatewayCommandBroadcastX(ch); err != nil {
		return err
	}

	// Queue Parameters
	name := pubsub.GetRMQGatewayCommandQ(g.Cfg.ID)
	opts := pubsub.QueueOpts{
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     false,
		Args: amqp.Table{
			"x-dead-letter-exchange": pubsub.XIonbusDlx,
		},
	}

	// Declare the command queue
	_, err = pubsub.DeclareQueue(ch, name, opts)
	if err != nil {
		return err
	}

	// Bind queue to gateway topic exchange
	key := pubsub.GetRMQGatewayCommandRK(g.Cfg.ID, "#")
	topicX := pubsub.GetRMQGatewayCommandTopicX()
	if err := pubsub.BindQueue(ch, name, key, topicX); err != nil {
		return err
	}

	// Bind queue to broadcast exchange
	broadcastX := pubsub.GetRMQGatewayCommandBroadcastX()
	if err := pubsub.BindQueue(ch, name, "", broadcastX); err != nil {
		return err
	}

	// Subscribe to gateway commands sent by server
	if err := pubsub.SubscribeRMQ(
		g.Ctx,
		ch,
		pubsub.RMQSubOpts{
			QueueName: name,
		},
		models.ContentJSON,
		g.HandlerRMQGatewayCommands,
	); err != nil {
		return err
	}

	return nil
}

// Creates a channel to send responses on
func (g *Gateway) setupRMQResponses() error {
	ch, err := pubsub.OpenChannel(g.RMQConn)
	if err != nil {
		return err
	}
	g.ResponseCh = ch

	if err := pubsub.DeclareGatewayResponseTopicX(ch); err != nil {
		return err
	}

	name := pubsub.GetRMQGatewayResponseQ(g.Cfg.ID)
	opts := pubsub.QueueOpts{
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     false,
		Args: amqp.Table{
			"x-dead-letter-exchange": pubsub.XIonbusDlx,
		},
	}

	// Declare the Response Queue
	_, err = pubsub.DeclareQueue(ch, name, opts)
	if err != nil {
		return err
	}

	// Bind to Queue to server responses key
	key := pubsub.GetRMQServerResponseRK("*", "#")
	topicX := pubsub.GetRMQGatewayResponseTopicX()
	if err := pubsub.BindQueue(ch, name, key, topicX); err != nil {
		return err
	}

	// Subscribe to responses published
	if err := pubsub.SubscribeRMQ(
		g.Ctx,
		ch,
		pubsub.RMQSubOpts{
			QueueName: name,
		},
		models.ContentJSON,
		func(msg any) pubsub.AckType {
			log.Println(msg)
			return pubsub.Ack
		},
	); err != nil {
		return err
	}

	return nil
}
