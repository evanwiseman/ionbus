package pubsub

import "github.com/evanwiseman/ionbus/internal/models"

type Publisher interface {
	Publish(topic string, payload []byte, contentType models.ContentType) error
}

type Subscriber interface {
	Subscribe(topic string, handler func([]byte) error) error
}
