package handlers

import (
	"leeg/model"
	"leeg/svc"
	"leeg/views/pages"
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

	round, err := rh.service.GetRound(leegID, roundID)
	if err != nil {
		return err
	}
	if open {
		err = Render(w, r, pages.RoundContent(round, round.AsRef()))
		if err != nil {
			return err
		}
	} else {
		err := Render(w, r, pages.RoundContent(model.Round{}, round.AsRef()))
		if err != nil {
			return err
		}
	}

	return Render(w, r, pages.RoundHeader(leegID, round.AsRef(), open, true))
}
