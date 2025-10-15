package gateway

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/config"
)

func StartMQTT() (mqtt.Client, error) {
	config := config.LoadMQTTConfig("node")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Url)
	opts.SetKeepAlive(config.KeepAlive)
	opts.SetCleanSession(config.CleanSession)
	opts.SetClientID(config.ClientID)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}
