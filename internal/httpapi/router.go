package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(leagueH *LeagueHandler, matchH *MatchHandler, predictH *PredictHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Post("/league/reset", leagueH.Reset)
	r.Get("/league/table", leagueH.Table)
	r.Get("/league/week", leagueH.Week)
	r.Post("/league/next-week", leagueH.NextWeek)
	r.Post("/league/play-all", leagueH.PlayAll)

	r.Get("/matches", matchH.List)
	r.Put("/matches/{id}", matchH.EditResult)

	r.Get("/predictions", predictH.Predictions)

	return r
}
