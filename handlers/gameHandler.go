package handlers

import (
	"context"
	"leeg/model"
	"leeg/svc"
	"leeg/views/components"
	"leeg/views/components/forms"
	"log/slog"
	"net/http"
)

type GameHandler struct {
	service svc.LeegService
}

func (g GameHandler) HandleGetGame(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	roundID := r.PathValue("roundID")
	gameID := r.PathValue("gameID")
	editing := r.URL.Query().Get("editing") == "true"

	if leegID == "" || gameID == "" || roundID == "" {
		return hxRedirect(w, r, "/")
	}
	game, teams, err := g.service.GetGame(leegID, roundID, gameID)
	if err != nil {
		return err
	}
	nav := model.Nav{LeegID: leegID, RoundID: roundID}

	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)

	return Render(w, r.WithContext(ctx), components.Game(game, teams, editing, false))
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
	teamA := r.FormValue("teamA")
	teamB := r.FormValue("teamB")
	if (teamA != "" && teamB == "") || (teamB != "" && teamA == "") {
		slog.Error("either both teams must be referenced in update, or neither")
		return hxRedirect(w, r, "/")
	}

	if teamA == "" && winnerID == "" {
		slog.Error("either teamIDs or a winner must be referenced in update")
		return hxRedirect(w, r, "/")
	}

	var game model.Game
	var allTeams model.TeamList
	var updatedTeams []model.Team

	if winnerID != "" {
		game, allTeams, updatedTeams, err = g.service.ResolveGame(leegID, gameID, winnerID)
		if err != nil {
			return err
		}
	} else {
		game, allTeams, updatedTeams, err = g.service.RematchGame(leegID, roundID, gameID, teamA, teamB)
		if err != nil {
			return err
		}
	}

	nav := model.Nav{LeegID: leegID, RoundID: roundID}
	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)

	err = Render(w, r.WithContext(ctx), components.Game(game, allTeams.AsEntityList(), false, false))
	if err != nil {
		return err
	}

	for _, team := range updatedTeams {
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
		var teams model.EntityRefList
		if teamA == teamB {
			teams, err = g.service.GetTeams(leegID)
			if err != nil {
				return err
			}
			w.WriteHeader(http.StatusBadRequest)
			return Render(w, r.WithContext(ctx), forms.RecordGameForm(leegID, roundID, teams, teamA, teamB, map[string]string{"teamB": "a team can't play itself"}, false, false))
		}
		round, game, err = g.service.RecordMatchup(leegID, roundID, teamA, teamB)
		if err != nil {
			return err
		}
		return Render(w, r.WithContext(ctx), components.GameAndControls(game, round))
	}

	return Render(w, r.WithContext(ctx), components.GameAndControls(game, round))
}
