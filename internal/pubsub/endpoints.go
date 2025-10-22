package pubsub

const (
	ClientsPrefix   = "clients"
	CommandsPrefix  = "commands"
	GatewaysPrefix  = "gateways"
	ServerPrefix    = "server"
	ResponsesPrefix = "responses"
	TelemetryPrefix = "telemetry"
)

const (
	ExchangeIonbusTopic     = "ionbus_topic"
	ExchangeIonbusDirect    = "ionbus_direct"
	ExchangeIonbusBroadcast = "ionbus_broadcast"
	ExchangeIonbusDlx       = "ionbus_dlx"
)

const (
	QueueIonbusDlq = "ionbus_dlq"
)
