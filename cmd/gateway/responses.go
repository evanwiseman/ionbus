package main

import (
	"encoding/json"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) SendServerResponse(res models.Response) error {
	// Marhsal the response
	payload, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	// Wrap response in message envelope
	msg := models.Message{
		SourceID:     g.Cfg.ID,
		SourceDevice: g.Cfg.Device,
		Version:      g.Cfg.Version,
		Payload:      payload,
	}

	// Determine routing
	exchange := pubsub.RServerResTX()
	key := pubsub.RServerResTRK(res.TargetID, res.Method)

	return g.RMQ.ResponsePublisher.Publish(
		pubsub.RMQPubOpts{
			Exchange: exchange,
			Key:      key,
		},
		msg,
	)
}

func (g *Gateway) SendClientResponse(res models.Response) error {
	// Marhsal the response
	payload, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	// Wrap response in message envelope
	msg := models.Message{
		SourceID:     g.Cfg.ID,
		SourceDevice: g.Cfg.Device,
		Version:      g.Cfg.Version,
		Payload:      payload,
	}

	// Determine routing
	topic := pubsub.MClientResT(res.TargetID, res.Method)
	return g.MQTT.ResponsePublisher.Publish(
		pubsub.MQTTPubOpts{
			Topic:    topic,
			QoS:      byte(1),
			Retained: false,
		},
		msg,
	)
}
