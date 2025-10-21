package util

import (
	"github.com/evanwiseman/ionbus/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func OpenChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func ConnectRMQ(cfg config.RMQConfig) (*amqp.Connection, error) {
	conn, err := amqp.Dial(cfg.GetUrl())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
