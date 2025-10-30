package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Gateway struct {
	Ctx  context.Context
	Cfg  *GatewayConfig
	MQTT *GatewayMQTT
	RMQ  *GatewayRMQ
}

type GatewayMQTT struct {
	Client             mqtt.Client
	RequestPublisher   *pubsub.MQTTPublisher
	RequestSubscriber  *pubsub.MQTTSubscriber
	ResponsePublisher  *pubsub.MQTTPublisher
	ResponseSubscriber *pubsub.MQTTSubscriber
}

func (g *GatewayMQTT) Close() {
	g.Client.Disconnect(250)
}

type GatewayRMQ struct {
	Conn               *amqp.Connection
	DeadCh             *amqp.Channel
	RequestPublisher   *pubsub.RMQPublisher
	RequestSubscriber  *pubsub.RMQSubscriber
	ResponsePublisher  *pubsub.RMQPublisher
	ResponseSubscriber *pubsub.RMQSubscriber
}

func (g *GatewayRMQ) Close() {
	if g.DeadCh != nil {
		g.DeadCh.Close()
	}
	if g.RequestPublisher != nil {
		g.RequestPublisher.Close()
	}
	if g.RequestSubscriber != nil {
		g.RequestSubscriber.Close()
	}
	if g.ResponsePublisher != nil {
		g.ResponsePublisher.Close()
	}
	if g.ResponseSubscriber != nil {
		g.ResponseSubscriber.Close()
	}
	if g.Conn != nil {
		g.Conn.Close()
	}
}

func NewGateway(ctx context.Context, cfg *GatewayConfig) (*Gateway, error) {
	gateway := &Gateway{
		Ctx:  ctx,
		Cfg:  cfg,
		MQTT: &GatewayMQTT{},
		RMQ:  &GatewayRMQ{},
	}
	// MQTT
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.GetUrl()).
		SetKeepAlive(cfg.MQTT.KeepAlive).
		SetCleanSession(cfg.MQTT.CleanSession).
		SetClientID(cfg.ID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("ERROR: MQTT connection lost: %v", err)
		}).
		SetOnConnectHandler(func(client mqtt.Client) {
			log.Printf("INFO: MQTT connected successfully")
			if err := gateway.setupMQTTRequests(); err != nil {
				log.Printf("ERROR: Failed to setup MQTT requests: %v", err)
			}
			if err := gateway.setupMQTTResponses(); err != nil {
				log.Printf("ERROR: Failed to setup MQTT responses: %v", err)
			}
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			log.Printf("WARN: MQTT reconnecting...")
		}).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(10 * time.Second)

	client := mqtt.NewClient(opts)
	gateway.MQTT.Client = client
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to establish connection to mqtt: %w", token.Error())
	}

	// RabbitMQ
	conn, err := amqp.Dial(cfg.RMQ.GetUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to rabbitmq: %w", err)
	}
	gateway.RMQ.Conn = conn

	// Setup infrastructure

	if err := gateway.setupRMQ(); err != nil {
		return nil, fmt.Errorf("failed to setup rabbitmq: %w", err)
	}

	return gateway, nil
}

func (g *Gateway) Start() error {
	g.RequestServerIdentifiers("*")
	g.RequestClientIdentifiers("+")
	return nil
}

func (g *Gateway) Close() {
	g.MQTT.Close()
	g.RMQ.Close()
}

func (g *Gateway) setupMQTTRequests() error {
	requestPublisher := pubsub.NewMQTTPublisher(g.Ctx, g.MQTT.Client)
	g.MQTT.RequestPublisher = requestPublisher

	requestSubscriber := pubsub.NewMQTTSubscriber(g.Ctx, g.MQTT.Client)
	if err := requestSubscriber.Subscribe(
		pubsub.MQTTSubOpts{
			Topic: pubsub.MGatewayReqT(g.Cfg.ID, "#"),
			QoS:   byte(1),
		},
		g.HandlerRequests,
	); err != nil {
		return err
	}
	if err := requestSubscriber.Subscribe(
		pubsub.MQTTSubOpts{
			Topic: pubsub.MGatewayReqB("#"),
			QoS:   byte(1),
		},
		g.HandlerRequests,
	); err != nil {
		return err
	}
	g.MQTT.RequestSubscriber = requestSubscriber

	return nil
}

func (g *Gateway) setupMQTTResponses() error {
	responsePublisher := pubsub.NewMQTTPublisher(g.Ctx, g.MQTT.Client)
	g.MQTT.ResponsePublisher = responsePublisher

	responseSubscriber := pubsub.NewMQTTSubscriber(g.Ctx, g.MQTT.Client)
	if err := responseSubscriber.Subscribe(
		pubsub.MQTTSubOpts{
			Topic: pubsub.MGatewayResT(g.Cfg.ID, "#"),
			QoS:   byte(1),
		},
		g.HandlerResponses,
	); err != nil {
		return err
	}
	g.MQTT.ResponseSubscriber = responseSubscriber

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

	requestPublisher := pubsub.NewRMQPublisher(g.Ctx, pubCh)
	g.RMQ.RequestPublisher = requestPublisher

	requestSubscriber := pubsub.NewRMQSubscriber(g.Ctx, subCh)
	requestSubscriber.Subscribe(
		pubsub.RMQSubOpts{
			QueueName:     name,
			PrefetchCount: 10,
			AutoAck:       false,
		},
		g.HandlerRequests,
	)
	g.RMQ.RequestSubscriber = requestSubscriber

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

	responsePublisher := pubsub.NewRMQPublisher(g.Ctx, pubCh)
	g.RMQ.ResponsePublisher = responsePublisher

	responseSubscriber := pubsub.NewRMQSubscriber(g.Ctx, subCh)
	if err := responseSubscriber.Subscribe(
		pubsub.RMQSubOpts{
			QueueName:     name,
			PrefetchCount: 10,
			AutoAck:       false,
		},
		g.HandlerResponses,
	); err != nil {
		return err
	}
	g.RMQ.ResponseSubscriber = responseSubscriber

	// if err := g.RMQ.ResponseFlow.Sub.Subscribe(
	// 	pubsub.RMQSubOpts{
	// 		QueueName: name,
	// 	},
	// ); err != nil {
	// 	return err
	// }

	// g.RMQ.ResponseFlow.Sub.Mux.HandleFunc(
	// 	pubsub.RGatewayResTRK(g.Cfg.ID, string(models.MethodGetIdentifiers)),
	// 	g.HandleIdentifierResponse,
	// )

	return nil
}
