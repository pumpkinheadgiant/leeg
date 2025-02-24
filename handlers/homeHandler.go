package handlers

import (
	"net/http"

	"phg.com/leeg/svc"
	"phg.com/leeg/views/pages"
)

type HomeHandler struct {
	service svc.LeegService
}

func (h HomeHandler) HandleGetHome(w http.ResponseWriter, r *http.Request) error {

	leegs, err := h.service.GetLeegs()
	if err != nil {
		return err
	}

	return Render(w, r, pages.HomePage(leegs))
}
