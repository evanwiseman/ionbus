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
	Port          int
	WebSocketPort int
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

func LoadMQTTConfig() (MQTTConfig, error) {
	mqttSchema := os.Getenv("MQTT_SCHEMA")
	mqttBroker := os.Getenv("MQTT_BROKER")
	mqttPort, err := strconv.Atoi(os.Getenv("MQTT_PORT"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttWebSocketPort, err := strconv.Atoi(os.Getenv("MQTT_WS_PORT"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	mqttQoSInt, err := strconv.Atoi(os.Getenv("MQTT_QOS"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttQoS := byte(mqttQoSInt)
	mqttKeepAlive, err := time.ParseDuration(os.Getenv("MQTT_KEEPALIVE"))
	if err != nil {
		return MQTTConfig{}, err
	}
	mqttCleanSession, err := strconv.ParseBool(os.Getenv("MQTT_CLEAN_SESSION"))
	if err != nil {
		return MQTTConfig{}, err
	}

	return MQTTConfig{
		Schema:        mqttSchema,
		Broker:        mqttBroker,
		Port:          mqttPort,
		WebSocketPort: mqttWebSocketPort,
		Username:      mqttUsername,
		Password:      mqttPassword,
		QoS:           mqttQoS,
		KeepAlive:     mqttKeepAlive,
		CleanSession:  mqttCleanSession,
	}, nil
}

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

type DBConfig struct {
	Schema   string
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	SSLMode  string
}

func (dbc DBConfig) GetUrl() string {
	return fmt.Sprintf(
		"%v://%v:%v@%v:%v/%v@sslmode=%v",
		dbc.Schema,
		dbc.Username,
		dbc.Password,
		dbc.Host,
		dbc.Port,
		dbc.Name,
		dbc.SSLMode,
	)
}

func LoadDBConfig() (DBConfig, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return DBConfig{}, err
	}
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	return DBConfig{
		Host:     dbHost,
		Port:     dbPort,
		Username: dbUsername,
		Password: dbPassword,
		Name:     dbName,
		SSLMode:  dbSSLMode,
	}, nil
}

type LogConfig struct {
	Level string
}

type ClientConfig struct {
	ID   string
	MQTT MQTTConfig
}

func LoadClientConfig() (ClientConfig, error) {
	id := os.Getenv("CLIENT_ID")
	mqttConfig, err := LoadMQTTConfig()
	if err != nil {
		return ClientConfig{}, err
	}

	return ClientConfig{
		ID:   id,
		MQTT: mqttConfig,
	}, nil
}

type GatewayConfig struct {
	ID   string
	MQTT MQTTConfig
	RMQ  RMQConfig
}

func LoadGatewayConfig() (GatewayConfig, error) {
	id := os.Getenv("GATEWAY_ID")
	mqttConfig, err := LoadMQTTConfig()
	if err != nil {
		return GatewayConfig{}, err
	}

	rmqConfig, err := LoadRMQConfig()
	if err != nil {
		return GatewayConfig{}, err
	}

	return GatewayConfig{
		ID:   id,
		MQTT: mqttConfig,
		RMQ:  rmqConfig,
	}, nil
}

type ServerConfig struct {
	ID  string
	RMQ RMQConfig
	DB  DBConfig
}

func LoadServerConfig() (ServerConfig, error) {
	id := os.Getenv("SERVER_ID")
	rmqConfig, err := LoadRMQConfig()
	if err != nil {
		return ServerConfig{}, err
	}

	dbConfig, err := LoadDBConfig()
	if err != nil {
		return ServerConfig{}, err
	}

	return ServerConfig{
		ID:  id,
		RMQ: rmqConfig,
		DB:  dbConfig,
	}, nil
}
