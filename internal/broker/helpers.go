package broker

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/evanwiseman/ionbus/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func StartMQTT(cfg *config.MQTTConfig, clientID string) (client mqtt.Client, err error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.GetUrl()).
		SetKeepAlive(cfg.KeepAlive).
		SetCleanSession(cfg.CleanSession).
		SetClientID(clientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("ERROR: MQTT connection lost: %v", err)
		}).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("INFO: MQTT connected successfully")
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			log.Printf("WARN: MQTT reconnecting...")
		}).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(10 * time.Second)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("error: unable to connect to mqtt: %w", token.Error())
	}

	return client, nil
}

func StartRMQ(cfg *config.RMQConfig) (conn *amqp.Connection, err error) {
	conn, err = amqp.Dial(cfg.GetUrl())
	if err != nil {
		return nil, fmt.Errorf("error: unable to connect to rabbitmq: %w", err)
	}

	// Setup dead lettering
	deadCh, err := OpenChannel(conn)
	if err != nil {
		return nil, err
	}
	if err := DeclareDLX(deadCh); err != nil {
		return nil, err
	}
	if err := DeclareAndBindDLQ(deadCh); err != nil {
		return nil, err
	}

	return conn, nil
}

type QueueOpts struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

func DeclareQueue(ch *amqp.Channel, name string, opts QueueOpts) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare %s queue: %w", name, err)
	}

	return q, nil
}

func BindQueue(ch *amqp.Channel, name, key, exchange string) error {
	if err := ch.QueueBind(
		name,
		key,
		exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind %s queue to %s key on %s exchange: %w", name, key, exchange, err)
	}

	return nil
}

func OpenChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("unable to open channel: %w", err)
	}
	return ch, nil
}

// ========================
// Dead Letter
// ========================

func DeclareDLX(
	ch *amqp.Channel,
) error {
	err := ch.ExchangeDeclare(
		XIonbusDlx,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %w", err)
	}
	return nil
}

func DeclareAndBindDLQ(
	ch *amqp.Channel,
) error {
	_, err := ch.QueueDeclare(
		QIonbusDlq,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("unable to declare %v: %w", QIonbusDlq, err)
	}

	// Bind the DLQ to the DLX using the routing key
	err = ch.QueueBind(QIonbusDlq, "#", XIonbusDlx, false, nil)
	if err != nil {
		return fmt.Errorf(
			"unable to bind queue %v with key %v to %v: %w",
			QIonbusDlq, "#", XIonbusDlx, err,
		)
	}

	return nil
}
