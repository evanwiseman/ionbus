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
	Port          string
	WebSocketPort string
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

func LoadMQTTConfig() *MQTTConfig {
	qos, err := strconv.Atoi(os.Getenv("MQTT_QOS"))
	if err != nil {
		qos = 1
	}
	keepAlive, err := time.ParseDuration(os.Getenv("MQTT_KEEPALIVE"))
	if err != nil {
		keepAlive = time.Duration(30 * time.Second)
	}
	cleanSession, err := strconv.ParseBool(os.Getenv("MQTT_CLEAN_SESSION"))
	if err != nil {
		cleanSession = false
	}

	return &MQTTConfig{
		Schema:        os.Getenv("MQTT_SCHEMA"),
		Broker:        os.Getenv("MQTT_BROKER"),
		Port:          os.Getenv("MQTT_PORT"),
		WebSocketPort: os.Getenv("MQTT_WS_PORT"),
		Username:      os.Getenv("MQTT_USERNAME"),
		Password:      os.Getenv("MQTT_PASSWORD"),
		QoS:           byte(qos),
		KeepAlive:     keepAlive,
		CleanSession:  cleanSession,
	}
}
