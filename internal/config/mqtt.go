package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type MQTTConfig struct {
	Schema        string
	Broker        string
	Port          int
	WebSocketPort int
	Username      string
	Password      string
	QoS           byte
	KeepAlive     time.Duration
	CleanSession  bool
	ClientID      string
}

func (mc MQTTConfig) GetUrl() string {
	return fmt.Sprintf("%v://%v:%v", mc.Schema, mc.Broker, mc.Port)
}

func LoadMQTTConfig() (MQTTConfig, error) {
	mqttSchema := os.Getenv("MQTT_SCHEMA")
	mqttBroker := os.Getenv("MQTT_BROKER")
	mqttPort, err := strconv.Atoi(os.Getenv("MQTT_PORT"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttWebSocketPort, err := strconv.Atoi(os.Getenv("MQTT_WS_PORT"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	mqttQoSInt, err := strconv.Atoi(os.Getenv("MQTT_QOS"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttQoS := byte(mqttQoSInt)
	mqttKeepAlive, err := time.ParseDuration(os.Getenv("MQTT_KEEPALIVE"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttCleanSession, err := strconv.ParseBool(os.Getenv("MQTT_CLEAN_SESSION"))
	if err != nil {
		return MQTTConfig{}, err
	}

	return MQTTConfig{
		Schema:        mqttSchema,
		Broker:        mqttBroker,
		Port:          mqttPort,
		WebSocketPort: mqttWebSocketPort,
		Username:      mqttUsername,
		Password:      mqttPassword,
		QoS:           mqttQoS,
		KeepAlive:     mqttKeepAlive,
		CleanSession:  mqttCleanSession,
	}, nil
}
