package models

import "github.com/google/uuid"

type Handler struct {
	ID         uuid.UUID `json:"id"`
	Method     string    `json:"method"`
	Language   string    `json:"language"`
	Extension  string    `json:"extension"`
	Program    []byte    `json:"program"`
	EntryPoint string    `json:"entry_point"`
	Persistent bool      `json:"persistent"`
}
