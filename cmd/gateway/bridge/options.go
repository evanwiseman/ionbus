package bridge

import "github.com/evanwiseman/ionbus/internal/models"

type RMQBridgeOptions struct {
	QueueName   string
	ContentType models.ContentType
	MQTTTopic   string
	Qos         byte
}

type MQTTBridgeOptions struct {
	Topic       string
	ContentType models.ContentType
	RMQExchange string
	RMQKey      string
}
