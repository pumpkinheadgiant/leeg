package svc

import (
	"encoding/json"
	"errors"
	"fmt"

	"leeg/model"

	"go.etcd.io/bbolt"
)

func (b BBoltService) CreateRandomGame(leegID string, roundNumber int) (model.Round, model.Game, error) {
	var game model.Game
	var round model.Round
	return round, game, b.Db.Update(func(tx *bbolt.Tx) error {
		leegData, err := b.DataForLeeg(tx, leegID)
		if err != nil {
			return err
		}
		leeg := leegData.Leeg
		if roundNumber < 1 || roundNumber > leeg.TotalRounds() {
			return fmt.Errorf("invalid roundNumber: %v", roundNumber)
		}
		// round, err := leeg.GetCurrentRound()
		if err != nil {
			return err
		}
		// game := round.GetR
		return nil
	})
}

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

	gamesBucket := leegBucket.Bucket([]byte(gamesBucketKey))
	if gamesBucket == nil {
		return leegData, errors.New("failed to load games bucket for leeg")
	}
	leegData.GamesBucket = gamesBucket

	return leegData, nil
}

func (b BBoltService) CreateLeeg(request model.LeegCreateRequest) (model.EntityRef, error) {
	var leegRef model.EntityRef
	return leegRef, b.Db.Update(func(tx *bbolt.Tx) error {
		leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
		if leegsBucket == nil {
			return errors.New("failed to retrieve leegsBucket")
		}
		newLeegID := model.NewId()

		var rounds = []model.Round{}
		for i := range request.RoundCount {
			var round = model.Round{
				Active:        i == 0, // round 1 will be the initial active round
				RoundNumber:   i + 1,
				LeegID:        newLeegID,
				Games:         []model.Game{},
				GamesPerRound: request.TeamCount / 2,
				TeamsPlayed:   model.EntityRefList{},
			}
			rounds = append(rounds, round)
		}
		var teams = []model.Team{}
		for i := range request.TeamCount {
			var team = model.Team{
				ID:   model.NewId(),
				Name: fmt.Sprintf("%v %v", request.TeamDescriptor, i+1),
			}
			teams = append(teams, team)
		}

		var newLeeg = model.Leeg{
			ID:             newLeegID,
			Name:           request.Name,
			TeamDescriptor: request.TeamDescriptor,
			Teams:          teams,
			Rounds:         rounds,
		}
		leegBucket, err := leegsBucket.CreateBucket([]byte(newLeeg.ID))
		if err != nil {
			return err
		}
		dataBucket, err := leegBucket.CreateBucket([]byte(dataBucketKey))
		if err != nil {
			return err
		}
		_, err = leegBucket.CreateBucket([]byte(gamesBucketKey))
		if err != nil {
			return err
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
