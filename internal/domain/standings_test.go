package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int { return &v }

func TestComputeStandings(t *testing.T) {
	teams := []Team{
		{ID: 1, Name: "A", Strength: 80},
		{ID: 2, Name: "B", Strength: 70},
		{ID: 3, Name: "C", Strength: 60},
		{ID: 4, Name: "D", Strength: 50},
	}

	tests := []struct {
		name    string
		matches []Match
		check   func(t *testing.T, rows []StandingsRow)
	}{
		{
			name:    "empty matches returns zero rows for all teams",
			matches: []Match{},
			check: func(t *testing.T, rows []StandingsRow) {
				require.Len(t, rows, 4)
				for _, r := range rows {
					assert.Equal(t, 0, r.Pts)
					assert.Equal(t, 0, r.P)
					assert.Equal(t, 0, r.GF)
				}
			},
		},
		{
			name: "one win and one loss",
			matches: []Match{
				{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 2, HomeGoals: intPtr(3), AwayGoals: intPtr(1), Played: true},
			},
			check: func(t *testing.T, rows []StandingsRow) {
				require.Len(t, rows, 4)
				// A should be first with 3 points
				assert.Equal(t, 1, rows[0].TeamID)
				assert.Equal(t, 3, rows[0].Pts)
				assert.Equal(t, 1, rows[0].W)
				assert.Equal(t, 3, rows[0].GF)
				assert.Equal(t, 1, rows[0].GA)
				assert.Equal(t, 2, rows[0].GD)
				// B has 0 points, 1 loss — find by ID (others also have 0 pts)
				var b *StandingsRow
				for i := range rows {
					if rows[i].TeamID == 2 {
						b = &rows[i]
						break
					}
				}
				require.NotNil(t, b)
				assert.Equal(t, 0, b.Pts)
				assert.Equal(t, 1, b.L)
				assert.Equal(t, 1, b.GF)
				assert.Equal(t, 3, b.GA)
			},
		},
		{
			name: "one draw",
			matches: []Match{
				{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 2, HomeGoals: intPtr(2), AwayGoals: intPtr(2), Played: true},
			},
			check: func(t *testing.T, rows []StandingsRow) {
				assert.Equal(t, 1, rows[0].Pts)
				assert.Equal(t, 1, rows[0].D)
				assert.Equal(t, 1, rows[1].Pts)
				assert.Equal(t, 1, rows[1].D)
			},
		},
		{
			name: "tied on points separated by goal difference",
			matches: []Match{
				// A beats C 4-0 → GD +4
				{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 3, HomeGoals: intPtr(4), AwayGoals: intPtr(0), Played: true},
				// B beats D 2-1 → GD +1
				{ID: 2, Week: 1, HomeTeamID: 2, AwayTeamID: 4, HomeGoals: intPtr(2), AwayGoals: intPtr(1), Played: true},
			},
			check: func(t *testing.T, rows []StandingsRow) {
				// Both A and B have 3 pts, A has better GD
				assert.Equal(t, 1, rows[0].TeamID)
				assert.Equal(t, 2, rows[1].TeamID)
				assert.Equal(t, 3, rows[0].Pts)
				assert.Equal(t, 3, rows[1].Pts)
				assert.Greater(t, rows[0].GD, rows[1].GD)
			},
		},
		{
			name: "three teams tied on points and GD separated by GF",
			matches: []Match{
				// A beats D 3-1 → GD +2, GF 3
				{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 4, HomeGoals: intPtr(3), AwayGoals: intPtr(1), Played: true},
				// B beats D 4-2 → GD +2, GF 4
				{ID: 2, Week: 2, HomeTeamID: 2, AwayTeamID: 4, HomeGoals: intPtr(4), AwayGoals: intPtr(2), Played: true},
				// C beats D 5-3 → GD +2, GF 5
				{ID: 3, Week: 3, HomeTeamID: 3, AwayTeamID: 4, HomeGoals: intPtr(5), AwayGoals: intPtr(3), Played: true},
			},
			check: func(t *testing.T, rows []StandingsRow) {
				// All three have 3 pts, GD +2 — rank by GF: C(5) > B(4) > A(3)
				assert.Equal(t, 3, rows[0].TeamID) // C
				assert.Equal(t, 2, rows[1].TeamID) // B
				assert.Equal(t, 1, rows[2].TeamID) // A
				assert.Equal(t, rows[0].Pts, rows[1].Pts)
				assert.Equal(t, rows[1].Pts, rows[2].Pts)
				assert.Equal(t, rows[0].GD, rows[1].GD)
				assert.Equal(t, rows[1].GD, rows[2].GD)
				assert.Greater(t, rows[0].GF, rows[1].GF)
				assert.Greater(t, rows[1].GF, rows[2].GF)
			},
		},
		{
			name: "unplayed matches are ignored",
			matches: []Match{
				{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 2, HomeGoals: nil, AwayGoals: nil, Played: false},
			},
			check: func(t *testing.T, rows []StandingsRow) {
				for _, r := range rows {
					assert.Equal(t, 0, r.P)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rows := ComputeStandings(teams, tc.matches)
			tc.check(t, rows)
		})
	}
}
