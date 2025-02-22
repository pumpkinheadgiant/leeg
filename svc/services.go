package svc

import (
	"go.etcd.io/bbolt"
	"phg.com/leeg/model"
)

type BBoltService struct {
	Db *bbolt.DB
}

type LeegData interface {
	GetLeegs() ([]model.Leeg, error)
}

const LeegsBucketKey = "leegs"
