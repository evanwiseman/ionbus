package main

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

type ClientMQTT struct {
	Client       mqtt.Client
	RequestFlow  *pubsub.MQTTFlow
	ResponseFlow *pubsub.MQTTFlow
}

type Client struct {
	Ctx  context.Context
	Cfg  ClientConfig
	MQTT ClientMQTT
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
		Ctx: ctx,
		Cfg: *cfg,
		MQTT: ClientMQTT{
			Client: mqttClient,
		},
	}

	// Setup Infrastructure
	if err := client.setupMQTT(); err != nil {
		return nil, fmt.Errorf("failed to setup mqtt: %w", err)
	}

	return client, nil
}

func (c *Client) Start() error {
	// c.RequestGatewayIdentifiers("", nil, "device initialization")
	return nil
}

func (c *Client) Close() {
	if c.MQTT.Client != nil {
		c.MQTT.Client.Disconnect(250)
	}
}

func (c *Client) setupMQTT() error {
	if err := c.setupRequests(); err != nil {
		return fmt.Errorf("failed to setup commands: %w", err)
	}

	if err := c.setupResponses(); err != nil {
		return fmt.Errorf("failed to setup response: %w", err)
	}

	return nil
}

func (c *Client) setupRequests() error {
	topic := pubsub.MClientReqT(c.Cfg.ID, "#")
	qos := byte(1)

	requestFlow := pubsub.NewMQTTFlow(c.Ctx, c.MQTT.Client)
	c.MQTT.RequestFlow = requestFlow
	if err := requestFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: topic,
		QoS:   qos,
	}); err != nil {
		return err
	}

	broadcast := pubsub.MClientReqB("#")
	if err := requestFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: broadcast,
		QoS:   qos,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) setupResponses() error {
	topic := pubsub.MClientResT(c.Cfg.ID, "#")
	qos := byte(1)

	responseFlow := pubsub.NewMQTTFlow(c.Ctx, c.MQTT.Client)
	c.MQTT.ResponseFlow = responseFlow
	if err := responseFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: topic,
		QoS:   qos,
	}); err != nil {
		return err
	}

	broadcast := pubsub.MClientReqB("#")
	if err := responseFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: broadcast,
		QoS:   qos,
	}); err != nil {
		return err
	}

	return nil
}
