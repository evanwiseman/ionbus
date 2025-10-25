package models

import "time"

type Command struct {
	Name   string      `json:"name"`
	Sender string      `json:"sender"`
	Args   interface{} `json:"args"`
}

type TargetType string

const (
	AllIdentifiers     = "all_identifiers"
	ClientIdentifiers  = "client_identifiers"
	GatewayIdentifiers = "gateway_identifiers"
)

type RequestArgs struct {
	Target    TargetType             `json:"target"`
	Filters   map[string]interface{} `json:"filters"`
	Timestamp time.Time              `json:"timestamp"`
	Reason    string                 `json:"reason,omitempty"`
}

type ResponseType string

const (
	ResponseIdentifier = "identifier"
)

type Response struct {
	Name      string       `json:"name"`
	Status    string       `json:"status"`
	Type      ResponseType `json:"response_type,omitempty"`
	Data      interface{}  `json:"data,omitempty"`
	Error     string       `json:"error,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

type GatewayIdentifier struct {
	ID       string                 `json:"ID"`
	Version  string                 `json:"version,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ClientIdentifier struct {
	ID       string                 `json:"ID"`
	Version  string                 `json:"version,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
