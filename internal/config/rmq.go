package config

import (
	"fmt"
	"os"
	"strconv"
)

type RMQConfig struct {
	Schema   string
	Host     string
	Port     int
	Username string
	Password string
	VHost    string
	TLS      bool
}

func (rmq RMQConfig) GetUrl() string {
	return fmt.Sprintf("%v://%v:%v", rmq.Schema, rmq.Host, rmq.Port)
}

func LoadRMQConfig() (RMQConfig, error) {
	rmqSchema := os.Getenv("RMQ_SCHEMA")
	rmqHost := os.Getenv("RMQ_HOST")
	rmqPort, err := strconv.Atoi(os.Getenv("RMQ_PORT"))
	if err != nil {
		return RMQConfig{}, err
	}
	rmqUsername := os.Getenv("RMQ_USERNAME")
	rmqPassword := os.Getenv("RMQ_PASSWORD")
	rmqVHost := os.Getenv("RMQ_VHOST")
	rmqTLS, err := strconv.ParseBool(os.Getenv("RMQ_TLS"))
	if err != nil {
		return RMQConfig{}, err
	}

	return RMQConfig{
		Schema:   rmqSchema,
		Host:     rmqHost,
		Port:     rmqPort,
		Username: rmqUsername,
		Password: rmqPassword,
		VHost:    rmqVHost,
		TLS:      rmqTLS,
	}, nil
}
