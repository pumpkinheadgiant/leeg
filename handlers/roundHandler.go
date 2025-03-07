package handlers

import (
	"context"
	"fmt"
	"leeg/model"
	"leeg/svc"
	"leeg/views/components"
	"net/http"
)

type RoundHandler struct {
	service svc.LeegService
}

func (rh RoundHandler) HandleGetRound(w http.ResponseWriter, r *http.Request) error {
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
	open := r.URL.Query().Get("open") == "true"
	ctx := context.WithValue(r.Context(), model.ContextKey{}, leegID)

	round, games, err := rh.service.GetRound(leegID, roundID)
	if err != nil {
		return err
	}
	if open {
		if round.IsActive || round.Scheduled() {
			err = Render(w, r.WithContext(ctx), components.RoundContent(round, round.AsRef(), games))
			if err != nil {
				return err
			}
			return Render(w, r.WithContext(ctx), components.RoundHeader(leegID, round.AsRef(), open, true))
		} else {
			w.Header().Set("Leeg-Message", fmt.Sprintf("Round %v is not yet active", round.RoundNumber))
			w.Header().Set("Leeg-Status", "gray")
			return Render(w, r.WithContext(ctx), components.RoundContent(model.Round{}, round.AsRef(), map[string]model.Game{}))
		}
	} else {
		err := Render(w, r.WithContext(ctx), components.RoundContent(model.Round{}, round.AsRef(), map[string]model.Game{}))
		if err != nil {
			return err
		}
		return Render(w, r.WithContext(ctx), components.RoundHeader(leegID, round.AsRef(), open, true))
	}

}
