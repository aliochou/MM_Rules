#!/bin/bash

set -e

API_URL="http://localhost:8080/api/v1"
GAME_ID="game-1v3"

# Generate 4 unique player IDs
PLAYER_IDS=(
  "test-batch-1-$(date +%s%N | cut -b1-13)"
  "test-batch-2-$(date +%s%N | cut -b1-13)"
  "test-batch-3-$(date +%s%N | cut -b1-13)"
  "test-batch-4-$(date +%s%N | cut -b1-13)"
)

# Metadata compatible with 1v3 rules
METADATA='{"level": 30, "team_experience": 3, "communication": ["voice", "text"], "preferred_role": "support"}'

REQUEST_IDS=()
echo "Submitting 4 match requests for $GAME_ID..."
for PID in "${PLAYER_IDS[@]}"; do
  REQ=$(curl -s -X POST "$API_URL/match-request" \
    -H "Content-Type: application/json" \
    -d "{\"player_id\": \"$PID\", \"game_id\": \"$GAME_ID\", \"metadata\": $METADATA}")
  REQ_ID=$(echo "$REQ" | grep -o '"request_id":"[^"]*' | cut -d'"' -f4)
  echo "  Player $PID -> request_id: $REQ_ID"
  REQUEST_IDS+=("$REQ_ID")
done

echo "\nTriggering matchmaking for $GAME_ID..."
MATCHES=$(curl -s -X POST "$API_URL/process-matchmaking/$GAME_ID")
echo "\nMatchmaking result:"
echo "$MATCHES" | jq . || echo "$MATCHES"

echo "\nChecking match status for each request:"
for REQ_ID in "${REQUEST_IDS[@]}"; do
  STATUS=$(curl -s "$API_URL/match-status/$REQ_ID")
  echo "Request $REQ_ID: $STATUS"
done 