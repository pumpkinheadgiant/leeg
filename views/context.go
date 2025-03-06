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

func ToggleText(showOpen bool) string {
	if showOpen {
		return "open=true"
	} else {
		return "open=false"
	}
}
