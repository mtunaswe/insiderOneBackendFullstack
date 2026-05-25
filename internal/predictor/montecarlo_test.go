package predictor

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/mtunaswe/insider-league/internal/simulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonteCarloOdds(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "Chelsea", Strength: 85},
		{ID: 2, Name: "Manchester City", Strength: 80},
		{ID: 3, Name: "Arsenal", Strength: 70},
		{ID: 4, Name: "Liverpool", Strength: 60},
	}

	// Mid-season state: 3 weeks played, Chelsea leads
	standings := []domain.StandingsRow{
		{TeamID: 1, TeamName: "Chelsea", P: 3, W: 3, D: 0, L: 0, GF: 7, GA: 2, GD: 5, Pts: 9},
		{TeamID: 2, TeamName: "Manchester City", P: 3, W: 2, D: 0, L: 1, GF: 5, GA: 3, GD: 2, Pts: 6},
		{TeamID: 3, TeamName: "Arsenal", P: 3, W: 1, D: 0, L: 2, GF: 3, GA: 5, GD: -2, Pts: 3},
		{TeamID: 4, TeamName: "Liverpool", P: 3, W: 0, D: 0, L: 3, GF: 2, GA: 7, GD: -5, Pts: 0},
	}

	// Remaining 6 matches (weeks 4-6)
	remaining := []domain.Match{
		{ID: 7, Week: 4, HomeTeamID: 2, AwayTeamID: 1, Played: false},
		{ID: 8, Week: 4, HomeTeamID: 4, AwayTeamID: 3, Played: false},
		{ID: 9, Week: 5, HomeTeamID: 3, AwayTeamID: 1, Played: false},
		{ID: 10, Week: 5, HomeTeamID: 4, AwayTeamID: 2, Played: false},
		{ID: 11, Week: 6, HomeTeamID: 1, AwayTeamID: 4, Played: false},
		{ID: 12, Week: 6, HomeTeamID: 3, AwayTeamID: 2, Played: false},
	}

	factory := func(rng *rand.Rand) domain.MatchSimulator {
		return simulator.NewPoissonSimulator(rng)
	}

	mc := NewMonteCarlo(factory, 5000, 42)
	odds, err := mc.ChampionshipOdds(context.Background(), standings, remaining, teams)
	require.NoError(t, err)
	require.Len(t, odds, 4)

	t.Run("probabilities sum to 1.0", func(t *testing.T) {
		sum := 0.0
		for _, o := range odds {
			sum += o.Probability
		}
		assert.InDelta(t, 1.0, sum, 0.001)
	})

	t.Run("leading team has highest probability", func(t *testing.T) {
		// Chelsea (ID=1) leads with 9 pts, should have highest championship odds
		assert.Equal(t, 1, odds[0].TeamID, "Chelsea should have highest probability")
		assert.Greater(t, odds[0].Probability, odds[1].Probability)
	})

	t.Run("all probabilities are non-negative", func(t *testing.T) {
		for _, o := range odds {
			assert.False(t, math.Signbit(o.Probability), "probability should be non-negative for %s", o.TeamName)
		}
	})
}
