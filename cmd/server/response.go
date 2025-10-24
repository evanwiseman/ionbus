package main

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

func (s *Server) Server2GatewayResponse(res models.Response) error {
	key := pubsub.GetRMQServerResponseRK(s.Cfg.ID, "#")
	log.Printf("Sending response to %s", key)
	pubsub.PublishRMQ(
		s.Ctx,
		s.ResponseCh,
		pubsub.RMQPubOpts{
			Exchange: pubsub.GetRMQGatewayResponseTopicX(),
			Key:      key,
		},
		models.ContentJSON,
		res,
	)
	return nil
}
