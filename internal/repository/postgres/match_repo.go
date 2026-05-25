package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mtunaswe/insider-league/internal/domain"
)

type MatchRepo struct {
	pool *pgxpool.Pool
}

func NewMatchRepo(pool *pgxpool.Pool) *MatchRepo {
	return &MatchRepo{pool: pool}
}

func (r *MatchRepo) List(ctx context.Context, filter domain.MatchFilter) ([]domain.Match, error) {
	query := "SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, played FROM matches"
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Week != nil {
		conditions = append(conditions, fmt.Sprintf("week = $%d", argIdx))
		args = append(args, *filter.Week)
		argIdx++
	}
	if filter.Played != nil {
		conditions = append(conditions, fmt.Sprintf("played = $%d", argIdx))
		args = append(args, *filter.Played)
		argIdx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY week, id"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("match repo list: %w", err)
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		var m domain.Match
		if err := rows.Scan(&m.ID, &m.Week, &m.HomeTeamID, &m.AwayTeamID, &m.HomeGoals, &m.AwayGoals, &m.Played); err != nil {
			return nil, fmt.Errorf("match repo list scan: %w", err)
		}
		matches = append(matches, m)
	}
	return matches, rows.Err()
}

func (r *MatchRepo) GetByID(ctx context.Context, id int) (domain.Match, error) {
	var m domain.Match
	err := r.pool.QueryRow(ctx,
		"SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, played FROM matches WHERE id = $1", id).
		Scan(&m.ID, &m.Week, &m.HomeTeamID, &m.AwayTeamID, &m.HomeGoals, &m.AwayGoals, &m.Played)
	if err != nil {
		return m, fmt.Errorf("match repo get by id: %w", err)
	}
	return m, nil
}

func (r *MatchRepo) CreateBatch(ctx context.Context, matches []domain.Match) error {
	if len(matches) == 0 {
		return nil
	}

	query := "INSERT INTO matches (week, home_team_id, away_team_id, home_goals, away_goals, played) VALUES "
	var placeholders []string
	var args []interface{}
	argIdx := 1

	for _, m := range matches {
		placeholders = append(placeholders,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5))
		args = append(args, m.Week, m.HomeTeamID, m.AwayTeamID, m.HomeGoals, m.AwayGoals, m.Played)
		argIdx += 6
	}

	query += strings.Join(placeholders, ", ")
	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("match repo create batch: %w", err)
	}
	return nil
}

func (r *MatchRepo) UpdateResult(ctx context.Context, id int, homeGoals, awayGoals int) error {
	tag, err := r.pool.Exec(ctx,
		"UPDATE matches SET home_goals = $1, away_goals = $2, played = true WHERE id = $3",
		homeGoals, awayGoals, id)
	if err != nil {
		return fmt.Errorf("match repo update result: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("match repo update result: match %d not found", id)
	}
	return nil
}

func (r *MatchRepo) Truncate(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "TRUNCATE matches RESTART IDENTITY")
	if err != nil {
		return fmt.Errorf("match repo truncate: %w", err)
	}
	return nil
}
