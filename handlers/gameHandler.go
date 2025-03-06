package handlers

import (
	"leeg/svc"
	"leeg/views/components"
	"net/http"
)

type GameHandler struct {
	service svc.LeegService
}

func (g GameHandler) HandleGameCreationRequest(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	if leegID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}

	roundID := r.PathValue("roundID")
	if roundID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}

	round, game, err := g.service.CreateRandomGame(leegID, roundID)
	if err != nil {
		return err
	}

	return Render(w, r, components.GameAndControls(round.AsRef(), game, round))
}
