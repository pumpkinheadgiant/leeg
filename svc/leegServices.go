package svc

import (
	"encoding/json"
	"errors"
	"fmt"

	"leeg/model"
	"leeg/rando"

	"go.etcd.io/bbolt"
)

func (l LeegServices) RenameTeam(update model.TeamUpdateRequest) (model.Team, []model.Game, model.Round, bool, error) {

	var team model.Team
	var games []model.Game
	var activeRound model.Round
	var available = false
	return team, games, activeRound, available, l.Db.Update(func(tx *bbolt.Tx) error {
		leegDAO, err := l.GetLeegDAO(tx, update.LeegID)
		if err != nil {
			return err
		}
		leeg := leegDAO.Leeg
		available = leeg.TeamsMap.NameAvailable(update.TeamID, update.Name)
		if !available {
			return nil
		}
		team, err = leeg.TeamsMap.RenameTeam(update.TeamID, update.Name)
		if err != nil {
			return err
		}

		games, err = leegDAO.updateGamesForRenamedTeam(team.AsRef())
		if err != nil {
			return err
		}
		for _, roundRef := range leeg.Rounds {
			round, err := leegDAO.getRoundByID(roundRef.ID)
			if err != nil {
				return err
			}
			round.AllTeams = round.AllTeams.Update(team.AsRef())
			round.UnplayedTeams = round.UnplayedTeams.Update(team.AsRef())
			err = leegDAO.saveRound(round)
			if err != nil {
				return err
			}
			if round.IsActive {
				activeRound = round
			}
		}

		return leegDAO.saveLeeg(leeg)
	})
}

func (l LeegServices) ResolveGame(leegID string, gameID string, winnerID string) (model.Game, []model.Team, error) {
	var game model.Game
	var teams []model.Team
	return game, teams, l.Db.Update(func(tx *bbolt.Tx) error {
		leegDAO, err := l.GetLeegDAO(tx, leegID)
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
		leegDAO, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}

		game, err = leegDAO.getGameByID(gameID)
		return err

	})
}

func (l LeegServices) RecordMatchup(leegID string, roundID string, teamA string, teamB string) (model.Round, model.Game, error) {
	var round model.Round
	var game model.Game
	return round, game, l.Db.Update(func(tx *bbolt.Tx) error {
		leegDAO, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg := leegDAO.Leeg

		round, err = leegDAO.getRoundByID(roundID)
		if err != nil {
			return err
		}
		if len(round.Games) >= round.GamesPerRound {
			return errors.New("round is already full, unable to RecordMatchup")
		}
		gameNumber := len(round.Games) + 1
		game = model.Game{
			ID:          model.NewId(),
			Round:       round.AsRef(),
			GameNumber:  gameNumber,
			RoundNumber: round.RoundNumber,
			TeamA:       leeg.TeamsMap[teamA].AsRef(),
			TeamB:       leeg.TeamsMap[teamB].AsRef(),
		}
		round.UnplayedTeams = round.UnplayedTeams.Remove(teamA)
		round.UnplayedTeams = round.UnplayedTeams.Remove(teamB)
		round.Games = append(round.Games, game.AsRef())

		if len(round.Games) == round.GamesPerRound {
			var advanced = false
			round, advanced, err = leegDAO.advanceRound(round)
			if err != nil {
				return err
			}
			if advanced {
				leeg.ActiveRound = leeg.GetNextRound()
			}
			leeg.ActiveRound = round.AsRef()
		}
		err = leegDAO.saveGame(game)
		if err != nil {
			return err
		}
		err = leegDAO.saveRound(round)
		if err != nil {
			return err
		}
		return leegDAO.saveLeeg(leeg)
	})
}

func (l *LeegDAO) advanceRound(round model.Round) (model.Round, bool, error) {
	var advanced = false
	if len(round.Games) == round.GamesPerRound {
		// This Round is fully scheduled
		round.IsActive = false

		nextRoundRef := l.Leeg.GetNextRound()
		if nextRoundRef.ID == "" {
			// This Leeg is fully scheduled
			l.Leeg.Scheduled = true
		} else {
			// Next round becomes active
			nextRound, err := l.getRoundByID(nextRoundRef.ID)
			if err != nil {
				return round, advanced, err
			}
			nextRound.IsActive = true
			err = l.saveRound(nextRound)
			if err != nil {
				return round, advanced, err
			}
			advanced = true
		}
	}
	return round, advanced, nil
}

func (l LeegServices) CreateRandomGame(leegID string, roundID string) (model.Round, model.Game, error) {
	var game model.Game
	var round model.Round
	return round, game, l.Db.Update(func(tx *bbolt.Tx) error {
		leegDAO, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg := leegDAO.Leeg

		round, err = leegDAO.getRoundByID(roundID)
		if err != nil {
			return err
		}

		gameNumber := len(round.Games) + 1
		game, round.UnplayedTeams, err = newRandomMatchup(gameNumber, round.RoundNumber, round.UnplayedTeams, leeg.MatchupMap, l.Rando)
		if err != nil {
			return err
		}
		err = leegDAO.saveGame(game)
		if err != nil {
			return err
		}

		round.Games = append(round.Games, game.AsRef())
		var advanced = false
		round, advanced, err = leegDAO.advanceRound(round)
		if err != nil {
			return err
		}
		if advanced {
			leeg.ActiveRound = leeg.GetNextRound()
		}

		err = leegDAO.saveRound(round)
		if err != nil {
			return err
		}
		err = leeg.MatchupMap.RecordMatchup(game)
		if err != nil {
			return err
		}
		err = leegDAO.saveLeeg(leeg)
		return err
	})
}

func newRandomMatchup(gameNumber int, roundNumber int, eligibleTeams model.EntityRefList, leegMatchupMap map[string]model.EntityRefList, rando rando.RandoConfig) (model.Game, model.EntityRefList, error) {
	var game = model.Game{ID: model.NewId(), GameNumber: gameNumber, RoundNumber: roundNumber}
	if len(eligibleTeams) < 2 {
		return game, eligibleTeams, errors.New("must have at least two eligible teams to match")
	}
	attempts := 1
	for game.TeamA.ID == "" || game.TeamB.ID == "" || game.TeamA.ID == game.TeamB.ID || leegMatchupMap[game.TeamA.ID].HasID(game.TeamB.ID) && attempts < len(eligibleTeams)*2 {
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
		attempts++
	}

	eligibleTeams = eligibleTeams.Remove(game.TeamA.ID)
	eligibleTeams = eligibleTeams.Remove(game.TeamB.ID)

	return game, eligibleTeams, nil
}

func (b LeegServices) GetLeegDAO(tx *bbolt.Tx, leegID string) (LeegDAO, error) {
	leegDAO := LeegDAO{}

	leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
	if leegsBucket == nil {
		return leegDAO, errors.New("failed to load leegs bucket")
	}

	leegBucket := leegsBucket.Bucket([]byte(leegID))
	if leegsBucket == nil {
		return leegDAO, fmt.Errorf("failed to load leeg bucket with id %v", leegID)
	}
	leegDataBucket := leegBucket.Bucket([]byte(dataBucketKey))
	if leegDataBucket == nil {
		return leegDAO, errors.New("failed to retrieve leeg data bucket")
	}
	leegDAO.DataBucket = leegDataBucket

	var leeg model.Leeg
	var leegBytes = leegDataBucket.Get([]byte(leegDataID))
	if leegBytes == nil {
		return leegDAO, errors.New("failed to retrieve leeg data bytes")
	}
	err := json.Unmarshal(leegBytes, &leeg)
	if err != nil {
		return leegDAO, err
	}
	leegDAO.Leeg = leeg

	roundsBucket := leegBucket.Bucket([]byte(roundsBucketKey))
	if roundsBucket == nil {
		return leegDAO, errors.New("failed to load rounds bucket for leeg")
	}
	leegDAO.RoundsBucket = roundsBucket

	gameBucket := leegBucket.Bucket([]byte(gamesBucketKey))
	if gameBucket == nil {
		return leegDAO, errors.New("failed to load games bucket for leeg")
	}
	leegDAO.GamesBucket = gameBucket
	return leegDAO, nil
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
				AllTeams:      allTeamsList,
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

func (b LeegServices) GetTeams(leegID string) (model.EntityRefList, error) {
	var teams model.EntityRefList
	return teams, b.Db.View(func(tx *bbolt.Tx) error {
		dao, err := b.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		teams = dao.Leeg.TeamList()
		return nil
	})
}

func (b LeegServices) GetRound(leegID string, roundID string) (model.Round, map[string]model.Game, error) {
	var round model.Round
	var gamesByIDMap = map[string]model.Game{}

	return round, gamesByIDMap, b.Db.View(func(tx *bbolt.Tx) error {
		leegDAO, err := b.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		roundBytes := leegDAO.RoundsBucket.Get([]byte(roundID))
		if roundBytes == nil {
			return fmt.Errorf("couldn't locate round with id '%v'", roundID)
		}
		err = json.Unmarshal(roundBytes, &round)
		if err != nil {
			return err
		}
		for _, gameRef := range round.Games {
			game, err := leegDAO.getGameByID(gameRef.ID)
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
		leegDAO, err := b.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg = leegDAO.Leeg
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
