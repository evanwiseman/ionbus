package main

import (
	"log"
	"time"

	"github.com/evanwiseman/ionbus/internal/models"
)

func (c *Client) HandlerClientCommands(command models.Command) {
	switch command.Name {
	case "request":
		err := c.Client2GatewayResponse(models.Response{
			Name:      c.Cfg.ID,
			Status:    "",
			Data:      nil,
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Printf("Failed to send response: %v", err)
		}
		log.Printf("received request command: %v", command.Args)
	default:
		break
	}

}
