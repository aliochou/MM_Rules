#!/bin/bash

# MM-Rules Matchmaking API Demo Script
# This script demonstrates the basic usage of the matchmaking API

BASE_URL="http://localhost:8080/api/v1"
GAME_ID="demo-game"

echo "🚀 MM-Rules Matchmaking API Demo"
echo "=================================="

# Check if server is running
echo "📡 Checking server health..."
curl -s http://localhost:8080/health | jq .

echo -e "\n📋 Step 1: Creating game configuration..."
curl -X POST "$BASE_URL/rules/$GAME_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "teams": [
      { "name": "PlayerA", "size": 1 },
      { "name": "PlayerB", "size": 1 }
    ],
    "rules": [
      {
        "field": "level",
        "min": 1,
        "strict": false,
        "priority": 1
      }
    ]
  }' | jq .

echo -e "\n👥 Step 2: Creating match requests..."

# Create several match requests
echo "Creating player 1..."
RESPONSE1=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "player1",
    "game_id": "'$GAME_ID'",
    "metadata": {
      "level": 25,
      "inventory": ["itemA", "itemB"],
      "region": "us-west"
    }
  }')
REQUEST_ID1=$(echo $RESPONSE1 | jq -r '.request_id')
echo $RESPONSE1 | jq .

echo "Creating player 2..."
RESPONSE2=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "player2",
    "game_id": "'$GAME_ID'",
    "metadata": {
      "level": 30,
      "inventory": ["itemA"],
      "region": "us-west"
    }
  }')
REQUEST_ID2=$(echo $RESPONSE2 | jq -r '.request_id')
echo $RESPONSE2 | jq .

echo "Creating player 3..."
RESPONSE3=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "player3",
    "game_id": "'$GAME_ID'",
    "metadata": {
      "level": 22,
      "inventory": ["itemB"],
      "region": "us-east"
    }
  }')
REQUEST_ID3=$(echo $RESPONSE3 | jq -r '.request_id')
echo $RESPONSE3 | jq .

echo "Creating player 4..."
RESPONSE4=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "player4",
    "game_id": "'$GAME_ID'",
    "metadata": {
      "level": 28,
      "inventory": ["itemA", "itemC"],
      "region": "us-west"
    }
  }')
REQUEST_ID4=$(echo $RESPONSE4 | jq -r '.request_id')
echo $RESPONSE4 | jq .

echo -e "\n⏳ Step 3: Checking initial match status..."
echo "Player 1 status:"
curl -s "$BASE_URL/match-status/$REQUEST_ID1" | jq .

echo "Player 2 status:"
curl -s "$BASE_URL/match-status/$REQUEST_ID2" | jq .

echo -e "\n🔄 Step 4: Processing matchmaking..."
curl -X POST "$BASE_URL/process-matchmaking/$GAME_ID" | jq .

# Add a 1 second delay to allow status updates
sleep 1

echo -e "\n✅ Step 5: Checking final match status..."
echo "Player 1 status:"
curl -s "$BASE_URL/match-status/$REQUEST_ID1" | jq .

echo "Player 2 status:"
curl -s "$BASE_URL/match-status/$REQUEST_ID2" | jq .

echo "Player 3 status:"
curl -s "$BASE_URL/match-status/$REQUEST_ID3" | jq .

echo "Player 4 status:"
curl -s "$BASE_URL/match-status/$REQUEST_ID4" | jq .

echo -e "\n📊 Step 6: Getting system statistics..."
curl -s "$BASE_URL/stats" | jq .

echo -e "\n🎉 Demo completed!"
echo "Check the responses above to see how the matchmaking system works." 