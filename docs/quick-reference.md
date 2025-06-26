# Rule Management Quick Reference

## 🚀 Quick Start

```bash
# Start the server
make run

# Load predefined rules
make load-rules

# Test the rules
make test-rules
```

## 📋 Common Workflows

### **Add New Rule Set**

```bash
# 1. Create template
./scripts/manage-rules.sh template my-game.yaml

# 2. Edit template
vim my-game.yaml

# 3. Create game configuration
./scripts/manage-rules.sh create my-game my-game.yaml

# 4. Test
make test-rules
```

### **Edit Existing Rules**

```bash
# 1. Show current config
./scripts/manage-rules.sh show game-1v1

# 2. Edit configuration
vim updated-game-1v1.yaml

# 3. Update
./scripts/manage-rules.sh update game-1v1 updated-game-1v1.yaml

# 4. Test
make test-rules
```

### **Remove Rules**

```bash
# Delete game configuration
./scripts/manage-rules.sh delete game-1v1
```

## 🛠️ Management Script Commands

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List all games | `./scripts/manage-rules.sh list` |
| `show <id>` | Show game config | `./scripts/manage-rules.sh show game-1v1` |
| `create <id> [file]` | Create new game | `./scripts/manage-rules.sh create my-game config.yaml` |
| `update <id> <file>` | Update game | `./scripts/manage-rules.sh update game-1v1 new-config.yaml` |
| `delete <id>` | Delete game | `./scripts/manage-rules.sh delete game-1v1` |
| `template [file]` | Create template | `./scripts/manage-rules.sh template my-template.yaml` |
| `help` | Show help | `./scripts/manage-rules.sh help` |

## 📁 File Locations

| File | Purpose |
|------|---------|
| `config/game-rules.yaml` | Main configuration file |
| `examples/rules-demo.sh` | Demo script |
| `scripts/manage-rules.sh` | Management script |
| `scripts/test-rules.sh` | Test script |
| `docs/rule-sets.md` | Detailed documentation |

## 🔧 Make Commands

| Command | Description |
|---------|-------------|
| `make load-rules` | Load all rule sets |
| `make test-rules` | Test rule sets |
| `make demo-rules` | Run comprehensive demo |
| `make manage-rules` | Open management script |

## 📝 YAML Template

```yaml
game_id: "my-game"
description: "Description of the game"

teams:
  - name: "Team1"
    size: 1
  - name: "Team2"
    size: 1

rules:
  - field: "level"
    min: 10
    max: 50
    strict: false
    priority: 1
    relax_after: 30
    description: "Player level requirement"
```

## 🚨 Troubleshooting

### Server Not Running
```bash
make run
```

### Rule Not Working
```bash
# Check configuration
./scripts/manage-rules.sh show game-id

# Test rules
make test-rules
```

### Configuration Error
```bash
# Check YAML syntax
yq eval config/game-rules.yaml

# Check server health
curl http://localhost:8080/health
```

## 📞 Need Help?

- **Documentation**: `docs/rule-workflow.md`
- **Examples**: `examples/rules-demo.sh`
- **API Reference**: `README.md`
- **Management Script**: `./scripts/manage-rules.sh help` 