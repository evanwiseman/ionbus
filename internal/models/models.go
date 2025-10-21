package models

type Command struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}
