package main

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

type Client struct {
	Ctx        context.Context
	Cfg        ClientConfig
	MQTTClient mqtt.Client
}

func NewClient(ctx context.Context, cfg *ClientConfig) (*Client, error) {
	// MQTT
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.GetUrl()).
		SetKeepAlive(cfg.MQTT.KeepAlive).
		SetCleanSession(cfg.MQTT.CleanSession).
		SetClientID(cfg.ID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to establish connection to mqtt: %w", token.Error())
	}

	client := &Client{
		Ctx:        ctx,
		Cfg:        *cfg,
		MQTTClient: mqttClient,
	}

	// Setup Infrastructure
	if err := client.setupMQTT(); err != nil {
		return nil, fmt.Errorf("failed to setup mqtt: %w", err)
	}

	return client, nil
}

func (c *Client) Start() error {
	// TODO: Start client logic
	return nil
}

func (c *Client) Close() {
	if c.MQTTClient != nil {
		c.MQTTClient.Disconnect(250)
	}
}

func (c *Client) setupMQTT() error {
	if err := c.setupCommands(); err != nil {
		return fmt.Errorf("failed to setup commands: %w", err)
	}

	return nil
}

func (c *Client) setupCommands() error {
	topic := pubsub.GetMQTTClientCommandTopic(c.Cfg.ID, "#")
	qos := byte(1)

	err := pubsub.SubscribeMQTT(
		c.Ctx,
		c.MQTTClient,
		pubsub.MQTTSubOpts{
			Topic: topic,
			QoS:   qos,
		},
		models.ContentJSON,
		c.HandlerClientCommands,
	)
	if err != nil {
		return err
	}

	broadcast := pubsub.GetMQTTClientCommandBroadcast("#")
	err = pubsub.SubscribeMQTT(
		c.Ctx,
		c.MQTTClient,
		pubsub.MQTTSubOpts{
			Topic: broadcast,
			QoS:   qos,
		},
		models.ContentJSON,
		c.HandlerClientCommands,
	)
	if err != nil {
		return err
	}

	return nil
}
