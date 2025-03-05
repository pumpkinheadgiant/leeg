package handlers

import (
	"net/http"

	"leeg/svc"
	"leeg/views/pages"
)

type HomeHandler struct {
	services svc.LeegService
}

func (h HomeHandler) HandleGetHome(w http.ResponseWriter, r *http.Request) error {

	leegs, err := h.services.GetLeegs()
	if err != nil {
		return err
	}

	return Render(w, r, pages.HomePage(leegs))
}
