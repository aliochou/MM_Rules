# Developer Onboarding Guide

Welcome to the MM-Rules matchmaking backend! This guide will help you get up and running quickly.

## 🚀 Quick Start (5 minutes)

### 1. Prerequisites Check

```bash
# Check Go version (needs 1.21+)
go version

# Check if Redis is available
redis-cli ping

# Check if jq and yq are installed
jq --version
yq --version
```

### 2. Setup Development Environment

```bash
# Clone and setup
git clone <repository-url>
cd MM_Rules

# Install dependencies
go mod download

# Start Redis (if not running)
make redis

# Start the backend (loads rules automatically)
./scripts/start-server.sh
# or
make server-start
```

### 3. Load and Test Rules

# If you started the backend with './scripts/start-server.sh' or 'make server-start', rules are loaded automatically.
# If you used 'make run', load rules manually:
```bash
make load-rules
make test-rules
```

### 4. Verify Everything Works

```bash
# Check server health
curl http://localhost:8080/health

# List available rule sets
./scripts/manage-rules.sh list

# Run the demo
make demo-rules
```

## 📚 Understanding the System

### Architecture Overview

MM-Rules is a rule-driven matchmaking system with these key components:

- **API Layer** (`internal/api/`): HTTP handlers and routing
- **Rule Engine** (`internal/engine/`): Processes matchmaking rules
- **Matchmaker** (`internal/matchmaker/`): Core matchmaking logic
- **Storage** (`internal/storage/`): Redis-based data persistence
- **Allocation** (`internal/allocation/`): Game session allocation

### Key Concepts

1. **Game Configurations**: Define teams and rules for matchmaking
2. **Match Requests**: Player requests to join matchmaking
3. **Rules**: Conditions that players must meet to be matched
4. **Rule Relaxation**: Rules that become less strict over time
5. **Team Composition**: How players are grouped into teams

## 🛠️ Development Workflow

### Daily Development

```bash
# Start development environment
make redis
./scripts/start-server.sh   # or: make server-start

# If you used 'make run', load rules manually:
# make load-rules

# Make changes to code
# ...

# Test your changes
make test
make test-rules
```

### Adding New Features

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow Go coding standards
   - Add tests for new functionality
   - Update documentation if needed

3. **Test thoroughly**
   ```bash
   go test ./...
   make test-rules
   ```

4. **Commit and push**
   ```bash
   git add .
   git commit -m "Add feature: description"
   git push origin feature/your-feature-name
   ```

### Rule Development

#### Creating New Rule Sets

```bash
# Create a template
./scripts/manage-rules.sh template my-new-game.yaml

# Edit the template
vim my-new-game.yaml

# Create the game configuration
./scripts/manage-rules.sh create my-new-game my-new-game.yaml

# Test the rules
make test-rules
```

#### Modifying Existing Rules

```bash
# Show current configuration
./scripts/manage-rules.sh show game-1v1

# Create updated configuration
./scripts/manage-rules.sh template updated-game-1v1.yaml

# Edit the configuration
vim updated-game-1v1.yaml

# Update the game
./scripts/manage-rules.sh update game-1v1 updated-game-1v1.yaml

# Test the changes
make test-rules
```

## 🔧 Available Tools

### Make Commands

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `make run` | Start server | Daily development |
| `make build` | Build binary | Before deployment |
| `make test` | Run tests | After code changes |
| `make test-coverage` | Test with coverage | Before PR |
| `make load-rules` | Load rule sets | After rule changes |
| `make test-rules` | Test rule sets | After rule changes |
| `make demo-rules` | Run demo | Testing features |
| `make redis` | Start Redis | Development setup |
| `make monitoring` | Start monitoring | Debugging |

### Management Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `./scripts/manage-rules.sh` | Rule management | `./scripts/manage-rules.sh help` |
| `./scripts/test-rules.sh` | Rule testing | `make test-rules` |
| `./scripts/load-rules.sh` | Load configurations | `make load-rules` |

### API Testing

```bash
# Health check
curl http://localhost:8080/health

# Create match request
curl -X POST "http://localhost:8080/api/v1/match-request" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": "test-player",
    "game_id": "game-1v1",
    "metadata": {
      "level": 25,
      "region": "us-west",
      "skill_rating": 1500
    }
  }'

# Process matchmaking
curl -X POST "http://localhost:8080/api/v1/process-matchmaking/game-1v1"

# Get statistics
curl http://localhost:8080/api/v1/stats
```

## 📁 Project Structure

```
.
├── cmd/server/           # Application entry point
├── internal/             # Core application code
│   ├── api/             # HTTP handlers and routing
│   ├── allocation/      # Session allocation logic
│   ├── engine/          # Rule processing engine
│   ├── matchmaker/      # Core matchmaking logic
│   ├── models/          # Data structures
│   └── storage/         # Redis storage layer
├── config/              # Configuration files
│   ├── config.yaml      # Main system configuration
│   └── game-rules.yaml  # Predefined rule sets
├── scripts/             # Management scripts
├── examples/            # Example scripts
├── docs/                # Documentation
├── k8s/                 # Kubernetes manifests
└── tests/               # Test files
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/engine

# Run tests with coverage
make test-coverage

# Test rule sets
make test-rules
```

### Writing Tests

Follow these patterns for writing tests:

```go
func TestYourFunction(t *testing.T) {
    // Arrange
    input := "test input"
    
    // Act
    result := YourFunction(input)
    
    // Assert
    if result != "expected output" {
        t.Errorf("Expected %s, got %s", "expected output", result)
    }
}
```

### Testing Rules

Use the test scripts to validate rule behavior:

```bash
# Test predefined rules
make test-rules

# Test specific game
./scripts/manage-rules.sh show game-1v1
```

## 🔍 Debugging

### Common Issues

1. **Server won't start**
   ```bash
   # Check if Redis is running
   redis-cli ping
   
   # Check server logs
   make logs
   ```

2. **Rules not working**
   ```bash
   # Check rule configuration
   ./scripts/manage-rules.sh show game-id
   
   # Test rules
   make test-rules
   ```

3. **API errors**
   ```bash
   # Check server health
   curl http://localhost:8080/health
   
   # Check server logs
   make logs
   ```

### Debugging Tools

```bash
# View server logs
make logs

# Check Redis data
redis-cli keys "*"

# Monitor API requests
curl -v http://localhost:8080/api/v1/stats

# Check rule configurations
./scripts/manage-rules.sh list
```

## 📊 Monitoring

### Development Monitoring

```bash
# Start monitoring stack
make monitoring

# View metrics
curl http://localhost:8080/metrics

# Open Grafana (macOS)
make grafana-open
```

### Key Metrics

- `matchmaking_requests_total`: Total match requests
- `matchmaking_matches_total`: Total matches created
- `matchmaking_queue_size`: Current queue size
- `matchmaking_processing_duration`: Processing time

## 🚀 Deployment

### Local Testing

```bash
# Build and run with Docker
make docker-build
make docker-run
```

### Production Deployment

```bash
# Deploy to Kubernetes
make k8s-deploy

# Check deployment status
kubectl get pods -l app=mm-rules-matchmaking
```

## 📝 Documentation

### Key Documents

- [README.md](../README.md) - Main project documentation
- [Rule Sets Documentation](rule-sets.md) - Rule set specifications
- [Rule Management Workflow](rule-workflow.md) - Rule management guide
- [Quick Reference](quick-reference.md) - Command reference

### Contributing to Documentation

1. Update relevant documentation when adding features
2. Include examples in documentation
3. Keep the quick reference updated
4. Add comments to complex code sections

## 🤝 Team Collaboration

### Code Review Process

1. Create feature branch
2. Make changes and test
3. Create pull request
4. Request review from team
5. Address feedback
6. Merge after approval

### Communication

- Use clear commit messages
- Document breaking changes
- Update team on major changes
- Ask questions in team chat

## 🎯 Next Steps

1. **Explore the codebase**: Look at the main packages in `internal/`
2. **Try the demos**: Run `make demo-rules` to see the system in action
3. **Create your first rule**: Use the management scripts to create a custom rule set
4. **Join team discussions**: Participate in code reviews and planning

## 📞 Getting Help

- **Documentation**: Check the docs folder
- **Issues**: Open GitHub issues for bugs
- **Team Chat**: Ask questions in team communication channels
- **Code Review**: Get help during PR reviews

Welcome to the team! 🎉 