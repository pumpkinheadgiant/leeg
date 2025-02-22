package handlers

import (
	"net/http"

	"phg.com/leeg/svc"
	"phg.com/leeg/views/pages"
)

type HomeHandler struct {
	service svc.LeegData
}

func (h HomeHandler) HandleGetHome(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	// fake commit
	return Render(w, r.WithContext(ctx), pages.HomePage())
}
