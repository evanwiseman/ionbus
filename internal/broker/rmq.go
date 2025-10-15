package broker

import (
	"github.com/evanwiseman/ionbus/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func StartRMQ(rmqConfig config.RMQConfig) (*amqp.Connection, error) {
	conn, err := amqp.Dial(rmqConfig.GetUrl())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
