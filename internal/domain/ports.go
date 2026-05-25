package domain

import (
	"context"
	"time"
)

type TeamRepository interface {
	List(ctx context.Context) ([]Team, error)
	GetByID(ctx context.Context, id int) (Team, error)
}

type MatchFilter struct {
	Week   *int
	Played *bool
}

type MatchRepository interface {
	List(ctx context.Context, filter MatchFilter) ([]Match, error)
	GetByID(ctx context.Context, id int) (Match, error)
	CreateBatch(ctx context.Context, matches []Match) error
	UpdateResult(ctx context.Context, id int, homeGoals, awayGoals int) error
	Truncate(ctx context.Context) error
}

type MatchSimulator interface {
	Simulate(home, away Team) (homeGoals, awayGoals int)
}

type Predictor interface {
	Predict(ctx context.Context, teams []Team, matches []Match) (PredictionResult, error)
}

type Clock interface {
	Now() time.Time
}
