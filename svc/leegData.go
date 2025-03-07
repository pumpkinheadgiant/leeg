package svc

import (
	"encoding/json"
	"leeg/model"

	"go.etcd.io/bbolt"
)

type LeegDAO struct {
	Leeg         model.Leeg
	RoundsBucket *bbolt.Bucket
	DataBucket   *bbolt.Bucket
	GamesBucket  *bbolt.Bucket
}

func (l LeegDAO) updateGamesForRenamedTeam(teamRef model.EntityRef) ([]model.Game, error) {

	var updatedGames = []model.Game{}

	gamesCursor := l.GamesBucket.Cursor()
	var game = model.Game{}

	for key, value := gamesCursor.First(); key != nil; key, value = gamesCursor.Next() {
		err := json.Unmarshal(value, &game)
		if err != nil {
			return updatedGames, err
		}
		if game.RenameTeam(teamRef) {
			updatedGames = append(updatedGames, game)
			err = l.saveGame(game)
			if err != nil {
				return updatedGames, err
			}
		}
	}
	return updatedGames, nil
}
func (l LeegDAO) saveGame(game model.Game) error {
	gameBytes, err := json.Marshal(game)
	if err != nil {
		return err
	}
	return l.GamesBucket.Put([]byte(game.ID), gameBytes)
}

func (l LeegDAO) saveRound(round model.Round) error {
	roundBytes, err := json.Marshal(round)
	if err != nil {
		return err
	}
	return l.RoundsBucket.Put([]byte(round.ID), roundBytes)
}

func (l LeegDAO) getRoundByID(id string) (model.Round, error) {
	var round model.Round
	roundBytes := l.RoundsBucket.Get([]byte(id))
	return round, json.Unmarshal(roundBytes, &round)
}

func (l LeegDAO) getGameByID(id string) (model.Game, error) {
	var game model.Game
	gameBytes := l.GamesBucket.Get([]byte(id))
	return game, json.Unmarshal(gameBytes, &game)
}

func (l LeegDAO) getGameByIDMapForRound(round model.Round) (map[string]model.Game, error) {
	var gamesByIDMap = map[string]model.Game{}
	for _, game := range round.Games {
		game, err := l.getGameByID(game.ID)
		if err != nil {
			return gamesByIDMap, err
		}
		gamesByIDMap[game.ID] = game
	}

	return gamesByIDMap, nil
}

func (l LeegDAO) saveLeeg(leeg model.Leeg) error {
	leegBytes, err := json.Marshal(leeg)
	if err != nil {
		return err
	}
	return l.DataBucket.Put([]byte(leegDataID), leegBytes)
}
