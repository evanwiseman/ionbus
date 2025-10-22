package config

import (
	"fmt"
	"os"
	"strconv"
)

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
	dbSchema := os.Getenv("DB_SCHEMA")
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
		Schema:   dbSchema,
		Host:     dbHost,
		Port:     dbPort,
		Username: dbUsername,
		Password: dbPassword,
		Name:     dbName,
		SSLMode:  dbSSLMode,
	}, nil
}
