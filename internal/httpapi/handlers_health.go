package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/mtunaswe/insider-league/internal/domain"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

type LeagueStater interface {
	CurrentWeek(ctx context.Context) (int, error)
	Standings(ctx context.Context) ([]domain.StandingsRow, error)
}

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

type AppConfig struct {
	PredictIterations int
	SimSeed           int64
}

type HealthHandler struct {
	db        DBPinger
	league    LeagueStater
	build     BuildInfo
	config    AppConfig
	startTime time.Time
}

func NewHealthHandler(db DBPinger, league LeagueStater, build BuildInfo, config AppConfig, startTime time.Time) *HealthHandler {
	return &HealthHandler{
		db:        db,
		league:    league,
		build:     build,
		config:    config,
		startTime: startTime,
	}
}

type healthResponse struct {
	Status    string         `json:"status"`
	Version   string         `json:"version"`
	Commit    string         `json:"commit"`
	BuildTime string         `json:"build_time"`
	Uptime    float64        `json:"uptime_seconds"`
	Database  databaseStatus `json:"database"`
	League    *leagueStatus  `json:"league,omitempty"`
	Config    configStatus   `json:"config"`
}

type databaseStatus struct {
	Status string `json:"status"`
	PingUs int64  `json:"ping_microseconds,omitempty"`
	Error  string `json:"error,omitempty"`
}

type leagueStatus struct {
	Teams          int  `json:"teams"`
	CurrentWeek    int  `json:"current_week"`
	TotalWeeks     int  `json:"total_weeks"`
	MatchesPlayed  int  `json:"matches_played"`
	MatchesTotal   int  `json:"matches_total"`
	SeasonComplete bool `json:"season_complete"`
}

type configStatus struct {
	PredictIterations int   `json:"predict_iterations"`
	SimSeed           int64 `json:"sim_seed"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp := healthResponse{
		Status:    "ok",
		Version:   h.build.Version,
		Commit:    h.build.Commit,
		BuildTime: h.build.BuildTime,
		Uptime:    time.Since(h.startTime).Seconds(),
		Config: configStatus{
			PredictIterations: h.config.PredictIterations,
			SimSeed:           h.config.SimSeed,
		},
	}

	pingStart := time.Now()
	if err := h.db.Ping(ctx); err != nil {
		resp.Status = "degraded"
		resp.Database = databaseStatus{
			Status: "error",
			Error:  err.Error(),
		}
		writeJSON(w, http.StatusServiceUnavailable, resp)
		return
	}
	resp.Database = databaseStatus{
		Status: "ok",
		PingUs: time.Since(pingStart).Microseconds(),
	}

	ls := h.getLeagueStatus(ctx)
	resp.League = ls

	writeJSON(w, http.StatusOK, resp)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *HealthHandler) getLeagueStatus(ctx context.Context) *leagueStatus {
	standings, err := h.league.Standings(ctx)
	if err != nil {
		return nil
	}

	currentWeek, err := h.league.CurrentWeek(ctx)
	if err != nil {
		return nil
	}

	matchesPlayed := 0
	for _, row := range standings {
		matchesPlayed += row.P
	}
	matchesPlayed /= 2

	return &leagueStatus{
		Teams:          len(standings),
		CurrentWeek:    currentWeek,
		TotalWeeks:     domain.TotalWeeks,
		MatchesPlayed:  matchesPlayed,
		MatchesTotal:   domain.TotalMatches,
		SeasonComplete: currentWeek > domain.TotalWeeks,
	}
}
