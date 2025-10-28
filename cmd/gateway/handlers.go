package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) HandleIdentifierRequest(msg models.Message) {
	var req models.Request
	if err := models.Unmarshal(msg.Payload, models.ContentJSON, &req); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		return
	}

	res := models.Response{
		ID:           req.ID,
		TargetID:     req.SourceID,
		TargetDevice: req.SourceDevice,
		SourceID:     g.Cfg.ID,
		SourceDevice: models.DeviceGateway,
		Action:       req.Action,
		Timestamp:    time.Now(),
		Body:         models.IdentifierBody{ID: g.Cfg.ID},
	}

	switch req.SourceDevice {
	case models.DeviceServer:
		key := pubsub.RServerResTRK(req.SourceID, string(req.Action))
		log.Printf("Sending response to %s", key)
		if err := g.RMQ.RequestFlow.Pub.Publish(
			pubsub.RMQPubOpts{
				Exchange: pubsub.RServerResTX(),
				Key:      key,
			},
			models.ContentJSON,
			res,
		); err != nil {
			log.Printf("Failed to send response: %v", err)
		}
	case models.DeviceClient:
		topic := pubsub.MClientResT(req.SourceID, string(req.Action))
		log.Printf("Sending response to %s", topic)
		if err := g.MQTT.RequestFlow.Pub.Publish(
			pubsub.MQTTPubOpts{
				Topic: topic,
				QoS:   byte(1),
			},
			models.ContentJSON,
			res,
		); err != nil {
			log.Printf("Failed to send response: %v", err)
		}
	default:
		log.Printf("Unable to handle identifier request: unknown device")
	}
}

func (g *Gateway) HandleIdentifierResponse(msg models.Message) {
	log.Print(msg)
}
