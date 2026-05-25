package league

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/mtunaswe/insider-league/internal/domain"
)

type Service struct {
	teamRepo  domain.TeamRepository
	matchRepo domain.MatchRepository
	simulator domain.MatchSimulator
	rng       *rand.Rand
}

func NewService(teamRepo domain.TeamRepository, matchRepo domain.MatchRepository, simulator domain.MatchSimulator, rng *rand.Rand) *Service {
	return &Service{
		teamRepo:  teamRepo,
		matchRepo: matchRepo,
		simulator: simulator,
		rng:       rng,
	}
}

func (s *Service) Reset(ctx context.Context) error {
	if err := s.matchRepo.Truncate(ctx); err != nil {
		return fmt.Errorf("league reset: %w", err)
	}

	teams, err := s.teamRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("league reset: %w", err)
	}

	fixtures := GenerateFixtures(teams)
	if err := s.matchRepo.CreateBatch(ctx, fixtures); err != nil {
		return fmt.Errorf("league reset: %w", err)
	}
	return nil
}

func (s *Service) CurrentWeek(ctx context.Context) (int, error) {
	played := false
	matches, err := s.matchRepo.List(ctx, domain.MatchFilter{Played: &played})
	if err != nil {
		return 0, fmt.Errorf("league current week: %w", err)
	}

	if len(matches) == 0 {
		return domain.TotalWeeks + 1, nil
	}

	minWeek := matches[0].Week
	for _, m := range matches[1:] {
		if m.Week < minWeek {
			minWeek = m.Week
		}
	}
	return minWeek, nil
}

func (s *Service) Standings(ctx context.Context) ([]domain.StandingsRow, error) {
	teams, err := s.teamRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("league standings: %w", err)
	}

	played := true
	matches, err := s.matchRepo.List(ctx, domain.MatchFilter{Played: &played})
	if err != nil {
		return nil, fmt.Errorf("league standings: %w", err)
	}

	return domain.ComputeStandings(teams, matches), nil
}

type MatchResult struct {
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeGoals int    `json:"home_goals"`
	AwayGoals int    `json:"away_goals"`
}

type WeekSummary struct {
	Week    int           `json:"week"`
	Matches []MatchResult `json:"matches"`
}

func (s *Service) PlayNextWeek(ctx context.Context) (*WeekSummary, error) {
	week, err := s.CurrentWeek(ctx)
	if err != nil {
		return nil, err
	}
	if week > domain.TotalWeeks {
		return nil, ErrSeasonComplete
	}

	teams, err := s.teamRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("league play next week: %w", err)
	}
	teamMap := make(map[int]domain.Team, len(teams))
	for _, t := range teams {
		teamMap[t.ID] = t
	}

	matches, err := s.matchRepo.List(ctx, domain.MatchFilter{Week: &week})
	if err != nil {
		return nil, fmt.Errorf("league play next week: %w", err)
	}

	summary := &WeekSummary{Week: week}
	for _, m := range matches {
		if m.Played {
			continue
		}
		home := teamMap[m.HomeTeamID]
		away := teamMap[m.AwayTeamID]
		hg, ag := s.simulator.Simulate(home, away)

		if err := s.matchRepo.UpdateResult(ctx, m.ID, hg, ag); err != nil {
			return nil, fmt.Errorf("league play next week: %w", err)
		}

		summary.Matches = append(summary.Matches, MatchResult{
			HomeTeam:  home.Name,
			AwayTeam:  away.Name,
			HomeGoals: hg,
			AwayGoals: ag,
		})
	}

	return summary, nil
}

type PlayAllResult struct {
	Weeks      []WeekSummary         `json:"weeks"`
	FinalTable []domain.StandingsRow `json:"final_table"`
	Champion   string                `json:"champion"`
}

func (s *Service) PlayAll(ctx context.Context) (*PlayAllResult, error) {
	var summaries []WeekSummary
	for {
		summary, err := s.PlayNextWeek(ctx)
		if err == ErrSeasonComplete {
			break
		}
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, *summary)
	}

	standings, err := s.Standings(ctx)
	if err != nil {
		return nil, err
	}

	champion := ""
	if len(standings) > 0 {
		champion = standings[0].TeamName
	}

	return &PlayAllResult{
		Weeks:      summaries,
		FinalTable: standings,
		Champion:   champion,
	}, nil
}

func (s *Service) EditMatch(ctx context.Context, id int, homeGoals, awayGoals int) error {
	m, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("league edit match: %w", err)
	}
	if !m.Played {
		return ErrMatchNotPlayed
	}
	if err := s.matchRepo.UpdateResult(ctx, id, homeGoals, awayGoals); err != nil {
		return fmt.Errorf("league edit match: %w", err)
	}
	return nil
}
