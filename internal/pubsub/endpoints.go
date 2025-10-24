package pubsub

import "fmt"

// Queue names
const (
	BroadcastPrefix = "broadcast"
	CommandPrefix   = "command"
	RequestPrefix   = "request"
	ResponsePrefix  = "response"
	ClientPrefix    = "client"
	GatewayPrefix   = "gateway"
	ServerPrefix    = "server"
	QIonbusDlq      = "ionbus.dlq"
)

// Exchange names
const (
	XIonbusDlx = "ionbus.dlx"
)

// ========================
// Client Command
// ========================
func GetMQTTClientCommandTopic(clientID, action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", ClientPrefix, clientID, CommandPrefix, action)
}

func GetMQTTClientCommandBroadcast(action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, ClientPrefix, CommandPrefix, action)
}

func GetMQTTClientResponseTopic(clientID, action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", ClientPrefix, clientID, ResponsePrefix, action)
}

// ========================
// Gateway Command
// ========================

func GetRMQGatewayCommandTopicX() string {
	return fmt.Sprintf("%s.%s.topic", GatewayPrefix, CommandPrefix)
}

func GetRMQGatewayCommandBroadcastX() string {
	return fmt.Sprintf("%s.%s.broadcast", GatewayPrefix, CommandPrefix)
}

func GetRMQGatewayCommandQ(gatewayID string) string {
	return fmt.Sprintf("%s.%s", gatewayID, CommandPrefix)
}

func GetRMQGatewayCommandRK(gatewayID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", GatewayPrefix, gatewayID, CommandPrefix, action)
}

// ========================
// Server Command
// ========================

func GetRMQServerCommandTopicX() string {
	return fmt.Sprintf("%s.%s.topic", ServerPrefix, CommandPrefix)
}

func GetRMQServerCommandBroadcastX() string {
	return fmt.Sprintf("%s.%s.broadcast", ServerPrefix, CommandPrefix)
}

func GetRMQServerCommandQ(serverID string) string {
	return fmt.Sprintf("%s.%s", serverID, CommandPrefix)
}

func GetRMQServerCommandRK(serverID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", ServerPrefix, serverID, CommandPrefix, action)
}

// ========================
// Gateway Response
// ========================

func GetRMQGatewayResponseTopicX() string {
	return fmt.Sprintf("%s.%s.topic", GatewayPrefix, ResponsePrefix)
}

func GetRMQGatewayResponseQ(gatewayID string) string {
	return fmt.Sprintf("%s.%s", gatewayID, ResponsePrefix)
}

// The routing key the gateway posts to, the server listens to gatewayID='*'
func GetRMQGatewayResponseRKK(gatewayID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", GatewayPrefix, gatewayID, ResponsePrefix, action)
}

func GetMQTTGatewayResponseTopic(gatewayID, action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", GatewayPrefix, gatewayID, ResponsePrefix, action)
}

func GetMQTTGatewayResponseBroadcast(action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, GatewayPrefix, ResponsePrefix, action)
}

// ========================
// Server Response
// ========================

func GetRMQServerResponseTopicX() string {
	return fmt.Sprintf("%s.%s.topic", ServerPrefix, ResponsePrefix)
}

func GetRMQServerResponseQ(serverID string) string {
	return fmt.Sprintf("%s.%s", serverID, ResponsePrefix)
}

func GetRMQServerResponseRK(serverID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", ServerPrefix, serverID, ResponsePrefix, action)
}
