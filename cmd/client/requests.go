package main

import (
	"encoding/json"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (c *Client) SendGatewayRequest(req models.Request) error {
	var topic string
	if req.TargetID == "+" || req.TargetID == "" {
		topic = pubsub.MGatewayReqB(req.Method)
	} else {
		topic = pubsub.MGatewayReqT(req.TargetID, req.Method)
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error: failed to marshal request: %w", err)
	}

	msg := models.Message{
		SourceID:     c.Cfg.ID,
		SourceDevice: c.Cfg.Device,
		Version:      c.Cfg.Version,
		Payload:      payload,
	}

	return c.MQTT.RequestPublisher.Publish(
		pubsub.MQTTPubOpts{
			Topic:    topic,
			QoS:      byte(1),
			Retained: false,
		},
		msg,
	)
}

func (c *Client) RequestGatewayIdentifiers(gatewayID string) error {
	req := models.Request{
		Method:       string(models.MethodGetIdentifiers),
		TargetID:     gatewayID,
		TargetDevice: models.DeviceGateway,
		Payload:      nil,
	}

	return c.SendGatewayRequest(req)
}
