package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) HandleIdentifierRequest(msg models.Message) {
	var req models.Request
	if err := models.Unmarshal(msg.Payload, models.ContentJSON, &req); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		return
	}

	switch req.SourceDevice {
	case models.DeviceServer:
		res := models.Response{
			ID:           req.ID,
			TargetID:     req.SourceID,
			TargetDevice: req.SourceDevice,
			SourceID:     g.Cfg.ID,
			SourceDevice: models.DeviceGateway,
			Action:       req.Action,
			Timestamp:    time.Now(),
			Body:         models.IdentifierBody{ID: g.Cfg.ID},
		}
		key := pubsub.RServerResTRK(req.SourceID, string(models.ActionGetIdentifiers))
		log.Printf("Sending response to %s", key)
		if err := g.RMQ.RequestFlow.Pub.Publish(
			pubsub.RMQPubOpts{
				Exchange: pubsub.RServerResTX(),
				Key:      key,
			},
			models.ContentJSON,
			res,
		); err != nil {
			log.Printf("Failed to send response: %v", err)
		}
	default:
		log.Printf("Unable to handle identifier request: unknown device")
	}
}

func (g *Gateway) HandleIdentifierResponse(msg models.Message) {
	log.Print(msg)
}
