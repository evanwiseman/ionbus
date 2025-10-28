package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/google/uuid"
)

func (c *Client) SendGatewayRequest(req models.Request) error {
	if req.ID == "" {
		req.ID = uuid.NewString()
	}
	req.Timestamp = time.Now()

	var topic string
	if req.TargetID == "+" || req.TargetID == "" {
		topic = pubsub.MGatewayReqB(string(req.Action))
	} else {
		topic = pubsub.MGatewayReqT(req.TargetID, string(req.Action))
	}

	log.Printf("Sending request to '%s'", topic)
	if err := c.MQTT.RequestFlow.Pub.Publish(
		pubsub.MQTTPubOpts{
			Topic: topic,
			QoS:   byte(1),
		},
		models.ContentJSON,
		req,
	); err != nil {
		log.Printf("Failed to send request: %v", err)
		return err
	}

	return nil
}

func (c *Client) RequestGatewayIdentifiers(gatewayID string) error {
	req := models.Request{
		ID:           uuid.NewString(),
		SourceID:     c.Cfg.ID,
		SourceDevice: models.DeviceClient,
		TargetID:     gatewayID,
		TargetDevice: models.DeviceGateway,
		Action:       models.ActionGetIdentifiers,
	}

	return c.SendGatewayRequest(req)
}
