package svc

import (
	"leeg/model"
	"leeg/rando"

	"go.etcd.io/bbolt"
)

type LeegServices struct {
	Db    *bbolt.DB
	Rando rando.RandoConfig
}

type LeegService interface {
	CreateLeeg(request model.LeegCreateRequest) (model.EntityRef, error)
	CreateRandomGame(leegID string, roundID string) (model.Round, model.Game, error)
	GetGame(leegID string, roundID string, gameID string) (model.Game, error)
	GetLeeg(leegID string) (model.Leeg, error)
	GetLeegs() ([]model.EntityRef, error)
	GetRound(leegID string, roundID string) (model.Round, map[string]model.Game, error)
	RenameTeam(update model.TeamUpdateRequest) (model.Team, []model.Game, bool, error)
	ResolveGame(leegID string, gameID string, winnerID string) (model.Game, []model.Team, error)
}

const LeegsBucketKey = "leegs"
const leegDataID = "leeg"
const dataBucketKey = "data"
const roundsBucketKey = "rounds"
const gamesBucketKey = "games"
