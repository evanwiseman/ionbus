package amqpx

import (
	"context"
	"fmt"

	"github.com/evanwiseman/ionbus/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Publish[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T,
	contentType routing.ContentType,
) error {
	// Marshal val to JSON []byte
	data, err := routing.Marshal(val, contentType)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %v", err)
	}

	// Publish the message to the exchange
	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: string(contentType),
			Body:        data,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
