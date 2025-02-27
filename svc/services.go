package svc

import (
	"leeg/model"

	"go.etcd.io/bbolt"
)

type BBoltService struct {
	Db *bbolt.DB
}

type LeegService interface {
	GetLeegs() ([]model.EntityRef, error)
	GetLeeg(leegID string) (model.Leeg, error)
	CreateLeeg(request model.LeegCreateRequest) (model.EntityRef, error)
}

const LeegsBucketKey = "leegs"
const LeegDataKey = "leeg"
const dataBucketKey = "data"
const gamesBucketKey = "games"
