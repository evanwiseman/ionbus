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

func (g *Gateway) SendServerCommand(serverID string, command models.Command) error {
	key := pubsub.GetServerCommandRK(serverID, "#")

	err := pubsub.PublishRMQ(
		g.Ctx,
		g.CommandCh,
		pubsub.RMQPubOpts{
			Exchange: pubsub.GetServerCommandTopicX(),
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

func (g *Gateway) BroadcastServerCommand(command models.Command) error {
	err := pubsub.PublishRMQ(
		g.Ctx,
		g.CommandCh,
		pubsub.RMQPubOpts{
			Exchange: pubsub.GetServerCommandBroadcastX(),
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

func (g *Gateway) RequestServerIdentifiers(
	serverID string,
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
	if serverID == "*" || serverID == "" {
		return g.BroadcastServerCommand(cmd)
	}
	return g.SendServerCommand(serverID, cmd)
}

// ========================
// Handlers
// ========================

func (g *Gateway) HandlerGatewayCommands(command models.Command) pubsub.AckType {
	switch command.Name {
	case "request":
		// handle request
		err := g.Gateway2ServerResponse(models.Response{
			Name:      g.Cfg.ID,
			Status:    "",
			Data:      nil,
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Printf("failed to send response: %v", err)
		}
		log.Printf("received request command: %v", command.Args)
	case "restart":
		log.Printf("received restart command: %v", command.Args)
	case "update":
		log.Printf("received update command: %v", command.Args)
	default:
		return pubsub.NackDiscard
	}

	return pubsub.Ack
}
