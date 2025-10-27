package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/google/uuid"
)

func (s *Server) SendGatewayRequest(req models.Request) error {
	if req.ID == "" {
		req.ID = uuid.NewString()
	}
	req.Timestamp = time.Now()

	var exchange string
	var key string
	if req.TargetID == "*" || req.TargetID == "" {
		exchange = pubsub.RGatewayReqBX()
		key = pubsub.RGatewayReqBRK(string(models.ActionGetIdentifiers))
	} else {
		exchange = pubsub.RGatewayReqTX()
		key = pubsub.RGatewayReqTRK(req.TargetID, string(models.ActionGetIdentifiers))
	}

	log.Printf("Sending request to '%s' on '%s'", key, exchange)
	if err := s.RMQ.RequestFlow.Pub.Publish(
		pubsub.RMQPubOpts{
			Exchange: exchange,
			Key:      key,
		},
		models.ContentJSON,
		req,
	); err != nil {
		log.Printf("Failed to send request: %v", err)
		return err
	}

	return nil
}

func (s *Server) RequestGatewayIdentifiers(gatewayID string) error {
	req := models.Request{
		ID:           uuid.NewString(),
		SourceID:     s.Cfg.ID,
		SourceDevice: models.DeviceServer,
		TargetID:     gatewayID,
		TargetDevice: models.DeviceGateway,
		Action:       models.ActionGetIdentifiers,
	}

	return s.SendGatewayRequest(req)
}
