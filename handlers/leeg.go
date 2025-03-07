package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"leeg/rando"
	"leeg/svc"
	"leeg/svc/migration"

	"github.com/go-chi/chi/v5"
	"go.etcd.io/bbolt"
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
	services := svc.LeegServices{Db: database, Rando: rando.RandoConfig{}}

	homeHandler := HomeHandler{services}
	leegHandler := LeegHandler{services}
	gameHandler := GameHandler{services}
	roundHandler := RoundHandler{services}
	router := chi.NewMux()
	router.Handle("/*", publicHandler())

	router.Get("/", Make(homeHandler.HandleGetHome))

	router.Post("/leegs", Make(leegHandler.HandlePostLeeg))
	router.Get("/leegs/{leegID}", Make(leegHandler.HandleGetLeeg))

	router.Get("/leegs/{leegID}/rounds/{roundID}/games/{gameID}", Make(gameHandler.HandleGetGame))
	router.Post("/leegs/{leegID}/rounds/{roundID}/games", Make(gameHandler.HandleGameCreationRequest))
	router.Put("/leegs/{leegID}/rounds/{roundID}/games/{gameID}", Make(gameHandler.HandleGameUpdate))
	router.Get("/leegs/{leegID}/rounds/{roundID}", Make(roundHandler.HandleGetRound))

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
	fs := http.FileServer(http.FS(os.DirFS("public")))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Explicitly set MIME types for specific file extensions
		if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".ico") {
			w.Header().Set("Content-Type", "image/x-icon")
		}

		fs.ServeHTTP(w, r)
	})
}
