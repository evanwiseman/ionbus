package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ========================
// Server
// ========================

type ServerRMQ struct {
	Conn         *amqp.Connection
	DeadCh       *amqp.Channel
	RequestFlow  *pubsub.RMQFlow
	ResponseFlow *pubsub.RMQFlow
}

type Server struct {
	Ctx context.Context
	Cfg *ServerConfig
	DB  *sql.DB
	RMQ ServerRMQ
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
	if s.RMQ.DeadCh != nil {
		s.RMQ.DeadCh.Close()
	}
	if s.RMQ.Conn != nil {
		s.RMQ.Conn.Close()
	}
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
	s.RMQ.RequestFlow = pubsub.NewRMQFlow(s.Ctx, pubCh, subCh)
	if err := s.RMQ.RequestFlow.Sub.Subscribe(pubsub.RMQSubOpts{QueueName: name}); err != nil {
		return err
	}

	s.RMQ.RequestFlow.Sub.Mux.HandleFunc(
		pubsub.RServerReqTRK(s.Cfg.ID, "#"),
		s.HandleIdentifierRequest,
	)
	s.RMQ.RequestFlow.Sub.Mux.HandleFunc(
		pubsub.RServerReqBRK("#"),
		s.HandleIdentifierRequest,
	)

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

	// Subscribe to responses
	s.RMQ.ResponseFlow = pubsub.NewRMQFlow(s.Ctx, pubCh, subCh)
	if err := s.RMQ.ResponseFlow.Sub.Subscribe(
		pubsub.RMQSubOpts{
			QueueName: name,
		},
	); err != nil {
		return err
	}

	s.RMQ.ResponseFlow.Sub.Mux.HandleFunc(
		pubsub.RServerResTRK(s.Cfg.ID, string(models.ActionGetIdentifiers)),
		s.HandleIdentifierResponse,
	)

	return nil
}

// ========================
// Database
// ========================

func (s *Server) setupDatabase() error {
	// TODO
	return nil
}
