package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/evanwiseman/ionbus/internal/pubsub"
)

type Client struct {
	Ctx  context.Context
	Cfg  ClientConfig
	MQTT *ClientMQTT
}

type ClientMQTT struct {
	Client             mqtt.Client
	RequestPublisher   *pubsub.MQTTPublisher
	RequestSubscriber  *pubsub.MQTTSubscriber
	ResponsePublisher  *pubsub.MQTTPublisher
	ResponseSubscriber *pubsub.MQTTSubscriber
}

func (c *ClientMQTT) Close() {
	c.Client.Disconnect(250)
}

func NewClient(ctx context.Context, cfg *ClientConfig) (*Client, error) {
	client := &Client{
		Ctx:  ctx,
		Cfg:  *cfg,
		MQTT: &ClientMQTT{},
	}
	// MQTT
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.GetUrl()).
		SetKeepAlive(cfg.MQTT.KeepAlive).
		SetCleanSession(cfg.MQTT.CleanSession).
		SetClientID(cfg.ID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("ERROR: MQTT connection lost: %v", err)
		}).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("INFO: Client MQTT connected, setting up subscriptions...")
			if err := client.setupRequests(); err != nil {
				log.Printf("ERROR: Failed to setup MQTT requests: %v", err)
			}
			if err := client.setupResponses(); err != nil {
				log.Printf("ERROR: Failed to setup MQTT responses: %v", err)
			}
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			log.Printf("WARN: MQTT reconnecting...")
		}).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(10 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	client.MQTT.Client = mqttClient
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to establish connection to mqtt: %w", token.Error())
	}

	return client, nil
}

func (c *Client) Start() error {
	// Get all gateways
	c.RequestIdentifiers(models.DeviceGateway, "+")
	return nil
}

func (c *Client) Close() {
	c.MQTT.Close()
}

func (c *Client) setupRequests() error {
	requestPublisher := pubsub.NewMQTTPublisher(c.Ctx, c.MQTT.Client)
	c.MQTT.RequestPublisher = requestPublisher

	requestSubscriber := pubsub.NewMQTTSubscriber(c.Ctx, c.MQTT.Client)
	if err := requestSubscriber.Subscribe(
		pubsub.MQTTSubOpts{
			Topic: pubsub.MQTTTopic(c.Cfg.Device, c.Cfg.ID, models.ActionRequest, "#"),
			QoS:   byte(1),
		},
		c.HandlerRequests,
	); err != nil {
		return err
	}

	if err := requestSubscriber.Subscribe(
		pubsub.MQTTSubOpts{
			Topic: pubsub.MQTTBroadcast(c.Cfg.Device, models.ActionRequest, "#"),
			QoS:   byte(1),
		},
		c.HandlerRequests,
	); err != nil {
		return err
	}
	c.MQTT.RequestSubscriber = requestSubscriber

	return nil
}

func (c *Client) setupResponses() error {
	responsePublisher := pubsub.NewMQTTPublisher(c.Ctx, c.MQTT.Client)
	c.MQTT.ResponsePublisher = responsePublisher

	responseSubscriber := pubsub.NewMQTTSubscriber(c.Ctx, c.MQTT.Client)
	if err := responseSubscriber.Subscribe(
		pubsub.MQTTSubOpts{
			Topic: pubsub.MQTTTopic(c.Cfg.Device, c.Cfg.ID, models.ActionResponse, "#"),
			QoS:   byte(1),
		},
		c.HandlerResponses,
	); err != nil {
		return err
	}
	c.MQTT.ResponseSubscriber = responseSubscriber

	return nil
}
