package main

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
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
	c.RequestGatewayIdentifiers("+")
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

	c.MQTT.RequestFlow = pubsub.NewMQTTFlow(c.Ctx, c.MQTT.Client)

	if err := c.MQTT.RequestFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: pubsub.MClientReqT(c.Cfg.ID, "#"),
		QoS:   byte(1),
	}); err != nil {
		return err
	}

	if err := c.MQTT.RequestFlow.Sub.Subscribe(
		pubsub.MQTTSubOpts{
			Topic: pubsub.MClientReqB("#"),
			QoS:   byte(1),
		},
	); err != nil {
		return err
	}

	c.MQTT.RequestFlow.Sub.Mux.HandleFunc(
		pubsub.MClientReqT(c.Cfg.ID, string(models.ActionGetIdentifiers)),
		c.HandleIdentifierRequest,
	)

	c.MQTT.RequestFlow.Sub.Mux.HandleFunc(
		pubsub.MClientReqB(string(models.ActionGetIdentifiers)),
		c.HandleIdentifierRequest,
	)

	return nil
}

func (c *Client) setupResponses() error {
	c.MQTT.ResponseFlow = pubsub.NewMQTTFlow(c.Ctx, c.MQTT.Client)
	if err := c.MQTT.ResponseFlow.Sub.Subscribe(pubsub.MQTTSubOpts{
		Topic: pubsub.MClientResT(c.Cfg.ID, "#"),
		QoS:   byte(1),
	}); err != nil {
		return err
	}

	c.MQTT.ResponseFlow.Sub.Mux.HandleFunc(
		pubsub.MClientResT(c.Cfg.ID, string(models.ActionGetIdentifiers)),
		c.HandleIdentifierResponse,
	)

	return nil
}
