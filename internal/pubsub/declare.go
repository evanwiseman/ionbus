package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ========================
// Commands
// ========================

func DeclareGatewayCommandTopicX(
	ch *amqp.Channel,
) error {
	return ch.ExchangeDeclare(
		GetGatewayCommandTopicX(),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

func DeclareGatewayCommandBroadcastX(
	ch *amqp.Channel,
) error {
	return ch.ExchangeDeclare(
		GetGatewayCommandBroadcastX(),
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
}

// ========================
// Responses
// ========================

func DeclareGatewayResponseTopicX(
	ch *amqp.Channel,
) error {
	return ch.ExchangeDeclare(
		GetGatewayResponseTopicX(),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

// ========================
// Dead Letter
// ========================

func DeclareDLX(
	ch *amqp.Channel,
) error {
	return ch.ExchangeDeclare(
		XIonbusDlx,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
}

func DeclareAndBindDLQ(
	ch *amqp.Channel,
) error {
	_, err := ch.QueueDeclare(
		QIonbusDlq,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("unable to declare %v: %w", QIonbusDlq, err)
	}

	// Bind the DLQ to the DLX using the routing key
	err = ch.QueueBind(QIonbusDlq, "#", XIonbusDlx, false, nil)
	if err != nil {
		return fmt.Errorf(
			"unable to bind queue %v with key %v to %v: %w",
			QIonbusDlq, "#", XIonbusDlx, err,
		)
	}

	return nil
}
