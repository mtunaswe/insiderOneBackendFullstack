package httpapi

import (
	"net/http"

	"github.com/mtunaswe/insider-league/internal/domain"
)

type PredictHandler struct {
	predictor domain.Predictor
	teamRepo  domain.TeamRepository
	matchRepo domain.MatchRepository
}

func NewPredictHandler(predictor domain.Predictor, teamRepo domain.TeamRepository, matchRepo domain.MatchRepository) *PredictHandler {
	return &PredictHandler{
		predictor: predictor,
		teamRepo:  teamRepo,
		matchRepo: matchRepo,
	}
}

func (h *PredictHandler) Predictions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teams, err := h.teamRepo.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	played := true
	playedMatches, err := h.matchRepo.List(ctx, domain.MatchFilter{Played: &played})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	standings := domain.ComputeStandings(teams, playedMatches)

	unplayed := false
	remaining, err := h.matchRepo.List(ctx, domain.MatchFilter{Played: &unplayed})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	odds, err := h.predictor.ChampionshipOdds(ctx, standings, remaining, teams)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, PredictionsResponse{Predictions: odds})
}
