package main

import (
	"log"

	"github.com/evanwiseman/ionbus/internal/brokers/mqttx"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	log.Println("Starting ionbus client...")

	mqttConfig := config.LoadMQTTConfig("client")
	mqttClient, err := mqttx.NewClient(mqttConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", err)
	}
	defer mqttClient.Disconnect(250)
}
