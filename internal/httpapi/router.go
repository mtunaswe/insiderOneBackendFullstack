package httpapi

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mtunaswe/insider-league/web"
)

func NewRouter(leagueH *LeagueHandler, matchH *MatchHandler, predictH *PredictHandler, healthH *HealthHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", healthH.Health)

	r.Post("/league/reset", leagueH.Reset)
	r.Get("/league/table", leagueH.Table)
	r.Get("/league/week", leagueH.Week)
	r.Post("/league/next-week", leagueH.NextWeek)
	r.Post("/league/play-all", leagueH.PlayAll)

	r.Get("/matches", matchH.List)
	r.Put("/matches/{id}", matchH.EditResult)

	r.Get("/predictions", predictH.Predictions)

	staticFS, _ := fs.Sub(web.StaticFS, "static")
	fileServer := http.FileServer(http.FS(staticFS))

	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len("/assets"):]
		fileServer.ServeHTTP(w, r)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		index, _ := fs.ReadFile(staticFS, "index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
	})

	return r
}
