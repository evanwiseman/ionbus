package server

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/models"
)

type ServerConfig struct {
	ID      string
	Device  models.Device
	Version string
	RMQ     config.RMQConfig
	DB      config.DBConfig
}

func LoadServerConfig() (*ServerConfig, error) {
	id := os.Getenv("SERVER_ID")
	rmqConfig, err := config.LoadRMQConfig()
	if err != nil {
		return nil, err
	}

	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		ID:      id,
		Device:  models.DeviceServer,
		Version: "1.0",
		RMQ:     rmqConfig,
		DB:      dbConfig,
	}, nil
}
