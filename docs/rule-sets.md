# MM-Rules Predefined Rule Sets

This document describes the two predefined rule sets available in the MM-Rules matchmaking system.

## Overview

The MM-Rules system comes with two predefined rule sets designed for common matchmaking scenarios:

1. **1v1 Matchmaking** - Competitive head-to-head matches
2. **1v3 Matchmaking** - Team-based asymmetric matches

## Rule Set #1: 1v1 Matchmaking

**Game ID**: `game-1v1`  
**Description**: Competitive 1v1 matchmaking with skill-based matching

### Team Configuration

| Team Name | Size | Description |
|-----------|------|-------------|
| Player1   | 1    | First player |
| Player2   | 1    | Second player |

### Rules

#### 1. Level Range Rule
- **Field**: `level`
- **Min**: 10
- **Max**: 50
- **Strict**: false
- **Priority**: 1
- **Relax After**: 30 seconds
- **Description**: Players must be between level 10-50. This rule relaxes after 30 seconds to allow for faster matchmaking.

#### 2. Region Preference Rule
- **Field**: `region`
- **Equals**: `us-west`
- **Strict**: false
- **Priority**: 2
- **Relax After**: 60 seconds
- **Description**: Prefers players from the same region. Relaxes after 60 seconds to allow cross-region matches.

#### 3. Skill Rating Rule
- **Field**: `skill_rating`
- **Min**: 1000
- **Max**: 2000
- **Strict**: true
- **Priority**: 3
- **Description**: Players must have a skill rating between 1000-2000. This is a strict rule that never relaxes.

### Example Player Metadata

```json
{
  "level": 25,
  "region": "us-west",
  "skill_rating": 1500,
  "preferred_role": "attacker"
}
```

### Use Cases

- Competitive fighting games
- Chess matches
- 1v1 strategy games
- Skill-based competitive modes

## Rule Set #2: 1v3 Matchmaking

**Game ID**: `game-1v3`  
**Description**: Team-based 1v3 matchmaking with coordination requirements

### Team Configuration

| Team Name | Size | Description |
|-----------|------|-------------|
| Solo      | 1    | Single player team |
| Trio      | 3    | Three-player team |

### Rules

#### 1. Level Range Rule
- **Field**: `level`
- **Min**: 15
- **Max**: 60
- **Strict**: false
- **Priority**: 1
- **Relax After**: 45 seconds
- **Description**: Players must be between level 15-60. Relaxes after 45 seconds.

#### 2. Team Experience Rule
- **Field**: `team_experience`
- **Min**: 1
- **Strict**: false
- **Priority**: 2
- **Relax After**: 90 seconds
- **Description**: Players must have at least 1 unit of team experience. Relaxes after 90 seconds.

#### 3. Communication Rule
- **Field**: `communication`
- **Contains**: `voice`
- **Strict**: false
- **Priority**: 3
- **Relax After**: 120 seconds
- **Description**: Players must have voice communication capability. Relaxes after 120 seconds.

### Example Player Metadata

#### Solo Player
```json
{
  "level": 35,
  "team_experience": 5,
  "communication": ["voice", "text"],
  "preferred_role": "leader"
}
```

#### Trio Players
```json
{
  "level": 28,
  "team_experience": 3,
  "communication": ["voice"],
  "preferred_role": "support"
}
```

### Use Cases

- Asymmetric team games
- 1v3 survival modes
- Team coordination games
- Mentor/mentee scenarios

## Rule Relaxation Strategy

Both rule sets implement a progressive relaxation strategy:

1. **Immediate Matching**: Try to match players who meet all criteria
2. **Progressive Relaxation**: Rules relax over time to increase match probability
3. **Strict Rules**: Some rules never relax (e.g., skill rating in 1v1)

### Relaxation Timeline

#### 1v1 Matchmaking
- 0-30s: All rules active
- 30-60s: Level rule relaxed
- 60s+: Level and region rules relaxed
- Never: Skill rating rule remains strict

#### 1v3 Matchmaking
- 0-45s: All rules active
- 45-90s: Level rule relaxed
- 90-120s: Level and team experience rules relaxed
- 120s+: All rules relaxed except strict ones

## Usage

### Loading Rule Sets

```bash
# Load all predefined rule sets
make load-rules

# Or use the script directly
./scripts/load-rules.sh
```

### Testing Rule Sets

```bash
# Test the rule sets
make test-rules

# Or use the script directly
./scripts/test-rules.sh
```

### Demo

```bash
# Run the comprehensive demo
make demo-rules

# Or use the script directly
./examples/rules-demo.sh
```

## Configuration

Rule sets are defined in `config/game-rules.yaml` and can be customized:

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

## Customization

You can create custom rule sets by:

1. Adding new game configurations to `config/game-rules.yaml`
2. Using the API directly to create game configurations
3. Modifying existing rule sets to match your requirements

### Creating Custom Rules

```bash
curl -X POST "http://localhost:8080/api/v1/rules/my-custom-game" \
  -H "Content-Type: application/json" \
  -d '{
    "teams": [
      {"name": "TeamA", "size": 2},
      {"name": "TeamB", "size": 2}
    ],
    "rules": [
      {
        "field": "rank",
        "min": 1000,
        "strict": true,
        "priority": 1
      }
    ]
  }'
```

## Best Practices

1. **Start Strict**: Begin with strict rules and relax over time
2. **Prioritize Rules**: Use priority to control rule evaluation order
3. **Monitor Performance**: Track matchmaking success rates
4. **Test Thoroughly**: Use the test scripts to validate rule behavior
5. **Document Changes**: Keep rule configurations version controlled

## Troubleshooting

### Common Issues

1. **No Matches Created**: Check if rules are too strict
2. **Slow Matchmaking**: Consider reducing relaxation times
3. **Poor Match Quality**: Adjust rule priorities or criteria

### Debugging

```bash
# Check game configuration
curl "http://localhost:8080/api/v1/rules/game-1v1"

# Check player status
curl "http://localhost:8080/api/v1/match-status/request-id"

# View statistics
curl "http://localhost:8080/api/v1/stats"
``` 