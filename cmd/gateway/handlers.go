package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) HandlerGatewayCommands(command models.Command) pubsub.AckType {
	switch command.Name {
	case "request":
		// handle request
		err := g.SendGatewayResponseRMQ(models.Response{
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
