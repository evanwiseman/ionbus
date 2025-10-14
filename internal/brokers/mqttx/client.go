package mqttx

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/config"
)

func NewClient(mc config.MQTTConfig) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mc.Url)
	opts.SetKeepAlive(mc.KeepAlive)
	opts.SetCleanSession(mc.CleanSession)
	opts.SetClientID(mc.ClientID)
	opts.SetUsername(mc.Username)
	opts.SetPassword(mc.Password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}
