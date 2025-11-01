package models

import "encoding/json"

type Action string

const (
	ActionRequest  = "request"
	ActionResponse = "response"
)

type Method string

const (
	MethodGetIdentifiers Method = "GET_IDENTIFIERS"
)

type Device string

const (
	DeviceClient  = "client"
	DeviceGateway = "gateway"
	DeviceServer  = "server"
)

type Message struct {
	SourceID     string          `json:"source_id"`
	SourceDevice Device          `json:"source_device"`
	Version      string          `json:"version"`
	Payload      json.RawMessage `json:"payload"`
}

type Request struct {
	Method       string          `json:"method"`
	TargetID     string          `json:"target_id"`
	TargetDevice Device          `json:"target_device"`
	Payload      json.RawMessage `json:"payload"`
}

type Response struct {
	Method       string          `json:"method"`
	TargetID     string          `json:"target_id"`
	TargetDevice Device          `json:"target_device"`
	StatusCode   int             `json:"status_code,omitempty"`
	Error        string          `json:"error,omitempty"`
	Payload      json.RawMessage `json:"payload,omitempty"`
}
