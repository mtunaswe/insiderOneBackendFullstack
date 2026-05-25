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

func TestMonteCarloPredict(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "Chelsea", Strength: 85},
		{ID: 2, Name: "Manchester City", Strength: 80},
		{ID: 3, Name: "Arsenal", Strength: 70},
		{ID: 4, Name: "Liverpool", Strength: 60},
	}

	hg := func(v int) *int { return &v }

	// 3 weeks played
	played := []domain.Match{
		{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 4, HomeGoals: hg(3), AwayGoals: hg(0), Played: true},
		{ID: 2, Week: 1, HomeTeamID: 2, AwayTeamID: 3, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 3, Week: 2, HomeTeamID: 1, AwayTeamID: 3, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 4, Week: 2, HomeTeamID: 2, AwayTeamID: 4, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 5, Week: 3, HomeTeamID: 1, AwayTeamID: 2, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 6, Week: 3, HomeTeamID: 3, AwayTeamID: 4, HomeGoals: hg(1), AwayGoals: hg(0), Played: true},
	}

	remaining := []domain.Match{
		{ID: 7, Week: 4, HomeTeamID: 4, AwayTeamID: 1, Played: false},
		{ID: 8, Week: 4, HomeTeamID: 3, AwayTeamID: 2, Played: false},
		{ID: 9, Week: 5, HomeTeamID: 3, AwayTeamID: 1, Played: false},
		{ID: 10, Week: 5, HomeTeamID: 4, AwayTeamID: 2, Played: false},
		{ID: 11, Week: 6, HomeTeamID: 4, AwayTeamID: 3, Played: false},
		{ID: 12, Week: 6, HomeTeamID: 2, AwayTeamID: 1, Played: false},
	}

	allMatches := append(played, remaining...)

	factory := func(rng *rand.Rand) domain.MatchSimulator {
		return simulator.NewPoissonSimulator(rng)
	}

	mc := NewMonteCarlo(factory, 5000, 42)
	result, err := mc.Predict(context.Background(), teams, allMatches)
	require.NoError(t, err)

	t.Run("championship probabilities sum to 1.0", func(t *testing.T) {
		sum := 0.0
		for _, o := range result.ChampionshipOdds {
			sum += o.Probability
		}
		assert.InDelta(t, 1.0, sum, 0.001)
	})

	t.Run("leading team has highest probability", func(t *testing.T) {
		assert.Equal(t, 1, result.ChampionshipOdds[0].TeamID, "Chelsea should have highest probability")
		assert.Greater(t, result.ChampionshipOdds[0].Probability, result.ChampionshipOdds[1].Probability)
	})

	t.Run("all championship probabilities are non-negative", func(t *testing.T) {
		for _, o := range result.ChampionshipOdds {
			assert.False(t, math.Signbit(o.Probability), "probability should be non-negative for %s", o.TeamName)
		}
	})

	t.Run("remaining matches has correct count", func(t *testing.T) {
		assert.Len(t, result.MatchOdds, 6)
	})

	t.Run("match odds sum to 1.0 for each match", func(t *testing.T) {
		for _, mo := range result.MatchOdds {
			sum := mo.HomeWin + mo.Draw + mo.AwayWin
			assert.InDelta(t, 1.0, sum, 0.005, "match %d odds should sum to 1.0", mo.MatchID)
		}
	})

	t.Run("match odds sorted by week then id", func(t *testing.T) {
		for i := 1; i < len(result.MatchOdds); i++ {
			prev := result.MatchOdds[i-1]
			curr := result.MatchOdds[i]
			if prev.Week == curr.Week {
				assert.Less(t, prev.MatchID, curr.MatchID)
			} else {
				assert.Less(t, prev.Week, curr.Week)
			}
		}
	})

	t.Run("stronger home team has higher home_win", func(t *testing.T) {
		// Match 7: Liverpool(60) home vs Chelsea(85) away
		// Match 12: Man City(80) home vs Chelsea(85) away
		// Chelsea away in both, but against weaker Liverpool home_win should be lower
		var liverpoolHomeVsChelsea, manCityHomeVsChelsea domain.MatchOdds
		for _, mo := range result.MatchOdds {
			if mo.MatchID == 7 {
				liverpoolHomeVsChelsea = mo
			}
			if mo.MatchID == 12 {
				manCityHomeVsChelsea = mo
			}
		}
		assert.Less(t, liverpoolHomeVsChelsea.HomeWin, manCityHomeVsChelsea.HomeWin,
			"Man City should have higher home_win vs Chelsea than Liverpool does")
	})

	t.Run("same seed produces identical results", func(t *testing.T) {
		mc2 := NewMonteCarlo(factory, 5000, 42)
		result2, err := mc2.Predict(context.Background(), teams, allMatches)
		require.NoError(t, err)
		assert.Equal(t, result.MatchOdds, result2.MatchOdds)
		assert.Equal(t, result.ChampionshipOdds, result2.ChampionshipOdds)
	})

	t.Run("iterations field is set", func(t *testing.T) {
		assert.Equal(t, 5000, result.Iterations)
	})
}

func TestPredictSeasonComplete(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "Chelsea", Strength: 85},
		{ID: 2, Name: "Manchester City", Strength: 80},
		{ID: 3, Name: "Arsenal", Strength: 70},
		{ID: 4, Name: "Liverpool", Strength: 60},
	}

	hg := func(v int) *int { return &v }

	// All matches played
	allPlayed := []domain.Match{
		{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 4, HomeGoals: hg(3), AwayGoals: hg(0), Played: true},
		{ID: 2, Week: 1, HomeTeamID: 2, AwayTeamID: 3, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 3, Week: 2, HomeTeamID: 1, AwayTeamID: 3, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 4, Week: 2, HomeTeamID: 2, AwayTeamID: 4, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 5, Week: 3, HomeTeamID: 1, AwayTeamID: 2, HomeGoals: hg(2), AwayGoals: hg(1), Played: true},
		{ID: 6, Week: 3, HomeTeamID: 3, AwayTeamID: 4, HomeGoals: hg(1), AwayGoals: hg(0), Played: true},
		{ID: 7, Week: 4, HomeTeamID: 4, AwayTeamID: 1, HomeGoals: hg(0), AwayGoals: hg(2), Played: true},
		{ID: 8, Week: 4, HomeTeamID: 3, AwayTeamID: 2, HomeGoals: hg(1), AwayGoals: hg(1), Played: true},
		{ID: 9, Week: 5, HomeTeamID: 3, AwayTeamID: 1, HomeGoals: hg(0), AwayGoals: hg(1), Played: true},
		{ID: 10, Week: 5, HomeTeamID: 4, AwayTeamID: 2, HomeGoals: hg(0), AwayGoals: hg(3), Played: true},
		{ID: 11, Week: 6, HomeTeamID: 4, AwayTeamID: 3, HomeGoals: hg(1), AwayGoals: hg(2), Played: true},
		{ID: 12, Week: 6, HomeTeamID: 2, AwayTeamID: 1, HomeGoals: hg(0), AwayGoals: hg(1), Played: true},
	}

	factory := func(rng *rand.Rand) domain.MatchSimulator {
		return simulator.NewPoissonSimulator(rng)
	}

	mc := NewMonteCarlo(factory, 1000, 42)
	result, err := mc.Predict(context.Background(), teams, allPlayed)
	require.NoError(t, err)

	t.Run("remaining matches is empty", func(t *testing.T) {
		assert.Empty(t, result.MatchOdds)
	})

	t.Run("champion has 100% probability", func(t *testing.T) {
		assert.Equal(t, 1.0, result.ChampionshipOdds[0].Probability)
		assert.Equal(t, "Chelsea", result.ChampionshipOdds[0].TeamName)
	})
}

func TestPredictFreshReset(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "Chelsea", Strength: 85},
		{ID: 2, Name: "Manchester City", Strength: 80},
		{ID: 3, Name: "Arsenal", Strength: 70},
		{ID: 4, Name: "Liverpool", Strength: 60},
	}

	// All unplayed (fresh reset)
	allUnplayed := []domain.Match{
		{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 4, Played: false},
		{ID: 2, Week: 1, HomeTeamID: 2, AwayTeamID: 3, Played: false},
		{ID: 3, Week: 2, HomeTeamID: 1, AwayTeamID: 3, Played: false},
		{ID: 4, Week: 2, HomeTeamID: 2, AwayTeamID: 4, Played: false},
		{ID: 5, Week: 3, HomeTeamID: 1, AwayTeamID: 2, Played: false},
		{ID: 6, Week: 3, HomeTeamID: 3, AwayTeamID: 4, Played: false},
		{ID: 7, Week: 4, HomeTeamID: 4, AwayTeamID: 1, Played: false},
		{ID: 8, Week: 4, HomeTeamID: 3, AwayTeamID: 2, Played: false},
		{ID: 9, Week: 5, HomeTeamID: 3, AwayTeamID: 1, Played: false},
		{ID: 10, Week: 5, HomeTeamID: 4, AwayTeamID: 2, Played: false},
		{ID: 11, Week: 6, HomeTeamID: 4, AwayTeamID: 3, Played: false},
		{ID: 12, Week: 6, HomeTeamID: 2, AwayTeamID: 1, Played: false},
	}

	factory := func(rng *rand.Rand) domain.MatchSimulator {
		return simulator.NewPoissonSimulator(rng)
	}

	mc := NewMonteCarlo(factory, 1000, 42)
	result, err := mc.Predict(context.Background(), teams, allUnplayed)
	require.NoError(t, err)

	t.Run("remaining matches has 12 entries", func(t *testing.T) {
		assert.Len(t, result.MatchOdds, 12)
	})

	t.Run("all match odds sum to 1.0", func(t *testing.T) {
		for _, mo := range result.MatchOdds {
			sum := mo.HomeWin + mo.Draw + mo.AwayWin
			assert.InDelta(t, 1.0, sum, 0.005, "match %d", mo.MatchID)
		}
	})
}
