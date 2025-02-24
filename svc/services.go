package svc

import (
	"go.etcd.io/bbolt"
	"phg.com/leeg/model"
)

type BBoltService struct {
	Db *bbolt.DB
}

type LeegService interface {
	GetLeegs() ([]model.EntityRef, error)
	CreateLeeg(request model.LeegCreateRequest) (model.EntityRef, error)
}

const LeegsBucketKey = "leegs"
const LeegDataKey = "leeg"
const dataBucketKey = "data"
const teamsBucketKey = "teems"
