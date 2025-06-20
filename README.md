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

## Quick Start

### Prerequisites

- Go 1.21+
- Redis 6.0+
- Docker (optional)
- Kubernetes cluster (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd MM_Rules
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start Redis**
   ```bash
   docker run -d -p 6379:6379 redis:7-alpine
   ```

4. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

### Docker Deployment

1. **Build the image**
   ```bash
   docker build -t mm-rules-matchmaking .
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 \
     -e MM_RULES_REDIS_ADDR=host.docker.internal:6379 \
     mm-rules-matchmaking
   ```

### Kubernetes Deployment

1. **Apply the manifests**
   ```bash
   kubectl apply -f k8s/
   ```

2. **Check deployment status**
   ```bash
   kubectl get pods -l app=mm-rules-matchmaking
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

## Session Allocation

The system integrates with external allocation services via webhooks:

### Allocation Request Format

```json
{
  "match_id": "match-123",
  "game_id": "my-game",
  "players": ["player1", "player2"],
  "team_name": "Duo"
}
```

### Allocation Response Format

```json
{
  "success": true,
  "session": {
    "ip": "12.34.56.78",
    "port": 7777,
    "id": "session-123"
  }
}
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
├── k8s/               # Kubernetes manifests
├── Dockerfile         # Container definition
└── README.md          # This file
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/server cmd/server/main.go
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

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

[Add your license here]

## Support

For questions and support, please open an issue on GitHub. 