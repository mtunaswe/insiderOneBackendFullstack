package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/mtunaswe/insider-league/internal/league"
)

type MatchHandler struct {
	svc       *league.Service
	matchRepo domain.MatchRepository
}

func NewMatchHandler(svc *league.Service, matchRepo domain.MatchRepository) *MatchHandler {
	return &MatchHandler{svc: svc, matchRepo: matchRepo}
}

func (h *MatchHandler) List(w http.ResponseWriter, r *http.Request) {
	var filter domain.MatchFilter

	if weekStr := r.URL.Query().Get("week"); weekStr != "" {
		week, err := strconv.Atoi(weekStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid week parameter")
			return
		}
		filter.Week = &week
	}

	if playedStr := r.URL.Query().Get("played"); playedStr != "" {
		played, err := strconv.ParseBool(playedStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid played parameter")
			return
		}
		filter.Played = &played
	}

	matches, err := h.matchRepo.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if matches == nil {
		matches = []domain.Match{}
	}
	writeJSON(w, http.StatusOK, MatchListResponse{Matches: matches})
}

func (h *MatchHandler) EditResult(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	var req EditMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.HomeGoals == nil || req.AwayGoals == nil {
		writeError(w, http.StatusBadRequest, "home_goals and away_goals are required")
		return
	}

	if *req.HomeGoals < 0 || *req.HomeGoals > 20 || *req.AwayGoals < 0 || *req.AwayGoals > 20 {
		writeError(w, http.StatusBadRequest, "goals must be between 0 and 20")
		return
	}

	err = h.svc.EditMatch(r.Context(), id, *req.HomeGoals, *req.AwayGoals)
	if err != nil {
		if errors.Is(err, league.ErrMatchNotPlayed) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "match not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "match updated successfully"})
}
