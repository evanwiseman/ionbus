package main

import (
	"fmt"
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

// ========================
// Send
// ========================

func (s *Server) SendGatewayCommand(gatewayID string, command models.Command) error {
	key := fmt.Sprintf("%s.%s.%s.#", pubsub.GatewayPrefix, gatewayID, pubsub.CommandPrefix)

	err := pubsub.PublishRMQ(
		s.Ctx,
		s.CommandCh,
		pubsub.RMQPubOpts{
			Exchange: pubsub.GetGatewayCommandTopicX(),
			Key:      key,
		},
		models.ContentJSON,
		command,
	)

	if err != nil {
		return fmt.Errorf("failed to send command with key %s: %w", key, err)
	}
	return nil
}

func (s *Server) BroadcastGatewayCommand(command models.Command) error {
	err := pubsub.PublishRMQ(
		s.Ctx,
		s.CommandCh,
		pubsub.RMQPubOpts{
			Exchange: pubsub.GetGatewayCommandBroadcastX(),
			Key:      "", // ignored for fanout
		},
		models.ContentJSON,
		command,
	)

	if err != nil {
		return fmt.Errorf("failed broadcasting command: %w", err)
	}

	return nil
}

// ========================
// Requests
// ========================

func (s *Server) RequestGatewayIdentifiers(
	gatewayID string,
	filters map[string]interface{},
	reason string,
) error {
	cmd := models.Command{
		Name: "request",
		Args: models.RequestArgs{
			Filters:   filters,
			Timestamp: time.Now(),
			Reason:    reason,
		},
	}
	if gatewayID == "*" || gatewayID == "" {
		return s.BroadcastGatewayCommand(cmd)
	}
	return s.SendGatewayCommand(gatewayID, cmd)
}

// ========================
// Handlers
// ========================

func (s *Server) HandlerServerCommands(command models.Command) pubsub.AckType {
	switch command.Name {
	case "request":
		// handle request
		err := s.Server2GatewayResponse(models.Response{
			Name:      s.Cfg.ID,
			Status:    "",
			Data:      nil,
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Printf("failed to send response: %v", err)
		}
		log.Printf("received request command: %v", command.Args)
	default:
		return pubsub.NackDiscard
	}

	return pubsub.Ack
}
