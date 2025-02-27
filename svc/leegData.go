package svc

import (
	"leeg/model"

	"go.etcd.io/bbolt"
)

type LeegData struct {
	Leeg        model.Leeg
	GamesBucket *bbolt.Bucket
}
