package league

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/mtunaswe/insider-league/internal/simulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memTeamRepo struct {
	teams []domain.Team
}

func (r *memTeamRepo) List(_ context.Context) ([]domain.Team, error) {
	return r.teams, nil
}

func (r *memTeamRepo) GetByID(_ context.Context, id int) (domain.Team, error) {
	for _, t := range r.teams {
		if t.ID == id {
			return t, nil
		}
	}
	return domain.Team{}, fmt.Errorf("team %d not found", id)
}

type memMatchRepo struct {
	matches []domain.Match
	nextID  int
}

func (r *memMatchRepo) List(_ context.Context, filter domain.MatchFilter) ([]domain.Match, error) {
	var result []domain.Match
	for _, m := range r.matches {
		if filter.Week != nil && m.Week != *filter.Week {
			continue
		}
		if filter.Played != nil && m.Played != *filter.Played {
			continue
		}
		result = append(result, m)
	}
	return result, nil
}

func (r *memMatchRepo) GetByID(_ context.Context, id int) (domain.Match, error) {
	for _, m := range r.matches {
		if m.ID == id {
			return m, nil
		}
	}
	return domain.Match{}, fmt.Errorf("match %d not found", id)
}

func (r *memMatchRepo) CreateBatch(_ context.Context, matches []domain.Match) error {
	for _, m := range matches {
		r.nextID++
		m.ID = r.nextID
		r.matches = append(r.matches, m)
	}
	return nil
}

func (r *memMatchRepo) UpdateResult(_ context.Context, id int, homeGoals, awayGoals int) error {
	for i := range r.matches {
		if r.matches[i].ID == id {
			r.matches[i].HomeGoals = &homeGoals
			r.matches[i].AwayGoals = &awayGoals
			r.matches[i].Played = true
			return nil
		}
	}
	return fmt.Errorf("match %d not found", id)
}

func (r *memMatchRepo) Truncate(_ context.Context) error {
	r.matches = nil
	return nil
}

func TestEditMatchChangesStandings(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "Chelsea", Strength: 85},
		{ID: 2, Name: "Manchester City", Strength: 80},
		{ID: 3, Name: "Arsenal", Strength: 70},
		{ID: 4, Name: "Liverpool", Strength: 60},
	}

	teamRepo := &memTeamRepo{teams: teams}
	matchRepo := &memMatchRepo{}
	rng := rand.New(rand.NewSource(42))
	sim := simulator.NewPoissonSimulator(rng)

	svc := NewService(teamRepo, matchRepo, sim, rng)
	ctx := context.Background()

	// Reset to generate fixtures
	err := svc.Reset(ctx)
	require.NoError(t, err)

	// Play first week
	_, err = svc.PlayNextWeek(ctx)
	require.NoError(t, err)

	// Get standings before edit
	standingsBefore, err := svc.Standings(ctx)
	require.NoError(t, err)

	// Find a played match and change 3-2 → 0-5 (massive swing)
	played := true
	playedMatches, err := matchRepo.List(ctx, domain.MatchFilter{Played: &played})
	require.NoError(t, err)
	require.NotEmpty(t, playedMatches)

	matchToEdit := playedMatches[0]

	// First set the match to a known score
	err = matchRepo.UpdateResult(ctx, matchToEdit.ID, 3, 2)
	require.NoError(t, err)

	standingsAfterSet, err := svc.Standings(ctx)
	require.NoError(t, err)

	// Now edit it to 0-5 (reversing the winner)
	err = svc.EditMatch(ctx, matchToEdit.ID, 0, 5)
	require.NoError(t, err)

	standingsAfterEdit, err := svc.Standings(ctx)
	require.NoError(t, err)

	// The standings should have changed
	assert.NotEqual(t, standingsAfterSet, standingsAfterEdit, "standings should change after editing a match")

	// The team that was away (originally lost 2-3) should now have won (5-0)
	awayTeamID := matchToEdit.AwayTeamID
	var awayRowBefore, awayRowAfter domain.StandingsRow
	for _, r := range standingsBefore {
		if r.TeamID == awayTeamID {
			awayRowBefore = r
		}
	}
	for _, r := range standingsAfterEdit {
		if r.TeamID == awayTeamID {
			awayRowAfter = r
		}
	}

	// After edit, away team scored 5 goals
	assert.Equal(t, 5, awayRowAfter.GF)
	assert.Equal(t, 0, awayRowAfter.GA)
	_ = awayRowBefore
}

func TestEditMatchNotPlayedReturnsError(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "A", Strength: 80},
		{ID: 2, Name: "B", Strength: 70},
		{ID: 3, Name: "C", Strength: 60},
		{ID: 4, Name: "D", Strength: 50},
	}

	teamRepo := &memTeamRepo{teams: teams}
	matchRepo := &memMatchRepo{}
	rng := rand.New(rand.NewSource(42))
	sim := simulator.NewPoissonSimulator(rng)

	svc := NewService(teamRepo, matchRepo, sim, rng)
	ctx := context.Background()

	err := svc.Reset(ctx)
	require.NoError(t, err)

	// Try to edit an unplayed match
	unplayed := false
	unplayedMatches, err := matchRepo.List(ctx, domain.MatchFilter{Played: &unplayed})
	require.NoError(t, err)
	require.NotEmpty(t, unplayedMatches)

	err = svc.EditMatch(ctx, unplayedMatches[0].ID, 1, 1)
	assert.ErrorIs(t, err, ErrMatchNotPlayed)
}
