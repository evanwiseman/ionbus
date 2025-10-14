package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	ClientID      string
	Username      string
	Password      string
}

func LoadMQTTConfig(role string) MQTTConfig {
	mqttPort := getEnvInt("MQTT_PORT", 1883)
	mqttWS := getEnvInt("MQTT_WS_PORT", 9001)
	mqttUrl := getEnvString("MQTT_URL", fmt.Sprintf("mqtt://localhost:%d", mqttPort))
	mqttQoS := byte(getEnvInt("MQTT_QOS", 1))
	mqttKeepAlive := time.Duration(getEnvInt("MQTT_KEEPALIVE", 30)) * time.Second
	mqttCleanSession := getEnvBool("MQTT_CLEAN_SESSION", false)
	mqttClientID := fmt.Sprintf(
		"%s-%s-%s",
		getEnvString("MQTT_CLIENT_ID", "ionbus-node"),
		role,
		uuid.New().String(),
	)
	mqttUsername := getEnvString("MQTT_USERNAME", "")
	mqttPassword := getEnvString("MQTT_PASSWORD", "")

	return MQTTConfig{
		Port:          mqttPort,
		WebSocketPort: mqttWS,
		Url:           mqttUrl,
		QoS:           mqttQoS,
		KeepAlive:     mqttKeepAlive,
		CleanSession:  mqttCleanSession,
		ClientID:      mqttClientID,
		Username:      mqttUsername,
		Password:      mqttPassword,
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
	rabbitUrl := getEnvString("RABBIT_URL", fmt.Sprintf("amqp://guest:guest@localhost:%d/", rabbitPort))

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
	// ==== Postgres ====
	pgUser := getEnvString("POSTGRES_USER", "ionbus")
	pgPassword := getEnvString("POSTGRES_PASSWORD", "ionbus")
	pgPort := getEnvInt("POSTGRES_PORT", 5432)
	pgUrl := getEnvString(
		"POSTGRES_URL",
		fmt.Sprintf(
			"postgres://%v:%v@localhost:%v/ionbus?sslmode=disable",
			pgUser,
			pgPassword,
			pgPort,
		),
	)
	return PostgresConfig{
		User:     pgUser,
		Password: pgPassword,
		Port:     pgPort,
		Url:      pgUrl,
	}
}

// ========================
// Log Config
// ========================
type LogConfig struct {
	Level string
}

func LoadLogConfig() LogConfig {
	return LogConfig{
		Level: getEnvString("LOG_LEVEL", "debug"),
	}
}
