package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/evanwiseman/ionbus/internal/util"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	Ctx           context.Context
	Cfg           *ServerConfig
	RMQConn       *amqp.Connection
	RMQPublishCh  *amqp.Channel
	RMQResponseCh *amqp.Channel
	DB            *sql.DB
}

// Create a new server from the config passed in
func NewServer(ctx context.Context, cfg *ServerConfig) (*Server, error) {
	rmqConn, err := util.ConnectRMQ(cfg.RMQ)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to rabbitmq: %w", err)
	}
	db, err := util.ConnectDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to database: %w", err)
	}

	server := &Server{
		Ctx:          ctx,
		Cfg:          cfg,
		RMQConn:      rmqConn,
		RMQPublishCh: nil,
		DB:           db,
	}

	// Setup infrastructure
	rmqPub, err := server.setupRabbitMQ()
	if err != nil {
		return nil, fmt.Errorf("failed to setup rabbitmq: %w", err)
	}
	server.RMQPublishCh = rmqPub

	// Store the response channel!
	rmqResponseCh, _, err := server.setupResponses()
	if err != nil {
		return nil, fmt.Errorf("failed to setup responses: %w", err)
	}
	server.RMQResponseCh = rmqResponseCh

	if err := server.setupDatabase(); err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return server, nil
}

// Start begins server operations
func (s *Server) Start() error {
	s.RequestGatewayIdentifiersRMQ("*", nil, "device initialization")
	return nil
}

// Close gracefully shuts down the server
func (s *Server) Close() {
	if s.RMQPublishCh != nil {
		s.RMQPublishCh.Close()
	}
	if s.RMQResponseCh != nil {
		s.RMQResponseCh.Close()
	}
	if s.RMQConn != nil {
		s.RMQConn.Close()
	}
	if s.DB != nil {
		s.DB.Close()
	}
}

// Setup RabbitMQ topic exchange and dead letter exchange
func (s *Server) setupRabbitMQ() (*amqp.Channel, error) {
	ch, err := util.OpenChannel(s.RMQConn)
	if err != nil {
		return nil, fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	if err := pubsub.DeclareIonbusTopic(ch); err != nil {
		return nil, fmt.Errorf("failed to declare ionbus topic exchange: %w", err)
	}
	if err := pubsub.DeclareIonbusBroadcast(ch); err != nil {
		return nil, fmt.Errorf("failed to declare ionbus broadcast exchange")
	}

	if err := pubsub.DeclareDLX(ch); err != nil {
		return nil, fmt.Errorf("failed to declare dead letter exchange")
	}
	if err := pubsub.DeclareDLQ(ch); err != nil {
		return nil, fmt.Errorf("failed to declare dead letter queue")
	}

	return ch, nil
}

func (s *Server) setupDatabase() error {
	// TODO
	return nil
}

func (s *Server) setupResponses() (*amqp.Channel, amqp.Queue, error) {
	ch, err := util.OpenChannel(s.RMQConn)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to open channel: %w", err)
	}

	name := fmt.Sprintf("%s.%s", pubsub.GatewaysPrefix, pubsub.ResponsesPrefix)
	key := name

	q, err := pubsub.DeclareAndBindQueue(
		ch,
		pubsub.ExchangeIonbusDirect,
		name,
		pubsub.Durable,
		key,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to declare and bind direct exchange: %w", err)
	}

	if err := pubsub.SubscribeRMQ(
		s.Ctx,
		ch,
		pubsub.RMQSubscribeOptions{
			QueueName: name,
		},
		models.ContentJSON,
		func(msg any) pubsub.AckType {
			log.Println(msg)
			return pubsub.Ack
		},
	); err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to subscribe to queue: %w", err)
	}

	return ch, q, nil
}
