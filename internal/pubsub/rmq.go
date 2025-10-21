package pubsub

import (
	"context"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishRMQ[T any](
	ctx context.Context,
	ch *amqp.Channel,
	opts RMQPublishOptions,
	contentType models.ContentType,
	val T,
) error {
	// Marshal val to JSON []byte
	payload, err := models.Marshal(val, contentType)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	// Publish the message to the exchange
	err = ch.PublishWithContext(
		ctx,
		opts.Exchange,
		opts.Key,
		opts.Mandatory,
		opts.Immediate,
		amqp.Publishing{
			ContentType: string(contentType),
			Body:        payload,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func SubscribeRMQ[T any](
	ctx context.Context,
	ch *amqp.Channel,
	opts RMQSubscribeOptions,
	contentType models.ContentType,
	handler func(T) AckType,
) error {
	// Limit prefetch so other servers can clean process queue
	if err := ch.Qos(opts.PrefetchCount, opts.PrefetchSize, opts.QosGlobal); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Get a chan of deliveries by consuming message
	msgs, err := ch.Consume(
		opts.QueueName,
		opts.Consumer,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				// Unmarshal message
				var obj T
				if err := models.Unmarshal(msg.Body, contentType, &obj); err != nil {
					msg.Nack(false, false)
					continue
				}

				// Send the object of T to the handler
				ackType := handler(obj)
				switch ackType {
				case Ack:
					msg.Ack(false)
				case NackRequeue:
					msg.Nack(false, true)
				case NackDiscard:
					msg.Nack(false, false)
				}
			}
		}
	}()

	return nil
}
