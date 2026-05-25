package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mtunaswe/insider-league/internal/domain"
)

type TeamRepo struct {
	pool *pgxpool.Pool
}

func NewTeamRepo(pool *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{pool: pool}
}

func (r *TeamRepo) List(ctx context.Context) ([]domain.Team, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, name, strength FROM teams ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("team repo list: %w", err)
	}
	defer rows.Close()

	var teams []domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Strength); err != nil {
			return nil, fmt.Errorf("team repo list scan: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *TeamRepo) GetByID(ctx context.Context, id int) (domain.Team, error) {
	var t domain.Team
	err := r.pool.QueryRow(ctx, "SELECT id, name, strength FROM teams WHERE id = $1", id).
		Scan(&t.ID, &t.Name, &t.Strength)
	if err != nil {
		return t, fmt.Errorf("team repo get by id: %w", err)
	}
	return t, nil
}
