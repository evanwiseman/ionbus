#!/bin/bash

# Script to spawn 100 unique clients with different CLIENT_IDs
# Usage: ./spawn-clients.sh [number_of_clients]

set -e

# Configuration
NUM_CLIENTS=${1:-100}
CLIENT_BINARY="./cmd/client/client"
ENV_FILE="./cmd/client/.env"
LOG_DIR="./logs/clients"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Create log directory if it doesn't exist
mkdir -p "$LOG_DIR"

# Check if client binary exists, if not try to build it
if [ ! -f "$CLIENT_BINARY" ]; then
    echo -e "${YELLOW}Client binary not found. Building...${NC}"
    go build -o "$CLIENT_BINARY" ./cmd/client
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to build client${NC}"
        exit 1
    fi
    echo -e "${GREEN}Client built successfully${NC}"
fi

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    exit 1
fi

# Array to store PIDs
declare -a CLIENT_PIDS

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Shutting down all clients...${NC}"
    for pid in "${CLIENT_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null
        fi
    done
    echo -e "${GREEN}All clients stopped${NC}"
    exit 0
}

# Trap CTRL+C and other termination signals
trap cleanup SIGINT SIGTERM

echo -e "${BLUE}Starting $NUM_CLIENTS clients...${NC}"

# Spawn clients
for i in $(seq 1 $NUM_CLIENTS); do
    CLIENT_ID="client_${i}"
    LOG_FILE="$LOG_DIR/${CLIENT_ID}.log"
    
    # Export CLIENT_ID for this specific instance
    export CLIENT_ID="$CLIENT_ID"
    
    # Start the client in the background
    "$CLIENT_BINARY" > "$LOG_FILE" 2>&1 &
    
    CLIENT_PID=$!
    CLIENT_PIDS+=($CLIENT_PID)
    
    echo -e "${GREEN}Started client ${i}/${NUM_CLIENTS}${NC} - ID: ${CLIENT_ID}, PID: ${CLIENT_PID}"
    
    # Small delay to avoid overwhelming the system
    sleep 0.1
done

echo -e "\n${GREEN}All $NUM_CLIENTS clients started successfully!${NC}"
echo -e "${BLUE}Logs are available in: $LOG_DIR${NC}"
echo -e "${YELLOW}Press CTRL+C to stop all clients${NC}\n"

# Monitor clients
while true; do
    sleep 5
    
    # Check if any clients have died
    for i in "${!CLIENT_PIDS[@]}"; do
        pid="${CLIENT_PIDS[$i]}"
        if ! kill -0 "$pid" 2>/dev/null; then
            echo -e "${RED}Client with PID $pid has stopped${NC}"
            unset 'CLIENT_PIDS[$i]'
        fi
    done
    
    # If all clients are dead, exit
    if [ ${#CLIENT_PIDS[@]} -eq 0 ]; then
        echo -e "${RED}All clients have stopped${NC}"
        exit 1
    fi
done