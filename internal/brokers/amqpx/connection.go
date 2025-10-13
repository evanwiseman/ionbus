package amqpx

import (
	"github.com/evanwiseman/ionbus/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewConnection(rc config.RabbitConfig) (*amqp.Connection, error) {
	conn, err := amqp.Dial(rc.Url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
