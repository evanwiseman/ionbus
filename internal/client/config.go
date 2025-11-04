package client

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/models"
)

type ClientConfig struct {
	ID      string
	Device  models.Device
	Version string
	MQTT    config.MQTTConfig
}

func LoadClientConfig() (*ClientConfig, error) {
	id := os.Getenv("CLIENT_ID")
	mqttConfig, err := config.LoadMQTTConfig()
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		ID:      id,
		Device:  models.DeviceClient,
		Version: "1.0",
		MQTT:    mqttConfig,
	}, nil
}
