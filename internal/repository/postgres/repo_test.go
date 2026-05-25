package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.SkipNow()
	}
	pool, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestTeamRepo_List(t *testing.T) {
	pool := testPool(t)
	repo := NewTeamRepo(pool)

	teams, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, teams)
}

func TestMatchRepo_CreateAndList(t *testing.T) {
	pool := testPool(t)
	repo := NewMatchRepo(pool)
	ctx := context.Background()

	_ = repo.Truncate(ctx)

	matches := []domain.Match{
		{Week: 1, HomeTeamID: 1, AwayTeamID: 2, Played: false},
		{Week: 1, HomeTeamID: 3, AwayTeamID: 4, Played: false},
	}
	err := repo.CreateBatch(ctx, matches)
	require.NoError(t, err)

	all, err := repo.List(ctx, domain.MatchFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 2)

	week := 1
	filtered, err := repo.List(ctx, domain.MatchFilter{Week: &week})
	require.NoError(t, err)
	assert.Len(t, filtered, 2)

	err = repo.UpdateResult(ctx, all[0].ID, 2, 1)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, all[0].ID)
	require.NoError(t, err)
	assert.True(t, updated.Played)
	assert.Equal(t, 2, *updated.HomeGoals)
	assert.Equal(t, 1, *updated.AwayGoals)

	_ = repo.Truncate(ctx)
}
