#!/bin/bash

# Script to spawn multiple unique servers with different SERVER_IDs
# Usage: ./spawn-servers.sh [number_of_servers]

set -e

# Configuration
NUM_SERVERS=${1:-5}
SERVER_BINARY="./cmd/server/server"
ENV_FILE="./cmd/server/.env"
LOG_DIR="./logs/servers"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Create log directory if it doesn't exist
mkdir -p "$LOG_DIR"

# Check if server binary exists, if not try to build it
if [ ! -f "$SERVER_BINARY" ]; then
    echo -e "${YELLOW}Server binary not found. Building...${NC}"
    go build -o "$SERVER_BINARY" ./cmd/server
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to build server${NC}"
        exit 1
    fi
    echo -e "${GREEN}Server built successfully${NC}"
fi

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    exit 1
fi

# Array to store PIDs
declare -a SERVER_PIDS

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Shutting down all servers...${NC}"
    for pid in "${SERVER_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null
        fi
    done
    echo -e "${GREEN}All servers stopped${NC}"
    exit 0
}

# Trap CTRL+C and other termination signals
trap cleanup SIGINT SIGTERM

echo -e "${BLUE}Starting $NUM_SERVERS servers...${NC}"

# Spawn servers
for i in $(seq 1 $NUM_SERVERS); do
    SERVER_ID="server_${i}"
    LOG_FILE="$LOG_DIR/${SERVER_ID}.log"
    
    # Export SERVER_ID for this specific instance
    export SERVER_ID="$SERVER_ID"
    
    # Start the server in the background
    "$SERVER_BINARY" > "$LOG_FILE" 2>&1 &
    
    SERVER_PID=$!
    SERVER_PIDS+=($SERVER_PID)
    
    echo -e "${GREEN}Started server ${i}/${NUM_SERVERS}${NC} - ID: ${SERVER_ID}, PID: ${SERVER_PID}"
    
    # Small delay to avoid overwhelming the system
    sleep 0.1
done

echo -e "\n${GREEN}All $NUM_SERVERS servers started successfully!${NC}"
echo -e "${BLUE}Logs are available in: $LOG_DIR${NC}"
echo -e "${YELLOW}Press CTRL+C to stop all servers${NC}\n"

# Monitor servers
while true; do
    sleep 5
    
    # Check if any servers have died
    for i in "${!SERVER_PIDS[@]}"; do
        pid="${SERVER_PIDS[$i]}"
        if ! kill -0 "$pid" 2>/dev/null; then
            echo -e "${RED}Server with PID $pid has stopped${NC}"
            unset 'SERVER_PIDS[$i]'
        fi
    done
    
    # If all servers are dead, exit
    if [ ${#SERVER_PIDS[@]} -eq 0 ]; then
        echo -e "${RED}All servers have stopped${NC}"
        exit 1
    fi
done