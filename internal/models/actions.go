package models

type CommandType string

const (
	CommandRequest  = "request"
	CommandResponse = "response"
)

type MethodType string

const (
	MethodGetIdentifiers MethodType = "GET_IDENTIFIERS"
)
