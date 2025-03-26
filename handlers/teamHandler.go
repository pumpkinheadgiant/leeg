package handlers

import (
	"context"
	"leeg/model"
	"leeg/svc"
	"leeg/views/components"
	"leeg/views/components/forms"
	"net/http"
	"strings"
)

type TeamHandler struct {
	service svc.LeegService
}

func (t TeamHandler) HandleTeamUpdate(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	teamID := r.PathValue("teamID")

	if leegID == "" || teamID == "" {
		return hxRedirect(w, r, "/")
	}
	err := r.ParseForm()
	if err != nil {
		return err
	}
	name := strings.TrimSpace(r.FormValue("name"))
	nav := model.Nav{LeegID: leegID}
	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)

	teamRequest := model.TeamUpdateRequest{LeegID: leegID, TeamID: teamID, Name: name}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		errors := map[string]string{"name": "name cannot be empty"}
		return Render(w, r.WithContext(ctx), forms.TeamForm(teamRequest, errors, false, false))
	}

	team, games, activeRound, nameAvailable, err := t.service.RenameTeam(teamRequest)
	if err != nil {
		return err
	}
	if !nameAvailable {
		w.WriteHeader(http.StatusBadRequest)
		errors := map[string]string{"name": "name is in use"}
		return Render(w, r.WithContext(ctx), forms.TeamForm(teamRequest, errors, false, false))
	}
	err = Render(w, r.WithContext(ctx), components.Team(team, false))
	if err != nil {
		return err
	}
	for _, game := range games {
		err = Render(w, r.WithContext(ctx), components.Game(game, activeRound.AllTeams, false, true))
		if err != nil {
			return err
		}
	}

	return Render(w, r.WithContext(ctx), forms.RecordGameForm(leegID, activeRound.ID, activeRound.AllTeams, "", "", map[string]string{}, true, true))
}
