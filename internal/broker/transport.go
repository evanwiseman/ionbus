package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Transport interface {
	SendRequest(req models.Request) error
	SendResposne(res models.Response) error
}

type BaseTransport struct {
	DeviceConfig *config.DeviceConfig
}

type MQTTTransport struct {
	BaseTransport
	RequestPublisher   *MQTTPublisher
	RequestSubscriber  *MQTTSubscriber
	ResponsePublisher  *MQTTPublisher
	ResponseSubscriber *MQTTSubscriber
}

func NewMQTTTransport(ctx context.Context, cfg *config.DeviceConfig, client mqtt.Client) (*MQTTTransport, error) {
	requestPublisher := NewMQTTPublisher(ctx, client)
	requestSubscriber := NewMQTTSubscriber(ctx, client)
	// (Sub) Request Topic
	if err := requestSubscriber.Subscribe(
		MQTTSubOpts{
			Topic: MQTTTopic(cfg.DeviceType, cfg.DeviceID, models.CommandRequest, "#"),
			QoS:   byte(1),
		},
		func([]byte) error { return nil }, // TODO replace with handler
	); err != nil {
		return nil, err
	}
	// (Sub) Request Broadcast
	if err := requestSubscriber.Subscribe(
		MQTTSubOpts{
			Topic: MQTTBroadcast(cfg.DeviceType, models.CommandRequest, "#"),
			QoS:   byte(1),
		},
		func([]byte) error {
			return nil
		}, // TODO replace with handler
	); err != nil {
		return nil, err
	}

	responsePublisher := NewMQTTPublisher(ctx, client)
	responseSubscriber := NewMQTTSubscriber(ctx, client)
	if err := responseSubscriber.Subscribe(
		MQTTSubOpts{
			Topic: MQTTTopic(cfg.DeviceType, cfg.DeviceID, models.CommandResponse, "#"),
			QoS:   byte(1),
		},
		func([]byte) error { return nil },
	); err != nil {
		return nil, err
	}

	return &MQTTTransport{
		BaseTransport: BaseTransport{
			DeviceConfig: cfg,
		},
		RequestPublisher:   requestPublisher,
		RequestSubscriber:  requestSubscriber,
		ResponsePublisher:  responsePublisher,
		ResponseSubscriber: responseSubscriber,
	}, nil
}

func (t *MQTTTransport) SendRequest(req models.Request) error {
	var topic string
	if req.TargetID == "+" || req.TargetID == "" {
		topic = MQTTBroadcast(req.TargetDevice, models.CommandRequest, req.Method)
	} else {
		topic = MQTTTopic(req.TargetDevice, req.TargetID, models.CommandRequest, req.Method)
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error: failed to marshal request: %w", err)
	}

	msg := models.Message{
		SourceID:     t.DeviceConfig.DeviceID,
		SourceDevice: t.DeviceConfig.DeviceType,
		Version:      t.DeviceConfig.Version,
		Payload:      payload,
	}

	return t.RequestPublisher.Publish(
		MQTTPubOpts{
			Topic:    topic,
			QoS:      byte(1),
			Retained: false,
		},
		msg,
	)
}

func (t *MQTTTransport) SendResponse(res models.Response) error {
	// Marhsal the response
	payload, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Wrap response in message envelope
	msg := models.Message{
		SourceID:     t.DeviceConfig.DeviceID,
		SourceDevice: t.DeviceConfig.DeviceType,
		Version:      t.DeviceConfig.Version,
		Payload:      payload,
	}

	// Determine routing
	topic := MQTTTopic(res.TargetDevice, res.TargetID, models.CommandResponse, res.Method)
	return t.ResponsePublisher.Publish(
		MQTTPubOpts{
			Topic:    topic,
			QoS:      byte(1),
			Retained: false,
		},
		msg,
	)
}

type RMQTransport struct {
	BaseTransport
	RequestPublisher   *RMQPublisher
	RequestSubscriber  *RMQSubscriber
	ResponsePublisher  *RMQPublisher
	ResponseSubscriber *RMQSubscriber
}

func NewRMQTransport(ctx context.Context, cfg *config.DeviceConfig, conn *amqp.Connection) (*RMQTransport, error) {
	// ========================
	// Request Publisher
	// ========================
	reqPubCh, err := OpenChannel(conn)
	if err != nil {
		return nil, err
	}
	// Declare Exchanges
	reqTopicX := RMQTopicX(cfg.DeviceType, models.CommandRequest)
	if err := reqPubCh.ExchangeDeclare(
		reqTopicX,
		string(Topic),
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare %s topic exchange: %w", reqTopicX, err)
	}

	reqBroadcastX := RMQBroadcastX(cfg.DeviceType, models.CommandRequest)
	if err := reqPubCh.ExchangeDeclare(
		reqBroadcastX,
		string(Fanout),
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare %s broadcast exchange: %w", reqBroadcastX, err)
	}
	requestPublisher := NewRMQPublisher(ctx, reqPubCh)

	// ========================
	// Request Subscriber
	// ========================
	reqSubCh, err := OpenChannel(conn)
	if err != nil {
		return nil, err
	}

	// Queue Parameters
	reqName := RMQQueue(cfg.DeviceType, cfg.DeviceID, models.CommandRequest)
	reqOpts := QueueOpts{
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     false,
		Args: amqp.Table{
			"x-dead-letter-exchange": XIonbusDlx,
		},
	}

	// Declare the command queue
	_, err = DeclareQueue(reqSubCh, reqName, reqOpts)
	if err != nil {
		return nil, err
	}

	// Bind queue to gateway topic exchange
	topicRK := RMQTopicRK(cfg.DeviceType, cfg.DeviceID, models.CommandRequest, "#'")
	if err := BindQueue(
		reqSubCh,
		reqName,
		topicRK,
		reqTopicX,
	); err != nil {
		return nil, err
	}

	// Bind queue to broadcast exchange
	broadcastRK := RMQBroadcastRK(cfg.DeviceType, models.CommandRequest, "#")
	if err := BindQueue(
		reqSubCh,
		reqName,
		broadcastRK,
		reqBroadcastX,
	); err != nil {
		return nil, err
	}
	requestSubscriber := NewRMQSubscriber(ctx, reqSubCh)
	requestSubscriber.Subscribe(
		RMQSubOpts{QueueName: reqName},
		func([]byte) error {
			log.Print("received")
			return nil
		},
	)

	// ========================
	// Response Publisher
	// ========================
	resPubCh, err := OpenChannel(conn)
	if err != nil {
		return nil, err
	}
	// Declare the exchange
	resTopicX := RMQTopicX(cfg.DeviceType, models.CommandResponse)
	if err := resPubCh.ExchangeDeclare(
		resTopicX,
		string(Topic),
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare %s topic exchange: %w", resTopicX, err)
	}
	responsePublisher := NewRMQPublisher(ctx, resPubCh)

	// ========================
	// Response Subscriber
	// ========================
	resSubCh, err := OpenChannel(conn)
	if err != nil {
		return nil, err
	}
	// Queue parameters
	resName := RMQQueue(cfg.DeviceType, cfg.DeviceID, models.CommandResponse)
	resOpts := QueueOpts{
		Durable:    false,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     false,
		Args: amqp.Table{
			"x-dead-letter-exchange": XIonbusDlx,
		},
	}

	// Declare the Response Queue
	_, err = DeclareQueue(resSubCh, resName, resOpts)
	if err != nil {
		return nil, err
	}

	resTopicRK := RMQTopicRK(cfg.DeviceType, cfg.DeviceID, models.CommandResponse, "#")
	if err := BindQueue(
		resSubCh,
		resName,
		resTopicRK,
		resTopicX,
	); err != nil {
		return nil, err
	}
	responseSubscriber := NewRMQSubscriber(ctx, resSubCh)
	responseSubscriber.Subscribe(
		RMQSubOpts{QueueName: resName},
		func([]byte) error { return nil },
	)

	return &RMQTransport{
		BaseTransport: BaseTransport{
			DeviceConfig: cfg,
		},
		RequestPublisher:   requestPublisher,
		RequestSubscriber:  requestSubscriber,
		ResponsePublisher:  responsePublisher,
		ResponseSubscriber: responseSubscriber,
	}, nil
}

func (t *RMQTransport) SendRequest(req models.Request) error {
	// Marhsal the response
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error: failed to marshal request: %w", err)
	}

	var exchange string
	var key string

	// Broadcast or targeted request
	if req.TargetID == "*" || req.TargetID == "" {
		exchange = RMQBroadcastX(req.TargetDevice, models.CommandRequest)
		key = RMQBroadcastRK(req.TargetDevice, models.CommandRequest, req.Method)
	} else {
		exchange = RMQTopicX(req.TargetDevice, models.CommandRequest)
		key = RMQTopicRK(req.TargetDevice, req.TargetID, models.CommandRequest, req.Method)
	}

	msg := models.Message{
		SourceID:     t.DeviceConfig.DeviceID,
		SourceDevice: t.DeviceConfig.DeviceType,
		Version:      t.DeviceConfig.Version,
		Payload:      payload,
	}

	return t.RequestPublisher.Publish(
		RMQPubOpts{
			Exchange: exchange,
			Key:      key,
		},
		msg,
	)
}

func (t *RMQTransport) SendResponse(res models.Response) error {
	// Marhsal the response
	payload, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Wrap response in message envelope
	msg := models.Message{
		SourceID:     t.DeviceConfig.DeviceID,
		SourceDevice: t.DeviceConfig.DeviceType,
		Version:      t.DeviceConfig.Version,
		Payload:      payload,
	}

	// Determine routing
	exchange := RMQTopicX(res.TargetDevice, models.CommandResponse)
	key := RMQTopicRK(res.TargetDevice, res.TargetID, models.CommandResponse, res.Method)
	return t.ResponsePublisher.Publish(
		RMQPubOpts{
			Exchange: exchange,
			Key:      key,
		},
		msg,
	)
}
