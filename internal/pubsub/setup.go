package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareIonbusTopic(
	ch *amqp.Channel,
) error {
	return ch.ExchangeDeclare(
		ExchangeIonbusTopic,
		"topic",
		true,
		false,
		false,
		false,
		nil,
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
) error {
	_, err := ch.QueueDeclare(
		QueueIonbusDlq,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("unable to declare %v: %w", QueueIonbusDlq, err)
	}

	// Bind the DLQ to the DLX using the routing key
	err = ch.QueueBind(QueueIonbusDlq, "#", ExchangeIonbusDlx, false, nil)
	if err != nil {
		return fmt.Errorf(
			"unable to bind queue %v with key %v to %v: %w",
			QueueIonbusDlq, "#", ExchangeIonbusDlx, err,
		)
	}

	return nil
}

// Declares and binds a queue to the channel on the exchange.
// The queue name doesn't have to be unique, this method just
// ensures that the queue is declared before binding.
//
// queueType is specified as durable or transiet, if durable
// flags durable=true, autoDelete=false, exclusive=false,
// if transient flags durable=false, autodelete=true, exclusive=true
//
// # Routing key should make sense and follow a hierarchical structure
//
// Dead letter exchange is set by default to the ionbus_dlx bound to ionbus_dlq.
// Potential error on fail if dlx not setup.
func DeclareAndBindQueue(
	ch *amqp.Channel,
	exchange string,
	name string,
	queueType QueueType,
	key string,
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
	args := amqp.Table{
		"x-dead-letter-exchange": ExchangeIonbusDlx,
	}

	// Declare a new q
	q, err := ch.QueueDeclare(
		name,
		isDurable,
		isAutoDelete,
		isExclusive,
		false,
		args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("unable to declare %v: %w", name, err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(name, key, exchange, false, nil)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf(
			"unable to bind queue %v with key %v to %v: %w",
			name, key, exchange, err,
		)
	}

	// Return the channel and the queue
	return q, nil

}
