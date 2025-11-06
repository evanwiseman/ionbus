package broker

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
