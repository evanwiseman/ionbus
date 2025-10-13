package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanwiseman/ionbus/internal/brokers/amqpx"
	"github.com/evanwiseman/ionbus/internal/brokers/mqttx"
	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func cleanup() {
	log.Println("Stopping ionbus server...")
}

func main() {
	godotenv.Load()
	// Capture ctrl + c for cleanup
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctrlC
		cleanup()
		os.Exit(1)
	}()

	log.Println("Starting ionbus server...")

	// ========================
	// Start RabbitMQ
	// ========================
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
	mqttConfig := config.LoadMQTTConfig("server")
	mqttClient, err := mqttx.NewClient(mqttConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT: %v\n", err)
	}
	defer mqttClient.Disconnect(250)
	log.Println("Successfully connected to MQTT")

	// ========================
	// Open Postgres Database
	// ========================
	pgConfig := config.LoadPostgresConfig()
	db, err := sql.Open("postgres", pgConfig.Url)
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}
	defer db.Close()
	log.Println("Successfully opened database")

	cleanup()
}
