package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (s *Server) HandleIdentifierRequest(msg models.Message) {
	var req models.Request
	if err := models.Unmarshal(msg.Payload, models.ContentJSON, &req); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		return
	}

	switch req.SourceDevice {
	case models.DeviceGateway:
		res := models.Response{
			ID:           req.ID,
			TargetID:     req.SourceID,
			TargetDevice: req.SourceDevice,
			SourceID:     s.Cfg.ID,
			SourceDevice: models.DeviceServer,
			Action:       req.Action,
			Timestamp:    time.Now(),
			Body:         models.IdentifierBody{ID: s.Cfg.ID},
		}

		key := pubsub.RGatewayResTRK(req.SourceID, string(models.ActionGetIdentifiers))
		log.Printf("Sending response to %s", key)
		if err := s.RMQ.ResponseFlow.Pub.Publish(
			pubsub.RMQPubOpts{
				Exchange: pubsub.RGatewayResTX(),
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

func (s *Server) HandleIdentifierResponse(msg models.Message) {
	log.Print(msg)
}
