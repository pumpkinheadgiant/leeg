package svc

import (
	"encoding/json"
	"leeg/model"

	"go.etcd.io/bbolt"
)

type LeegData struct {
	Leeg         model.Leeg
	RoundsBucket *bbolt.Bucket
	DataBucket   *bbolt.Bucket
	GamesBucket  *bbolt.Bucket
}

func (l LeegData) saveGame(game model.Game) error {
	gameBytes, err := json.Marshal(game)
	if err != nil {
		return err
	}
	return l.GamesBucket.Put([]byte(game.ID), gameBytes)
}

func (l LeegData) saveRound(round model.Round) error {
	roundBytes, err := json.Marshal(round)
	if err != nil {
		return err
	}
	return l.RoundsBucket.Put([]byte(round.ID), roundBytes)
}

func (l LeegData) getRoundByID(id string) (model.Round, error) {
	var round model.Round
	roundBytes := l.RoundsBucket.Get([]byte(id))
	return round, json.Unmarshal(roundBytes, &round)
}

func (l LeegData) getGameByID(id string) (model.Game, error) {
	var game model.Game
	gameBytes := l.GamesBucket.Get([]byte(id))
	return game, json.Unmarshal(gameBytes, &game)
}

func (l LeegData) saveLeeg(leeg model.Leeg) error {

	return nil
}
