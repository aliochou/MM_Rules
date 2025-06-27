#!/bin/bash

# MM-Rules Server Stop Script

echo "🔄 Stopping MM-Rules Server..."

# Find and kill any process on port 8080
EXISTING_PID=$(lsof -ti:8080 2>/dev/null)

if [ ! -z "$EXISTING_PID" ]; then
    echo "📋 Found server process on port 8080 (PID: $EXISTING_PID)"
    echo "🔄 Terminating process..."
    kill -9 $EXISTING_PID 2>/dev/null
    sleep 2
    
    # Check if it's still running
    if lsof -ti:8080 > /dev/null 2>&1; then
        echo "❌ Failed to stop server process"
        exit 1
    else
        echo "✅ Server process terminated successfully"
    fi
else
    echo "✅ No server process found on port 8080"
fi 