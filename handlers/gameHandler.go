package handlers

import (
	"context"
	"leeg/model"
	"leeg/svc"
	"leeg/views/components"
	"leeg/views/components/forms"
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
		return hxRedirect(w, r, "/")
	}
	game, err := g.service.GetGame(leegID, roundID, gameID)
	if err != nil {
		return err
	}
	nav := model.Nav{LeegID: leegID, RoundID: roundID}

	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)

	return Render(w, r.WithContext(ctx), components.Game(game, !editing, false))
}

func (g GameHandler) HandleGameUpdate(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	roundID := r.PathValue("roundID")
	gameID := r.PathValue("gameID")

	if leegID == "" || gameID == "" || roundID == "" {
		return hxRedirect(w, r, "/")
	}

	err := r.ParseForm()
	if err != nil {
		return err
	}
	winnerID := r.FormValue("winner")
	if winnerID == "" {
		return hxRedirect(w, r, "/")
	}
	game, teams, err := g.service.ResolveGame(leegID, gameID, winnerID)
	if err != nil {
		return err
	}
	nav := model.Nav{LeegID: leegID, RoundID: roundID}
	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)

	err = Render(w, r.WithContext(ctx), components.Game(game, false, false))
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
		return hxRedirect(w, r, "/")
	}

	roundID := r.PathValue("roundID")
	if roundID == "" {
		return hxRedirect(w, r, "/")
	}

	err := r.ParseForm()
	if err != nil {
		return err
	}
	var round model.Round
	var game model.Game

	nav := model.Nav{LeegID: leegID, RoundID: roundID}
	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)

	teamA := r.FormValue("teamA")
	teamB := r.FormValue("teamB")

	if teamA == "" {
		round, game, err = g.service.CreateRandomGame(leegID, roundID)
		if err != nil {
			return err
		}
	} else {
		if teamA == teamB {
			teams, err := g.service.GetTeams(leegID)
			if err != nil {
				return err
			}
			w.WriteHeader(http.StatusBadRequest)
			return Render(w, r.WithContext(ctx), forms.GameForm(leegID, roundID, teams, teamA, teamB, map[string]string{"teamB": "a team can't play itself"}, false, false))
		}
		round, game, err = g.service.RecordMatchup(leegID, roundID, teamA, teamB)
		if err != nil {
			return err
		}
	}

	return Render(w, r.WithContext(ctx), components.GameAndControls(game, round))
}
