package broker

import (
	"fmt"

	"github.com/evanwiseman/ionbus/internal/models"
)

const (
	BroadcastPrefix = "broadcast"
	TopicPrefix     = "topic"

	XIonbusDlx = "ionbus.dlx"
	QIonbusDlq = "ionbus.dlq"
)

func MQTTTopic(device models.DeviceType, id string, command string, method string) string {
	return fmt.Sprintf("%s/%s/%s/%s", device, id, command, method)
}

func MQTTBroadcast(device models.DeviceType, command string, method string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, device, command, method)
}

func RMQTopicX(device models.DeviceType, command string) string {
	return fmt.Sprintf("%s.%s.%s", device, command, TopicPrefix)
}

func RMQBroadcastX(device models.DeviceType, command string) string {
	return fmt.Sprintf("%s.%s.%s", device, command, BroadcastPrefix)
}

func RMQTopicRK(device models.DeviceType, id string, command string, method string) string {
	return fmt.Sprintf("%s.%s.%s.%s", device, id, command, method)
}

func RMQBroadcastRK(device models.DeviceType, command string, method string) string {
	return fmt.Sprintf("%s.%s.%s.%s", BroadcastPrefix, device, command, method)
}

func RMQQueue(device models.DeviceType, id string, command string) string {
	return fmt.Sprintf("%s.%s.%s", device, id, command)
}
