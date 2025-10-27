package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueOpts struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

func DeclareQueue(ch *amqp.Channel, name string, opts QueueOpts) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare %s queue: %w", name, err)
	}

	return q, nil
}

func BindQueue(ch *amqp.Channel, name, key, exchange string) error {
	if err := ch.QueueBind(
		name,
		key,
		exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind %s queue to %s key on %s exchange: %w", name, key, exchange, err)
	}

	return nil
}

func OpenChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("unable to open channel: %w", err)
	}
	return ch, nil
}

// ========================
// Commands
// ========================

func DeclareGatewayCommandTopicX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		RGatewayReqTX(),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare gateway command topic: %w", err)
	}
	return nil
}

func DeclareGatewayCommandBroadcastX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		RGatewayReqBX(),
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare gateway command broadcast: %w", err)
	}
	return nil
}

func DeclareServerCommandTopicX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		RServerReqTX(),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare server command topic: %w", err)
	}
	return nil
}

func DeclareServerCommandBroadcastX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		RServerReqBX(),
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare server command broadcast")
	}
	return nil
}

// ========================
// Responses
// ========================

func DeclareGatewayResponseTopicX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		RGatewayResTX(),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare gateway response topic: %w", err)
	}
	return nil
}

func DeclareServerResponseTopicX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		RServerResTX(),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare server response topic: %w", err)
	}
	return nil
}

// ========================
// Dead Letter
// ========================

func DeclareDLX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		XIonbusDlx,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %w", err)
	}
	return nil
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
