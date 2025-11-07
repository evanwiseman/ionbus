package models

import "github.com/google/uuid"

type Handler struct {
	ID         uuid.UUID `json:"id"`
	Method     string    `json:"method"`
	RunCommand string    `json:"run_command"`
	Extension  string    `json:"extension"`
	Program    []byte    `json:"program,omitempty"`
	Version    string    `json:"version"`
}
