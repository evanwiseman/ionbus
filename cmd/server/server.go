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
	CommandCh  *amqp.Channel
	DeadCh     *amqp.Channel
	PublishCh  *amqp.Channel
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
	deadCh, err := pubsub.OpenChannel(s.Conn)
	if err != nil {
		return err
	}
	s.DeadCh = deadCh
	if err := pubsub.DeclareDLX(deadCh); err != nil {
		return err
	}
	if err := pubsub.DeclareAndBindDLQ(deadCh); err != nil {
		return err
	}

	pubCh, err := pubsub.OpenChannel(s.Conn)
	if err != nil {
		return err
	}
	s.PublishCh = pubCh

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
	ch, err := pubsub.OpenChannel(s.Conn)
	if err != nil {
		return err
	}
	s.CommandCh = ch

	// Ensure the server has a command exchange
	if err := pubsub.DeclareServerCommandTopicX(ch); err != nil {
		return err
	}
	if err := pubsub.DeclareServerCommandBroadcastX(ch); err != nil {
		return err
	}

	// Queue Parameters
	name := pubsub.GetServerCommandQ(s.Cfg.ID)
	opts := pubsub.QueueOpts{
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     false,
		Args: amqp.Table{
			"x-dead-letter-exchange": pubsub.XIonbusDlx,
		},
	}

	// Declare the commands queue
	_, err = pubsub.DeclareQueue(ch, name, opts)
	if err != nil {
		return err
	}

	// Bind queue to topic exchange
	key := pubsub.GetServerCommandRK(s.Cfg.ID, "#")
	topicX := pubsub.GetServerCommandTopicX()
	if err := pubsub.BindQueue(ch, name, key, topicX); err != nil {
		return err
	}

	// Bind queue to broadcast exchange
	broadcastX := pubsub.GetServerCommandBroadcastX()
	if err := pubsub.BindQueue(ch, name, "", broadcastX); err != nil {
		return err
	}

	// Subscribe to server commands sent by gateway
	if err := pubsub.SubscribeRMQ(
		s.Ctx,
		ch,
		pubsub.RMQSubOpts{
			QueueName: name,
		},
		models.ContentJSON,
		s.HandlerServerCommands,
	); err != nil {
		return err
	}

	return nil
}

func (s *Server) setupResponses() error {
	// Create the response channel
	ch, err := pubsub.OpenChannel(s.Conn)
	if err != nil {
		return err
	}
	s.ResponseCh = ch

	// Ensure the server has a response exchange
	if err := pubsub.DeclareServerResponseTopicX(ch); err != nil {
		return err
	}

	// Queue parameters
	name := pubsub.GetServerResponseQ(s.Cfg.ID)
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
	_, err = pubsub.DeclareQueue(ch, name, opts)
	if err != nil {
		return err
	}

	// Bind Queue to gateway responses
	key := pubsub.GetGatewayResponseRK("*", "#")
	topicX := pubsub.GetServerResponseTopicX()
	if err := pubsub.BindQueue(ch, name, key, topicX); err != nil {
		return err
	}

	// Subscribe to responses published
	if err := pubsub.SubscribeRMQ(
		s.Ctx,
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

// ========================
// Database
// ========================

func (s *Server) setupDatabase() error {
	// TODO
	return nil
}
