#!/bin/bash

# MM-Rules Client Test Script
# This script tests the client functionality with the backend

BASE_URL="http://localhost:8080"
CLIENT_URL="http://localhost:5173"

echo "🧪 MM-Rules Client Testing"
echo "=========================="

# Check if backend is running
echo "📡 Checking backend health..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Error: Backend is not running on localhost:8080"
    echo "Start the backend first: make run"
    exit 1
fi

echo "✅ Backend is running"

# Check if client is running
echo "📡 Checking client health..."
if ! curl -s http://localhost:5173 > /dev/null; then
    echo "❌ Error: Client is not running on localhost:5173"
    echo "Start the client first: cd mm-rules-client && npm run dev"
    exit 1
fi

echo "✅ Client is running"

# Load rule sets if not already loaded
echo "📋 Loading rule sets..."
make load-rules > /dev/null 2>&1

# Test 1v1 matchmaking
echo -e "\n🧪 Test 1: 1v1 Matchmaking"
echo "Creating 1v1 players..."

# Create two 1v1 players
PLAYER1_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "test1v1_1",
    "game_id": "game-1v1",
    "metadata": {
      "level": 25,
      "region": "us-west",
      "skill_rating": 1500
    }
  }')

PLAYER2_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "test1v1_2",
    "game_id": "game-1v1",
    "metadata": {
      "level": 30,
      "region": "us-west",
      "skill_rating": 1600
    }
  }')

REQUEST_ID1=$(echo $PLAYER1_RESPONSE | jq -r '.request_id')
REQUEST_ID2=$(echo $PLAYER2_RESPONSE | jq -r '.request_id')

echo "✅ Created players: $REQUEST_ID1, $REQUEST_ID2"

# Process matchmaking
echo "🔄 Processing 1v1 matchmaking..."
MATCHMAKING_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/process-matchmaking/game-1v1")
MATCHES_CREATED=$(echo "$MATCHMAKING_RESPONSE" | jq '.matches | length')

if [ "$MATCHES_CREATED" -gt 0 ]; then
    echo "✅ 1v1 matchmaking successful: Created $MATCHES_CREATED match(es)"
    echo "$MATCHMAKING_RESPONSE" | jq '.matches'
else
    echo "❌ 1v1 matchmaking failed: No matches created"
fi

# Test 1v3 matchmaking
echo -e "\n🧪 Test 2: 1v3 Matchmaking"
echo "Creating 1v3 players..."

# Create 1v3 players (1 solo + 3 trio)
SOLO_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "test1v3_solo",
    "game_id": "game-1v3",
    "metadata": {
      "level": 35,
      "team_experience": 5,
      "communication": ["voice", "text"]
    }
  }')

TRIO1_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "test1v3_trio1",
    "game_id": "game-1v3",
    "metadata": {
      "level": 28,
      "team_experience": 3,
      "communication": ["voice"]
    }
  }')

TRIO2_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "test1v3_trio2",
    "game_id": "game-1v3",
    "metadata": {
      "level": 32,
      "team_experience": 4,
      "communication": ["voice", "text"]
    }
  }')

TRIO3_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "test1v3_trio3",
    "game_id": "game-1v3",
    "metadata": {
      "level": 29,
      "team_experience": 2,
      "communication": ["voice"]
    }
  }')

SOLO_ID=$(echo $SOLO_RESPONSE | jq -r '.request_id')
TRIO1_ID=$(echo $TRIO1_RESPONSE | jq -r '.request_id')
TRIO2_ID=$(echo $TRIO2_RESPONSE | jq -r '.request_id')
TRIO3_ID=$(echo $TRIO3_RESPONSE | jq -r '.request_id')

echo "✅ Created players: Solo=$SOLO_ID, Trio=$TRIO1_ID,$TRIO2_ID,$TRIO3_ID"

# Process matchmaking
echo "🔄 Processing 1v3 matchmaking..."
MATCHMAKING_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/process-matchmaking/game-1v3")
MATCHES_CREATED=$(echo "$MATCHMAKING_RESPONSE" | jq '.matches | length')

if [ "$MATCHES_CREATED" -gt 0 ]; then
    echo "✅ 1v3 matchmaking successful: Created $MATCHES_CREATED match(es)"
    echo "$MATCHMAKING_RESPONSE" | jq '.matches'
else
    echo "❌ 1v3 matchmaking failed: No matches created"
fi

# Test client endpoints
echo -e "\n🧪 Test 3: Client API Endpoints"
echo "Testing client can access backend..."

# Test health endpoint
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
if echo "$HEALTH_RESPONSE" | jq -e '.status' > /dev/null; then
    echo "✅ Health endpoint accessible"
else
    echo "❌ Health endpoint not accessible"
fi

# Test stats endpoint
STATS_RESPONSE=$(curl -s "$BASE_URL/api/v1/stats")
if echo "$STATS_RESPONSE" | jq -e '.total_requests' > /dev/null; then
    echo "✅ Stats endpoint accessible"
    echo "📊 Current stats:"
    echo "$STATS_RESPONSE" | jq '.'
else
    echo "❌ Stats endpoint not accessible"
fi

echo -e "\n🎉 Client testing completed!"
echo ""
echo "📋 Next steps:"
echo "1. Open http://localhost:5173 in your browser"
echo "2. Click 'Join 1v1 Competitive' or 'Join 1v3 Team Battle'"
echo "3. Watch the matchmaking process in action"
echo ""
echo "💡 The client will automatically:"
echo "   - Generate random player metadata"
echo "   - Join the selected game mode"
echo "   - Display match information when found"
echo "   - Show team composition and player IDs" 