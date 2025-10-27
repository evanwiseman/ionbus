package pubsub

import (
	"context"
	"fmt"
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RMQPubOpts struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
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

type RMQFlow struct {
	Pub *RMQPublisher
	Sub *RMQSubscriber
}

func NewRMQFlow(ctx context.Context, pubCh *amqp.Channel, subCh *amqp.Channel) *RMQFlow {
	return &RMQFlow{
		Pub: NewRMQPublisher(ctx, pubCh),
		Sub: NewRMQSubscriber(ctx, subCh),
	}
}

type RMQSubscriber struct {
	Ctx context.Context
	Ch  *amqp.Channel
	Mux *MessageMux
}

func NewRMQSubscriber(ctx context.Context, ch *amqp.Channel) *RMQSubscriber {
	return &RMQSubscriber{Ctx: ctx, Ch: ch, Mux: NewMessageMux(".")}
}

func (s *RMQSubscriber) Subscribe(opts RMQSubOpts) error {
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
			case d, ok := <-msgs:
				if !ok {
					// Channel closed, exit
					return
				}
				topic := d.RoutingKey
				err := s.Mux.Dispatch(models.Message{
					Topic:   topic,
					Payload: d.Body,
					Source:  "rmq",
				})
				if err != nil {
					log.Printf("Failed to dispatch: %v", err)
				}
			}
		}
	}()

	return nil
}

type RMQPublisher struct {
	Ctx context.Context
	Ch  *amqp.Channel
}

func NewRMQPublisher(ctx context.Context, ch *amqp.Channel) *RMQPublisher {
	return &RMQPublisher{Ctx: ctx, Ch: ch}
}

func (p *RMQPublisher) Publish(opts RMQPubOpts, contentType models.ContentType, val any) error {
	payload, err := models.Marshal(val, contentType)
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
			ContentType: string(contentType),
			Body:        payload,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
