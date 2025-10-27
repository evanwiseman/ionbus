package models

import "time"

type ActionType string

const (
	ActionGetIdentifiers ActionType = "get_identifiers"
)

type DeviceType string

const (
	DeviceGateway = "gateway"
	DeviceServer  = "server"
)

// Request represents a generic request over MQTT or RMQ
type Request struct {
	ID           string            `json:"id"`             // Unique request ID for correlation
	SourceID     string            `json:"source_id"`      // e.g., "gateway", "server", "client"
	SourceDevice DeviceType        `json:"source_device"`  // the source devices type for routing
	TargetID     string            `json:"target_id"`      // Optional target ID
	TargetDevice DeviceType        `json:"target_device"`  // the target device we want
	Action       ActionType        `json:"action"`         // What the recipient should do
	Timestamp    time.Time         `json:"timestamp"`      // When the request was created
	Headers      map[string]string `json:"headers"`        // Optional metadata (like content type, auth, etc.)
	Body         any               `json:"body,omitempty"` // Payload, can be any struct
}

type Response struct {
	ID           string     `json:"id"`
	TargetID     string     `json:"target"`
	TargetDevice DeviceType `json:"target_device"`
	SourceID     string     `json:"source_id"`
	SourceDevice DeviceType `json:"source_device"`
	Action       ActionType `json:"action"`
	Timestamp    time.Time  `json:"timestamp"`
	Body         any        `json:"body,omitempty"`
	Error        string     `json:"error,omitempty"`
}

type IdentifierBody struct {
	ID string `json:"id"`
}
