package bridge

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Bridge struct {
	rmqCh      *amqp.Channel
	mqttClient mqtt.Client
}

func (b *Bridge) RMQToMQTT() error {
	return nil
}

func (b *Bridge) MQTTToRMQ() error {
	return nil
}
