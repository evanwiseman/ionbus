package pubsub

import (
	"context"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
	log.Printf("Publishing RMQ message %v...\n", val)
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

func PublishMQTT[T any](
	ctx context.Context,
	client mqtt.Client,
	opts MQTTPublishOptions,
	contentType models.ContentType,
	val T,
) error {
	payload, err := models.Marshal(val, contentType)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	log.Printf("Publishing MQTT message %v...\n", val)
	token := client.Publish(opts.Topic, opts.QoS, opts.Retained, payload)

	// Wait or cancel via context
	done := make(chan struct{}, 1)
	go func() {
		token.Wait()
		done <- struct{}{}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		if err := token.Error(); err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
		return nil
	}
}
