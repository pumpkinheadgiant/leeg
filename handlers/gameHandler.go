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
	game, err := g.service.GetGame(leegID, roundID, gameID)
	if err != nil {
		return err
	}
	nav := model.Nav{LeegID: leegID, RoundID: roundID}

	ctx := context.WithValue(r.Context(), model.ContextKey{}, nav)

	return Render(w, r.WithContext(ctx), components.Game(game, !editing))
}

func (g GameHandler) HandleGameUpdate(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	roundID := r.PathValue("roundID")
	gameID := r.PathValue("gameID")

	if leegID == "" || gameID == "" || roundID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}

	err := r.ParseForm()
	if err != nil {
		return err
	}
	winnerID := r.FormValue("winner")
	if winnerID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}
	game, teams, err := g.service.ResolveGame(leegID, gameID, winnerID)
	if err != nil {
		return err
	}
	nav := model.Nav{LeegID: leegID, RoundID: roundID}
	ctx := context.WithValue(r.Context(), model.ContextKey{}, nav)

	err = Render(w, r.WithContext(ctx), components.Game(game, false))
	if err != nil {
		return err
	}

	for _, team := range teams {
		err = Render(w, r.WithContext(ctx), components.Team(team, true))
		if err != nil {
			return err
		}
	}
	return nil
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

	nav := model.Nav{LeegID: leegID, RoundID: roundID}
	ctx := context.WithValue(r.Context(), model.ContextKey{}, nav)

	return Render(w, r.WithContext(ctx), components.GameAndControls(game, round))
}
