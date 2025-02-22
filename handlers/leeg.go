package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"go.etcd.io/bbolt"
	"phg.com/leeg/svc"
	"phg.com/leeg/svc/migration"
)

const DATAFILE_KEY = "datafile"

type LeegApp struct {
	router *chi.Mux
}

func (l *LeegApp) Init() error {
	database, err := l.initializeDB()
	if err != nil {
		return err
	}
	migrator := migration.Migrator{}
	err = migrator.Migrate(database)
	if err != nil {
		return err
	}
	service := svc.BBoltService{Db: database}
	homeHandler := HomeHandler{service}
	router := chi.NewMux()
	router.Handle("/*", publicHandler())

	router.Get("/", Make(homeHandler.HandleGetHome))
	l.router = router
	return nil
}

func (l LeegApp) Start() error {
	port := os.Getenv("LISTEN_PORT")
	slog.Info("starting slerver", "port", port)
	return http.ListenAndServe(port, l.router)
}

func (l *LeegApp) initializeDB() (*bbolt.DB, error) {
	dbFile := os.Getenv(DATAFILE_KEY)
	if dbFile == "" {
		return nil, fmt.Errorf("environment variable %s not set", DATAFILE_KEY)
	}
	db, err := bbolt.Open(dbFile, 0600, nil)
	return db, err
}

func publicHandler() http.Handler {
	slog.Info("building static files for development")
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public")))
}
