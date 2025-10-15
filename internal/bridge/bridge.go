package bridge

import (
	"context"
	"log"

	"github.com/evanwiseman/ionbus/internal/pubsub"
	"github.com/evanwiseman/ionbus/internal/routing"
)

// RMQToMQTT creates a unidirectional bridge that forwards messages from a RabbitMQ queue to an MQTT topic.
//
// This function subscribes to a RabbitMQ queue and publishes each received message to the specified
// MQTT topic. Messages are acknowledged based on publish success:
//   - Successful publish: Message is ACKed and removed from the RabbitMQ queue
//   - Failed publish: Message is NACKed with requeue, allowing retry
//
// Parameters:
//   - ctx: Context for cancellation. When cancelled, the subscription stops gracefully.
//   - subOpts: RabbitMQ subscription configuration (queue name, prefetch, QoS, etc.)
//   - pubOpts: MQTT publish configuration (topic, QoS level, retained flag)
//   - contentType: Serialization format (routing.ContentJSON or routing.ContentGOB)
//
// Prerequisites:
//   - RabbitMQ queue must exist (use ch.QueueDeclare before calling)
//   - RabbitMQ channel must be open and healthy
//   - MQTT client must be connected
//
// Error Handling:
//   - Returns error if RabbitMQ subscription fails to start
//   - Publish errors are logged and cause message requeue (not returned)
//   - Unmarshal errors are logged and cause message discard (NACK without requeue)
//
// Loop Prevention:
//
//	This is a ONE-WAY bridge. To avoid infinite loops, ensure the destination MQTT topic
//	does not bridge back to the source RabbitMQ queue. Use distinct topic/queue names for
//	each direction.
//
// Example - Device Commands Flow:
//
//	err := bridge.RMQToMQTT(
//	    ctx,
//	    pubsub.RMQSubscribeOptions{
//	        QueueName:     "commands.outbound",  // Commands waiting to be sent
//	        PrefetchCount: 100,                  // Process up to 100 at once
//	        AutoAck:       false,                // Manual ack for reliability
//	    },
//	    pubsub.MQTTPublishOptions{
//	        Topic:    "device/+/cmd",            // Publish to device command topic
//	        Qos:      1,                         // At least once delivery
//	        Retained: false,
//	    },
//	    routing.ContentJSON,
//	)
//
// Recommended Naming Pattern:
//
//	RMQ Queue:  "commands.outbound"   -> MQTT Topic: "device/+/cmd"
//	RMQ Queue:  "status.outbound"     -> MQTT Topic: "device/+/status"
//
// Anti-Pattern (CREATES LOOPS):
//
//	RMQ Queue: "data" -> MQTT Topic: "data" -> RMQ Queue: "data"
func (b *Bridge) RMQToMQTT(
	ctx context.Context,
	subOpts pubsub.RMQSubscribeOptions,
	pubOpts pubsub.MQTTPublishOptions,
	contentType routing.ContentType,
) error {
	return pubsub.SubscribeRMQ(
		ctx,
		b.RMQCh,
		subOpts,
		contentType,
		func(msg any) routing.AckType {
			if err := pubsub.PublishMQTT(
				ctx,
				b.MQTTClient,
				pubOpts,
				contentType,
				msg,
			); err != nil {
				log.Printf("Bridge RMQ->MQTT publish failed: %v", err)
				return routing.NackRequeue
			}
			return routing.Ack
		},
	)
}

// MQTTToRMQ creates a unidirectional bridge that forwards messages from an MQTT topic to a RabbitMQ exchange.
//
// This function subscribes to an MQTT topic and publishes each received message to the specified
// RabbitMQ exchange with a routing key. The MQTT subscription remains active until the context
// is cancelled.
//
// Parameters:
//   - ctx: Context for cancellation. When cancelled, unsubscribes from MQTT topic.
//   - subOpts: MQTT subscription configuration (topic pattern, QoS level)
//   - pubOpts: RabbitMQ publish configuration (exchange, routing key, mandatory/immediate flags)
//   - contentType: Serialization format (routing.ContentJSON or routing.ContentGOB)
//
// Prerequisites:
//   - RabbitMQ exchange must exist (use ch.ExchangeDeclare before calling)
//   - Target queue should be bound to the exchange with appropriate routing key
//   - MQTT client must be connected
//   - MQTT broker must be reachable
//
// Error Handling:
//   - Returns error if MQTT subscription fails to start or times out
//   - Publish errors are logged and trigger message requeue attempt (ACK type ignored for MQTT)
//   - Unmarshal errors are logged but message is lost (MQTT has no requeue mechanism)
//
// MQTT Limitations:
//
//	Unlike RabbitMQ, MQTT QoS 0/1/2 doesn't support selective requeue. The AckType return
//	from the handler is used for consistency but has no effect on MQTT message handling.
//	Consider using QoS 1 or 2 for important messages.
//
// Loop Prevention:
//
//	This is a ONE-WAY bridge. To avoid infinite loops, ensure the destination RabbitMQ
//	queue does not bridge back to the source MQTT topic. Use distinct topic/queue names
//	for each direction.
//
// Example - Device Telemetry Flow:
//
//	err := bridge.MQTTToRMQ(
//	    ctx,
//	    pubsub.MQTTSubscribeOptions{
//	        Topic: "device/+/telemetry",         // Subscribe to all device telemetry
//	        Qos:   1,                            // At least once delivery
//	    },
//	    pubsub.RMQPublishOptions{
//	        Exchange:  "telemetry",              // Publish to telemetry exchange
//	        Key:       "device.data.inbound",    // Routing key for processing queue
//	        Mandatory: false,
//	        Immediate: false,
//	    },
//	    routing.ContentJSON,
//	)
//
// Recommended Naming Pattern:
//
//	MQTT Topic: "device/+/telemetry"  -> RMQ Queue:  "telemetry.inbound"
//	MQTT Topic: "device/+/alert"      -> RMQ Queue:  "alerts.inbound"
//
// Anti-Pattern (CREATES LOOPS):
//
//	MQTT Topic: "sync" -> RMQ Queue: "sync" -> MQTT Topic: "sync"
func (b *Bridge) MQTTToRMQ(
	ctx context.Context,
	subOpts pubsub.MQTTSubscribeOptions,
	pubOpts pubsub.RMQPublishOptions,
	contentType routing.ContentType,
) error {
	return pubsub.SubscribeMQTT(
		ctx,
		b.MQTTClient,
		subOpts,
		contentType,
		func(msg any) routing.AckType {
			if err := pubsub.PublishRMQ(
				ctx,
				b.RMQCh,
				pubOpts,
				contentType,
				msg,
			); err != nil {
				log.Printf("Bridge MQTT->RMQ publish failed: %v", err)
				return routing.NackRequeue
			}
			return routing.Ack
		},
	)
}

// BidirectionalBridge sets up TWO separate one-way bridges with distinct paths.
// This prevents loops by using different topics/queues for each direction.
//
// IMPORTANT: Ensure RabbitMQ queues and exchanges are declared BEFORE calling this.
// The bridge doesn't create infrastructure - it only moves messages.
//
// Example usage:
//
//	// 1. Declare RabbitMQ infrastructure first
//	ch.QueueDeclare("device.commands.outbound", true, false, false, false, nil)
//	ch.QueueDeclare("device.telemetry.inbound", true, false, false, false, nil)
//	ch.ExchangeDeclare("telemetry", "topic", true, false, false, false, nil)
//	ch.QueueBind("device.telemetry.inbound", "device.#", "telemetry", false, nil)
//
//	// 2. Create bridge configuration
//	bridge.BidirectionalBridge(
//	  ctx,
//	  pubsub.RMQSubscribeOptions{QueueName: "device.commands.outbound"},
//	  pubsub.MQTTPublishOptions{Topic: "device/+/commands"},
//	  pubsub.MQTTSubscribeOptions{Topic: "device/+/telemetry"},
//	  pubsub.RMQPublishOptions{Exchange: "telemetry", Key: "device.telemetry"},
//	  routing.ContentJSON,
//	)
//
// Flow:
//
//	RMQ "device.commands.outbound" -> MQTT "device/+/commands"
//	MQTT "device/+/telemetry"      -> RMQ "device.telemetry.inbound"
//
// No loop possible because paths are completely separate.
func (b *Bridge) BidirectionalBridge(
	ctx context.Context,
	rmqSubOpts pubsub.RMQSubscribeOptions,
	mqttPubOpts pubsub.MQTTPublishOptions,
	mqttSubOpts pubsub.MQTTSubscribeOptions,
	rmqPubOpts pubsub.RMQPublishOptions,
	contentType routing.ContentType,
) error {
	// Start RMQ -> MQTT bridge
	errChan1 := make(chan error, 1)
	go func() {
		err := b.RMQToMQTT(ctx, rmqSubOpts, mqttPubOpts, contentType)
		if err != nil {
			errChan1 <- err
		}
	}()

	// Start MQTT -> RMQ bridge
	errChan2 := make(chan error, 1)
	go func() {
		err := b.MQTTToRMQ(ctx, mqttSubOpts, rmqPubOpts, contentType)
		if err != nil {
			errChan2 <- err
		}
	}()

	// Return first error encountered
	select {
	case err := <-errChan1:
		return err
	case err := <-errChan2:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
