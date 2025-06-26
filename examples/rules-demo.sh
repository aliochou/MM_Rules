#!/bin/bash

# MM-Rules Matchmaking API Demo Script
# This script demonstrates 1v1 and 1v3 matchmaking rule configurations

BASE_URL="http://localhost:8080/api/v1"
GAME_1V1="game-1v1"
GAME_1V3="game-1v3"

echo "🚀 MM-Rules Matchmaking API - 1v1 and 1v3 Rules Demo"
echo "====================================================="

# Check if server is running
echo "📡 Checking server health..."
curl -s http://localhost:8080/health | jq .

echo -e "\n📋 Step 1: Creating 1v1 game configuration..."
curl -X POST "$BASE_URL/rules/$GAME_1V1" \
  -H "Content-Type: application/json" \
  -d '{
    "teams": [
      { "name": "Player1", "size": 1 },
      { "name": "Player2", "size": 1 }
    ],
    "rules": [
      {
        "field": "level",
        "min": 10,
        "max": 50,
        "strict": false,
        "priority": 1,
        "relax_after": 30
      },
      {
        "field": "region",
        "equals": "us-west",
        "strict": false,
        "priority": 2,
        "relax_after": 60
      },
      {
        "field": "skill_rating",
        "min": 1000,
        "max": 2000,
        "strict": true,
        "priority": 3
      }
    ]
  }' | jq .

echo -e "\n📋 Step 2: Creating 1v3 game configuration..."
curl -X POST "$BASE_URL/rules/$GAME_1V3" \
  -H "Content-Type: application/json" \
  -d '{
    "teams": [
      { "name": "Solo", "size": 1 },
      { "name": "Trio", "size": 3 }
    ],
    "rules": [
      {
        "field": "level",
        "min": 15,
        "max": 60,
        "strict": false,
        "priority": 1,
        "relax_after": 45
      },
      {
        "field": "team_experience",
        "min": 1,
        "strict": false,
        "priority": 2,
        "relax_after": 90
      },
      {
        "field": "communication",
        "contains": "voice",
        "strict": false,
        "priority": 3,
        "relax_after": 120
      }
    ]
  }' | jq .

echo -e "\n👥 Step 3: Creating 1v1 match requests..."

# Create 1v1 players
echo "Creating 1v1 player 1..."
RESPONSE_1V1_1=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "1v1_player1",
    "game_id": "'$GAME_1V1'",
    "metadata": {
      "level": 25,
      "region": "us-west",
      "skill_rating": 1500,
      "preferred_role": "attacker"
    }
  }')
REQUEST_1V1_1=$(echo $RESPONSE_1V1_1 | jq -r '.request_id')
echo $RESPONSE_1V1_1 | jq .

echo "Creating 1v1 player 2..."
RESPONSE_1V1_2=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "1v1_player2",
    "game_id": "'$GAME_1V1'",
    "metadata": {
      "level": 30,
      "region": "us-west",
      "skill_rating": 1600,
      "preferred_role": "defender"
    }
  }')
REQUEST_1V1_2=$(echo $RESPONSE_1V1_2 | jq -r '.request_id')
echo $RESPONSE_1V1_2 | jq .

echo -e "\n👥 Step 4: Creating 1v3 match requests..."

# Create 1v3 players (1 solo + 3 trio)
echo "Creating 1v3 solo player..."
RESPONSE_1V3_SOLO=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "1v3_solo",
    "game_id": "'$GAME_1V3'",
    "metadata": {
      "level": 35,
      "team_experience": 5,
      "communication": ["voice", "text"],
      "preferred_role": "leader"
    }
  }')
REQUEST_1V3_SOLO=$(echo $RESPONSE_1V3_SOLO | jq -r '.request_id')
echo $RESPONSE_1V3_SOLO | jq .

echo "Creating 1v3 trio player 1..."
RESPONSE_1V3_TRIO1=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "1v3_trio1",
    "game_id": "'$GAME_1V3'",
    "metadata": {
      "level": 28,
      "team_experience": 3,
      "communication": ["voice"],
      "preferred_role": "support"
    }
  }')
REQUEST_1V3_TRIO1=$(echo $RESPONSE_1V3_TRIO1 | jq -r '.request_id')
echo $RESPONSE_1V3_TRIO1 | jq .

echo "Creating 1v3 trio player 2..."
RESPONSE_1V3_TRIO2=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "1v3_trio2",
    "game_id": "'$GAME_1V3'",
    "metadata": {
      "level": 32,
      "team_experience": 4,
      "communication": ["voice", "text"],
      "preferred_role": "attacker"
    }
  }')
REQUEST_1V3_TRIO2=$(echo $RESPONSE_1V3_TRIO2 | jq -r '.request_id')
echo $RESPONSE_1V3_TRIO2 | jq .

echo "Creating 1v3 trio player 3..."
RESPONSE_1V3_TRIO3=$(curl -s -X POST "$BASE_URL/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "1v3_trio3",
    "game_id": "'$GAME_1V3'",
    "metadata": {
      "level": 29,
      "team_experience": 2,
      "communication": ["voice"],
      "preferred_role": "defender"
    }
  }')
REQUEST_1V3_TRIO3=$(echo $RESPONSE_1V3_TRIO3 | jq -r '.request_id')
echo $RESPONSE_1V3_TRIO3 | jq .

echo -e "\n⏳ Step 5: Checking initial match status..."
echo "1v1 Player 1 status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V1_1" | jq .

echo "1v1 Player 2 status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V1_2" | jq .

echo "1v3 Solo status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V3_SOLO" | jq .

echo -e "\n🔄 Step 6: Processing 1v1 matchmaking..."
curl -X POST "$BASE_URL/process-matchmaking/$GAME_1V1" | jq .

echo -e "\n🔄 Step 7: Processing 1v3 matchmaking..."
curl -X POST "$BASE_URL/process-matchmaking/$GAME_1V3" | jq .

# Add a 1 second delay to allow status updates
sleep 1

echo -e "\n✅ Step 8: Checking final match status..."
echo "1v1 Player 1 status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V1_1" | jq .

echo "1v1 Player 2 status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V1_2" | jq .

echo "1v3 Solo status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V3_SOLO" | jq .

echo "1v3 Trio Player 1 status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V3_TRIO1" | jq .

echo "1v3 Trio Player 2 status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V3_TRIO2" | jq .

echo "1v3 Trio Player 3 status:"
curl -s "$BASE_URL/match-status/$REQUEST_1V3_TRIO3" | jq .

echo -e "\n📊 Step 9: Getting system statistics..."
curl -s "$BASE_URL/stats" | jq .

echo -e "\n🎉 Rules Demo completed!"
echo "Check the responses above to see how different rule sets work for 1v1 and 1v3 matchmaking." 