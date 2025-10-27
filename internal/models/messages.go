package models

type Message struct {
	Topic   string
	Payload []byte
	Source  string
}
