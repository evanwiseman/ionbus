package services

import (
	"context"
	"database/sql"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/broker"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/models"
	_ "github.com/mattn/go-sqlite3"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Gateway struct {
	Cfg           *config.GatewayConfig
	Ctx           context.Context
	Client        mqtt.Client
	Conn          *amqp.Connection
	DB            *sql.DB
	MQTTTransport *broker.MQTTTransport
	RMQTransport  *broker.RMQTransport
}

func (g *Gateway) Start() error {
	g.MQTTTransport.SendRequest(models.Request{
		Method:       "GET_IDENTIFIERS",
		TargetID:     "+",
		TargetDevice: models.DeviceNode,
		Payload:      nil,
	})
	g.RMQTransport.SendRequest(models.Request{
		Method:       "GET_IDENTIFIERS",
		TargetID:     "*",
		TargetDevice: models.DeviceServer,
		Payload:      nil,
	})
	return nil
}

func (g *Gateway) Stop() error {
	g.Client.Disconnect(250)
	if err := g.Conn.Close(); err != nil {
		return err
	}
	if err := g.DB.Close(); err != nil {
		return err
	}
	return nil
}

func NewGateway(ctx context.Context, cfg *config.GatewayConfig) (*Gateway, error) {
	client, err := broker.StartMQTT(cfg.MQTT, cfg.Device.DeviceID)
	if err != nil {
		return nil, err
	}

	conn, err := broker.StartRMQ(cfg.RMQ)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", cfg.SQLite.Path)
	if err != nil {
		return nil, fmt.Errorf("error: failed to open SQLite database: %w", err)
	}

	mqttTransport, err := broker.NewMQTTTransport(ctx, cfg.Device, client)
	if err != nil {
		return nil, fmt.Errorf("error: failed to create mqtt transport: %w", err)
	}

	rmqTransport, err := broker.NewRMQTransport(ctx, cfg.Device, conn)
	if err != nil {
		return nil, fmt.Errorf("error: faile dto create rmq transport: %w", err)
	}

	gateway := &Gateway{
		Ctx:           ctx,
		Cfg:           cfg,
		Client:        client,
		Conn:          conn,
		DB:            db,
		MQTTTransport: mqttTransport,
		RMQTransport:  rmqTransport,
	}

	return gateway, nil
}
