package svc

import (
	"go.etcd.io/bbolt"
	"phg.com/leeg/model"
)

func (b BBoltService) GetLeegs() ([]model.Leeg, error) {
	var leegs []model.Leeg
	return leegs, b.Db.View(func(tx *bbolt.Tx) error {

		return nil
	})
}
