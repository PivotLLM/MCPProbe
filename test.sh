#!/bin/sh
# Test script for MCPProbe against MCPFusion
# Requires .env file with SERVER_URL and APIKEY

set -e

if [ ! -f .env ]; then
    echo "Error: .env file not found"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' .env | xargs)

if [ -z "$SERVER_URL" ] || [ -z "$APIKEY" ]; then
    echo "Error: SERVER_URL and APIKEY must be set in .env"
    exit 1
fi

echo "Building..."
go build -o probe
echo ""

echo "=== Enumerate MCPFusion Services ==="
echo ""
./probe -url "${SERVER_URL}/mcp" \
    -transport http \
    -headers "Authorization:Bearer ${APIKEY}" \
    -list-only

echo ""
echo "=== Call health_status Tool ==="
echo ""
./probe -url "${SERVER_URL}/mcp" \
    -transport http \
    -headers "Authorization:Bearer ${APIKEY}" \
    -call "health_status" \
    -params "{}"
