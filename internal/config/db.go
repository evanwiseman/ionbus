package config

import (
	"fmt"
	"os"
)

type PostgresConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode  string
}

func (c PostgresConfig) GetUrl() string {
	return fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v@sslmode=%v",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

func LoadPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USERNAME"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DATABASE"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
	}
}

type SQLiteConfig struct {
	Path string
}

func LoadSQLiteConfig() *SQLiteConfig {
	return &SQLiteConfig{
		Path: os.Getenv("SQLITE_PATH"),
	}
}
