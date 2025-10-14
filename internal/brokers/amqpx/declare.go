package amqpx

import (
	"github.com/evanwiseman/ionbus/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	ch *amqp.Channel,
	exchange,
	queueName,
	key string,
	queueType routing.SimpleQueueType, // represents "durable" or "transient"
) (amqp.Queue, error) {
	// Get parameters from queueType
	var isDurable bool
	var isAutoDelete bool
	var isExclusive bool
	switch queueType {
	case routing.Durable:
		isDurable = true
		isAutoDelete = false
		isExclusive = false
	case routing.Transient:
		isDurable = false
		isAutoDelete = true
		isExclusive = true
	}

	// Declare the Queue
	q, err := ch.QueueDeclare(
		queueName,
		isDurable,
		isAutoDelete,
		isExclusive,
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, err
	}

	// Bind the Queue to the Exchange
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return amqp.Queue{}, err
	}

	// Return the queue and nil
	return q, nil
}
