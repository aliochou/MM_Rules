#!/bin/bash

# MM-Rules Rule Set Test Script
# This script tests the predefined rule sets to ensure they work correctly

BASE_URL="http://localhost:8080/api/v1"
GAME_1V1="game-1v1"
GAME_1V3="game-1v3"

echo "🧪 MM-Rules Rule Set Testing"
echo "============================"

# Check if server is running
echo "📡 Checking server health..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Error: Server is not running on localhost:8080"
    echo "Start the server first: make run"
    exit 1
fi

echo "✅ Server is running"

# Function to test rule set
test_rule_set() {
    local game_id=$1
    local test_name=$2
    local players_json=$3
    local expected_matches=$4
    
    echo -e "\n🧪 Testing $test_name ($game_id)..."
    
    # Create players
    local request_ids=()
    for player_data in $players_json; do
        response=$(curl -s -X POST "$BASE_URL/match-request" \
            -H "Content-Type: application/json" \
            -d "$player_data")
        
        if echo "$response" | jq -e '.error' > /dev/null; then
            echo "❌ Failed to create player: $(echo "$response" | jq -r '.error')"
            return 1
        fi
        
        request_id=$(echo "$response" | jq -r '.request_id')
        request_ids+=("$request_id")
        echo "✅ Created player: $request_id"
    done
    
    # Process matchmaking
    echo "🔄 Processing matchmaking..."
    matchmaking_response=$(curl -s -X POST "$BASE_URL/process-matchmaking/$game_id")
    
    if echo "$matchmaking_response" | jq -e '.error' > /dev/null; then
        echo "❌ Matchmaking failed: $(echo "$matchmaking_response" | jq -r '.error')"
        return 1
    fi
    
    # Check results
    sleep 1
    matches_created=$(echo "$matchmaking_response" | jq '.matches | length')
    
    if [ "$matches_created" -eq "$expected_matches" ]; then
        echo "✅ Test passed: Created $matches_created matches (expected $expected_matches)"
        echo "$matchmaking_response" | jq '.matches'
    else
        echo "❌ Test failed: Created $matches_created matches (expected $expected_matches)"
        echo "$matchmaking_response" | jq .
        return 1
    fi
    
    # Check player statuses
    echo "📊 Checking player statuses..."
    for request_id in "${request_ids[@]}"; do
        status_response=$(curl -s "$BASE_URL/match-status/$request_id")
        status=$(echo "$status_response" | jq -r '.status')
        echo "  Player $request_id: $status"
    done
    
    return 0
}

# Test 1v1 rule set
echo -e "\n📋 Test 1: 1v1 Matchmaking"
test_1v1_players=(
    '{"player_id": "test1v1_1", "game_id": "'$GAME_1V1'", "metadata": {"level": 25, "region": "us-west", "skill_rating": 1500}}'
    '{"player_id": "test1v1_2", "game_id": "'$GAME_1V1'", "metadata": {"level": 30, "region": "us-west", "skill_rating": 1600}}'
)

test_rule_set "$GAME_1V1" "1v1 Matchmaking" "${test_1v1_players[*]}" 1

# Test 1v3 rule set
echo -e "\n📋 Test 2: 1v3 Matchmaking"
test_1v3_players=(
    '{"player_id": "test1v3_solo", "game_id": "'$GAME_1V3'", "metadata": {"level": 35, "team_experience": 5, "communication": ["voice", "text"]}}'
    '{"player_id": "test1v3_trio1", "game_id": "'$GAME_1V3'", "metadata": {"level": 28, "team_experience": 3, "communication": ["voice"]}}'
    '{"player_id": "test1v3_trio2", "game_id": "'$GAME_1V3'", "metadata": {"level": 32, "team_experience": 4, "communication": ["voice", "text"]}}'
    '{"player_id": "test1v3_trio3", "game_id": "'$GAME_1V3'", "metadata": {"level": 29, "team_experience": 2, "communication": ["voice"]}}'
)

test_rule_set "$GAME_1V3" "1v3 Matchmaking" "${test_1v3_players[*]}" 1

# Test rule violations
echo -e "\n📋 Test 3: Rule Violations (1v1)"
test_violation_players=(
    '{"player_id": "violation1", "game_id": "'$GAME_1V1'", "metadata": {"level": 5, "region": "us-west", "skill_rating": 1500}}'
    '{"player_id": "violation2", "game_id": "'$GAME_1V1'", "metadata": {"level": 25, "region": "us-west", "skill_rating": 500}}'
)

echo "🧪 Testing rule violations (should not create matches)..."
for player_data in "${test_violation_players[@]}"; do
    response=$(curl -s -X POST "$BASE_URL/match-request" \
        -H "Content-Type: application/json" \
        -d "$player_data")
    request_id=$(echo "$response" | jq -r '.request_id')
    echo "✅ Created player with violations: $request_id"
done

matchmaking_response=$(curl -s -X POST "$BASE_URL/process-matchmaking/$GAME_1V1")
matches_created=$(echo "$matchmaking_response" | jq '.matches | length')

if [ "$matches_created" -eq 0 ]; then
    echo "✅ Rule violation test passed: No matches created (as expected)"
else
    echo "❌ Rule violation test failed: Created $matches_created matches (expected 0)"
fi

echo -e "\n📊 Final Statistics..."
curl -s "$BASE_URL/stats" | jq .

echo -e "\n🎉 Rule set testing completed!" 