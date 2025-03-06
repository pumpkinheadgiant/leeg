package main

import (
	"fmt"
	"log"
	"log/slog"

	"leeg/handlers"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("problem loading env", "err", err)
		log.Fatal("exiting due to env load failure")
	}

	fmt.Println(fmt.Sprintf("let the leeg begin!"))
	leegApp := handlers.LeegApp{}
	leegApp.Init()
	err := leegApp.Start()

	slog.Error("app exited", "err", err)
}
