package config

type NodeConfig struct {
	Device *DeviceConfig
	MQTT   *MQTTConfig
	SQLite *SQLiteConfig
}

func LoadNodeConfig() *NodeConfig {
	return &NodeConfig{
		Device: LoadDeviceConfig(),
		MQTT:   LoadMQTTConfig(),
		SQLite: LoadSQLiteConfig(),
	}
}
