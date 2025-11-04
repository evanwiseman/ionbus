package gateway

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/evanwiseman/ionbus/internal/models"
)

func (g *Gateway) HandlerRequests(data []byte) error {
	// Unmarshal the message
	var msg models.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Unmarshal request
	var req models.Request
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf(
		"Received request: Method=%s, Source=%s (%s)",
		req.Method, msg.SourceID, msg.SourceDevice,
	)

	switch models.Method(req.Method) {
	case models.MethodGetIdentifiers:
		return g.handleGetIdentifiers(msg, req)
	default:
		log.Printf("Unknown method: %s", req.Method)
		return nil // Don't error on unknown methods, just ignore
	}
}

func (g *Gateway) handleGetIdentifiers(msg models.Message, req models.Request) error {
	payload, err := json.Marshal(g.Cfg.ID)
	if err != nil {
		return fmt.Errorf("error: failed to marshal id: %w", err)
	}

	// Create response with server identifier
	res := models.Response{
		Method:       req.Method,
		TargetID:     msg.SourceID, // Reply to the sender
		TargetDevice: msg.SourceDevice,
		Payload:      payload,
	}

	switch msg.SourceDevice {
	case models.DeviceClient:
		return g.SendClientResponse(res)
	case models.DeviceServer:
		return g.SendServerResponse(res)
	default:
		return fmt.Errorf("error: unsupported source")
	}
	// Send the response
}

func (g *Gateway) HandlerResponses(data []byte) error {
	var msg models.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	var res models.Response
	if err := json.Unmarshal(msg.Payload, &res); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Printf(
		"Received response: Method=%s, Source=%s (%s)",
		res.Method, msg.SourceID, msg.SourceDevice,
	)

	switch models.Method(res.Method) {
	case models.MethodGetIdentifiers:
		return g.handleIdentifiersResponse(msg, res)
	}
	return nil
}

func (g *Gateway) handleIdentifiersResponse(msg models.Message, res models.Response) error {
	// Parse the identifier body
	bodyBytes, err := json.Marshal(res.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	var id string
	if err := json.Unmarshal(bodyBytes, &id); err != nil {
		return fmt.Errorf("failed to unmarshal identifier body: %w", err)
	}

	log.Printf(
		"%s %s identified with ID: %s",
		msg.SourceDevice, msg.SourceID, id,
	)

	// TODO: Store identifier in database
	// TODO: Update status/registry

	return nil
}
