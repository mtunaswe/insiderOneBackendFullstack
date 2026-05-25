package simulator

import (
	"math"
	"math/rand"

	"github.com/mtunaswe/insider-league/internal/domain"
)

const (
	baseExpectedGoals      = 1.35
	homeAdvantageFactor    = 1.15
	awayDisadvantageFactor = 0.95
	maxGoals               = 9
)

type PoissonSimulator struct {
	rng *rand.Rand
}

func NewPoissonSimulator(rng *rand.Rand) domain.MatchSimulator {
	return &PoissonSimulator{rng: rng}
}

func (s *PoissonSimulator) Simulate(home, away domain.Team) (homeGoals, awayGoals int) {
	lambdaHome := baseExpectedGoals * (float64(home.Strength) / float64(away.Strength)) * homeAdvantageFactor
	lambdaAway := baseExpectedGoals * (float64(away.Strength) / float64(home.Strength)) * awayDisadvantageFactor

	homeGoals = clamp(poisson(s.rng, lambdaHome), 0, maxGoals)
	awayGoals = clamp(poisson(s.rng, lambdaAway), 0, maxGoals)
	return homeGoals, awayGoals
}

func poisson(rng *rand.Rand, lambda float64) int {
	l := math.Exp(-lambda)
	k := 0
	p := 1.0
	for {
		k++
		p *= rng.Float64()
		if p <= l {
			break
		}
	}
	return k - 1
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
