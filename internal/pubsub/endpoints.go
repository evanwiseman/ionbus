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
// Gateway Command
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
// Server Command
// ========================

func GetServerCommandTopicX() string {
	return fmt.Sprintf("%s.%s.topic", ServerPrefix, CommandPrefix)
}

func GetServerCommandBroadcastX() string {
	return fmt.Sprintf("%s.%s.broadcast", ServerPrefix, CommandPrefix)
}

func GetServerCommandQ(serverID string) string {
	return fmt.Sprintf("%s.%s", serverID, CommandPrefix)
}

func GetServerCommandRK(serverID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", ServerPrefix, serverID, CommandPrefix, action)
}

// ========================
// Gateway Response
// ========================

func GetGatewayResponseTopicX() string {
	return fmt.Sprintf("%s.%s.topic", GatewayPrefix, ResponsePrefix)
}

func GetGatewayResponseQ(gatewayID string) string {
	return fmt.Sprintf("%s.%s", gatewayID, ResponsePrefix)
}

// The routing key the gateway posts to, the server listens to gatewayID='*'
func GetGatewayResponseRK(gatewayID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", GatewayPrefix, gatewayID, ResponsePrefix, action)
}

// ========================
// Server Response
// ========================

func GetServerResponseTopicX() string {
	return fmt.Sprintf("%s.%s.topic", ServerPrefix, ResponsePrefix)
}

func GetServerResponseQ(serverID string) string {
	return fmt.Sprintf("%s.%s", serverID, ResponsePrefix)
}

func GetServerResponseRK(serverID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", ServerPrefix, serverID, ResponsePrefix, action)
}
