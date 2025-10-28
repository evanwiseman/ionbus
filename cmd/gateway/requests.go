package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/google/uuid"
)

func (g *Gateway) SendServerRequest(req models.Request) error {
	if req.ID == "" {
		req.ID = uuid.NewString()
	}
	req.Timestamp = time.Now()

	var exchange string
	var key string
	if req.TargetID == "*" || req.TargetID == "" {
		exchange = pubsub.RServerReqBX()
		key = pubsub.RServerReqBRK(string(req.Action))
	} else {
		exchange = pubsub.RServerReqTX()
		key = pubsub.RServerReqTRK(req.TargetID, string(req.Action))
	}

	log.Printf("Sending request to '%s' on '%s'", key, exchange)
	if err := g.RMQ.RequestFlow.Pub.Publish(
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

func (g *Gateway) RequestServerIdentifiers(serverID string) error {
	req := models.Request{
		ID:           uuid.NewString(),
		SourceID:     g.Cfg.ID,
		SourceDevice: models.DeviceGateway,
		TargetID:     serverID,
		TargetDevice: models.DeviceServer,
		Action:       models.ActionGetIdentifiers,
	}

	return g.SendServerRequest(req)
}

func (g *Gateway) SendClientRequest(req models.Request) error {
	if req.ID == "" {
		req.ID = uuid.NewString()
	}
	req.Timestamp = time.Now()

	var topic string
	if req.TargetID == "+" || req.TargetID == "" {
		topic = pubsub.MClientReqB(string(req.Action))
	} else {
		topic = pubsub.MClientReqT(g.Cfg.ID, string(req.Action))
	}

	log.Printf("Sending request to '%s'", topic)
	if err := g.MQTT.RequestFlow.Pub.Publish(
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

func (g *Gateway) RequestClientIdentifiers(clientID string) error {
	req := models.Request{
		ID:           uuid.NewString(),
		SourceID:     g.Cfg.ID,
		SourceDevice: models.DeviceGateway,
		TargetID:     clientID,
		TargetDevice: models.DeviceClient,
		Action:       models.ActionGetIdentifiers,
	}

	return g.SendClientRequest(req)
}
