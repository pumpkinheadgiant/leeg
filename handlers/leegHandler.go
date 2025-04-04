package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"leeg/model"
	"leeg/svc"
	"leeg/views/components/forms"
	"leeg/views/pages"
)

type LeegHandler struct {
	service svc.LeegService
}

func (l LeegHandler) HandleGetLeeg(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	if leegID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}
	leeg, err := l.service.GetLeeg(leegID)
	if err != nil {
		return err
	}
	nav := model.Nav{LeegID: leegID}
	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)

	return Render(w, r.WithContext(ctx), pages.LeegPage(leeg))
}

func (l LeegHandler) HandleCopyLeeg(w http.ResponseWriter, r *http.Request) error {
	leegID := r.PathValue("leegID")
	if leegID == "" {
		w.WriteHeader(http.StatusNotFound)
		return hxRedirect(w, r, "/")
	}

	newLeeg, err := l.service.CopyLeeg(leegID)
	if err != nil {
		return err
	}
	nav := model.Nav{LeegID: leegID}
	ctx := context.WithValue(r.Context(), model.NavContextKey{}, nav)
	return hxRedirect(w, r.WithContext(ctx), fmt.Sprintf("/leegs/%v", newLeeg.ID))

}

func (l LeegHandler) HandlePostLeeg(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	name := r.FormValue("name")
	descriptor := r.FormValue("teamDescriptor")
	teamCount := 0
	teamCountString := r.FormValue("teamCount")
	if teamCountString != "" {
		teamCount, err = strconv.Atoi(teamCountString)
		if err != nil {
			return err
		}
	}
	roundCount := 0
	roundCountString := r.FormValue("roundCount")
	if roundCountString != "" {
		roundCount, err = strconv.Atoi(roundCountString)
		if err != nil {
			return err
		}
	}
	createRequest := model.LeegCreateRequest{
		Name:           name,
		TeamCount:      teamCount,
		TeamDescriptor: descriptor,
		RoundCount:     roundCount,
	}
	errors := createRequest.ValidateAndNormalize()
	if len(errors) > 0 {
		w.Header().Set("HX-Reswap", "outerHTML")
		w.WriteHeader(http.StatusBadRequest)
		return Render(w, r, forms.LeegForm(createRequest, errors, false, false))
	}
	leegRef, err := l.service.CreateLeeg(createRequest)
	if err != nil {
		return err
	}

	err = Render(w, r, pages.LeegLink(leegRef))
	if err != nil {
		return err
	}
	return Render(w, r, forms.LeegForm(model.LeegCreateRequest{TeamDescriptor: "Team", TeamCount: 4, RoundCount: 3}, map[string]string{}, true, true))
}
