package pubsub

type RMQPublishOptions struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
}

type MQTTPublishOptions struct {
	Topic    string
	Qos      byte
	Retained bool
}
