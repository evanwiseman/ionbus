package node

import (
	"github.com/evanwiseman/ionbus/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func StartRMQ() (*amqp.Connection, error) {
	rc := config.LoadRabbitConfig()
	conn, err := amqp.Dial(rc.Url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
