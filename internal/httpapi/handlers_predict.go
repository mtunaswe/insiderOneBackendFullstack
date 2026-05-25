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

	matches, err := h.matchRepo.List(ctx, domain.MatchFilter{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result, err := h.predictor.Predict(ctx, teams, matches)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	remaining := result.MatchOdds
	if remaining == nil {
		remaining = []domain.MatchOdds{}
	}

	writeJSON(w, http.StatusOK, PredictionsResponse{
		ChampionshipOdds: result.ChampionshipOdds,
		RemainingMatches: remaining,
		Iterations:       result.Iterations,
	})
}
