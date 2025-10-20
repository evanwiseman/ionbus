package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

func cleanup() {

}

func main() {
	// Listen for OS signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %s, shutting down...\n", sig)
		cancel()
	}()

	// Run main program logic
	run(ctx)

	// Perform cleanup after run returns
	cleanup()
}

func run(ctx context.Context) {
	log.Println("Starting ionbus client...")

	// Load from .env update
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	cfg, err := LoadClientConfig()
	if err != nil {
		log.Fatalf("Failed to get server config: %v\n", err)
	}

	// ========================
	// Start MQTT
	// ========================
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTT.GetUrl())
	opts.SetKeepAlive(cfg.MQTT.KeepAlive)
	opts.SetCleanSession(cfg.MQTT.CleanSession)
	opts.SetClientID(cfg.ID)
	opts.SetUsername(cfg.MQTT.Username)
	opts.SetPassword(cfg.MQTT.Password)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", token.Error())
	}
	defer mqttClient.Disconnect(250)
	log.Println("Successfully connected to MQTT")

	// ========================
	// Wait for cancellation
	// ========================
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
