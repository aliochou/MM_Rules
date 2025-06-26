# Rule Management Workflow

This document outlines the complete workflow for adding, editing, and removing rules in the MM-Rules system.

## 🎯 Overview

There are multiple ways to manage rules depending on your needs:

1. **Development/Testing**: Edit YAML files and reload
2. **Production/Real-time**: Use API directly or management script
3. **Team Collaboration**: Use version-controlled configuration files

## 🔄 Workflow Options

### **Option 1: YAML Configuration (Recommended for Development)**

**Best for**: Development, testing, team collaboration, version control

#### Adding a New Rule Set

1. **Create a new configuration file**:
   ```bash
   # Create a template
   ./scripts/manage-rules.sh template my-new-game.yaml
   
   # Edit the template
   vim my-new-game.yaml
   ```

2. **Load the configuration**:
   ```bash
   # Load from YAML file
   ./scripts/manage-rules.sh create my-new-game my-new-game.yaml
   
   # Or reload all configurations
   make load-rules
   ```

3. **Test the new rules**:
   ```bash
   make test-rules
   ```

#### Editing Existing Rules

1. **Edit the YAML file**:
   ```bash
   vim config/game-rules.yaml
   ```

2. **Reload the configuration**:
   ```bash
   make load-rules
   ```

3. **Test the changes**:
   ```bash
   make test-rules
   ```

#### Removing Rules

1. **Remove from YAML file**:
   ```bash
   vim config/game-rules.yaml
   # Delete the game configuration
   ```

2. **Delete from API**:
   ```bash
   ./scripts/manage-rules.sh delete game-id
   ```

### **Option 2: Management Script (Recommended for Production)**

**Best for**: Production environments, real-time updates, automation

#### Adding a New Rule Set

```bash
# Create a template
./scripts/manage-rules.sh template my-game.yaml

# Edit the template
vim my-game.yaml

# Create the game configuration
./scripts/manage-rules.sh create my-game my-game.yaml
```

#### Editing Existing Rules

```bash
# Show current configuration
./scripts/manage-rules.sh show game-1v1

# Edit configuration file
vim updated-game-1v1.yaml

# Update the configuration
./scripts/manage-rules.sh update game-1v1 updated-game-1v1.yaml
```

#### Removing Rules

```bash
# Delete a game configuration
./scripts/manage-rules.sh delete game-1v1
```

### **Option 3: Direct API (Advanced Users)**

**Best for**: Automation, CI/CD, programmatic updates

#### Adding a New Rule Set

```bash
curl -X POST "http://localhost:8080/api/v1/rules/my-game" \
  -H "Content-Type: application/json" \
  -d '{
    "teams": [
      {"name": "Team1", "size": 2},
      {"name": "Team2", "size": 2}
    ],
    "rules": [
      {
        "field": "level",
        "min": 10,
        "max": 50,
        "strict": false,
        "priority": 1,
        "relax_after": 30
      }
    ]
  }'
```

#### Updating Existing Rules

```bash
curl -X PUT "http://localhost:8080/api/v1/rules/game-1v1" \
  -H "Content-Type: application/json" \
  -d '{
    "teams": [...],
    "rules": [...]
  }'
```

#### Removing Rules

```bash
curl -X DELETE "http://localhost:8080/api/v1/rules/game-1v1"
```

## 🛠️ Management Script Commands

The `./scripts/manage-rules.sh` script provides a user-friendly interface:

```bash
# List all game configurations
./scripts/manage-rules.sh list

# Show specific game configuration
./scripts/manage-rules.sh show game-1v1

# Create new game configuration
./scripts/manage-rules.sh create my-game my-config.yaml

# Update existing game configuration
./scripts/manage-rules.sh update game-1v1 updated-config.yaml

# Delete game configuration
./scripts/manage-rules.sh delete game-1v1

# Create template file
./scripts/manage-rules.sh template my-template.yaml

# Show help
./scripts/manage-rules.sh help
```

## 📋 Team Workflow Examples

### **Scenario 1: Adding a New Game Mode**

**Team Member A** (Game Designer):
1. Creates a new rule set specification
2. Creates a YAML configuration file
3. Tests locally
4. Commits to version control

**Team Member B** (DevOps):
1. Reviews the configuration
2. Deploys to staging
3. Runs integration tests
4. Deploys to production

```bash
# Team Member A
./scripts/manage-rules.sh template battle-royale.yaml
# Edit battle-royale.yaml with game designer specs
./scripts/manage-rules.sh create battle-royale battle-royale.yaml
make test-rules
git add battle-royale.yaml
git commit -m "Add battle royale game mode"

# Team Member B
git pull
make load-rules
make test-rules
# Deploy to production
```

### **Scenario 2: Hotfix for Rule Issue**

**Team Member A** (Developer):
1. Identifies rule issue in production
2. Creates hotfix configuration
3. Tests the fix
4. Deploys immediately

```bash
# Create hotfix
./scripts/manage-rules.sh show game-1v1
# Create hotfix config
./scripts/manage-rules.sh update game-1v1 hotfix-config.yaml
# Verify fix
make test-rules
```

### **Scenario 3: A/B Testing Rules**

**Team Member A** (Data Scientist):
1. Creates variant rule sets
2. Deploys both versions
3. Monitors performance
4. Chooses winning variant

```bash
# Create variant A
./scripts/manage-rules.sh create game-1v1-variant-a variant-a.yaml

# Create variant B  
./scripts/manage-rules.sh create game-1v1-variant-b variant-b.yaml

# Monitor and compare
./scripts/manage-rules.sh list
```

## 🔧 Configuration File Format

### YAML Structure

```yaml
game_id: "my-game"
description: "Description of the game"

teams:
  - name: "Team1"
    size: 2
  - name: "Team2"
    size: 2

rules:
  - field: "level"
    min: 10
    max: 50
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
```

### Rule Properties

| Property | Type | Description | Required |
|----------|------|-------------|----------|
| `field` | string | Metadata field to evaluate | Yes |
| `min` | int | Minimum value (for numeric fields) | No |
| `max` | int | Maximum value (for numeric fields) | No |
| `equals` | string | Exact string match | No |
| `contains` | string | Array contains value | No |
| `strict` | bool | If true, rule failure prevents matching | Yes |
| `priority` | int | Higher priority rules evaluated first | Yes |
| `relax_after` | int | Seconds after which rule relaxes | No |
| `description` | string | Human-readable description | No |

## 🚀 Best Practices

### **Development Workflow**

1. **Use YAML files** for development and testing
2. **Version control** all configuration files
3. **Test thoroughly** before deploying
4. **Document changes** in commit messages

### **Production Workflow**

1. **Use management script** for real-time updates
2. **Backup configurations** before major changes
3. **Monitor performance** after rule changes
4. **Have rollback plan** ready

### **Team Collaboration**

1. **Code review** configuration changes
2. **Use feature branches** for new rule sets
3. **Test in staging** before production
4. **Communicate changes** to the team

## 🔍 Troubleshooting

### Common Issues

1. **Rule not working**:
   ```bash
   # Check current configuration
   ./scripts/manage-rules.sh show game-id
   
   # Test with sample data
   make test-rules
   ```

2. **Configuration not loading**:
   ```bash
   # Check server health
   curl http://localhost:8080/health
   
   # Check YAML syntax
   yq eval config/game-rules.yaml
   ```

3. **API errors**:
   ```bash
   # Check server logs
   make logs
   
   # Validate JSON format
   jq . your-config.json
   ```

### Debugging Commands

```bash
# List all configurations
./scripts/manage-rules.sh list

# Show specific configuration
./scripts/manage-rules.sh show game-id

# Check server status
curl http://localhost:8080/health

# View statistics
curl http://localhost:8080/api/v1/stats
```

## 📚 Related Documentation

- [Rule Sets Documentation](rule-sets.md) - Detailed rule set specifications
- [API Documentation](../README.md#api-reference) - Complete API reference
- [Configuration Guide](../README.md#configuration) - System configuration options 