#!/bin/bash

# MM-Rules Rule Management Script
# This script provides an easy interface for managing game rules

BASE_URL="http://localhost:8080/api/v1"
CONFIG_FILE="config/game-rules.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if server is running
check_server() {
    if ! curl -s http://localhost:8080/health > /dev/null; then
        print_status $RED "❌ Error: Server is not running on localhost:8080"
        print_status $YELLOW "Start the server first: make run"
        exit 1
    fi
}

# Function to list all game configurations
list_games() {
    print_status $BLUE "📋 Available Game Configurations:"
    echo "=================================="
    
    # Get list of games from YAML
    if command -v yq &> /dev/null; then
        games=$(yq eval '.games | keys | .[]' "$CONFIG_FILE" 2>/dev/null)
        for game_key in $games; do
            game_id=$(yq eval ".games.$game_key.game_id" "$CONFIG_FILE" 2>/dev/null)
            description=$(yq eval ".games.$game_key.description" "$CONFIG_FILE" 2>/dev/null)
            echo "  - $game_id: $description"
        done
    else
        print_status $YELLOW "⚠️  yq not installed. Install it to see game configurations from YAML."
    fi
    
    # Also try to get from API
    print_status $BLUE "\n📡 Active Game Configurations (from API):"
    echo "============================================="
    response=$(curl -s "$BASE_URL/rules")
    if echo "$response" | jq -e '.error' > /dev/null; then
        print_status $YELLOW "No active configurations found or API not available"
    else
        echo "$response" | jq -r '.games[]?.game_id // empty' 2>/dev/null || print_status $YELLOW "No active configurations found"
    fi
}

# Function to show game configuration
show_game() {
    local game_id=$1
    
    if [ -z "$game_id" ]; then
        print_status $RED "❌ Error: Game ID is required"
        echo "Usage: $0 show <game_id>"
        exit 1
    fi
    
    print_status $BLUE "📋 Configuration for $game_id:"
    echo "=================================="
    
    response=$(curl -s "$BASE_URL/rules/$game_id")
    if echo "$response" | jq -e '.error' > /dev/null; then
        print_status $RED "❌ Error: $(echo "$response" | jq -r '.error')"
        exit 1
    else
        echo "$response" | jq .
    fi
}

# Function to create new game configuration
create_game() {
    local game_id=$1
    local config_file=$2
    
    if [ -z "$game_id" ]; then
        print_status $RED "❌ Error: Game ID is required"
        echo "Usage: $0 create <game_id> [config_file]"
        exit 1
    fi
    
    if [ -n "$config_file" ] && [ -f "$config_file" ]; then
        # Use provided config file
        config_json=$(yq eval -o=json "$config_file" 2>/dev/null)
        if [ $? -ne 0 ]; then
            print_status $RED "❌ Error: Invalid YAML file or yq not installed"
            exit 1
        fi
    else
        # Create default configuration
        config_json='{
          "teams": [
            {"name": "Team1", "size": 1},
            {"name": "Team2", "size": 1}
          ],
          "rules": [
            {
              "field": "level",
              "min": 1,
              "max": 100,
              "strict": false,
              "priority": 1,
              "relax_after": 30
            }
          ]
        }'
        print_status $YELLOW "⚠️  Using default configuration. Provide a config file for custom rules."
    fi
    
    print_status $BLUE "📋 Creating game configuration for $game_id..."
    response=$(curl -s -X POST "$BASE_URL/rules/$game_id" \
        -H "Content-Type: application/json" \
        -d "$config_json")
    
    if echo "$response" | jq -e '.error' > /dev/null; then
        print_status $RED "❌ Failed to create $game_id: $(echo "$response" | jq -r '.error')"
        exit 1
    else
        print_status $GREEN "✅ Successfully created $game_id"
        echo "$response" | jq .
    fi
}

# Function to update game configuration
update_game() {
    local game_id=$1
    local config_file=$2
    
    if [ -z "$game_id" ] || [ -z "$config_file" ]; then
        print_status $RED "❌ Error: Game ID and config file are required"
        echo "Usage: $0 update <game_id> <config_file>"
        exit 1
    fi
    
    if [ ! -f "$config_file" ]; then
        print_status $RED "❌ Error: Config file not found: $config_file"
        exit 1
    fi
    
    config_json=$(yq eval -o=json "$config_file" 2>/dev/null)
    if [ $? -ne 0 ]; then
        print_status $RED "❌ Error: Invalid YAML file or yq not installed"
        exit 1
    fi
    
    print_status $BLUE "📋 Updating game configuration for $game_id..."
    response=$(curl -s -X PUT "$BASE_URL/rules/$game_id" \
        -H "Content-Type: application/json" \
        -d "$config_json")
    
    if echo "$response" | jq -e '.error' > /dev/null; then
        print_status $RED "❌ Failed to update $game_id: $(echo "$response" | jq -r '.error')"
        exit 1
    else
        print_status $GREEN "✅ Successfully updated $game_id"
        echo "$response" | jq .
    fi
}

# Function to delete game configuration
delete_game() {
    local game_id=$1
    
    if [ -z "$game_id" ]; then
        print_status $RED "❌ Error: Game ID is required"
        echo "Usage: $0 delete <game_id>"
        exit 1
    fi
    
    print_status $YELLOW "⚠️  Are you sure you want to delete $game_id? (y/N)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_status $BLUE "🗑️  Deleting game configuration for $game_id..."
        response=$(curl -s -X DELETE "$BASE_URL/rules/$game_id")
        
        if echo "$response" | jq -e '.error' > /dev/null; then
            print_status $RED "❌ Failed to delete $game_id: $(echo "$response" | jq -r '.error')"
            exit 1
        else
            print_status $GREEN "✅ Successfully deleted $game_id"
        fi
    else
        print_status $BLUE "❌ Deletion cancelled"
    fi
}

# Function to create a template configuration file
create_template() {
    local template_name=$1
    
    if [ -z "$template_name" ]; then
        template_name="new-game-config.yaml"
    fi
    
    cat > "$template_name" << 'EOF'
# Game Configuration Template
# Copy this template and modify it for your new game

game_id: "my-new-game"
description: "Description of your new game"

teams:
  - name: "Team1"
    size: 1
  - name: "Team2"
    size: 1

rules:
  - field: "level"
    min: 1
    max: 100
    strict: false
    priority: 1
    relax_after: 30
    description: "Player level requirement"
  
  - field: "region"
    equals: "us-west"
    strict: false
    priority: 2
    relax_after: 60
    description: "Region preference"
  
  - field: "skill_rating"
    min: 1000
    max: 2000
    strict: true
    priority: 3
    description: "Skill rating requirement (strict)"
EOF

    print_status $GREEN "✅ Created template: $template_name"
    print_status $BLUE "💡 Edit this file and use: $0 create <game_id> $template_name"
}

# Function to show help
show_help() {
    echo "MM-Rules Rule Management Script"
    echo "==============================="
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  list                    - List all game configurations"
    echo "  show <game_id>          - Show configuration for a specific game"
    echo "  create <game_id> [file] - Create new game configuration"
    echo "  update <game_id> <file> - Update existing game configuration"
    echo "  delete <game_id>        - Delete game configuration"
    echo "  template [filename]     - Create a template configuration file"
    echo "  help                    - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 list"
    echo "  $0 show game-1v1"
    echo "  $0 create my-game my-config.yaml"
    echo "  $0 update game-1v1 updated-config.yaml"
    echo "  $0 delete old-game"
    echo "  $0 template my-new-game.yaml"
    echo ""
    echo "Configuration Files:"
    echo "  - Use YAML format for configuration files"
    echo "  - See config/game-rules.yaml for examples"
    echo "  - Use 'template' command to create a starting template"
}

# Main script logic
check_server

case "${1:-help}" in
    "list")
        list_games
        ;;
    "show")
        show_game "$2"
        ;;
    "create")
        create_game "$2" "$3"
        ;;
    "update")
        update_game "$2" "$3"
        ;;
    "delete")
        delete_game "$2"
        ;;
    "template")
        create_template "$2"
        ;;
    "help"|*)
        show_help
        ;;
esac 