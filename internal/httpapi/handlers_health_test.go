package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

type mockLeagueStater struct {
	standings   []domain.StandingsRow
	currentWeek int
	err         error
}

func (m *mockLeagueStater) CurrentWeek(_ context.Context) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.currentWeek, nil
}

func (m *mockLeagueStater) Standings(_ context.Context) ([]domain.StandingsRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.standings, nil
}

func TestHealthHandler_HappyPath(t *testing.T) {
	pinger := &mockPinger{}
	leagueState := &mockLeagueStater{
		currentWeek: 3,
		standings: []domain.StandingsRow{
			{TeamID: 1, TeamName: "Chelsea", P: 4},
			{TeamID: 2, TeamName: "Manchester City", P: 4},
			{TeamID: 3, TeamName: "Arsenal", P: 4},
			{TeamID: 4, TeamName: "Liverpool", P: 4},
		},
	}

	h := NewHealthHandler(pinger, leagueState, BuildInfo{
		Version:   "0.1.0",
		Commit:    "abc1234",
		BuildTime: "2026-05-25T12:00:00Z",
	}, AppConfig{
		PredictIterations: 10000,
		SimSeed:           42,
	}, time.Now().Add(-10*time.Second))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp healthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "0.1.0", resp.Version)
	assert.Equal(t, "abc1234", resp.Commit)
	assert.Equal(t, "2026-05-25T12:00:00Z", resp.BuildTime)
	assert.Greater(t, resp.Uptime, 0.0)
	assert.Equal(t, "ok", resp.Database.Status)
	assert.GreaterOrEqual(t, resp.Database.PingUs, int64(0))
	assert.NotNil(t, resp.League)
	assert.Equal(t, 4, resp.League.Teams)
	assert.Equal(t, 3, resp.League.CurrentWeek)
	assert.Equal(t, 6, resp.League.TotalWeeks)
	assert.Equal(t, 8, resp.League.MatchesPlayed)
	assert.Equal(t, 12, resp.League.MatchesTotal)
	assert.False(t, resp.League.SeasonComplete)
	assert.Equal(t, 10000, resp.Config.PredictIterations)
	assert.Equal(t, int64(42), resp.Config.SimSeed)
}

func TestHealthHandler_SeasonComplete(t *testing.T) {
	pinger := &mockPinger{}
	leagueState := &mockLeagueStater{
		currentWeek: 7,
		standings: []domain.StandingsRow{
			{TeamID: 1, TeamName: "Chelsea", P: 6},
			{TeamID: 2, TeamName: "Manchester City", P: 6},
			{TeamID: 3, TeamName: "Arsenal", P: 6},
			{TeamID: 4, TeamName: "Liverpool", P: 6},
		},
	}

	h := NewHealthHandler(pinger, leagueState, BuildInfo{}, AppConfig{}, time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp healthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, 7, resp.League.CurrentWeek)
	assert.Equal(t, 12, resp.League.MatchesPlayed)
	assert.Equal(t, 12, resp.League.MatchesTotal)
	assert.True(t, resp.League.SeasonComplete)
}

func TestHealthHandler_DBDown(t *testing.T) {
	pinger := &mockPinger{err: errors.New("connection refused")}
	leagueState := &mockLeagueStater{}

	h := NewHealthHandler(pinger, leagueState, BuildInfo{
		Version: "0.1.0",
		Commit:  "abc1234",
	}, AppConfig{
		PredictIterations: 10000,
		SimSeed:           42,
	}, time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp healthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "degraded", resp.Status)
	assert.Equal(t, "error", resp.Database.Status)
	assert.Equal(t, "connection refused", resp.Database.Error)
	assert.Nil(t, resp.League)
}

func TestHealthHandler_UptimeMonotonic(t *testing.T) {
	pinger := &mockPinger{}
	leagueState := &mockLeagueStater{
		currentWeek: 1,
		standings: []domain.StandingsRow{
			{TeamID: 1, P: 0},
			{TeamID: 2, P: 0},
			{TeamID: 3, P: 0},
			{TeamID: 4, P: 0},
		},
	}

	h := NewHealthHandler(pinger, leagueState, BuildInfo{}, AppConfig{}, time.Now().Add(-5*time.Second))

	req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec1 := httptest.NewRecorder()
	h.Health(rec1, req1)

	var resp1 healthResponse
	require.NoError(t, json.NewDecoder(rec1.Body).Decode(&resp1))

	time.Sleep(10 * time.Millisecond)

	req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec2 := httptest.NewRecorder()
	h.Health(rec2, req2)

	var resp2 healthResponse
	require.NoError(t, json.NewDecoder(rec2.Body).Decode(&resp2))

	assert.GreaterOrEqual(t, resp2.Uptime, resp1.Uptime)
}

func TestReadyHandler_OK(t *testing.T) {
	pinger := &mockPinger{}
	h := NewHealthHandler(pinger, &mockLeagueStater{}, BuildInfo{}, AppConfig{}, time.Now())

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestReadyHandler_DBDown(t *testing.T) {
	pinger := &mockPinger{err: errors.New("timeout")}
	h := NewHealthHandler(pinger, &mockLeagueStater{}, BuildInfo{}, AppConfig{}, time.Now())

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}
