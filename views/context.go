package views

import (
	"context"
	"leeg/model"
)

func LeegID(ctx context.Context) string {
	if nav, exists := ctx.Value(model.NavContextKey{}).(model.Nav); exists {
		return nav.LeegID
	} else {
		return ""
	}
}

func RoundID(ctx context.Context) string {
	if nav, exists := ctx.Value(model.NavContextKey{}).(model.Nav); exists {
		return nav.RoundID
	} else {
		return ""
	}
}

func ToggleOpen(showOpen bool) string {
	if showOpen {
		return "open=true"
	} else {
		return "open=false"
	}
}

func ToggleEdit(edit bool) string {
	if edit {
		return "edit=true"
	} else {
		return "edit=false"
	}
}
