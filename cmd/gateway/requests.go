package main

import (
	"encoding/json"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) SendServerRequest(req models.Request) error {
	var exchange string
	var key string
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
		SourceID:     g.Cfg.ID,
		SourceDevice: g.Cfg.Device,
		Version:      g.Cfg.Version,
		Payload:      payload,
	}

	return g.RMQ.RequestPublisher.Publish(
		pubsub.RMQPubOpts{
			Exchange: exchange,
			Key:      key,
		},
		msg,
	)
}

func (g *Gateway) RequestServerIdentifiers(serverID string) error {
	req := models.Request{
		Method:       string(models.MethodGetIdentifiers),
		TargetID:     serverID,
		TargetDevice: models.DeviceServer,
		Payload:      nil,
	}

	return g.SendServerRequest(req)
}

func (g *Gateway) SendClientRequest(req models.Request) error {
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
		SourceID:     g.Cfg.ID,
		SourceDevice: g.Cfg.Device,
		Version:      g.Cfg.Version,
		Payload:      payload,
	}

	return g.MQTT.RequestPublisher.Publish(
		pubsub.MQTTPubOpts{
			Topic:    topic,
			QoS:      byte(1),
			Retained: false,
		},
		msg,
	)
}

func (g *Gateway) RequestClientIdentifiers(clientID string) error {
	req := models.Request{
		Method:       string(models.MethodGetIdentifiers),
		TargetID:     clientID,
		TargetDevice: models.DeviceClient,
		Payload:      nil,
	}

	return g.SendClientRequest(req)
}
