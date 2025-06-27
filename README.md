# MM-Rules Matchmaking Backend

A scalable, rule-driven matchmaking backend for multiplayer games built in Go. This system provides a custom rule-processing engine with JSON-based configuration, designed for Kubernetes deployment and horizontal scalability.

## Features

- **REST API**: Full HTTP API for match requests, game configuration, and status queries
- **JSON Rule Engine**: Configurable matchmaking rules with support for relaxation and strict matching
- **Team Composition**: Support for arbitrary team sizes and compositions (1v1, 2v2, 1v99, etc.)
- **Rule Relaxation**: Automatic rule relaxation after configurable time periods
- **Session Allocation**: Integration with external allocation services via webhooks
- **Redis Storage**: Fast, scalable storage for match requests and game configurations
- **Kubernetes Ready**: Designed for cloud-native deployment with health checks and metrics
- **Observability**: Prometheus metrics, structured logging, and health endpoints
- **Predefined Rule Sets**: Ready-to-use configurations for common matchmaking scenarios
- **Rule Management Tools**: Scripts and utilities for easy rule configuration and testing

## Quick Start

### Prerequisites

- Go 1.21+
- Redis 6.0+
- Docker (optional)
- Kubernetes cluster (optional)
- `jq` and `yq` (for rule management scripts)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd MM_Rules
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install optional tools** (for rule management)
   ```bash
   # macOS
   brew install jq yq
   
   # Ubuntu/Debian
   sudo apt-get install jq
   wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq
   chmod +x /usr/bin/yq
   ```

### Local Development

1. **Start Redis**
   ```bash
   # Using Docker
   docker run -d -p 6379:6379 redis:7-alpine
   
   # Or using make command
   make redis
   ```

2. **Run the server (loads rules automatically)**
   ```bash
   ./scripts/start-server.sh
   # or
   make server-start
   
   # If you use 'make run' or 'go run ...', you must load rules manually:
   # make load-rules
   ```

3. **Test the system**
   ```bash
   make test-rules
   ```

4. **Run the demo**
   ```bash
   make demo-rules
   ```

The server will start on `http://localhost:8080`

### First Steps

1. **Check server health**
   ```bash
   curl http://localhost:8080/health
   ```

2. **View available rule sets**
   ```bash
   ./scripts/manage-rules.sh list
   ```

3. **Create a match request**
   ```bash
   curl -X POST "http://localhost:8080/api/v1/match-request" \
     -H "Content-Type: application/json" \
     -d '{
       "player_id": "player1",
       "game_id": "game-1v1",
       "metadata": {
         "level": 25,
         "region": "us-west",
         "skill_rating": 1500
       }
     }'
   ```

## Rule Management

### Overview

The MM-Rules system provides multiple ways to manage matchmaking rules:

- **Predefined Rule Sets**: Ready-to-use configurations for common scenarios
- **YAML Configuration**: Version-controlled rule definitions
- **Management Scripts**: Easy-to-use command-line tools
- **Direct API**: Programmatic rule management

### Quick Rule Management

```bash
# List all game configurations
./scripts/manage-rules.sh list

# Show specific game configuration
./scripts/manage-rules.sh show game-1v1

# Create new game configuration
./scripts/manage-rules.sh template my-game.yaml
# Edit my-game.yaml
./scripts/manage-rules.sh create my-game my-game.yaml

# Update existing configuration
./scripts/manage-rules.sh update game-1v1 updated-config.yaml

# Delete game configuration
./scripts/manage-rules.sh delete game-1v1
```

### Available Commands

| Command | Description |
|---------|-------------|
| `make load-rules` | Load all predefined rule sets |
| `make test-rules` | Test rule sets with sample data |
| `make demo-rules` | Run comprehensive demo |
| `make manage-rules` | Open rule management interface |

### Rule Management Workflow

For detailed information about rule management workflows, see:
- [Rule Management Workflow](docs/rule-workflow.md) - Comprehensive guide
- [Quick Reference](docs/quick-reference.md) - Command reference
- [Rule Sets Documentation](docs/rule-sets.md) - Detailed specifications

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Game Client   │    │   Game Client   │    │   Game Client   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │    MM-Rules API Server    │
                    │  (Load Balanced, 3x)      │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │        Redis Cluster      │
                    │   (Match Requests &       │
                    │    Game Configs)          │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │    Allocation Service     │
                    │  (Agones, Unity, etc.)    │
                    └───────────────────────────┘
```

## API Reference

### Match Requests

#### Create Match Request
```http
POST /api/v1/match-request
Content-Type: application/json

{
  "player_id": "abc123",
  "game_id": "my-cool-game",
  "metadata": {
    "level": 25,
    "inventory": ["itemA", "itemB"],
    "region": "us-west"
  }
}
```

**Response:**
```json
{
  "request_id": "uuid-here",
  "status": "pending"
}
```

#### Get Match Status
```http
GET /api/v1/match-status/{request_id}
```

**Response:**
```json
{
  "status": "matched",
  "team": "Solo",
  "session": {
    "ip": "12.34.56.78",
    "port": 7777,
    "id": "session-123"
  }
}
```

### Game Configuration

#### Upload Game Rules
```http
POST /api/v1/rules/{game_id}
Content-Type: application/json

{
  "teams": [
    { "name": "Solo", "size": 1 },
    { "name": "Duo", "size": 2 },
    { "name": "Squad", "size": 4 }
  ],
  "rules": [
    {
      "field": "level",
      "min": 20,
      "strict": true,
      "priority": 10
    },
    {
      "field": "inventory",
      "contains": "itemA",
      "relax_after": 10,
      "priority": 5
    },
    {
      "field": "region",
      "equals": "us-west",
      "strict": false,
      "priority": 1
    }
  ]
}
```

### Matchmaking Processing

#### Process Matchmaking
```http
POST /api/v1/process-matchmaking/{game_id}
```

**Response:**
```json
{
  "message": "Matchmaking processed successfully",
  "matches": [
    {
      "match_id": "match-123",
      "team_name": "Duo",
      "players": ["player1", "player2"],
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### Health & Metrics

#### Health Check
```http
GET /health
```

#### Metrics (Prometheus)
```http
GET /metrics
```

#### Statistics
```http
GET /api/v1/stats
```

## Configuration

The system can be configured via environment variables or a YAML config file:

### Environment Variables

- `MM_RULES_SERVER_PORT`: Server port (default: 8080)
- `MM_RULES_REDIS_ADDR`: Redis address (default: localhost:6379)
- `MM_RULES_REDIS_PASSWORD`: Redis password
- `MM_RULES_REDIS_DB`: Redis database (default: 0)
- `MM_RULES_ALLOCATION_WEBHOOK_URL`: Allocation service webhook URL
- `MM_RULES_LOG_LEVEL`: Log level (debug, info, warn, error)

### Config File

Create `config/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug

redis:
  addr: localhost:6379
  password: ""
  db: 0

allocation:
  webhook_url: http://localhost:8081/allocate

log:
  level: info

matchmaking:
  process_interval: 5
  max_wait_time: 300
  allocation:
    max_retries: 3
    retry_delay: 1s
```

## Rule Engine

The rule engine supports various types of rules:

### Rule Types

1. **Numeric Range**: `min` and `max` values
2. **String Matching**: `equals` for exact matches
3. **Array Contains**: `contains` for array membership
4. **Relaxation**: `relax_after` seconds to automatically relax rules

### Rule Properties

- `field`: The metadata field to evaluate
- `strict`: If true, rule failure prevents matching
- `priority`: Higher priority rules are evaluated first
- `relax_after`: Seconds after which the rule is relaxed

### Example Rules

```json
[
  {
    "field": "level",
    "min": 20,
    "max": 50,
    "strict": true,
    "priority": 10
  },
  {
    "field": "inventory",
    "contains": "premium_item",
    "relax_after": 30,
    "priority": 5
  },
  {
    "field": "region",
    "equals": "us-west",
    "strict": false,
    "priority": 1
  }
]
```

## Predefined Rule Sets

The system comes with two predefined rule sets for common matchmaking scenarios:

### Rule Set #1: 1v1 Matchmaking (`game-1v1`)

**Description**: Competitive 1v1 matchmaking with skill-based matching

**Teams**:
- Player1 (size: 1)
- Player2 (size: 1)

**Rules**:
1. **Level Range**: Players must be level 10-50 (relaxes after 30s)
2. **Region Preference**: Prefer same region (relaxes after 60s)
3. **Skill Rating**: Must be 1000-2000 (strict rule)

**Example Player Metadata**:
```json
{
  "level": 25,
  "region": "us-west",
  "skill_rating": 1500,
  "preferred_role": "attacker"
}
```

### Rule Set #2: 1v3 Matchmaking (`game-1v3`)

**Description**: Team-based 1v3 matchmaking with coordination requirements

**Teams**:
- Solo (size: 1)
- Trio (size: 3)

**Rules**:
1. **Level Range**: Players must be level 15-60 (relaxes after 45s)
2. **Team Experience**: Minimum experience of 1 (relaxes after 90s)
3. **Communication**: Must have voice capability (relaxes after 120s)

**Example Player Metadata**:
```json
{
  "level": 35,
  "team_experience": 5,
  "communication": ["voice", "text"],
  "preferred_role": "leader"
}
```

### Loading Predefined Rules

Use the configuration loader script to apply these rule sets:

```bash
# Load all predefined rule sets
make load-rules

# Or run the demo script that includes both rule sets
make demo-rules
```

### Configuration File

Rule sets are defined in `config/game-rules.yaml`:

```yaml
games:
  game-1v1:
    game_id: "game-1v1"
    description: "1v1 competitive matchmaking with skill-based matching"
    teams:
      - name: "Player1"
        size: 1
      - name: "Player2"
        size: 1
    rules:
      - field: "level"
        min: 10
        max: 50
        strict: false
        priority: 1
        relax_after: 30
```

## Development

### Project Structure

```
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/            # HTTP handlers and routing
│   ├── allocation/     # Session allocation logic
│   ├── engine/         # Rule processing engine
│   ├── matchmaker/     # Core matchmaking logic
│   ├── models/         # Data structures
│   └── storage/        # Redis storage layer
├── config/             # Configuration files
│   ├── config.yaml     # Main system configuration
│   └── game-rules.yaml # Predefined rule sets
├── scripts/            # Management scripts
│   ├── load-rules.sh   # Load rule configurations
│   ├── test-rules.sh   # Test rule sets
│   └── manage-rules.sh # Rule management interface
├── examples/           # Example scripts
│   ├── demo.sh         # Basic demo
│   └── rules-demo.sh   # Comprehensive rule demo
├── docs/               # Documentation
│   ├── rule-sets.md    # Rule set specifications
│   ├── rule-workflow.md # Rule management workflow
│   └── quick-reference.md # Quick reference guide
├── k8s/               # Kubernetes manifests
├── Dockerfile         # Container definition
└── README.md          # This file
```

### Development Workflow

1. **Start development environment**
   ```bash
   make redis                # Start Redis
   ./scripts/start-server.sh # Start server and load rules automatically
   # or
   make server-start
   
   # If you use 'make run', load rules manually:
   # make load-rules
   ```

2. **Test rules**
   ```bash
   make test-rules           # Test rule sets
   ```

3. **Create new rules**
   ```bash
   ./scripts/manage-rules.sh template my-game.yaml
   # Edit my-game.yaml
   ./scripts/manage-rules.sh create my-game my-game.yaml
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Start the server locally |
| `make build` | Build the application |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage |
| `make load-rules` | Load predefined rule sets |
| `make test-rules` | Test rule sets |
| `make demo-rules` | Run comprehensive demo |
| `make manage-rules` | Open rule management interface |
| `make redis` | Start Redis container |
| `make redis-stop` | Stop Redis container |
| `make monitoring` | Start monitoring stack |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker container |
| `make k8s-deploy` | Deploy to Kubernetes |
| `make help` | Show all available commands |

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
make test-coverage

# Test specific package
go test ./internal/engine

# Test rule sets
make test-rules
```

### Building

```bash
# Build for local platform
go build -o bin/server cmd/server/main.go

# Or use make
make build

# Build Docker image
make docker-build
```

## Monitoring & Observability

### Metrics

The system exposes Prometheus metrics at `/metrics`:

- `matchmaking_requests_total`: Total match requests
- `matchmaking_matches_total`: Total matches created
- `matchmaking_queue_size`: Current queue size per game
- `matchmaking_processing_duration`: Matchmaking processing time

### Logging

Structured JSON logging with correlation IDs for request tracing.

### Health Checks

- `/health`: Basic health check
- `/metrics`: Prometheus metrics endpoint

### Monitoring Stack

Start the monitoring stack for development:

```bash
# Start Prometheus and Grafana
make monitoring

# View logs
make monitoring-logs

# Open Grafana (macOS)
make grafana-open
```

## Deployment

### Docker Deployment

1. **Build the image**
   ```bash
   make docker-build
   ```

2. **Run the container**
   ```bash
   make docker-run
   ```

### Kubernetes Deployment

1. **Apply the manifests**
   ```bash
   make k8s-deploy
   ```

2. **Check deployment status**
   ```bash
   kubectl get pods -l app=mm-rules-matchmaking
   ```

## Scaling

### Horizontal Scaling

The system is designed for horizontal scaling:

1. **Stateless Design**: All state is stored in Redis
2. **Load Balancing**: Multiple instances can be deployed behind a load balancer
3. **Redis Cluster**: For high availability and performance

### Performance Considerations

- **Redis Pipelining**: Batch operations for better performance
- **Connection Pooling**: Efficient Redis connection management
- **Async Processing**: Non-blocking matchmaking operations

## Security

- **Input Validation**: All API inputs are validated
- **Rate Limiting**: Can be added via middleware
- **Authentication**: Can be integrated with existing auth systems
- **HTTPS**: Recommended for production deployments

## Troubleshooting

### Common Issues

1. **Server won't start**: Check if Redis is running
2. **Rules not working**: Use `make test-rules` to validate
3. **Configuration errors**: Check YAML syntax with `yq eval config/game-rules.yaml`

### Debugging Commands

```bash
# Check server health
curl http://localhost:8080/health

# View server logs
make logs

# Check rule configurations
./scripts/manage-rules.sh list

# Test specific game
./scripts/manage-rules.sh show game-1v1
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Development Guidelines

- Follow Go coding standards
- Add tests for new features
- Update documentation for API changes
- Use the rule management scripts for testing

## Documentation

- [Rule Sets Documentation](docs/rule-sets.md) - Detailed rule set specifications
- [Rule Management Workflow](docs/rule-workflow.md) - Complete workflow guide
- [Quick Reference](docs/quick-reference.md) - Command reference
- [API Reference](#api-reference) - Complete API documentation

## License

[Add your license here]

## Support

For questions and support, please open an issue on GitHub.

## 🖥️ Web Client for Matchmaking Testing

The project includes a modern web client for testing and demonstrating the matchmaking system.

### How to Run the Client

1. Open a terminal and start the backend server (`make run` in the project root).
2. In another terminal, start the client:
   ```sh
   cd mm-rules-client
   npm install
   npm run dev
   ```
3. Open your browser to the URL shown in the terminal (usually http://localhost:5173/).

### What the Client Demonstrates
- Join 1v1 or 1v3 game modes with random player metadata.
- See real-time match status and detailed match information (match ID, all players, team, etc.).
- Multiple browser windows can be used to simulate multiple players.
- The client will show full match details for all players as soon as they are matched.

### How to Test
- Open two browser windows and join the same game mode (e.g., 1v1) in both.
- Both clients will be matched together and see the same detailed match info.
- Try 1v3 mode to see team-based matching.

This client is a great way to validate the end-to-end matchmaking flow and visualize the results interactively. 