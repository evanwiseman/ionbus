package bridge

import (
	"context"
	"log"

	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/evanwiseman/ionbus/internal/routing"
)

func (b *Bridge) RMQToMQTT(
	ctx context.Context,
	subOpts pubsub.RMQSubscribeOptions,
	pubOpts pubsub.MQTTPublishOptions,
	contentType routing.ContentType,
) error {
	return pubsub.SubscribeRMQ(
		ctx,
		b.RMQCh,
		subOpts,
		contentType,
		func(msg any) routing.AckType {
			if err := pubsub.PublishMQTT(
				ctx,
				b.MQTTClient,
				pubOpts,
				contentType,
				msg,
			); err != nil {
				log.Printf("Bridge RMQ->MQTT publish failed: %v", err)
				return routing.NackRequeue
			}
			return routing.Ack
		},
	)
}

func (b *Bridge) MQTTToRMQ(
	ctx context.Context,
	subOpts pubsub.MQTTSubscribeOptions,
	pubOpts pubsub.RMQPublishOptions,
	contentType routing.ContentType,
) error {
	return pubsub.SubscribeMQTT(
		ctx,
		b.MQTTClient,
		subOpts,
		contentType,
		func(msg any) routing.AckType {
			if err := pubsub.PublishRMQ(
				ctx,
				b.RMQCh,
				pubOpts,
				contentType,
				msg,
			); err != nil {
				log.Printf("Bridge MQTT->RMQ publish failed: %v", err)
				return routing.NackRequeue
			}
			return routing.Ack
		},
	)
}
