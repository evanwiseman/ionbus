package main

import (
	"encoding/json"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (s *Server) SendGatewayResponse(res models.Response) error {
	// Marhsal the response
	payload, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	// Wrap response in message envelope
	msg := models.Message{
		SourceID:     s.Cfg.ID,
		SourceDevice: s.Cfg.Device,
		Version:      s.Cfg.Version,
		Payload:      payload,
	}

	// Determine routing
	exchange := pubsub.RGatewayResTX()
	key := pubsub.RGatewayResTRK(res.TargetID, res.Method)
	return s.RMQ.ResponsePublisher.Publish(
		pubsub.RMQPubOpts{
			Exchange: exchange,
			Key:      key,
		},
		msg,
	)
}
