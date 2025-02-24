package handlers

import (
	"net/http"
	"strconv"

	"phg.com/leeg/model"
	"phg.com/leeg/svc"
	"phg.com/leeg/views/components/forms"
	"phg.com/leeg/views/pages"
)

type LeegHandler struct {
	service svc.LeegService
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
	return Render(w, r, forms.LeegForm(model.LeegCreateRequest{TeamDescriptor: "Team"}, map[string]string{}, true, true))
}
