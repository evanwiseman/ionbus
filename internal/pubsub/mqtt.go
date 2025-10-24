package pubsub

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
)

type MQTTPubOpts struct {
	Topic    string
	QoS      byte
	Retained bool
}

type MQTTSubOpts struct {
	Topic string
	QoS   byte
}

func PublishMQTT[T any](
	ctx context.Context,
	client mqtt.Client,
	opts MQTTPubOpts,
	contentType models.ContentType,
	val T,
) error {
	payload, err := models.Marshal(val, contentType)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	log.Printf("Publishing MQTT message %v to %s...", val, opts.Topic)
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

func SubscribeMQTT[T any](
	ctx context.Context,
	client mqtt.Client,
	opts MQTTSubOpts,
	contentType models.ContentType,
	handler func(T), // for consistency with rmq
) error {
	log.Printf("Subscribing to MQTT topic %s...", opts.Topic)
	token := client.Subscribe(opts.Topic, opts.QoS, func(client mqtt.Client, msg mqtt.Message) {
		var obj T

		if err := models.Unmarshal(msg.Payload(), contentType, &obj); err != nil {
			log.Printf("Failed to unmarhsal message: %v\n", err)
		}

		handler(obj) // Ignore ack type, paho handles it
	})

	// Wait for subscription acknowledgment with a timeout
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("timeout waiting for subscription to topic %s", opts.Topic)
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Run a goroutine to handle context cancellation
	go func() {
		<-ctx.Done()
		log.Printf("Context cancelled, unsubscribing from topic %s", opts.Topic)
		unsub := client.Unsubscribe(opts.Topic)
		unsub.WaitTimeout(3 * time.Second)
	}()

	return nil
}
