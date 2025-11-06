package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/broker"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/models"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	Cfg          *config.ServerConfig
	Ctx          context.Context
	Conn         *amqp.Connection
	DB           *sql.DB
	RMQTransport *broker.RMQTransport
}

// Start begins server operations
func (s *Server) Start() error {
	s.RMQTransport.SendRequest(models.Request{
		Method:       "GET_IDENTIFIERS",
		TargetID:     "*",
		TargetDevice: models.DeviceGateway,
		Payload:      nil,
	})
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	if err := s.Conn.Close(); err != nil {
		return err
	}
	if err := s.DB.Close(); err != nil {
		return err
	}
	return nil
}

// Create a new server from the config passed in
func NewServer(ctx context.Context, cfg *config.ServerConfig) (*Server, error) {
	conn, err := broker.StartRMQ(cfg.RMQ)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", cfg.PG.GetUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	rmqTransport, err := broker.NewRMQTransport(ctx, cfg.Device, conn)
	if err != nil {
		return nil, fmt.Errorf("error: faile dto create rmq transport: %w", err)
	}

	server := &Server{
		Cfg:          cfg,
		Ctx:          ctx,
		Conn:         conn,
		DB:           db,
		RMQTransport: rmqTransport,
	}

	// Setup infrastructure
	if err := server.setupDatabase(); err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return server, nil
}

// ========================
// Database
// ========================

func (s *Server) setupDatabase() error {
	// TODO
	return nil
}
