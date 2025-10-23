package main

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) SendGatewayResponseRMQ(res models.Response) error {
	key := pubsub.GetGatewayResponseRK(g.Cfg.ID, "#")
	log.Printf("Sending response to %s", key)
	pubsub.PublishRMQ(
		g.Ctx,
		g.ResponseCh,
		pubsub.RMQPublishOptions{
			Exchange: pubsub.GetGatewayResponseTopicX(),
			Key:      key,
		},
		models.ContentJSON,
		res,
	)
	return nil
}
