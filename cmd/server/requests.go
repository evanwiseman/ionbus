package main

import (
	"encoding/json"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (s *Server) SendGatewayRequest(req models.Request) error {
	var exchange string
	var key string

	// Broadcast or targeted request
	if req.TargetID == "*" || req.TargetID == "" {
		exchange = pubsub.RMQBroadcastX(req.TargetDevice, models.ActionRequest)
		key = pubsub.RMQBroadcastRK(req.TargetDevice, models.ActionRequest, req.Method)
	} else {
		exchange = pubsub.RMQTopicX(req.TargetDevice, models.ActionRequest)
		key = pubsub.RMQTopicRK(req.TargetDevice, req.TargetID, models.ActionRequest, req.Method)
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error: failed to marshal request: %w", err)
	}

	msg := models.Message{
		SourceID:     s.Cfg.ID,
		SourceDevice: s.Cfg.Device,
		Version:      s.Cfg.Version,
		Payload:      payload,
	}

	return s.RMQ.RequestPublisher.Publish(
		pubsub.RMQPubOpts{
			Exchange: exchange,
			Key:      key,
		},
		msg,
	)
}

func (s *Server) RequestGatewayIdentifiers(gatewayID string) error {
	req := models.Request{
		Method:       string(models.MethodGetIdentifiers),
		TargetID:     gatewayID,
		TargetDevice: models.DeviceGateway,
		Payload:      nil,
	}

	return s.SendGatewayRequest(req)
}
