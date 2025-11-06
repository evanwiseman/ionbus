package config

type ServerConfig struct {
	Device *DeviceConfig
	RMQ    *RMQConfig
	PG     *PostgresConfig
}

func LoadServerConfig() *ServerConfig {
	return &ServerConfig{
		Device: LoadDeviceConfig(),
		RMQ:    LoadRMQConfig(),
		PG:     LoadPostgresConfig(),
	}
}
