#!/bin/bash

# MM-Rules Server Management Script

echo "🔄 Starting MM-Rules Server..."

# Kill any existing process on port 8080
echo "📋 Checking for existing server process..."
EXISTING_PID=$(lsof -ti:8080 2>/dev/null)

if [ ! -z "$EXISTING_PID" ]; then
    echo "⚠️  Found existing process on port 8080 (PID: $EXISTING_PID)"
    echo "🔄 Killing existing process..."
    kill -9 $EXISTING_PID 2>/dev/null
    sleep 2
    echo "✅ Existing process terminated"
else
    echo "✅ No existing process found on port 8080"
fi

# Start the server
echo "🚀 Starting new server process..."
cd "$(dirname "$0")/.."
go run cmd/server/main.go &

# Get the PID of the new process
SERVER_PID=$!
echo "✅ Server started with PID: $SERVER_PID"

# Wait a moment and check if it's running
sleep 3
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "🎉 Server is running successfully!"
    echo "📊 Health check: http://localhost:8080/health"
    echo "📈 Metrics: http://localhost:8080/metrics"
    echo ""
    echo "🔄 Loading rules via ./scripts/load-rules.sh ..."
    ./scripts/load-rules.sh
    echo "✅ Rules loaded."
    echo "To stop the server, run: kill $SERVER_PID"
    echo "Or use: ./scripts/stop-server.sh"
else
    echo "❌ Server failed to start properly"
    exit 1
fi 