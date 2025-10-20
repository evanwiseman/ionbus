package models

type GatewayCommand struct {
	Name string                 `json:"name"`
	ID   string                 `json:"id"` // gateway ID
	Args map[string]interface{} `json:"args"`
}

type ClientCommand struct {
	Name string                 `json:"name"`
	ID   string                 `json:"client_id"` // client ID
	Args map[string]interface{} `json:"args"`
}
