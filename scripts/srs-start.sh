#!/bin/bash

# Start SRS media server for local development
# This script starts SRS in the background with the local configuration

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

SRS_BIN="$PROJECT_DIR/.srs/usr/local/srs/objs/srs"
SRS_CONF="$PROJECT_DIR/srs.local.conf"
SRS_PID_FILE="$PROJECT_DIR/.srs/srs.pid"
SRS_LOG_FILE="$PROJECT_DIR/.srs/srs.log"

# Create directories if they don't exist
mkdir -p "$PROJECT_DIR/.srs/hls-cache"

# Check if SRS is already running
if [ -f "$SRS_PID_FILE" ]; then
    PID=$(cat "$SRS_PID_FILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "SRS is already running (PID: $PID)"
        exit 0
    else
        echo "Removing stale PID file"
        rm -f "$SRS_PID_FILE"
    fi
fi

# Check if binary exists
if [ ! -f "$SRS_BIN" ]; then
    echo "Error: SRS binary not found at $SRS_BIN"
    echo "Please run 'make srs-install' first"
    exit 1
fi

# Check if config exists
if [ ! -f "$SRS_CONF" ]; then
    echo "Error: SRS config not found at $SRS_CONF"
    exit 1
fi

echo "Starting SRS media server..."
echo "  Binary:  $SRS_BIN"
echo "  Config:  $SRS_CONF"
echo "  Log:     $SRS_LOG_FILE"
echo ""

# Start SRS in the background
nohup "$SRS_BIN" -c "$SRS_CONF" > "$SRS_LOG_FILE" 2>&1 &
SRS_PID=$!

# Save PID
echo $SRS_PID > "$SRS_PID_FILE"

# Wait a moment and check if it started successfully
sleep 1

if ps -p "$SRS_PID" > /dev/null 2>&1; then
    echo "✓ SRS started successfully (PID: $SRS_PID)"
    echo ""
    echo "RTMP URL: rtmp://localhost:1935/live/{streamKey}"
    echo "HLS URL:  http://localhost:8080/live/{streamKey}.m3u8"
    echo "HTTP API: http://localhost:1985/api/v1/"
    echo ""
    echo "To stop: make srs-stop"
    echo "To view logs: tail -f .srs/srs.log"
else
    echo "✗ Failed to start SRS"
    echo "Check logs at: $SRS_LOG_FILE"
    rm -f "$SRS_PID_FILE"
    exit 1
fi
