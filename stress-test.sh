#!/bin/sh
# Stress test script for MCPProbe against MCPFusion health endpoint
# Usage: ./stress-test.sh [repeat] [concurrent]
#   repeat:     total number of calls (default: 500)
#   concurrent: number of parallel workers (default: 25)
# Requires .env file with SERVER_URL and APIKEY

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

REPEAT=${1:-500}
CONCURRENT=${2:-25}

echo "Building..."
go build -o probe
echo ""

./probe -url "${SERVER_URL}/mcp" \
    -transport http \
    -headers "Authorization:Bearer ${APIKEY}" \
    -call "health_status" \
    -params "{}" \
    -repeat "$REPEAT" \
    -concurrent "$CONCURRENT"
