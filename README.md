# Insider Backend Case

This project is a Go REST API for a 4-team football league simulation. It plays weekly fixtures, calculates the league table with Premier League rules, estimates championship probabilities after week 4, supports playing the full league automatically, and recalculates standings after edited match results.

## Tech Stack

- Go 1.23+
- Gin
- PostgreSQL
- Docker Compose
- Interface-based service/repository design

## League Rules

- Teams: Chelsea, Arsenal, Manchester City, Liverpool
- Double round-robin fixture
- 6 weeks total
- 2 matches per week
- Win: 3 points
- Draw: 1 point
- Loss: 0 points
- Ranking: points, goal difference, goals scored, team name

The match simulation is deterministic. Given the same database state and seed values, the generated scores are the same.

Championship predictions use a deterministic Monte Carlo simulation. After the 4th completed week, the API simulates the remaining fixtures many times and returns championship probability percentages.

## Project Structure

```text
cmd/api                 API entrypoint
internal/config         environment configuration
internal/domain         shared domain models
internal/http           Gin handlers and routes
internal/repository     repository interface
internal/repository/postgres
                        PostgreSQL implementation
internal/service        league, simulation, prediction, standings logic
migrations              SQL schema
sql                     example SQL queries used by the application
```

## Run With Docker

```bash
docker compose up --build
```

The API runs at:

```text
http://localhost:8080
```

PostgreSQL runs at:

```text
localhost:5432
```

Database credentials:

```text
user: insider
password: insider
database: insider_league
```

## First Request

Initialize or reset the league:

```bash
curl -X POST http://localhost:8080/api/leagues/reset
```

This clears all teams and matches, then inserts the default teams and fixtures.

## API Endpoints

### Health Check

```http
GET /health
```

### Reset League

```http
POST /api/leagues/reset
```

Resets the database and returns the current league snapshot.

### Current League Snapshot

```http
GET /api/leagues/current
```

Returns standings, all matches, current completed week, total weeks, and predictions when available.

### Standings

```http
GET /api/standings
```

### All Matches

```http
GET /api/matches
```

### Matches By Week

```http
GET /api/matches/weeks/4
```

### Play Next Week

```http
POST /api/simulation/play-next-week
```

Plays the next unplayed week and returns the played week plus the updated league snapshot.

### Play All Remaining Weeks

```http
POST /api/simulation/play-all
```

Automatically plays all remaining matches until the league is finished. This covers the first extra requirement in the case.

### Predictions

```http
GET /api/predictions
```

Before week 4 is completed, this returns an empty list with a message. From week 4 onward, it returns Monte Carlo championship probabilities.

### Edit Match Result

```http
PUT /api/matches/1/result
Content-Type: application/json

{
  "home_score": 2,
  "away_score": 1
}
```

The API marks the match as played, saves the new score, and recalculates standings and predictions from match data. This covers the second extra requirement in the case.

## Example Flow

```bash
curl -X POST http://localhost:8080/api/leagues/reset

curl -X POST http://localhost:8080/api/simulation/play-next-week
curl -X POST http://localhost:8080/api/simulation/play-next-week
curl -X POST http://localhost:8080/api/simulation/play-next-week
curl -X POST http://localhost:8080/api/simulation/play-next-week

curl http://localhost:8080/api/predictions

curl -X PUT http://localhost:8080/api/matches/1/result \
  -H "Content-Type: application/json" \
  -d '{"home_score": 0, "away_score": 3}'

curl http://localhost:8080/api/standings

curl -X POST http://localhost:8080/api/simulation/play-all
```

## SQL Files

- Schema: `migrations/001_init.sql`
- Example queries: `sql/queries.sql`

The standings are intentionally not stored as a table. They are calculated from played matches so edited match results always produce the correct league table.

## Environment Variables

```text
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=insider
DB_PASSWORD=insider
DB_NAME=insider_league
DB_SSLMODE=disable
SIMULATION_SEED=20260518
PREDICTION_SEED=42024
PREDICTION_ITERATIONS=10000
```

## Run Tests Locally

If Go is installed locally:

```bash
go test ./...
```

In an environment without local Go installed, the project can still be built and run through Docker:

```bash
docker compose up --build
```
