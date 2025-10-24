package main

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (c *Client) Client2GatewayResponse(res models.Response) error {
	topic := pubsub.GetMQTTGatewayResponseBroadcast("")
	log.Println(topic)
	pubsub.PublishMQTT(
		c.Ctx,
		c.MQTTClient,
		pubsub.MQTTPubOpts{
			Topic: topic,
			QoS:   1,
		},
		models.ContentJSON,
		res,
	)
	return nil
}
