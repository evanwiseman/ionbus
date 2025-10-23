package pubsub

import "fmt"

// Queue names
const (
	CommandPrefix  = "command"
	RequestPrefix  = "request"
	ResponsePrefix = "response"
	GatewayPrefix  = "gateway"
	ServerPrefix   = "server"
	QIonbusDlq     = "ionbus.dlq"
)

// Exchange names
const (
	XIonbusDlx = "ionbus.dlx"
)

// ========================
// Commands
// ========================

func GetGatewayCommandTopicX() string {
	return fmt.Sprintf("%s.%s.topic", GatewayPrefix, CommandPrefix)
}

func GetGatewayCommandBroadcastX() string {
	return fmt.Sprintf("%s.%s.broadcast", GatewayPrefix, CommandPrefix)
}

func GetGatewayCommandQ(gatewayID string) string {
	return fmt.Sprintf("%s.%s", gatewayID, CommandPrefix)
}

func GetGatewayCommandRK(gatewayID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", GatewayPrefix, gatewayID, CommandPrefix, action)
}

// ========================
// Responses
// ========================

func GetGatewayResponseTopicX() string {
	return fmt.Sprintf("%s.%s.topic", GatewayPrefix, ResponsePrefix)
}

// Lives in the server, collects the gateway responses
func GetGatewayResponseQ(serverID string) string {
	return fmt.Sprintf("%s.%s.%s", serverID, GatewayPrefix, ResponsePrefix)
}

// The routing key the gateway posts to, the server listens to gatewayID='*'
func GetGatewayResponseRK(gatewayID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", GatewayPrefix, gatewayID, ResponsePrefix, action)
}
