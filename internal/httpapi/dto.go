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
	Predictions []domain.ChampionshipOdds `json:"predictions"`
}

type MatchListResponse struct {
	Matches []domain.Match `json:"matches"`
}
