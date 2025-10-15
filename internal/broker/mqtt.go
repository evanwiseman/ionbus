package broker

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/config"
)

func StartMQTT(mqttConfig config.MQTTConfig, clientID string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttConfig.GetUrl())
	opts.SetKeepAlive(mqttConfig.KeepAlive)
	opts.SetCleanSession(mqttConfig.CleanSession)
	opts.SetClientID(clientID)
	opts.SetUsername(mqttConfig.Username)
	opts.SetPassword(mqttConfig.Password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}
