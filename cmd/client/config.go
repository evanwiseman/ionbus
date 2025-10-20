package main

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/config"
)

type ClientConfig struct {
	ID   string
	MQTT config.MQTTConfig
}

func LoadClientConfig() (ClientConfig, error) {
	id := os.Getenv("CLIENT_ID")
	mqttConfig, err := config.LoadMQTTConfig()
	if err != nil {
		return ClientConfig{}, err
	}

	return ClientConfig{
		ID:   id,
		MQTT: mqttConfig,
	}, nil
}
