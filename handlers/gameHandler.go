package handlers

import (
	"leeg/svc"
	"net/http"
)

type GameHandler struct {
	service svc.LeegService
}

func (g GameHandler) HandlePostGameRequest(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	if leegID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}
	var roundNumber = 0

	roundNumberValue := r.PathValue("roundNumber")
	if roundNumberValue == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}

	_, _, err := g.service.CreateRandomGame(leegID, roundNumber)
	if err != nil {
		return err
	}

	return nil
}
