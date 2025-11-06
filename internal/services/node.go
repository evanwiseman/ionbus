package services

import (
	"context"
	"database/sql"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/broker"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/models"
)

type Node struct {
	Cfg           *config.NodeConfig
	Ctx           context.Context
	Client        mqtt.Client
	DB            *sql.DB
	MQTTTransport *broker.MQTTTransport
}

func (n *Node) Start() error {
	// Get all gateways
	n.MQTTTransport.SendRequest(models.Request{
		Method:       "GET_IDENTIFIERS",
		TargetID:     "+",
		TargetDevice: models.DeviceGateway,
		Payload:      nil,
	})
	return nil
}

func (n *Node) Stop() error {
	n.Client.Disconnect(250)
	return nil
}

func NewNode(ctx context.Context, cfg *config.NodeConfig) (*Node, error) {
	// ========================
	// MQTT
	// ========================
	client, err := broker.StartMQTT(cfg.MQTT, cfg.Device.DeviceID)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", cfg.SQLite.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	mqttTransport, err := broker.NewMQTTTransport(ctx, cfg.Device, client)
	if err != nil {
		return nil, fmt.Errorf("error: failed to setup mqtt transport: %w", err)
	}

	node := &Node{
		Ctx:           ctx,
		Cfg:           cfg,
		Client:        client,
		MQTTTransport: mqttTransport,
		DB:            db,
	}

	return node, nil
}
