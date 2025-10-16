package routing

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

type ContentType string

const (
	ContentJSON ContentType = "application/json"
	ContentGOB  ContentType = "application/gob"
)

func Marshal(v any, contentType ContentType) ([]byte, error) {
	switch contentType {
	case ContentJSON:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return data, nil

	case ContentGOB:
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(v); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func Unmarshal(data []byte, contentType ContentType, v any) error {
	switch contentType {
	case ContentJSON:
		return json.Unmarshal(data, v)
	case ContentGOB:
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		return dec.Decode(v)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

type ExchangeType string

const (
	Direct  ExchangeType = "direct"
	Fanout  ExchangeType = "fanout"
	Headers ExchangeType = "headers"
	Topic   ExchangeType = "topic"
)

type QueueType int

const (
	Durable QueueType = iota
	Transient
)
