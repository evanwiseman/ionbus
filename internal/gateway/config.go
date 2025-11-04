package gateway

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/models"
)

type GatewayConfig struct {
	ID      string
	Device  models.Device
	Version string
	MQTT    config.MQTTConfig
	RMQ     config.RMQConfig
}

func LoadGatewayConfig() (*GatewayConfig, error) {
	id := os.Getenv("GATEWAY_ID")
	mqttConfig, err := config.LoadMQTTConfig()
	if err != nil {
		return nil, err
	}

	rmqConfig, err := config.LoadRMQConfig()
	if err != nil {
		return nil, err
	}

	cfg := GatewayConfig{
		ID:      id,
		Device:  models.DeviceGateway,
		Version: "1.0",
		MQTT:    mqttConfig,
		RMQ:     rmqConfig,
	}

	return &cfg, nil
}
