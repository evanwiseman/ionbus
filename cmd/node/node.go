package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanwiseman/ionbus/internal/brokers/amqpx"
	"github.com/evanwiseman/ionbus/internal/brokers/mqttx"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/joho/godotenv"
)

func cleanup() {
	log.Print("Stopping ionbus node... ")
}

func main() {
	godotenv.Load()

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
	log.Println("Starting ionbus node...")

	// ========================
	// Start RabbitMQ
	// ========================
	log.Println("Connecting to RabbitMQ...")
	rabbitConfig := config.LoadRabbitConfig()
	rabbitConn, err := amqpx.NewConnection(rabbitConfig)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer rabbitConn.Close()
	log.Println("Sucessfully connected to RabbitMQ")

	// ========================
	// Start MQTT
	// ========================
	log.Println("Connecting to MQTT...")
	mqttConfig := config.LoadMQTTConfig("client")
	mqttClient, err := mqttx.NewClient(mqttConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", err)
	}
	defer mqttClient.Disconnect(250)

	// ========================
	// Subscribe to MQTT Topics
	// ========================

	// ========================
	// Confirm hub is started
	// ========================
	log.Println("Successfully started ionbus hub")

	// ========================
	// Wait for cancellation
	// ========================
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
