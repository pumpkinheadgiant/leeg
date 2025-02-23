package svc

import (
	"encoding/json"
	"errors"

	"go.etcd.io/bbolt"
	"phg.com/leeg/model"
)

func (b BBoltService) GetLeegs() ([]model.Leeg, error) {
	var leegs []model.Leeg
	return leegs, b.Db.View(func(tx *bbolt.Tx) error {
		leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
		if leegsBucket == nil {
			return errors.New("failed to retrieve leegsBucket")
		}
		var leeg model.Leeg
		leegsCursor := leegsBucket.Cursor()
		for key, value := leegsCursor.First(); key != nil; key, value = leegsCursor.Next() {
			err := json.Unmarshal(value, &leeg)
			if err != nil {
				return err
			}
			leegs = append(leegs, leeg)
		}
		return nil
	})
}
