package domain

const (
	PointsWin  = 3
	PointsDraw = 1
	PointsLoss = 0
	TotalWeeks = 6
	TeamsCount = 4
)

type StandingsRow struct {
	TeamID   int    `json:"team_id"`
	TeamName string `json:"team_name"`
	P        int    `json:"played"`
	W        int    `json:"wins"`
	D        int    `json:"draws"`
	L        int    `json:"losses"`
	GF       int    `json:"goals_for"`
	GA       int    `json:"goals_against"`
	GD       int    `json:"goal_difference"`
	Pts      int    `json:"points"`
}

type ChampionshipOdds struct {
	TeamID      int     `json:"team_id"`
	TeamName    string  `json:"team_name"`
	Probability float64 `json:"probability"`
}
