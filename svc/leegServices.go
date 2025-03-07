package svc

import (
	"encoding/json"
	"errors"
	"fmt"

	"leeg/model"
	"leeg/rando"

	"go.etcd.io/bbolt"
)

func (l LeegServices) RenameTeam(leegID string, teamID string, name string) (model.Team, []model.Game, bool, error) {

	var team model.Team
	var games []model.Game
	var available = false
	return team, games, available, l.Db.View(func(tx *bbolt.Tx) error {
		leegDAO, err := l.DataForLeeg(tx, leegID)
		if err != nil {
			return err
		}
		leeg := leegDAO.Leeg
		available = leeg.TeamsMap.NameAvailable(teamID, name)
		if !available {
			return nil
		}
		team, err = leeg.TeamsMap.RenameTeam(teamID, name)
		if err != nil {
			return err
		}

		games, err = leegDAO.updateGamesForRenamedTeam(team.AsRef())
		if err != nil {
			return err
		}

		return nil
	})
}

func (l LeegServices) ResolveGame(leegID string, gameID string, winnerID string) (model.Game, []model.Team, error) {
	var game model.Game
	var teams []model.Team
	return game, teams, l.Db.Update(func(tx *bbolt.Tx) error {
		leegDAO, err := l.DataForLeeg(tx, leegID)
		if err != nil {
			return err
		}
		leeg := leegDAO.Leeg

		game, err = leegDAO.getGameByID(gameID)
		if err != nil {
			return err
		}
		teamA := leeg.TeamsMap[game.TeamA.ID]
		teamB := leeg.TeamsMap[game.TeamB.ID]

		teamAWins := teamA.ID == winnerID
		if teamAWins {
			teamA.TeamsDefeated = append(teamA.TeamsDefeated, teamB.AsRef())
			teamB.TeamsDefeatedBy = append(teamB.TeamsDefeatedBy, teamA.AsRef())
			game.Winner = teamA.AsRef()
		} else {
			teamA.TeamsDefeatedBy = append(teamA.TeamsDefeatedBy, teamB.AsRef())
			teamB.TeamsDefeated = append(teamB.TeamsDefeated, teamA.AsRef())
			game.Winner = teamB.AsRef()
		}

		err = leegDAO.saveGame(game)
		if err != nil {
			return err
		}

		leeg.TeamsMap[teamA.ID] = teamA
		leeg.TeamsMap[teamB.ID] = teamB
		teams = append(teams, teamA, teamB)
		return leegDAO.saveLeeg(leeg)
	})
}

func (l LeegServices) GetGame(leegID string, roundID string, gameID string) (model.Game, error) {
	var game model.Game
	return game, l.Db.View(func(tx *bbolt.Tx) error {
		leegDAO, err := l.DataForLeeg(tx, leegID)
		if err != nil {
			return err
		}

		game, err = leegDAO.getGameByID(gameID)
		return err

	})
}

func (l LeegServices) CreateRandomGame(leegID string, roundID string) (model.Round, model.Game, error) {
	var game model.Game
	var round model.Round
	return round, game, l.Db.Update(func(tx *bbolt.Tx) error {
		leegData, err := l.DataForLeeg(tx, leegID)
		if err != nil {
			return err
		}
		leeg := leegData.Leeg

		round, err = leegData.getRoundByID(roundID)
		if err != nil {
			return err
		}

		gameNumber := len(round.Games) + 1
		game, round.UnplayedTeams, err = newRandomMatchup(gameNumber, round.RoundNumber, round.UnplayedTeams, leeg.MatchupMap, l.Rando)
		if err != nil {
			return err
		}
		err = leegData.saveGame(game)
		if err != nil {
			return err
		}

		round.Games = append(round.Games, game.AsRef())

		if len(round.Games) == round.GamesPerRound {
			// This Round is fully scheduled
			round.IsActive = false

			nextRoundRef := leeg.GetNextRound()
			if nextRoundRef.ID == "" {
				// This Leeg is fully scheduled
				leeg.Scheduled = true
				leeg.ActiveRound = nextRoundRef
			} else {
				// Next round becomes active
				nextRound, err := leegData.getRoundByID(nextRoundRef.ID)
				if err != nil {
					return err
				}
				nextRound.IsActive = true
				err = leegData.saveRound(nextRound)
				if err != nil {
					return err
				}
				leeg.ActiveRound = nextRoundRef
			}
		}
		err = leegData.saveRound(round)
		if err != nil {
			return err
		}
		err = leeg.MatchupMap.RecordMatchup(game)
		if err != nil {
			return err
		}
		err = leegData.saveLeeg(leeg)
		return err
	})
}

func newRandomMatchup(gameNumber int, roundNumber int, eligibleTeams model.EntityRefList, leegMatchupMap map[string]model.EntityRefList, rando rando.RandoConfig) (model.Game, model.EntityRefList, error) {
	var game = model.Game{ID: model.NewId(), GameNumber: gameNumber, RoundNumber: roundNumber}
	if len(eligibleTeams) < 2 {
		return game, eligibleTeams, errors.New("must have at least two eligible teams to match")
	}
	for game.TeamA.ID == "" || game.TeamB.ID == "" || game.TeamA.ID == game.TeamB.ID || leegMatchupMap[game.TeamA.ID].HasID(game.TeamB.ID) {
		teamA, err := rando.RandomEntity(eligibleTeams)
		if err != nil {
			return game, eligibleTeams, err
		}
		teamB, err := rando.RandomEntity(eligibleTeams)
		if err != nil {
			return game, eligibleTeams, err
		}
		game.TeamA = teamA
		game.TeamB = teamB
	}
	eligibleTeams = eligibleTeams.Remove(game.TeamA.ID)
	eligibleTeams = eligibleTeams.Remove(game.TeamB.ID)

	return game, eligibleTeams, nil
}

func (b LeegServices) DataForLeeg(tx *bbolt.Tx, leegID string) (LeegDAO, error) {
	leegData := LeegDAO{}

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
	leegData.DataBucket = leegDataBucket

	var leeg model.Leeg
	var leegBytes = leegDataBucket.Get([]byte(leegDataID))
	if leegBytes == nil {
		return leegData, errors.New("failed to retrieve leeg data bytes")
	}
	err := json.Unmarshal(leegBytes, &leeg)
	if err != nil {
		return leegData, err
	}
	leegData.Leeg = leeg

	roundsBucket := leegBucket.Bucket([]byte(roundsBucketKey))
	if roundsBucket == nil {
		return leegData, errors.New("failed to load rounds bucket for leeg")
	}
	leegData.RoundsBucket = roundsBucket

	gameBucket := leegBucket.Bucket([]byte(gamesBucketKey))
	if gameBucket == nil {
		return leegData, errors.New("failed to load games bucket for leeg")
	}
	leegData.GamesBucket = gameBucket
	return leegData, nil
}

func (b LeegServices) CreateLeeg(request model.LeegCreateRequest) (model.EntityRef, error) {
	var leegRef model.EntityRef
	return leegRef, b.Db.Update(func(tx *bbolt.Tx) error {
		leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
		if leegsBucket == nil {
			return errors.New("failed to retrieve leegsBucket")
		}

		newLeegID := model.NewId()

		leegBucket, err := leegsBucket.CreateBucket([]byte(newLeegID))
		if err != nil {
			return err
		}
		dataBucket, err := leegBucket.CreateBucket([]byte(dataBucketKey))
		if err != nil {
			return err
		}
		roundsBucket, err := leegBucket.CreateBucket([]byte(roundsBucketKey))
		if err != nil {
			return err
		}
		_, err = leegBucket.CreateBucket([]byte(gamesBucketKey))
		if err != nil {
			return err
		}

		var teamsMap = map[string]model.Team{}
		var allTeamsList = model.EntityRefList{}

		for i := range request.TeamCount {
			var team = model.Team{
				ID:   model.NewId(),
				Name: fmt.Sprintf("%v %v", request.TeamDescriptor, i+1),
			}
			teamsMap[team.ID] = team
			allTeamsList = append(allTeamsList, team.AsRef())
		}

		var newLeeg = model.Leeg{
			ID:             newLeegID,
			Name:           request.Name,
			TeamDescriptor: request.TeamDescriptor,
			TeamsMap:       teamsMap,
			MatchupMap:     model.MatchupMap{},
		}

		for i := range request.RoundCount {
			var round = model.Round{
				ID:            model.NewId(),
				RoundNumber:   i + 1,
				LeegID:        newLeegID,
				Games:         model.EntityRefList{},
				GamesPerRound: request.TeamCount / 2,
				UnplayedTeams: allTeamsList,
			}

			roundRef := round.AsRef()
			if i == 0 {
				newLeeg.ActiveRound = roundRef
				round.IsActive = true
			}

			roundBytes, err := json.Marshal(round)
			if err != nil {
				return err
			}
			err = roundsBucket.Put([]byte(round.ID), roundBytes)
			if err != nil {
				return err
			}

			newLeeg.Rounds = append(newLeeg.Rounds, roundRef)
		}
		leegBytes, err := json.Marshal(newLeeg)
		if err != nil {
			return err
		}
		err = dataBucket.Put([]byte(leegDataID), leegBytes)
		if err != nil {
			return err
		}
		leegRef = newLeeg.AsRef()
		return nil
	})
}

func (b LeegServices) GetRound(leegID string, roundID string) (model.Round, map[string]model.Game, error) {
	var round model.Round
	var gamesByIDMap = map[string]model.Game{}

	return round, gamesByIDMap, b.Db.View(func(tx *bbolt.Tx) error {
		leegData, err := b.DataForLeeg(tx, leegID)
		if err != nil {
			return err
		}
		roundBytes := leegData.RoundsBucket.Get([]byte(roundID))
		if roundBytes == nil {
			return fmt.Errorf("couldn't locate round with id '%v'", roundID)
		}
		err = json.Unmarshal(roundBytes, &round)
		if err != nil {
			return err
		}
		for _, gameRef := range round.Games {
			game, err := leegData.getGameByID(gameRef.ID)
			if err != nil {
				return err
			}
			gamesByIDMap[game.ID] = game
		}
		return nil
	})
}

func (b LeegServices) GetLeeg(leegID string) (model.Leeg, error) {
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

func (b LeegServices) GetLeegs() ([]model.EntityRef, error) {
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
				leegBytes := leegDataBucket.Get([]byte(leegDataID))
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
