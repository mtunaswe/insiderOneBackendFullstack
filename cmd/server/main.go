package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mtunaswe/insider-league/internal/config"
	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/mtunaswe/insider-league/internal/httpapi"
	"github.com/mtunaswe/insider-league/internal/league"
	"github.com/mtunaswe/insider-league/internal/predictor"
	"github.com/mtunaswe/insider-league/internal/repository/postgres"
	"github.com/mtunaswe/insider-league/internal/simulator"
	"github.com/mtunaswe/insider-league/pkg/db"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	runMigrations(cfg.DatabaseURL())

	teamRepo := postgres.NewTeamRepo(pool)
	matchRepo := postgres.NewMatchRepo(pool)

	seed := parseSeed(cfg.SimSeed)
	rng := rand.New(rand.NewSource(seed))

	sim := simulator.NewPoissonSimulator(rng)
	svc := league.NewService(teamRepo, matchRepo, sim, rng)

	iterations := 10000
	if v := os.Getenv("PREDICT_ITERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			iterations = n
		}
	}
	simFactory := func(r *rand.Rand) domain.MatchSimulator {
		return simulator.NewPoissonSimulator(r)
	}
	pred := predictor.NewMonteCarlo(simFactory, iterations, seed)

	leagueH := httpapi.NewLeagueHandler(svc)
	matchH := httpapi.NewMatchHandler(svc, matchRepo)
	predictH := httpapi.NewPredictHandler(pred, teamRepo, matchRepo)

	router := httpapi.NewRouter(leagueH, matchH, predictH)

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func runMigrations(dbURL string) {
	m, err := migrate.New("file:///migrations", dbURL)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations applied successfully")
}

func parseSeed(s string) int64 {
	if s == "" || s == "0" {
		return time.Now().UnixNano()
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Now().UnixNano()
	}
	return n
}
