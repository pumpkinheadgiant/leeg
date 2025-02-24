package svc

import (
	"go.etcd.io/bbolt"
	"phg.com/leeg/model"
)

type LeegData struct {
	Leeg       model.Leeg
	TeamBucket *bbolt.Bucket
}
