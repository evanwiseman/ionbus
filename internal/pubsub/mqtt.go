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

type MQTTFlow struct {
	Pub *MQTTPublisher
	Sub *MQTTSubscriber
}

func NewMQTTFlow(ctx context.Context, client mqtt.Client) *MQTTFlow {
	return &MQTTFlow{
		Pub: NewMQTTPublisher(ctx, client),
		Sub: NewMQTTSubscriber(ctx, client, NewMessageMux("/")),
	}
}

type MQTTSubscriber struct {
	Ctx    context.Context
	Client mqtt.Client
	Mux    *MessageMux
}

func NewMQTTSubscriber(ctx context.Context, client mqtt.Client, mux *MessageMux) *MQTTSubscriber {
	return &MQTTSubscriber{Ctx: ctx, Client: client, Mux: mux}
}

func (s *MQTTSubscriber) Subscribe(opts MQTTSubOpts) error {
	token := s.Client.Subscribe(opts.Topic, opts.QoS, func(client mqtt.Client, msg mqtt.Message) {
		s.Mux.Dispatch(models.Message{
			Topic:   msg.Topic(),
			Payload: msg.Payload(),
			Source:  "mqtt",
		})
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
		<-s.Ctx.Done()
		log.Printf("Context cancelled, unsubscribing from topic %s", opts.Topic)
		unsub := s.Client.Unsubscribe(opts.Topic)
		unsub.WaitTimeout(3 * time.Second)
	}()

	return nil
}

type MQTTPublisher struct {
	Ctx    context.Context
	Client mqtt.Client
}

func NewMQTTPublisher(ctx context.Context, client mqtt.Client) *MQTTPublisher {
	return &MQTTPublisher{Ctx: ctx, Client: client}
}

func (p *MQTTPublisher) Publish(opts MQTTPubOpts, contentType models.ContentType, val any) error {
	payload, err := models.Marshal(val, contentType)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	token := p.Client.Publish(opts.Topic, opts.QoS, opts.Retained, payload)

	// Wait or cancel via context
	done := make(chan struct{}, 1)
	go func() {
		token.Wait()
		done <- struct{}{}
		close(done)
	}()

	select {
	case <-p.Ctx.Done():
		return p.Ctx.Err()
	case <-done:
		if err := token.Error(); err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
		return nil
	}
}
