package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func cleanup() {
	log.Println("Stopping ionbus server...")
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
	log.Println("Starting ionbus server...")

	// Load from .env update
	if err := godotenv.Load("./cmd/server/.env"); err != nil {
		log.Fatalf("Failed to load .env: %v", err)
	}
	cfg, err := LoadServerConfig()
	if err != nil {
		log.Fatalf("Failed to load server config: %v", err)
	}

	// Create a new server
	server, err := NewServer(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create new server: %v", err)
	}
	defer server.Close()

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Println("Successfully started ionbus server")

	// Wait for cancellation
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
