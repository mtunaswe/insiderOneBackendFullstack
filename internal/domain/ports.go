package domain

import (
	"context"
	"time"
)

type TeamRepository interface {
	GetAll(ctx context.Context) ([]Team, error)
	DeleteAll(ctx context.Context) error
	InsertAll(ctx context.Context, teams []Team) error
}

type MatchRepository interface {
	GetAll(ctx context.Context) ([]Match, error)
	GetByWeek(ctx context.Context, week int) ([]Match, error)
	GetPlayed(ctx context.Context) ([]Match, error)
	GetUnplayed(ctx context.Context) ([]Match, error)
	GetByID(ctx context.Context, id int) (Match, error)
	InsertAll(ctx context.Context, matches []Match) error
	Update(ctx context.Context, m Match) error
	DeleteAll(ctx context.Context) error
}

type MatchSimulator interface {
	Simulate(home, away Team) (homeGoals, awayGoals int)
}

type Predictor interface {
	ChampionshipOdds(ctx context.Context, standings []StandingsRow, remaining []Match, teams []Team) ([]ChampionshipOdds, error)
}

type Clock interface {
	Now() time.Time
}
