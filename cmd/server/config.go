package main

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/config"
)

type ServerConfig struct {
	ID  string
	RMQ config.RMQConfig
	DB  config.DBConfig
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
		ID:  id,
		RMQ: rmqConfig,
		DB:  dbConfig,
	}, nil
}
