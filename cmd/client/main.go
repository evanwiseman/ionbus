package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func cleanup() {
	log.Println("Stopping ionbus client...")
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
	if err := godotenv.Load("./cmd/client/.env"); err != nil {
		log.Fatalf("Failed to load .env: %v", err)
	}
	cfg, err := LoadClientConfig()
	if err != nil {
		log.Fatalf("Failed to load cilent config: %v", err)
	}

	// Create a new client
	client, err := NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create new client: %v", err)
	}
	defer client.Close()

	time.Sleep(500 * time.Millisecond)

	// Start the client
	if err := client.Start(); err != nil {
		log.Fatalf("failed to start client: %v", err)
	}
	log.Println("Successfully started ionbus client")

	// Wait for cancellation
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
