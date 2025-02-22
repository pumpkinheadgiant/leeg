package main

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
	"phg.com/leeg/handlers"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("problem loading env", "err", err)

	}

	fmt.Println(fmt.Sprintf("let the leeg begin!"))
	leegApp := handlers.LeegApp{}
	leegApp.Init()
	err := leegApp.Start()

	slog.Error("app exited", "err", err)
}
