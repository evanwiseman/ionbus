package pubsub

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeRMQ[T any](
	ctx context.Context,
	conn *amqp.Connection,
	ch *amqp.Channel,
	queueName string,
	prefetch int,
	contentType routing.ContentType,
	handler func(T) routing.AckType,
) error {
	// Limit prefetch so other servers can clean process queue
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
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
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping consumer")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Channel closed, consumer stopping")
					return
				}

				// Unmarshal message
				var obj T
				if err := routing.Unmarshal(msg.Body, contentType, &obj); err != nil {
					log.Printf("Failed to unmarshal message: %v\n", err)
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
		}
	}()

	return nil
}

func SubscribeMQTT[T any](
	client mqtt.Client,
	topic string,
	qos byte,
	contentType routing.ContentType,
	handler func(T) routing.AckType, // for consistency with rmq
) error {
	token := client.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		var obj T

		if err := routing.Unmarshal(msg.Payload(), contentType, &obj); err != nil {
			log.Printf("Failed to unmarhsal message: %v\n", err)
		}

		_ = handler(obj) // Ignore ack type, paho handles it
	})

	// Wait for subscription acknowledgment with a timeout
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("timeout waiting for subscription to topic %s", topic)
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	return nil
}
