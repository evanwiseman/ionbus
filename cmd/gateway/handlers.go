package main

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func HandlerCommands(command models.Command) pubsub.AckType {
	switch command.Name {
	case "request":
		// handle request
		log.Printf("received request command: %v", command.Params)
	case "restart":
		log.Printf("received restart command: %v", command.Params)
	case "update":
		log.Printf("received update command: %v", command.Params)
	default:
		return pubsub.NackDiscard
	}

	return pubsub.Ack
}
