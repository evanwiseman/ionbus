package main

import (
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

// SendCommand publishes a command to a specific gateway
func (s *Server) SendCommand(gatewayID string, command models.Command) error {
	key := fmt.Sprintf("%s.%s.%s.#", pubsub.GatewaysPrefix, gatewayID, pubsub.CommandsPrefix)

	err := pubsub.PublishRMQ(
		s.Ctx,
		s.RMQPub,
		pubsub.RMQPublishOptions{
			Exchange: pubsub.ExchangeIonbusTopic,
			Key:      key,
		},
		models.ContentJSON,
		command,
	)

	if err != nil {
		return fmt.Errorf("failed sending command to gateway %s: %w", gatewayID, err)
	}

	return nil
}

// BroadcastCommand publishes a command to all gateways
func (s *Server) BroadcastCommand(command models.Command) error {
	err := pubsub.PublishRMQ(
		s.Ctx,
		s.RMQPub,
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
