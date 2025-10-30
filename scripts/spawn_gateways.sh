#!/bin/bash

# Script to spawn multiple unique gateways with different GATEWAY_IDs
# Usage: ./spawn-gateways.sh [number_of_gateways]

set -e

# Configuration
NUM_GATEWAYS=${1:-10}
GATEWAY_BINARY="./cmd/gateway/gateway"
ENV_FILE="./cmd/gateway/.env"
LOG_DIR="./logs/gateways"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Create log directory if it doesn't exist
mkdir -p "$LOG_DIR"

# Check if gateway binary exists, if not try to build it
if [ ! -f "$GATEWAY_BINARY" ]; then
    echo -e "${YELLOW}Gateway binary not found. Building...${NC}"
    go build -o "$GATEWAY_BINARY" ./cmd/gateway
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to build gateway${NC}"
        exit 1
    fi
    echo -e "${GREEN}Gateway built successfully${NC}"
fi

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    exit 1
fi

# Array to store PIDs
declare -a GATEWAY_PIDS

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Shutting down all gateways...${NC}"
    for pid in "${GATEWAY_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null
        fi
    done
    echo -e "${GREEN}All gateways stopped${NC}"
    exit 0
}

# Trap CTRL+C and other termination signals
trap cleanup SIGINT SIGTERM

echo -e "${BLUE}Starting $NUM_GATEWAYS gateways...${NC}"

# Spawn gateways
for i in $(seq 1 $NUM_GATEWAYS); do
    GATEWAY_ID="gateway_${i}"
    LOG_FILE="$LOG_DIR/${GATEWAY_ID}.log"
    
    # Export GATEWAY_ID for this specific instance
    export GATEWAY_ID="$GATEWAY_ID"
    
    # Start the gateway in the background
    "$GATEWAY_BINARY" > "$LOG_FILE" 2>&1 &
    
    GATEWAY_PID=$!
    GATEWAY_PIDS+=($GATEWAY_PID)
    
    echo -e "${GREEN}Started gateway ${i}/${NUM_GATEWAYS}${NC} - ID: ${GATEWAY_ID}, PID: ${GATEWAY_PID}"
    
    # Small delay to avoid overwhelming the system
    sleep 0.1
done

echo -e "\n${GREEN}All $NUM_GATEWAYS gateways started successfully!${NC}"
echo -e "${BLUE}Logs are available in: $LOG_DIR${NC}"
echo -e "${YELLOW}Press CTRL+C to stop all gateways${NC}\n"

# Monitor gateways
while true; do
    sleep 5
    
    # Check if any gateways have died
    for i in "${!GATEWAY_PIDS[@]}"; do
        pid="${GATEWAY_PIDS[$i]}"
        if ! kill -0 "$pid" 2>/dev/null; then
            echo -e "${RED}Gateway with PID $pid has stopped${NC}"
            unset 'GATEWAY_PIDS[$i]'
        fi
    done
    
    # If all gateways are dead, exit
    if [ ${#GATEWAY_PIDS[@]} -eq 0 ]; then
        echo -e "${RED}All gateways have stopped${NC}"
        exit 1
    fi
done