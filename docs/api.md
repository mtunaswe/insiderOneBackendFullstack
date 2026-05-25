# API Reference

Base URL: `http://localhost:8080`

All responses are JSON. Errors follow [RFC 7807](https://tools.ietf.org/html/rfc7807) Problem Details format.

---

## GET /health

Health check with database connectivity status.

**Response 200:**
```json
{
  "status": "ok",
  "db": "ok"
}
```

**Response 503 (DB unreachable):**
```json
{
  "status": "degraded",
  "db": "error"
}
```

---

## POST /league/reset

Wipe all match data, regenerate fixtures for a fresh season.

**Request:** No body required.

**Response 200:**
```json
{
  "message": "league reset successfully"
}
```

---

## GET /league/table

Current league standings sorted by Points > Goal Difference > Goals For.

**Response 200:**
```json
{
  "table": [
    {
      "team_id": 1,
      "team_name": "Chelsea",
      "played": 4,
      "wins": 3,
      "draws": 1,
      "losses": 0,
      "goals_for": 11,
      "goals_against": 3,
      "goal_difference": 8,
      "points": 10
    },
    {
      "team_id": 3,
      "team_name": "Arsenal",
      "played": 4,
      "wins": 2,
      "draws": 1,
      "losses": 1,
      "goals_for": 7,
      "goals_against": 5,
      "goal_difference": 2,
      "points": 7
    }
  ]
}
```

---

## GET /league/week

Returns the current (next unplayed) week and total weeks in the season.

**Response 200:**
```json
{
  "current_week": 3,
  "total_weeks": 6
}
```

If the season is complete, `current_week` will be 7 (one past total).

---

## POST /league/next-week

Simulate the next unplayed week. Returns match results for that week.

**Response 200:**
```json
{
  "week": 3,
  "matches": [
    {
      "home_team": "Chelsea",
      "away_team": "Liverpool",
      "home_goals": 2,
      "away_goals": 1
    },
    {
      "home_team": "Arsenal",
      "away_team": "Manchester City",
      "home_goals": 0,
      "away_goals": 0
    }
  ]
}
```

**Response 409 (season already complete):**
```json
{
  "type": "about:blank",
  "title": "Conflict",
  "detail": "season is complete",
  "status": 409
}
```

---

## POST /league/play-all

Simulate all remaining weeks. Returns per-week breakdown, final table, and champion.

**Query parameters:**
- `from_current=false` — reset before playing (optional)

**Response 200:**
```json
{
  "weeks": [
    {
      "week": 3,
      "matches": [
        {
          "home_team": "Chelsea",
          "away_team": "Liverpool",
          "home_goals": 1,
          "away_goals": 0
        }
      ]
    }
  ],
  "final_table": [
    {
      "team_id": 1,
      "team_name": "Chelsea",
      "played": 6,
      "wins": 5,
      "draws": 1,
      "losses": 0,
      "goals_for": 14,
      "goals_against": 3,
      "goal_difference": 11,
      "points": 16
    }
  ],
  "champion": "Chelsea"
}
```

**Response 409 (season already complete):**
```json
{
  "type": "about:blank",
  "title": "Conflict",
  "detail": "season is complete",
  "status": 409
}
```

---

## GET /matches

List all matches with optional filters.

**Query parameters:**
- `week` (integer) — filter by week number
- `played` (boolean) — filter by played status

**Examples:**
- `GET /matches` — all 12 matches
- `GET /matches?week=2` — matches in week 2
- `GET /matches?played=true` — only completed matches

**Response 200:**
```json
{
  "matches": [
    {
      "id": 1,
      "week": 1,
      "home_team_id": 1,
      "away_team_id": 4,
      "home_goals": 3,
      "away_goals": 1,
      "played": true
    },
    {
      "id": 2,
      "week": 1,
      "home_team_id": 2,
      "away_team_id": 3,
      "home_goals": null,
      "away_goals": null,
      "played": false
    }
  ]
}
```

---

## PUT /matches/{id}

Edit the score of a played match. Standings and predictions recalculate on next request.

**Request body:**
```json
{
  "home_goals": 2,
  "away_goals": 3
}
```

**Validation:**
- `home_goals` and `away_goals` are required
- Values must be integers between 0 and 20

**Response 200:**
```json
{
  "message": "match updated successfully"
}
```

**Response 400 (invalid input):**
```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "detail": "goals must be between 0 and 20",
  "status": 400
}
```

**Response 404 (match not found):**
```json
{
  "type": "about:blank",
  "title": "Not Found",
  "detail": "match not found",
  "status": 404
}
```

**Response 409 (match not yet played):**
```json
{
  "type": "about:blank",
  "title": "Conflict",
  "detail": "match has not been played yet",
  "status": 409
}
```

---

## GET /predictions

Championship probability for each team calculated via Monte Carlo simulation (10,000 iterations by default). Available after at least 4 weeks have been played for meaningful results.

**Response 200:**
```json
{
  "predictions": [
    {
      "team_id": 1,
      "team_name": "Chelsea",
      "probability": 0.45
    },
    {
      "team_id": 3,
      "team_name": "Arsenal",
      "probability": 0.28
    },
    {
      "team_id": 2,
      "team_name": "Manchester City",
      "probability": 0.22
    },
    {
      "team_id": 4,
      "team_name": "Liverpool",
      "probability": 0.05
    }
  ]
}
```

Probabilities sum to approximately 1.0 (minor floating-point variance is expected).
