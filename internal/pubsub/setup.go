package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareCommandExchange(
	ch *amqp.Channel,
) error {
	return ch.ExchangeDeclare(
		ExchangeCommandsTopic,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // args
	)
}

func DeclareDLX(
	ch *amqp.Channel,
) error {
	return ch.ExchangeDeclare(
		ExchangeIonbusDlx,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
}

func DeclareDLQ(
	ch *amqp.Channel,
	queueName string,
	routingKey string,
) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("unable to declare %v: %w", queueName, err)
	}

	// Bind the DLQ to the DLX using the routing key
	exchangeName := ExchangeIonbusDlx
	err = ch.QueueBind(queueName, routingKey, exchangeName, false, nil)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf(
			"unable to bind queue %v with key %v to %v: %w",
			queueName, routingKey, exchangeName, err,
		)
	}

	return q, nil
}

func DeclareAndBindQueue(
	ch *amqp.Channel,
	exchangeName string,
	queueName string,
	queueType QueueType,
	routingKey string,
	args amqp.Table,
) (amqp.Queue, error) {
	var isDurable bool
	var isAutoDelete bool
	var isExclusive bool
	switch queueType {
	case Durable:
		isDurable = true
		isAutoDelete = false
		isExclusive = false
	case Transient:
		isDurable = false
		isAutoDelete = true
		isExclusive = true
	}

	// Ensure DLX is set if not already provided
	if args == nil {
		args = amqp.Table{}
	}
	if _, exists := args["x-dead-letter-exchange"]; !exists {
		args["x-dead-letter-exchange"] = ExchangeIonbusDlx
	}

	// Declare a new q
	q, err := ch.QueueDeclare(
		queueName,
		isDurable,
		isAutoDelete,
		isExclusive,
		false,
		args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("unable to declare %v: %w", queueName, err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(queueName, routingKey, exchangeName, false, nil)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf(
			"unable to bind queue %v with key %v to %v: %w",
			queueName, routingKey, exchangeName, err,
		)
	}

	// Return the channel and the queue
	return q, nil

}
