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
	CopyLeeg(leegID string) (model.Leeg, error)
	CreateLeeg(request model.LeegCreateRequest) (model.EntityRef, error)
	CreateRandomGame(leegID string, roundID string) (model.Round, model.Game, error)
	GetGame(leegID string, roundID string, gameID string) (model.Game, model.EntityRefList, error)
	GetLeeg(leegID string) (model.Leeg, error)
	GetLeegs() ([]model.EntityRef, error)
	GetRound(leegID string, roundID string) (model.Round, map[string]model.Game, error)
	GetTeams(leegID string) (model.EntityRefList, error)
	RecordMatchup(leegID string, roundID string, teamAID string, teamBID string, winner string) (model.Round, model.Game, []model.Team, model.RecordsMap, error)
	RematchGame(leegID string, roundID string, gameID string, teamA string, teamB string) (model.Game, model.RecordsMap, []model.Team, []model.Team, error)
	RenameTeam(update model.TeamUpdateRequest) (model.Team, model.Record, []model.Game, model.Round, bool, error)
	ResolveGame(leegID string, gameID string, winnerID string) (model.Game, []model.Team, []model.Team, model.RecordsMap, error)
}

const LeegsBucketKey = "leegs"
const leegDataID = "leeg"
const dataBucketKey = "data"
const roundsBucketKey = "rounds"
const gamesBucketKey = "games"
