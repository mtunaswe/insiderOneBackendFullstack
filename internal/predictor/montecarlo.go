package predictor

import (
	"context"
	"math/rand"
	"sort"

	"github.com/mtunaswe/insider-league/internal/domain"
)

type SimulatorFactory func(rng *rand.Rand) domain.MatchSimulator

type MonteCarlo struct {
	simFactory SimulatorFactory
	iterations int
	seed       int64
}

func NewMonteCarlo(simFactory SimulatorFactory, iterations int, seed int64) domain.Predictor {
	return &MonteCarlo{
		simFactory: simFactory,
		iterations: iterations,
		seed:       seed,
	}
}

func (mc *MonteCarlo) Predict(ctx context.Context, teams []domain.Team, matches []domain.Match) (domain.PredictionResult, error) {
	teamMap := make(map[int]domain.Team, len(teams))
	for _, t := range teams {
		teamMap[t.ID] = t
	}

	var played, remaining []domain.Match
	for _, m := range matches {
		if m.Played {
			played = append(played, m)
		} else {
			remaining = append(remaining, m)
		}
	}

	standings := domain.ComputeStandings(teams, played)

	wins := make(map[int]float64, len(teams))

	type matchStats struct {
		homeWins  int
		draws     int
		awayWins  int
		homeGoals int
		awayGoals int
	}
	matchAccum := make(map[int]*matchStats, len(remaining))
	for _, m := range remaining {
		matchAccum[m.ID] = &matchStats{}
	}

	for i := 0; i < mc.iterations; i++ {
		rng := rand.New(rand.NewSource(mc.seed + int64(i)))
		sim := mc.simFactory(rng)

		simulated := make([]domain.Match, len(remaining))
		for j, m := range remaining {
			home := teamMap[m.HomeTeamID]
			away := teamMap[m.AwayTeamID]
			hg, ag := sim.Simulate(home, away)
			simulated[j] = domain.Match{
				ID:         m.ID,
				Week:       m.Week,
				HomeTeamID: m.HomeTeamID,
				AwayTeamID: m.AwayTeamID,
				HomeGoals:  &hg,
				AwayGoals:  &ag,
				Played:     true,
			}

			stats := matchAccum[m.ID]
			stats.homeGoals += hg
			stats.awayGoals += ag
			if hg > ag {
				stats.homeWins++
			} else if hg == ag {
				stats.draws++
			} else {
				stats.awayWins++
			}
		}

		finalStandings := addResultsToStandings(standings, simulated)

		topPts := finalStandings[0].Pts
		topGD := finalStandings[0].GD
		topGF := finalStandings[0].GF

		var tiedAtTop []int
		for _, row := range finalStandings {
			if row.Pts == topPts && row.GD == topGD && row.GF == topGF {
				tiedAtTop = append(tiedAtTop, row.TeamID)
			} else {
				break
			}
		}

		share := 1.0 / float64(len(tiedAtTop))
		for _, id := range tiedAtTop {
			wins[id] += share
		}
	}

	odds := make([]domain.ChampionshipOdds, 0, len(teams))
	for _, t := range teams {
		odds = append(odds, domain.ChampionshipOdds{
			TeamID:      t.ID,
			TeamName:    t.Name,
			Probability: wins[t.ID] / float64(mc.iterations),
		})
	}
	sort.Slice(odds, func(i, j int) bool {
		return odds[i].Probability > odds[j].Probability
	})

	n := float64(mc.iterations)
	matchOdds := make([]domain.MatchOdds, 0, len(remaining))
	for _, m := range remaining {
		stats := matchAccum[m.ID]
		matchOdds = append(matchOdds, domain.MatchOdds{
			MatchID:           m.ID,
			Week:              m.Week,
			HomeName:          teamMap[m.HomeTeamID].Name,
			AwayName:          teamMap[m.AwayTeamID].Name,
			HomeWin:           float64(stats.homeWins) / n,
			Draw:              float64(stats.draws) / n,
			AwayWin:           float64(stats.awayWins) / n,
			ExpectedHomeGoals: float64(stats.homeGoals) / n,
			ExpectedAwayGoals: float64(stats.awayGoals) / n,
		})
	}
	sort.Slice(matchOdds, func(i, j int) bool {
		if matchOdds[i].Week != matchOdds[j].Week {
			return matchOdds[i].Week < matchOdds[j].Week
		}
		return matchOdds[i].MatchID < matchOdds[j].MatchID
	})

	return domain.PredictionResult{
		ChampionshipOdds: odds,
		MatchOdds:        matchOdds,
		Iterations:       mc.iterations,
	}, nil
}

func addResultsToStandings(current []domain.StandingsRow, simulated []domain.Match) []domain.StandingsRow {
	rowMap := make(map[int]*domain.StandingsRow, len(current))
	for _, r := range current {
		c := r
		rowMap[r.TeamID] = &c
	}

	for _, m := range simulated {
		home := rowMap[m.HomeTeamID]
		away := rowMap[m.AwayTeamID]
		hg, ag := *m.HomeGoals, *m.AwayGoals

		home.P++
		away.P++
		home.GF += hg
		home.GA += ag
		away.GF += ag
		away.GA += hg

		switch {
		case hg > ag:
			home.W++
			home.Pts += domain.PointsWin
			away.L++
		case hg < ag:
			away.W++
			away.Pts += domain.PointsWin
			home.L++
		default:
			home.D++
			away.D++
			home.Pts += domain.PointsDraw
			away.Pts += domain.PointsDraw
		}
	}

	rows := make([]domain.StandingsRow, 0, len(current))
	for _, r := range rowMap {
		r.GD = r.GF - r.GA
		rows = append(rows, *r)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Pts != rows[j].Pts {
			return rows[i].Pts > rows[j].Pts
		}
		if rows[i].GD != rows[j].GD {
			return rows[i].GD > rows[j].GD
		}
		return rows[i].GF > rows[j].GF
	})

	return rows
}
