package pubsub

import "fmt"

const (
	BroadcastPrefix = "broadcast"
	TopicPrefix     = "topic"

	RequestPrefix  = "request"
	ResponsePrefix = "response"

	ClientPrefix  = "client"
	GatewayPrefix = "gateway"
	ServerPrefix  = "server"

	XIonbusDlx = "ionbus.dlx"
	QIonbusDlq = "ionbus.dlq"
)

// ========================
// Client
// ========================
func MClientReqT(clientID, action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", ClientPrefix, clientID, RequestPrefix, action)
}

func MClientReqB(action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, ClientPrefix, RequestPrefix, action)
}

func MClientResT(clientID, action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", ClientPrefix, clientID, ResponsePrefix, action)
}

func MClientResB(action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, ClientPrefix, ResponsePrefix, action)
}

// ========================
// Gateway Requests
// ========================

// MQTT Request Topic
func MGatewayReqT(gatewayID, action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", GatewayPrefix, gatewayID, RequestPrefix, action)
}

// MQTT Request Broadcast
func MGatewayReqB(action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, ClientPrefix, RequestPrefix, action)
}

// RabbitMQ Request Topic Exchange
func RGatewayReqTX() string {
	return fmt.Sprintf("%s.%s.%s", GatewayPrefix, RequestPrefix, TopicPrefix)
}

// RabbitMQ Request Broadcast Exchange
func RGatewayReqBX() string {
	return fmt.Sprintf("%s.%s.%s", GatewayPrefix, RequestPrefix, BroadcastPrefix)
}

// RabbitMQ Request Queue
func RGatewayReqQ(gatewayID string) string {
	return fmt.Sprintf("%s.%s", gatewayID, RequestPrefix)
}

// RabbitMQ Request Topic Routing Key
func RGatewayReqTRK(gatewayID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", GatewayPrefix, gatewayID, RequestPrefix, action)
}

// RabbitMQ Request Broadcast Routing Key
func RGatewayReqBRK(action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", BroadcastPrefix, GatewayPrefix, RequestPrefix, action)
}

// ========================
// Server Requests
// ========================

// RabbitMQ Server Request Topic Exchange
func RServerReqTX() string {
	return fmt.Sprintf("%s.%s.%s", ServerPrefix, RequestPrefix, TopicPrefix)
}

// RabbitMQ Server Request Broadcast Exchange
func RServerReqBX() string {
	return fmt.Sprintf("%s.%s.%s", ServerPrefix, RequestPrefix, BroadcastPrefix)
}

// RabbitMQ Server Request Queue
func RServerReqQ(serverID string) string {
	return fmt.Sprintf("%s.%s", serverID, RequestPrefix)
}

// RabbitMQ Server Request Topic Routing Key
func RServerReqTRK(serverID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", ServerPrefix, serverID, RequestPrefix, action)
}

// RabbitMQ Server Request Broadcast Routing Key
func RServerReqBRK(action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", BroadcastPrefix, ServerPrefix, RequestPrefix, action)
}

// ========================
// Gateway Response
// ========================

func RGatewayResTX() string {
	return fmt.Sprintf("%s.%s.%s", GatewayPrefix, ResponsePrefix, TopicPrefix)
}

func RGatewayResQ(gatewayID string) string {
	return fmt.Sprintf("%s.%s", gatewayID, ResponsePrefix)
}

func RGatewayResTRK(gatewayID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", GatewayPrefix, gatewayID, ResponsePrefix, action)
}

func MGatewayResT(gatewayID, action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", GatewayPrefix, gatewayID, ResponsePrefix, action)
}

func MGatewayResB(action string) string {
	return fmt.Sprintf("%s/%s/%s/%s", BroadcastPrefix, GatewayPrefix, ResponsePrefix, action)
}

// ========================
// Server Response
// ========================

func RServerResTX() string {
	return fmt.Sprintf("%s.%s.%s", ServerPrefix, ResponsePrefix, TopicPrefix)
}

func RServerResQ(serverID string) string {
	return fmt.Sprintf("%s.%s", serverID, ResponsePrefix)
}

func RServerResTRK(serverID, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s", ServerPrefix, serverID, ResponsePrefix, action)
}
