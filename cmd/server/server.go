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

	// Setup Gateway Responses
	if err := s.setupResponses(); err != nil {
		return fmt.Errorf("failed to setup responses: %w", err)
	}

	return nil
}

func (s *Server) setupResponses() error {
	ch, err := s.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	s.ResponseCh = ch

	if err := pubsub.DeclareGatewayResponseTopicX(ch); err != nil {
		return fmt.Errorf("failed to declare request topic: %w", err)
	}

	queueName := pubsub.GetGatewayResponseQ(s.Cfg.ID)
	routingKey := pubsub.GetGatewayResponseRK("*", "#")

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

	if err := ch.QueueBind(
		queueName,
		routingKey,
		pubsub.GetGatewayResponseTopicX(),
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind %s to topic exchange: %w", queueName, err)
	}

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
