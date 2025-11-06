package config

type GatewayConfig struct {
	Device *DeviceConfig
	MQTT   *MQTTConfig
	RMQ    *RMQConfig
	SQLite *SQLiteConfig
}

func LoadGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		Device: LoadDeviceConfig(),
		MQTT:   LoadMQTTConfig(),
		RMQ:    LoadRMQConfig(),
		SQLite: LoadSQLiteConfig(),
	}
}
