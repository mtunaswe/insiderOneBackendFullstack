package httpapi

import "github.com/mtunaswe/insider-league/internal/domain"

type WeekResponse struct {
	CurrentWeek int `json:"current_week"`
	TotalWeeks  int `json:"total_weeks"`
}

type EditMatchRequest struct {
	HomeGoals *int `json:"home_goals"`
	AwayGoals *int `json:"away_goals"`
}

type TableResponse struct {
	Table []domain.StandingsRow `json:"table"`
}

type PredictionsResponse struct {
	ChampionshipOdds []domain.ChampionshipOdds `json:"championship_odds"`
	RemainingMatches []domain.MatchOdds        `json:"remaining_matches"`
	Iterations       int                       `json:"iterations"`
}

type MatchListResponse struct {
	Matches []domain.Match `json:"matches"`
}
