package routing

type ContentType string

const (
	ContentJSON ContentType = "application/json"
	ContentGOB  ContentType = "application/gob"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

type Message struct {
	Topic       string
	Payload     []byte
	ContentType ContentType
}
