package handlers

import (
	"leeg/svc"
	"leeg/views/pages"
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

	round, _, err := g.service.CreateRandomGame(leegID, roundID)
	if err != nil {
		return err
	}

	return Render(w, r, pages.Round(round))
}
