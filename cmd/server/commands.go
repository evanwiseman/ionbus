package main

import (
	"fmt"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

// ========================
// Send
// ========================

// SendCommandRMQ publishes a command to a specific gateway
func (s *Server) SendCommandRMQ(key string, command models.Command) error {
	err := pubsub.PublishRMQ(
		s.Ctx,
		s.RMQPublishCh,
		pubsub.RMQPublishOptions{
			Exchange: pubsub.ExchangeIonbusTopic,
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

func (s *Server) SendGatewayCommandRMQ(gatewayID string, command models.Command) error {
	key := fmt.Sprintf("%s.%s.%s.#", pubsub.GatewaysPrefix, gatewayID, pubsub.CommandsPrefix)
	return s.SendCommandRMQ(key, command)
}

func (s *Server) SendClientCommandRMQ(gatewayID string, clientID string, command models.Command) error {
	key := fmt.Sprintf(
		"%s.%s.%s.%s.%s.#",
		pubsub.GatewaysPrefix,
		gatewayID,
		pubsub.ClientsPrefix,
		clientID,
		pubsub.CommandsPrefix,
	)
	return s.SendCommandRMQ(key, command)
}

// ========================
// Broadcast
// ========================

// BroadcastCommandRMQ publishes a command to all gateways
func (s *Server) BroadcastCommandRMQ(command models.Command) error {
	err := pubsub.PublishRMQ(
		s.Ctx,
		s.RMQPublishCh,
		pubsub.RMQPublishOptions{
			Exchange: pubsub.ExchangeIonbusBroadcast,
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

func (s *Server) BroadcastGatewayCommandRMQ(command models.Command) error {
	return s.BroadcastCommandRMQ(command)
}

func (s *Server) BraodcastClientCommandRMQ(command models.Command) error {
	return s.BroadcastCommandRMQ(command)
}

// ========================
// Requests
// ========================

func (s *Server) RequestGatewayIdentifiersRMQ(
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
		return s.BroadcastGatewayCommandRMQ(cmd)
	}
	return s.SendGatewayCommandRMQ(gatewayID, cmd)
}

func (s *Server) RequestClientIdentifiersRMQ(
	gatewayID string,
	clientID string,
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
	if clientID == "*" || clientID == "" {
		return s.BraodcastClientCommandRMQ(cmd)
	}
	return s.SendClientCommandRMQ(gatewayID, clientID, cmd)
}
