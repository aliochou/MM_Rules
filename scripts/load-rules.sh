#!/bin/bash

# MM-Rules Configuration Loader
# This script loads game rule configurations from config/game-rules.yaml and applies them to the API

BASE_URL="http://localhost:8080/api/v1"
CONFIG_FILE="config/game-rules.yaml"

# Check if yq is installed (for YAML parsing)
if ! command -v yq &> /dev/null; then
    echo "❌ Error: yq is required but not installed."
    echo "Install yq: https://github.com/mikefarah/yq#install"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "❌ Error: jq is required but not installed."
    echo "Install jq: https://stedolan.github.io/jq/download/"
    exit 1
fi

echo "🚀 MM-Rules Configuration Loader"
echo "================================"

# Check if server is running
echo "📡 Checking server health..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Error: Server is not running on localhost:8080"
    echo "Start the server first: make run"
    exit 1
fi

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Error: Configuration file not found: $CONFIG_FILE"
    exit 1
fi

echo "✅ Server is running"
echo "✅ Configuration file found"

# Function to create game configuration
create_game_config() {
    local game_id=$1
    local config_json=$2
    
    echo "📋 Creating configuration for $game_id..."
    response=$(curl -s -X POST "$BASE_URL/rules/$game_id" \
        -H "Content-Type: application/json" \
        -d "$config_json")
    
    if echo "$response" | jq -e '.error' > /dev/null; then
        echo "❌ Failed to create $game_id: $(echo "$response" | jq -r '.error')"
        return 1
    else
        echo "✅ Successfully created $game_id"
        echo "$response" | jq .
        return 0
    fi
}

# Load and process each game configuration
echo -e "\n🔄 Loading game configurations..."

# Get list of games from YAML
games=$(yq eval '.games | keys | .[]' "$CONFIG_FILE")

for game_key in $games; do
    echo -e "\n--- Processing $game_key ---"
    
    # Extract game configuration
    game_id=$(yq eval ".games.$game_key.game_id" "$CONFIG_FILE")
    description=$(yq eval ".games.$game_key.description" "$CONFIG_FILE")
    
    echo "Game ID: $game_id"
    echo "Description: $description"
    
    # Build JSON configuration
    config_json=$(yq eval -o=json ".games.$game_key" "$CONFIG_FILE")
    
    # Create the game configuration
    if create_game_config "$game_id" "$config_json"; then
        echo "✅ $game_key configuration loaded successfully"
    else
        echo "❌ Failed to load $game_key configuration"
    fi
done

echo -e "\n📊 Summary:"
echo "============="

# List all configured games
echo "Configured games:"
for game_key in $games; do
    game_id=$(yq eval ".games.$game_key.game_id" "$CONFIG_FILE")
    description=$(yq eval ".games.$game_key.description" "$CONFIG_FILE")
    echo "  - $game_id: $description"
done

echo -e "\n🎉 Configuration loading completed!"
echo "You can now use these game IDs for matchmaking:"
for game_key in $games; do
    game_id=$(yq eval ".games.$game_key.game_id" "$CONFIG_FILE")
    echo "  - $game_id"
done

echo -e "\n💡 Next steps:"
echo "1. Run the demo script: ./examples/rules-demo.sh"
echo "2. Or create match requests manually using the game IDs above" 