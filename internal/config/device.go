package config

import (
	"os"

	"github.com/evanwiseman/ionbus/internal/models"
)

type DeviceConfig struct {
	DeviceID   string
	DeviceType models.DeviceType
	Version    string
}

func LoadDeviceConfig() *DeviceConfig {
	return &DeviceConfig{
		DeviceID:   os.Getenv("DEVICE_ID"),
		DeviceType: models.DeviceType(os.Getenv("DEVICE_TYPE")),
		Version:    os.Getenv("VERSION"),
	}
}
