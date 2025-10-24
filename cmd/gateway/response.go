package main

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) Gateway2ServerResponse(res models.Response) error {
	key := pubsub.GetRMQGatewayCommandRK(g.Cfg.ID, "#")
	log.Printf("Sending response to %s", key)
	pubsub.PublishRMQ(
		g.Ctx,
		g.ResponseCh,
		pubsub.RMQPubOpts{
			Exchange: pubsub.GetRMQServerResponseTopicX(),
			Key:      key,
		},
		models.ContentJSON,
		res,
	)
	return nil
}
