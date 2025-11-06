package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evanwiseman/ionbus/internal/config"
	"github.com/evanwiseman/ionbus/internal/services"
	"github.com/joho/godotenv"
)

func cleanup() {
	log.Println("Stopping ionbus node...")
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
	log.Println("Starting ionbus node...")

	// Load from .env update
	if err := godotenv.Load("./cmd/node/.env"); err != nil {
		log.Fatalf("Failed to load .env: %v", err)
	}
	cfg := config.LoadNodeConfig()
	// Create a new node
	node, err := services.NewNode(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create new node: %v", err)
	}
	defer node.Stop()

	time.Sleep(500 * time.Millisecond)

	// Start the node
	if err := node.Start(); err != nil {
		log.Fatalf("failed to start node: %v", err)
	}
	log.Println("Successfully started ionbus node")

	// Wait for cancellation
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
