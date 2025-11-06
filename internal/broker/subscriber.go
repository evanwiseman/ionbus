package broker

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Subscriber interface {
	Subscribe(opts any, handler func([]byte) error) error
}

type MQTTSubOpts struct {
	Topic string
	QoS   byte
}

type MQTTSubscriber struct {
	Subscriber
	Ctx    context.Context
	Client mqtt.Client
}

func NewMQTTSubscriber(ctx context.Context, client mqtt.Client) *MQTTSubscriber {
	return &MQTTSubscriber{Ctx: ctx, Client: client}
}

func (s *MQTTSubscriber) Subscribe(opts MQTTSubOpts, handler func([]byte) error) error {
	token := s.Client.Subscribe(opts.Topic, opts.QoS, func(_ mqtt.Client, msg mqtt.Message) {
		if err := handler(msg.Payload()); err != nil {
			log.Printf("handler failed: %v", err)
		}
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

type RMQSubOpts struct {
	QueueName     string
	Consumer      string
	AutoAck       bool
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
	PrefetchCount int
	PrefetchSize  int
	QosGlobal     bool
	Args          amqp.Table
}

type RMQSubscriber struct {
	Subscriber
	Ctx context.Context
	Ch  *amqp.Channel
}

func NewRMQSubscriber(ctx context.Context, ch *amqp.Channel) *RMQSubscriber {
	return &RMQSubscriber{Ctx: ctx, Ch: ch}
}

func (s *RMQSubscriber) Subscribe(opts RMQSubOpts, handler func([]byte) error) error {
	// Limit prefetch so other servers can clean process queue
	if err := s.Ch.Qos(opts.PrefetchCount, opts.PrefetchSize, opts.QosGlobal); err != nil {
		return fmt.Errorf("failed to set qos: %w", err)
	}

	// Get a chan of deliveries by consuming message
	msgs, err := s.Ch.Consume(
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
			case <-s.Ctx.Done():
				log.Printf("Context cancelled, stopping consumer for queue %s", opts.QueueName)
				// Optionally: cancel the consumer on the channel
				if err := s.Ch.Cancel(opts.Consumer, false); err != nil {
					log.Printf("Failed to cancel consumer: %v", err)
				}
				return
			case msg, ok := <-msgs:
				if !ok {
					// Channel closed, exit
					return
				}

				err := handler(msg.Body)
				if err != nil {
					log.Printf("Failed to handle: %v", err)
					msg.Nack(false, false)
				}
				msg.Ack(false)
			}
		}
	}()

	return nil
}
func (s *RMQSubscriber) Close() {
	if s.Ch != nil {
		s.Ch.Close()
	}
}
