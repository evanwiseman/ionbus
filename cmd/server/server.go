package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/evanwiseman/ionbus/internal/util"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	Ctx     context.Context
	Cfg     *ServerConfig
	RMQConn *amqp.Connection
	RMQPub  *amqp.Channel
	DB      *sql.DB
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
		Ctx:     ctx,
		Cfg:     cfg,
		RMQConn: rmqConn,
		RMQPub:  nil,
		DB:      db,
	}

	// Setup infrastructure
	rmqPub, err := server.setupRabbitMQ()
	if err != nil {
		return nil, fmt.Errorf("failed to setup rabbitmq: %w", err)
	}
	server.RMQPub = rmqPub

	if err := server.setupDatabase(); err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return server, nil
}

// Start begins server operations
func (s *Server) Start() error {
	// Send initial broadcast to discover all gateways
	return nil
}

// Close gracefully shuts down the server
func (s *Server) Close() {
	if s.RMQPub != nil {
		s.RMQPub.Close()
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
