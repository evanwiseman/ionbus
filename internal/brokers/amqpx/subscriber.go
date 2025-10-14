package amqpx

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Subscribe[T any](conn *amqp.Connection,
	ch *amqp.Channel,
	queueName string,
	contentType routing.ContentType,
	handler func(T) routing.AckType,
) error {
	// Limit prefetch so other servers can clean process queue
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	// Get a chan of deliveries by consuming message
	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			// Unmarshal message
			var obj T
			if err = routing.Unmarshal(msg.Body, contentType, &obj); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Nack(false, false)
				continue
			}

			// Send the object of T to the handler
			ackType := handler(obj)
			switch ackType {
			case routing.Ack:
				msg.Ack(false)
			case routing.NackRequeue:
				msg.Nack(false, true)
			case routing.NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()

	return nil
}
