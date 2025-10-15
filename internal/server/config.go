package server

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/config"
)

type ServerConfig struct {
	ID  string
	RMQ config.RMQConfig
	DB  config.DBConfig
}

func LoadServerConfig() (ServerConfig, error) {
	id := os.Getenv("SERVER_ID")
	rmqConfig, err := config.LoadRMQConfig()
	if err != nil {
		return ServerConfig{}, err
	}

	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		return ServerConfig{}, err
	}

	return ServerConfig{
		ID:  id,
		RMQ: rmqConfig,
		DB:  dbConfig,
	}, nil
}
