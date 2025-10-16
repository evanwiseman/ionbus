package bridge

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Bridge struct {
	RMQCh      *amqp.Channel
	MQTTClient mqtt.Client
}

type RMQBridgeOptions struct {
	QueueName   string
	ContentType routing.ContentType
	MQTTTopic   string
	Qos         byte
}

type MQTTBridgeOptions struct {
	Topic       string
	ContentType routing.ContentType
	RMQExchange string
	RMQKey      string
}

type RouteType string

const (
	RouteRMQ  = "rmq"
	RouteMQTT = "mqtt"
)

type Route struct {
	Type RouteType
	Name string
}
