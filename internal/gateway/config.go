package gateway

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/config"
)

type GatewayConfig struct {
	ID   string
	MQTT config.MQTTConfig
	RMQ  config.RMQConfig
}

func LoadGatewayConfig() (GatewayConfig, error) {
	id := os.Getenv("GATEWAY_ID")
	mqttConfig, err := config.LoadMQTTConfig()
	if err != nil {
		return GatewayConfig{}, err
	}

	rmqConfig, err := config.LoadRMQConfig()
	if err != nil {
		return GatewayConfig{}, err
	}

	return GatewayConfig{
		ID:   id,
		MQTT: mqttConfig,
		RMQ:  rmqConfig,
	}, nil
}
