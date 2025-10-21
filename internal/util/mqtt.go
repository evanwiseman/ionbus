package util

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/config"
)

// Connects to the MQTT broker using the Gateway Config's Parameters
func ConnectMQTT(cfg config.MQTTConfig, clientID string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.GetUrl()).
		SetKeepAlive(cfg.KeepAlive).
		SetCleanSession(cfg.CleanSession).
		SetClientID(clientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}
