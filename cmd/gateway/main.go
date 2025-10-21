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
	log.Print("Stopping ionbus gateway... ")
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
	log.Println("Starting ionbus gateway...")

	if err := godotenv.Load("./cmd/gateway/.env"); err != nil {
		log.Fatalf("Failed to load .env: %v", err)
	}
	cfg, err := LoadGatewayConfig()
	if err != nil {
		log.Fatalf("Failed to load gateway config: %v", err)
	}

	gateway, err := NewGateway(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}
	defer gateway.Close()

	if err := gateway.Start(); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
	log.Println("Successfully started ionbus gateway")

	// ========================
	// Wait for cancellation
	// ========================
	<-ctx.Done()
	log.Println("Context cancelled, shutting down gracefully...")
}
