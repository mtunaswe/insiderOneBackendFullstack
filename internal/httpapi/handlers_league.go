package httpapi

import (
	"errors"
	"net/http"

	"github.com/mtunaswe/insider-league/internal/league"
)

type LeagueHandler struct {
	svc *league.Service
}

func NewLeagueHandler(svc *league.Service) *LeagueHandler {
	return &LeagueHandler{svc: svc}
}

func (h *LeagueHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Reset(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "league reset successfully"})
}

func (h *LeagueHandler) Table(w http.ResponseWriter, r *http.Request) {
	standings, err := h.svc.Standings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, TableResponse{Table: standings})
}

func (h *LeagueHandler) Week(w http.ResponseWriter, r *http.Request) {
	week, err := h.svc.CurrentWeek(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, WeekResponse{
		CurrentWeek: week,
		TotalWeeks:  6,
	})
}

func (h *LeagueHandler) NextWeek(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.PlayNextWeek(r.Context())
	if err != nil {
		if errors.Is(err, league.ErrSeasonComplete) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *LeagueHandler) PlayAll(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.svc.PlayAll(r.Context())
	if err != nil {
		if errors.Is(err, league.ErrSeasonComplete) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}
