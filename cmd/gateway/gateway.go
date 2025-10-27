package main

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

type GatewayMQTT struct {
	Client       mqtt.Client
	RequestFlow  *pubsub.MQTTFlow
	ResponseFlow *pubsub.MQTTFlow
}

type GatewayRMQ struct {
	Conn         *amqp.Connection
	DeadCh       *amqp.Channel
	RequestFlow  *pubsub.RMQFlow
	ResponseFlow *pubsub.RMQFlow
}

type Gateway struct {
	Ctx  context.Context
	Cfg  *GatewayConfig
	MQTT *GatewayMQTT
	RMQ  *GatewayRMQ
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
		Ctx: ctx,
		Cfg: cfg,
		MQTT: &GatewayMQTT{
			Client: client,
		},
		RMQ: &GatewayRMQ{
			Conn: conn,
		},
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
	g.RequestServerIdentifiers("*")
	// g.RequestClientIdentifiers("+", nil, "device initialization")
	return nil
}

func (g *Gateway) Close() {
	if g.MQTT.Client != nil {
		g.MQTT.Client.Disconnect(250)
	}
	if g.RMQ.DeadCh != nil {
		g.RMQ.DeadCh.Close()
	}

	if g.RMQ.Conn != nil {
		g.RMQ.Conn.Close()
	}
}

func (g *Gateway) setupMQTT() error {
	g.setupMQTTRequests()
	g.setupMQTTResponses()
	return nil
}

func (g *Gateway) setupMQTTRequests() error {
	topic := pubsub.MGatewayReqT(g.Cfg.ID, "#")
	qos := byte(1)

	requestFlow := pubsub.NewMQTTFlow(g.Ctx, g.MQTT.Client)
	g.MQTT.RequestFlow = requestFlow
	if err := requestFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: topic,
		QoS:   qos,
	}); err != nil {
		return err
	}

	broadcast := pubsub.MGatewayReqB("#")
	if err := requestFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: broadcast,
		QoS:   qos,
	}); err != nil {
		return err
	}

	return nil
}

func (g *Gateway) setupMQTTResponses() error {
	topic := pubsub.MGatewayResT(g.Cfg.ID, "#")
	qos := byte(1)

	responseFlow := pubsub.NewMQTTFlow(g.Ctx, g.MQTT.Client)
	g.MQTT.ResponseFlow = responseFlow
	if err := responseFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: topic,
		QoS:   qos,
	}); err != nil {
		return nil
	}

	broadcast := pubsub.MGatewayResB("#")
	if err := responseFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: broadcast,
		QoS:   qos,
	}); err != nil {
		return nil
	}

	return nil
}

// Setup RabbitMQ topic exchange and dead letter exchange
func (g *Gateway) setupRMQ() error {
	// Setup dead lettering
	deadCh, err := pubsub.OpenChannel(g.RMQ.Conn)
	if err != nil {
		return err
	}
	g.RMQ.DeadCh = deadCh
	if err := pubsub.DeclareDLX(deadCh); err != nil {
		return err
	}
	if err := pubsub.DeclareAndBindDLQ(deadCh); err != nil {
		return err
	}

	// Setup commands for gateway from the server
	if err := g.setupRMQRequests(); err != nil {
		return fmt.Errorf("failed to setup commands: %w", err)
	}

	if err := g.setupRMQResponses(); err != nil {
		return fmt.Errorf("failed to setup responses: %w", err)
	}

	return nil
}

// Creates a commands queue to listen for server commands
func (g *Gateway) setupRMQRequests() error {
	pubCh, err := pubsub.OpenChannel(g.RMQ.Conn)
	if err != nil {
		return err
	}
	subCh, err := pubsub.OpenChannel(g.RMQ.Conn)
	if err != nil {
		return err
	}

	// Declare Exchanges
	if err := pubsub.DeclareGatewayCommandTopicX(subCh); err != nil {
		return err
	}
	if err := pubsub.DeclareGatewayCommandBroadcastX(subCh); err != nil {
		return err
	}

	// Queue Parameters
	name := pubsub.RGatewayReqQ(g.Cfg.ID)
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
	_, err = pubsub.DeclareQueue(subCh, name, opts)
	if err != nil {
		return err
	}

	// Bind queue to gateway topic exchange
	if err := pubsub.BindQueue(
		subCh,
		name,
		pubsub.RGatewayReqTRK(g.Cfg.ID, "#"),
		pubsub.RGatewayReqTX(),
	); err != nil {
		return err
	}

	// Bind queue to broadcast exchange
	if err := pubsub.BindQueue(
		subCh,
		name,
		pubsub.RGatewayReqBRK("#"),
		pubsub.RGatewayReqBX(),
	); err != nil {
		return err
	}

	g.RMQ.RequestFlow = pubsub.NewRMQFlow(g.Ctx, pubCh, subCh)
	if err := g.RMQ.RequestFlow.Sub.Subscribe(pubsub.RMQSubOpts{QueueName: name}); err != nil {
		return err
	}

	g.RMQ.RequestFlow.Sub.Mux.HandleFunc(
		pubsub.RGatewayReqTRK(g.Cfg.ID, string(models.ActionGetIdentifiers)),
		g.HandleIdentifierRequest,
	)
	g.RMQ.RequestFlow.Sub.Mux.HandleFunc(
		pubsub.RGatewayReqBRK(string(models.ActionGetIdentifiers)),
		g.HandleIdentifierRequest,
	)

	return nil
}

// Creates a channel to send responses on
func (g *Gateway) setupRMQResponses() error {
	pubCh, err := pubsub.OpenChannel(g.RMQ.Conn)
	if err != nil {
		return err
	}
	subCh, err := pubsub.OpenChannel(g.RMQ.Conn)
	if err != nil {
		return err
	}

	// Declare the exchange
	if err := pubsub.DeclareGatewayResponseTopicX(subCh); err != nil {
		return err
	}

	// Queue parameters
	name := pubsub.RGatewayResQ(g.Cfg.ID)
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
	_, err = pubsub.DeclareQueue(subCh, name, opts)
	if err != nil {
		return err
	}

	if err := pubsub.BindQueue(
		subCh,
		name,
		pubsub.RGatewayResTRK(g.Cfg.ID, "#"),
		pubsub.RGatewayResTX(),
	); err != nil {
		return err
	}

	g.RMQ.ResponseFlow = pubsub.NewRMQFlow(g.Ctx, pubCh, subCh)
	if err := g.RMQ.ResponseFlow.Sub.Subscribe(
		pubsub.RMQSubOpts{
			QueueName: name,
		},
	); err != nil {
		return err
	}

	g.RMQ.ResponseFlow.Sub.Mux.HandleFunc(
		pubsub.RGatewayResTRK(g.Cfg.ID, string(models.ActionGetIdentifiers)),
		g.HandleIdentifierResponse,
	)

	return nil
}
