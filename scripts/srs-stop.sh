#!/bin/bash

# Stop SRS media server

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

SRS_PID_FILE="$PROJECT_DIR/.srs/srs.pid"

if [ ! -f "$SRS_PID_FILE" ]; then
    echo "SRS is not running (no PID file found)"
    exit 0
fi

PID=$(cat "$SRS_PID_FILE")

if ! ps -p "$PID" > /dev/null 2>&1; then
    echo "SRS is not running (stale PID file)"
    rm -f "$SRS_PID_FILE"
    exit 0
fi

echo "Stopping SRS (PID: $PID)..."
kill "$PID"

# Wait for process to terminate
for i in {1..10}; do
    if ! ps -p "$PID" > /dev/null 2>&1; then
        echo "✓ SRS stopped successfully"
        rm -f "$SRS_PID_FILE"
        exit 0
    fi
    sleep 0.5
done

# Force kill if still running
if ps -p "$PID" > /dev/null 2>&1; then
    echo "Force stopping SRS..."
    kill -9 "$PID"
    rm -f "$SRS_PID_FILE"
    echo "✓ SRS stopped (forced)"
else
    echo "✓ SRS stopped successfully"
    rm -f "$SRS_PID_FILE"
fi
