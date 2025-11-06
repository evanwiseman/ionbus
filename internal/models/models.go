package models

import "encoding/json"

type CommandType string

const (
	CommandRequest  = "request"
	CommandResponse = "response"
)

type MethodType string

const (
	MethodGetIdentifiers MethodType = "GET_IDENTIFIERS"
)

type DeviceType string

const (
	DeviceNode    = "node"
	DeviceGateway = "gateway"
	DeviceServer  = "server"
)

type Message struct {
	SourceID     string          `json:"source_id"`
	SourceDevice DeviceType      `json:"source_device"`
	Version      string          `json:"version"`
	Payload      json.RawMessage `json:"payload"`
}

type Request struct {
	Method       string          `json:"method"`
	TargetID     string          `json:"target_id"`
	TargetDevice DeviceType      `json:"target_device"`
	Payload      json.RawMessage `json:"payload"`
}

type Response struct {
	Method       string          `json:"method"`
	TargetID     string          `json:"target_id"`
	TargetDevice DeviceType      `json:"target_device"`
	StatusCode   int             `json:"status_code,omitempty"`
	Error        string          `json:"error,omitempty"`
	Payload      json.RawMessage `json:"payload,omitempty"`
}
