package views

import (
	"context"
	"leeg/model"
)

func LeegID(ctx context.Context) string {
	if leegID, exists := ctx.Value(model.ContextKey{}).(string); exists {
		return leegID
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
