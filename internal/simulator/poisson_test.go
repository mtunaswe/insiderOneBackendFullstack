package simulator

import (
	"math/rand"
	"testing"

	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestEqualStrengthTeams(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	sim := NewPoissonSimulator(rng)

	home := domain.Team{ID: 1, Name: "A", Strength: 75}
	away := domain.Team{ID: 2, Name: "B", Strength: 75}

	const n = 10000
	homeWins, awayWins := 0, 0
	for i := 0; i < n; i++ {
		hg, ag := sim.Simulate(home, away)
		if hg > ag {
			homeWins++
		} else if ag > hg {
			awayWins++
		}
	}

	homeRate := float64(homeWins) / float64(n)
	awayRate := float64(awayWins) / float64(n)
	// With home advantage, home wins slightly more, but roughly balanced
	// Each side should win between 25% and 55%
	assert.InDelta(t, homeRate, awayRate, 0.15, "win rates should be roughly balanced for equal teams")
}

func TestStrongerTeamWinsMore(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	sim := NewPoissonSimulator(rng)

	chelsea := domain.Team{ID: 1, Name: "Chelsea", Strength: 85}
	liverpool := domain.Team{ID: 2, Name: "Liverpool", Strength: 60}

	const n = 10000
	chelseaWins, liverpoolWins := 0, 0
	for i := 0; i < n; i++ {
		hg, ag := sim.Simulate(chelsea, liverpool)
		if hg > ag {
			chelseaWins++
		} else if ag > hg {
			liverpoolWins++
		}
	}

	assert.Greater(t, chelseaWins, liverpoolWins, "stronger team should win more often")
}

func TestAverageGoalsInRange(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	sim := NewPoissonSimulator(rng)

	home := domain.Team{ID: 1, Name: "A", Strength: 75}
	away := domain.Team{ID: 2, Name: "B", Strength: 70}

	const n = 10000
	totalGoals := 0
	for i := 0; i < n; i++ {
		hg, ag := sim.Simulate(home, away)
		totalGoals += hg + ag
	}

	avg := float64(totalGoals) / float64(n)
	assert.Greater(t, avg, 2.0, "average goals should be at least 2.0")
	assert.Less(t, avg, 3.5, "average goals should be at most 3.5")
}
