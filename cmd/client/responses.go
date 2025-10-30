package main

import (
	"encoding/json"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (c *Client) SendGatewayResponse(res models.Response) error {
	// Marhsal the response
	payload, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	// Wrap response in message envelope
	msg := models.Message{
		SourceID:     c.Cfg.ID,
		SourceDevice: c.Cfg.Device,
		Version:      c.Cfg.Version,
		Payload:      payload,
	}

	// Determine routing
	topic := pubsub.MGatewayResT(res.TargetID, res.Method)
	return c.MQTT.ResponsePublisher.Publish(
		pubsub.MQTTPubOpts{
			Topic:    topic,
			QoS:      byte(1),
			Retained: false,
		},
		msg,
	)
}
