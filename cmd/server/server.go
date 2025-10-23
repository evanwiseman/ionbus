package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ========================
// Server
// ========================

type Server struct {
	Ctx        context.Context
	Cfg        *ServerConfig
	DB         *sql.DB
	Conn       *amqp.Connection
	DeadCh     *amqp.Channel
	CommandCh  *amqp.Channel
	ResponseCh *amqp.Channel
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
		Ctx:  ctx,
		Cfg:  cfg,
		DB:   db,
		Conn: conn,
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
	s.RequestGatewayIdentifiers("*", nil, "device initialization")
	return nil
}

// Close gracefully shuts down the server
func (s *Server) Close() {
	if s.DeadCh != nil {
		s.DeadCh.Close()
	}
	if s.CommandCh != nil {
		s.CommandCh.Close()
	}
	if s.ResponseCh != nil {
		s.ResponseCh.Close()
	}
	if s.Conn != nil {
		s.Conn.Close()
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
	ch, err := s.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	s.DeadCh = ch
	if err := pubsub.DeclareDLX(ch); err != nil {
		return fmt.Errorf("failed to declare dead letter exchange")
	}
	if err := pubsub.DeclareAndBindDLQ(ch); err != nil {
		return fmt.Errorf("failed to declare dead letter queue")
	}

	// Setup Commands (Commands to the server sent from other devices, really only requests)
	if err := s.setupCommands(); err != nil {
		return fmt.Errorf("failed to setup command responses: %w", err)
	}

	// Setup Responses (Responses from other devices, that the server reads from)
	if err := s.setupResponses(); err != nil {
		return fmt.Errorf("failed to setup responses: %w", err)
	}

	return nil
}

func (s *Server) setupCommands() error {
	ch, err := s.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	s.CommandCh = ch

	// Ensure the server has a command exchange
	if err := pubsub.DeclareServerCommandTopicX(ch); err != nil {
		return fmt.Errorf("failed to declare command topic: %w", err)
	}
	if err := pubsub.DeclareServerCommandBroadcastX(ch); err != nil {
		return fmt.Errorf("failed to declare command broadcast: %w", err)
	}

	// Create the server command queue
	queueName := pubsub.GetServerCommandQ(s.Cfg.ID)
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
		pubsub.GetServerCommandRK(s.Cfg.ID, "#"),
		pubsub.GetServerCommandTopicX(),
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind to topic exchange: %w", err)
	}

	// Bind queue to broadcast exchange
	if err := ch.QueueBind(
		queueName,
		"",
		pubsub.GetServerCommandBroadcastX(),
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind to broadcast exchange: %w", err)
	}

	// Subscribe to server commands sent by gateway
	if err := pubsub.SubscribeRMQ(
		s.Ctx,
		ch,
		pubsub.RMQSubscribeOptions{
			QueueName: queueName,
		},
		models.ContentJSON,
		s.HandlerServerCommands,
	); err != nil {
		return fmt.Errorf("failed to subscribe to command queue: %w", err)
	}

	return nil
}

func (s *Server) setupResponses() error {
	// Create the response channel
	ch, err := s.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	s.ResponseCh = ch

	// Ensure the server has a response exchange
	if err := pubsub.DeclareServerResponseTopicX(ch); err != nil {
		return fmt.Errorf("failed to declare request topic: %w", err)
	}

	// Create the server response queue
	queueName := pubsub.GetServerResponseQ(s.Cfg.ID)
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

	// Bind Queue to gateway responses
	routingKey := pubsub.GetGatewayResponseRK("*", "#")
	if err := ch.QueueBind(
		queueName,
		routingKey,
		pubsub.GetServerResponseTopicX(),
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind %s to topic exchange: %w", queueName, err)
	}

	// Subscribe to responses published
	if err := pubsub.SubscribeRMQ(
		s.Ctx,
		ch,
		pubsub.RMQSubscribeOptions{
			QueueName: queueName,
		},
		models.ContentJSON,
		func(msg any) pubsub.AckType {
			log.Println(msg)
			return pubsub.Ack
		},
	); err != nil {
		return fmt.Errorf("failed to subscribe to queue: %w", err)
	}

	return nil
}

// ========================
// Database
// ========================

func (s *Server) setupDatabase() error {
	// TODO
	return nil
}
