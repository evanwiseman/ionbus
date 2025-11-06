package broker

import (
	"context"
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(opts any, msg any) error
}

type MQTTPubOpts struct {
	Topic    string
	QoS      byte
	Retained bool
}

type MQTTPublisher struct {
	Publisher
	Ctx    context.Context
	Client mqtt.Client
}

func NewMQTTPublisher(ctx context.Context, client mqtt.Client) *MQTTPublisher {
	return &MQTTPublisher{Ctx: ctx, Client: client}
}

func (p *MQTTPublisher) Publish(opts MQTTPubOpts, msg any) error {
	payload, err := json.Marshal(msg)
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

type RMQPubOpts struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
}

type RMQPublisher struct {
	Publisher
	Ctx context.Context
	Ch  *amqp.Channel
}

func NewRMQPublisher(ctx context.Context, ch *amqp.Channel) *RMQPublisher {
	return &RMQPublisher{Ctx: ctx, Ch: ch}
}

func (p *RMQPublisher) Publish(opts RMQPubOpts, msg any) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	if err := p.Ch.PublishWithContext(
		p.Ctx,
		opts.Exchange,
		opts.Key,
		opts.Mandatory,
		opts.Immediate,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (p *RMQPublisher) Close() {
	if p.Ch != nil {
		p.Ch.Close()
	}
}
