package config

import (
	"fmt"
	"os"
	"strconv"
)

type RMQConfig struct {
	Schema   string
	Host     string
	Port     string
	Username string
	Password string
	VHost    string
	TLS      bool
}

func (c RMQConfig) GetUrl() string {
	return fmt.Sprintf("%v://%v:%v", c.Schema, c.Host, c.Port)
}

func LoadRMQConfig() *RMQConfig {
	tls, err := strconv.ParseBool(os.Getenv("RMQ_TLS"))
	if err != nil {
		tls = false
	}

	return &RMQConfig{
		Schema:   os.Getenv("RMQ_SCHEMA"),
		Host:     os.Getenv("RMQ_HOST"),
		Port:     os.Getenv("RMQ_PORT"),
		Username: os.Getenv("RMQ_USERNAME"),
		Password: os.Getenv("RMQ_PASSWORD"),
		VHost:    os.Getenv("RMQ_VHOST"),
		TLS:      tls,
	}
}
