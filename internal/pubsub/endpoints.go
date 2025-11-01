package pubsub

import (
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
)

const (
	BroadcastPrefix = "broadcast"
	TopicPrefix     = "topic"

	RequestPrefix  = "request"
	ResponsePrefix = "response"

	ClientPrefix  = "client"
	GatewayPrefix = "gateway"
	ServerPrefix  = "server"

	XIonbusDlx = "ionbus.dlx"
	QIonbusDlq = "ionbus.dlq"
)

func MQTTTopic(device models.Device, id string, action string, method string) string {
	return fmt.Sprintf("%s/%s/%s/%s", device, id, action, method)
}

func MQTTBroadcast(device models.Device, action string, method string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, device, action, method)
}

func RMQTopicX(device models.Device, action string) string {
	return fmt.Sprintf("%s.%s.%s", device, action, TopicPrefix)
}

func RMQBroadcastX(device models.Device, action string) string {
	return fmt.Sprintf("%s.%s.%s", device, action, BroadcastPrefix)
}

func RMQTopicRK(device models.Device, id string, action string, method string) string {
	return fmt.Sprintf("%s.%s.%s.%s", device, id, action, method)
}

func RMQBroadcastRK(device models.Device, action string, method string) string {
	return fmt.Sprintf("%s.%s.%s.%s", BroadcastPrefix, device, action, method)
}

func RMQQueue(device models.Device, id string, action string) string {
	return fmt.Sprintf("%s.%s.%s", device, id, action)
}
