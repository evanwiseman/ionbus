package pubsub

import (
	"fmt"
	"strings"
	"sync"

	"github.com/evanwiseman/ionbus/internal/models"
)

type HandlerFunc func(msg models.Message)

type MessageMux struct {
	mu        sync.RWMutex
	handlers  map[string]HandlerFunc
	delimiter string // "/" for MQTT, "." for RMQ
}

func NewMessageMux(delim string) *MessageMux {
	return &MessageMux{
		handlers:  make(map[string]HandlerFunc),
		delimiter: delim,
	}
}

func (m *MessageMux) Handle(pattern string, h HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[pattern] = h
}

func (m *MessageMux) HandleFunc(pattern string, fn func(models.Message)) {
	m.Handle(pattern, fn)
}

func (m *MessageMux) Dispatch(msg models.Message) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for pattern, handler := range m.handlers {
		if match(pattern, msg.Topic, m.delimiter) {
			handler(msg)
			return nil
		}
	}
	return fmt.Errorf("no handler for topic %q", msg.Topic)
}

func match(pattern, topic, delim string) bool {
	partsP := strings.Split(pattern, delim)
	partsT := strings.Split(topic, delim)

	for i := 0; i < len(partsP) && i < len(partsT); i++ {
		switch partsP[i] {
		case "+", "*": // support MQTT and RMQ wildcard styles
			continue
		case "#":
			return true
		}
		if partsP[i] != partsT[i] {
			return false
		}
	}
	return len(partsP) == len(partsT)
}
