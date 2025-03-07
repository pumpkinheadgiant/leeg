package handlers

import (
	"context"
	"leeg/model"
	"leeg/svc"
	"leeg/views/components"
	"net/http"
)

type GameHandler struct {
	service svc.LeegService
}

func (g GameHandler) HandleGetGame(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	roundID := r.PathValue("roundID")
	gameID := r.PathValue("gameID")
	editing := r.URL.Query().Get("edit") == "true"

	if leegID == "" || gameID == "" || roundID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}
	round, game, err := g.service.GetGame(leegID, roundID, gameID)
	if err != nil {
		return err
	}
	ctx := context.WithValue(r.Context(), model.ContextKey{}, leegID)

	return Render(w, r.WithContext(ctx), components.Game(round.AsRef(), game, !editing))
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
