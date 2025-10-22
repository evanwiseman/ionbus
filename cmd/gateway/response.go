package main

import (
	"fmt"
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (g *Gateway) SendResponseRMQ(key string, res models.Response) error {
	err := pubsub.PublishRMQ(
		g.Ctx,
		g.RMQPublishCh,
		pubsub.RMQPublishOptions{
			Exchange: pubsub.ExchangeIonbusDirect,
			Key:      key,
		},
		models.ContentJSON,
		res,
	)

	if err != nil {
		return fmt.Errorf("failed to send response with key %s: %w", key, err)
	}
	return nil
}

func (g *Gateway) SendGatewayResponseRMQ(res models.Response) error {
	key := fmt.Sprintf("%s.%s", pubsub.GatewaysPrefix, pubsub.ResponsesPrefix)
	log.Printf("Sending response to %s", key)
	return g.SendResponseRMQ(key, res)
}
