package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RMQPublishOptions struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
}

type RMQSubscribeOptions struct {
	QueueName     string
	Consumer      string
	AutoAck       bool
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
	PrefetchCount int
	PrefetchSize  int
	QosGlobal     bool
	Args          amqp.Table
}

type MQTTPublishOptions struct {
	Topic    string
	QoS      byte
	Retained bool
}

type MQTTSubscribeOptions struct {
	Topic string
	QoS   byte
}
