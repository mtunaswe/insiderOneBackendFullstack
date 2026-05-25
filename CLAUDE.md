# Insider League Simulator — Project Context

## What this project is
A REST API in Go that simulates a 4-team football mini-league using
Premier League rules. It plays weekly fixtures, maintains a league table,
predicts championship probabilities, and lets users edit past results
(standings auto-recalculate). Frontend is intentionally out of scope —
all interaction is via HTTP endpoints, testable with curl/Postman.

This is a submission for the Insider Backend/Full-Stack Development
Intern hiring case. Constraints from the brief are non-negotiable.

## Hard constraints (from the brief)
- Language MUST be Go. Do not introduce another language for the core service.
- Interface-based design and struct composition — favour interfaces at
  every seam (repo, simulator, predictor, clock/RNG).
- No frontend. Endpoints must be self-explanatory and Postman-testable.
- SQL schema and queries must be delivered as plain .sql files.
- Setup must be one command (`docker compose up`).
- Deployment instructions must be included; a live link is a bonus.

## League rules (Premier League subset)
- 4 teams, double round-robin → 12 matches over 6 weeks (2 per week).
- Win = 3, Draw = 1, Loss = 0.
- Tiebreakers in order: Points → Goal Difference → Goals For.
- Goal difference = Goals For − Goals Against.

## Tech stack
- Go 1.22+
- chi router (github.com/go-chi/chi/v5)
- pgx v5 (github.com/jackc/pgx/v5) for Postgres
- golang-migrate for SQL migrations
- testify for tests
- PostgreSQL 16
- Docker + docker compose for local + deploy

## Folder layout (target)
.
├── cmd/server/main.go              # entrypoint
├── internal/
│   ├── domain/                     # pure types + interfaces, no I/O
│   │   ├── team.go
│   │   ├── match.go
│   │   ├── league.go
│   │   └── ports.go                # interface definitions
│   ├── simulator/                  # match engine (Poisson)
│   ├── predictor/                  # Monte Carlo championship odds
│   ├── league/                     # league service: standings, week flow
│   ├── repository/postgres/        # DB adapters implementing ports
│   ├── httpapi/                    # handlers, router, DTOs, middleware
│   └── config/                     # env loading
├── migrations/                     # *.up.sql / *.down.sql
├── docs/
│   ├── api.md
│   └── postman_collection.json
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── .env.example
├── go.mod / go.sum
└── README.md

## Coding conventions
- All external dependencies behind interfaces in `internal/domain/ports.go`.
- No `panic` outside `main.go`; return errors.
- Handlers are thin: parse → call service → write JSON. No business logic.
- Services depend on interfaces, never on concrete repo types.
- Inject `*rand.Rand` and `clock` interface so simulations are testable.
- Table-driven tests for the simulator and standings calculator.
- Errors wrapped with `fmt.Errorf("...: %w", err)`.
- JSON field names: snake_case.
- HTTP status codes: 200 OK, 201 Created, 400 Bad Request, 404 Not Found,
  409 Conflict (e.g. playing a week that's already played), 500 Server Error.

## Domain shape (canonical)
- Team{ID, Name, Strength int /*1-100*/}
- Match{ID, Week int, HomeTeamID, AwayTeamID, HomeGoals *int, AwayGoals *int, Played bool}
- StandingsRow{TeamID, TeamName, P, W, D, L, GF, GA, GD, Pts}
- ChampionshipOdds{TeamID, TeamName, Probability float64 /*0..1*/}

## Endpoints (target)
POST   /league/reset                 → wipe + reseed teams + regenerate fixtures
GET    /league/table                 → current standings
GET    /league/week                  → current week index + total weeks
GET    /matches                      → all matches (filter: ?week=N, ?played=true)
POST   /league/next-week             → play the next unplayed week
POST   /league/play-all              → play every remaining week
PUT    /matches/{id}                 → edit a played match's score; recalculates
GET    /predictions                  → championship probabilities (Monte Carlo)

## Acceptance criteria (definition of done)
- `docker compose up` from a fresh clone brings up DB + app + applies migrations.
- All endpoints work via the included Postman collection.
- `go test ./...` is green.
- After 4 weeks the predictions endpoint returns probabilities that sum to ~1.0.
- Editing a match via PUT updates the standings and predictions on the next call.
- README explains: prerequisites, run, test, endpoints, deployment.

## How to work on this repo
- Always run `go fmt ./... && go vet ./...` before declaring a step done.
- After every step, run `go build ./...` and `go test ./...`.
- Don't introduce new external dependencies without noting why in the PR/commit.
- Migrations are append-only; never edit a committed migration.