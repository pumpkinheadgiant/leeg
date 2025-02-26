package svc

import (
	"encoding/json"
	"errors"
	"fmt"

	"leeg/model"

	"go.etcd.io/bbolt"
)

func (b BBoltService) DataForLeeg(tx *bbolt.Tx, leegID string) (LeegData, error) {
	leegData := LeegData{}

	leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
	if leegsBucket == nil {
		return leegData, errors.New("failed to load leegs bucket")
	}

	leegBucket := leegsBucket.Bucket([]byte(leegID))
	if leegsBucket == nil {
		return leegData, fmt.Errorf("failed to load leeg bucket with id %v", leegID)
	}
	leegDataBucket := leegBucket.Bucket([]byte(dataBucketKey))
	if leegDataBucket == nil {
		return leegData, errors.New("failed to retrieve leeg data bucket")
	}

	var leeg model.Leeg
	var leegBytes = leegDataBucket.Get([]byte(LeegDataKey))
	if leegBytes == nil {
		return leegData, errors.New("failed to retrieve leeg data bytes")
	}
	err := json.Unmarshal(leegBytes, &leeg)
	if err != nil {
		return leegData, err
	}
	leegData.Leeg = leeg

	teamsBucket := leegBucket.Bucket([]byte(teamsBucketKey))
	if teamsBucket == nil {
		return leegData, errors.New("failed to load teamsBucket for leeg bucket")
	}
	leegData.TeamBucket = teamsBucket

	return leegData, nil

}

func (b BBoltService) CreateLeeg(request model.LeegCreateRequest) (model.EntityRef, error) {
	var leegRef model.EntityRef
	return leegRef, b.Db.Update(func(tx *bbolt.Tx) error {
		leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
		if leegsBucket == nil {
			return errors.New("failed to retrieve leegsBucket")
		}

		var newLeeg = model.Leeg{
			ID:             model.NewId(),
			Name:           request.Name,
			TeamDescriptor: request.TeamDescriptor,
			Teams:          []model.EntityRef{},
			Rounds:         []model.Round{},
		}
		leegBucket, err := leegsBucket.CreateBucket([]byte(newLeeg.ID))
		if err != nil {
			return err
		}
		dataBucket, err := leegBucket.CreateBucket([]byte(dataBucketKey))
		if err != nil {
			return err
		}
		teamsBucket, err := leegBucket.CreateBucket([]byte(teamsBucketKey))
		if err != nil {
			return err
		}
		for i := range request.TeamCount {
			var team = model.Team{
				ID:   model.NewId(),
				Name: fmt.Sprintf("%v %v", request.TeamDescriptor, i+1),
			}
			teamBytes, err := json.Marshal(team)
			if err != nil {
				return err
			}
			err = teamsBucket.Put([]byte(team.ID), teamBytes)
			if err != nil {
				return err
			}
			newLeeg.Teams = append(newLeeg.Teams, team.AsRef())
		}
		leegBytes, err := json.Marshal(newLeeg)
		if err != nil {
			return err
		}
		err = dataBucket.Put([]byte(LeegDataKey), leegBytes)
		if err != nil {
			return err
		}
		leegRef = newLeeg.AsRef()
		return nil
	})
}

func (b BBoltService) GetLeeg(leegID string) (model.Leeg, error) {
	var leeg model.Leeg
	return leeg, b.Db.View(func(tx *bbolt.Tx) error {
		leegData, err := b.DataForLeeg(tx, leegID)
		if err != nil {
			return err
		}
		leeg = leegData.Leeg
		return nil
	})
}

func (b BBoltService) GetLeegs() ([]model.EntityRef, error) {
	var leegs []model.EntityRef
	return leegs, b.Db.View(func(tx *bbolt.Tx) error {
		leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
		if leegsBucket == nil {
			return errors.New("failed to retrieve leegsBucket")
		}
		leegsCursor := leegsBucket.Cursor()
		for leegID, leegBucket := leegsCursor.First(); leegID != nil; leegID, leegBucket = leegsCursor.Next() {
			if leegBucket == nil {
				leegBucket := leegsBucket.Bucket(leegID)
				if leegBucket == nil {
					return errors.New("failed to retrieve leeg Bucket")
				}
				leegDataBucket := leegBucket.Bucket([]byte(dataBucketKey))
				if leegDataBucket == nil {
					return errors.New("failed to retrieve leeg data Bucket")
				}
				leegBytes := leegDataBucket.Get([]byte(LeegDataKey))
				if leegBytes == nil {
					return errors.New("failed to retrieve leeg from data Bucket")
				}
				var leeg model.Leeg
				err := json.Unmarshal(leegBytes, &leeg)
				if err != nil {
					return err
				}
				leegs = append(leegs, leeg.AsRef())
			}
		}
		return nil
	})
}
