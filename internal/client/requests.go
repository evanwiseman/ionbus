package client

import (
	"encoding/json"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (c *Client) SendRequest(req models.Request) error {
	var topic string
	if req.TargetID == "+" || req.TargetID == "" {
		topic = pubsub.MQTTBroadcast(req.TargetDevice, models.ActionRequest, req.Method)
	} else {
		topic = pubsub.MQTTTopic(req.TargetDevice, req.TargetID, models.ActionRequest, req.Method)
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

func (c *Client) RequestIdentifiers(device models.Device, id string) error {
	req := models.Request{
		Method:       string(models.MethodGetIdentifiers),
		TargetID:     id,
		TargetDevice: device,
		Payload:      nil,
	}

	return c.SendRequest(req)
}
