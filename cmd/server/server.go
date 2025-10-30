package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/pubsub"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ========================
// Server
// ========================

type Server struct {
	Ctx context.Context
	Cfg *ServerConfig
	DB  *sql.DB
	RMQ ServerRMQ
}

type ServerRMQ struct {
	Conn               *amqp.Connection
	DeadCh             *amqp.Channel
	RequestPublisher   *pubsub.RMQPublisher
	RequestSubscriber  *pubsub.RMQSubscriber
	ResponsePublisher  *pubsub.RMQPublisher
	ResponseSubscriber *pubsub.RMQSubscriber
}

func (s *ServerRMQ) Close() {
	if s.DeadCh != nil {
		s.DeadCh.Close()
	}
	if s.RequestPublisher != nil {
		s.RequestPublisher.Close()
	}
	if s.RequestSubscriber != nil {
		s.RequestSubscriber.Close()
	}
	if s.ResponsePublisher != nil {
		s.ResponsePublisher.Close()
	}
	if s.ResponseSubscriber != nil {
		s.ResponseSubscriber.Close()
	}
	if s.Conn != nil {
		s.Conn.Close()
	}
}

// Create a new server from the config passed in
func NewServer(ctx context.Context, cfg *ServerConfig) (*Server, error) {
	db, err := sql.Open(cfg.DB.Schema, cfg.DB.GetUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to database: %w", err)
	}

	conn, err := amqp.Dial(cfg.RMQ.GetUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to rabbitmq: %w", err)
	}

	server := &Server{
		Ctx: ctx,
		Cfg: cfg,
		DB:  db,
		RMQ: ServerRMQ{
			Conn: conn,
		},
	}

	// Setup infrastructure
	if err := server.setupDatabase(); err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	if err := server.setupRabbitMQ(); err != nil {
		return nil, fmt.Errorf("failed to setup rabbitmq: %w", err)
	}

	return server, nil
}

// Start begins server operations
func (s *Server) Start() error {
	s.RequestGatewayIdentifiers("*")
	return nil
}

// Close gracefully shuts down the server
func (s *Server) Close() {
	s.RMQ.Close()
	if s.DB != nil {
		s.DB.Close()
	}
}

// ========================
// RabbitMQ
// ========================

// Setup RabbitMQ topic exchange and dead letter exchange
func (s *Server) setupRabbitMQ() error {
	deadCh, err := pubsub.OpenChannel(s.RMQ.Conn)
	if err != nil {
		return err
	}
	s.RMQ.DeadCh = deadCh
	if err := pubsub.DeclareDLX(deadCh); err != nil {
		return err
	}
	if err := pubsub.DeclareAndBindDLQ(deadCh); err != nil {
		return err
	}

	// Setup Requests
	if err := s.setupRequests(); err != nil {
		return fmt.Errorf("failed to setup requests: %w", err)
	}

	// Setup Responses
	if err := s.setupResponses(); err != nil {
		return fmt.Errorf("failed to setup responses: %w", err)
	}

	return nil
}

func (s *Server) setupRequests() error {
	pubCh, err := pubsub.OpenChannel(s.RMQ.Conn)
	if err != nil {
		return err
	}
	subCh, err := pubsub.OpenChannel(s.RMQ.Conn)
	if err != nil {
		return err
	}

	// Ensure the server has a request exchange
	if err := pubsub.DeclareServerCommandTopicX(pubCh); err != nil {
		return err
	}
	if err := pubsub.DeclareServerCommandBroadcastX(pubCh); err != nil {
		return err
	}

	// Queue Parameters
	name := pubsub.RServerReqQ(s.Cfg.ID)
	opts := pubsub.QueueOpts{
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     false,
		Args: amqp.Table{
			"x-dead-letter-exchange": pubsub.XIonbusDlx,
		},
	}

	// Declare the request queue
	_, err = pubsub.DeclareQueue(subCh, name, opts)
	if err != nil {
		return err
	}

	// Bind queue to topic exchange
	if err := pubsub.BindQueue(
		subCh,
		name,
		pubsub.RServerReqTRK(s.Cfg.ID, "#"),
		pubsub.RServerReqTX(),
	); err != nil {
		return err
	}

	// Bind queue to broadcast exchange
	if err := pubsub.BindQueue(
		subCh,
		name,
		pubsub.RServerReqBRK("#"),
		pubsub.RServerReqBX(),
	); err != nil {
		return err
	}

	// Subscribe to requests
	requestPublisher := pubsub.NewRMQPublisher(s.Ctx, pubCh)
	s.RMQ.RequestPublisher = requestPublisher

	requestSubscriber := pubsub.NewRMQSubscriber(s.Ctx, subCh)
	if err := requestSubscriber.Subscribe(
		pubsub.RMQSubOpts{
			QueueName:     name,
			PrefetchCount: 10,
			AutoAck:       false,
		},
		s.HandlerRequests,
	); err != nil {
		return err
	}
	s.RMQ.RequestSubscriber = requestSubscriber

	return nil
}

func (s *Server) setupResponses() error {
	// Create the response channel
	pubCh, err := pubsub.OpenChannel(s.RMQ.Conn)
	if err != nil {
		return err
	}
	subCh, err := pubsub.OpenChannel(s.RMQ.Conn)
	if err != nil {
		return err
	}

	// Ensure the server has a response exchange
	if err := pubsub.DeclareServerResponseTopicX(subCh); err != nil {
		return err
	}

	// Queue parameters
	name := pubsub.RServerResQ(s.Cfg.ID)
	opts := pubsub.QueueOpts{
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     false,
		Args: amqp.Table{
			"x-dead-letter-exchange": pubsub.XIonbusDlx,
		},
	}

	// Declare the response queue
	_, err = pubsub.DeclareQueue(subCh, name, opts)
	if err != nil {
		return err
	}

	// Bind Queue to gateway responses
	if err := pubsub.BindQueue(
		subCh,
		name,
		pubsub.RServerResTRK(s.Cfg.ID, "#"),
		pubsub.RServerResTX(),
	); err != nil {
		return err
	}

	responsePublisher := pubsub.NewRMQPublisher(s.Ctx, pubCh)
	s.RMQ.ResponsePublisher = responsePublisher

	responseSubscriber := pubsub.NewRMQSubscriber(s.Ctx, subCh)
	if err := responseSubscriber.Subscribe(
		pubsub.RMQSubOpts{
			QueueName:     name,
			PrefetchCount: 10,
			AutoAck:       false,
		},
		s.HandlerResponses,
	); err != nil {
		return err
	}
	s.RMQ.ResponseSubscriber = responseSubscriber

	return nil
}

// ========================
// Database
// ========================

func (s *Server) setupDatabase() error {
	// TODO
	return nil
}
