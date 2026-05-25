package domain

type Match struct {
	ID         int  `json:"id"`
	Week       int  `json:"week"`
	HomeTeamID int  `json:"home_team_id"`
	AwayTeamID int  `json:"away_team_id"`
	HomeGoals  *int `json:"home_goals"`
	AwayGoals  *int `json:"away_goals"`
	Played     bool `json:"played"`
}
