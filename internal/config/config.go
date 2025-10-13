package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ========================
// Helpers
// ========================
func getEnvString(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

// ========================
// MQTT Config
// ========================
type MQTTConfig struct {
	Port          int
	WebSocketPort int
	Url           string
	QoS           byte
	KeepAlive     time.Duration
	CleanSession  bool
}

func LoadMQTTConfig() MQTTConfig {
	mqttPort := getEnvInt("MQTT_PORT", 1883)
	mqttWS := getEnvInt("MQTT_WS_PORT", 9001)
	mqttUrl := getEnvString("MQTT_URL", fmt.Sprintf("mqtt://mqtt-broker:%d", mqttPort))
	mqttQoS := byte(getEnvInt("MQTT_QOS", 1))
	mqttKeepAlive := time.Duration(getEnvInt("MQTT_KEEPALIVE", 30)) * time.Second
	mqttCleanSession := getEnvBool("MQTT_CLEAN_SESSION", false)
	return MQTTConfig{
		Port:          mqttPort,
		WebSocketPort: mqttWS,
		Url:           mqttUrl,
		QoS:           mqttQoS,
		KeepAlive:     mqttKeepAlive,
		CleanSession:  mqttCleanSession,
	}
}

// ========================
// RabbitMQ Config
// ========================
type RabbitConfig struct {
	Port int
	Url  string
}

func LoadRabbitConfig() RabbitConfig {
	rabbitPort := getEnvInt("RABBIT_PORT", 5672)
	rabbitUrl := getEnvString("RABBIT_URL", fmt.Sprintf("amqp://guest:guest@rabbitmq:%d/", rabbitPort))
	return RabbitConfig{
		Port: rabbitPort,
		Url:  rabbitUrl,
	}
}

// ========================
// Postgres Config
// ========================
type PostgresConfig struct {
	User     string
	Password string
	Port     int
	Url      string
}

func LoadPostgresConfig() PostgresConfig {
	postgresUser := getEnvString("POSTGRES_USER", "ionbus")
	postgresPassword := getEnvString("POSTGRES_PASSWORD", "ionbus")
	postgresPort := getEnvInt("POSTGRES_PORT", 5432)
	postgresUrl := getEnvString(
		"POSTGRES_URL",
		fmt.Sprintf(
			"postgres://%v:%v@postgres:%v/ionbus?sslmode=disable",
			postgresUser,
			postgresPassword,
			postgresPort,
		),
	)
	return PostgresConfig{
		User:     postgresUser,
		Password: postgresPassword,
		Port:     postgresPort,
		Url:      postgresUrl,
	}
}

// ========================
// Log Config
// ========================
type LogConfig struct {
	Level string
}

func LoadLogConfig() LogConfig {
	logLevel := getEnvString("LOG_LEVEL", "debug")
	return LogConfig{
		Level: logLevel,
	}
}

// ========================
// Server Config
// ========================
type ServerConfig struct {
	MQTT     MQTTConfig
	Rabbit   RabbitConfig
	Postgres PostgresConfig
	Log      LogConfig
}

func LoadServerConfig() ServerConfig {
	mqttConfig := LoadMQTTConfig()
	rabbitConfig := LoadRabbitConfig()
	postgresConfig := LoadPostgresConfig()
	logConfig := LoadLogConfig()

	return ServerConfig{
		MQTT:     mqttConfig,
		Rabbit:   rabbitConfig,
		Postgres: postgresConfig,
		Log:      logConfig,
	}
}
