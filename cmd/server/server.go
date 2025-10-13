package main

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/config"
)

func main() {
	log.Println("Starting ionbus server...")

	serverConfig := config.LoadServerConfig()
}
