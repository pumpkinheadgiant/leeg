package svc

import (
	"encoding/json"
	"errors"
	"fmt"

	"leeg/model"
	"leeg/rando"

	"go.etcd.io/bbolt"
)

func (l LeegServices) RenameTeam(update model.TeamUpdateRequest) (model.Team, model.Record, []model.Game, model.Round, bool, error) {

	var team model.Team
	var games []model.Game
	var activeRound model.Round
	var available = false
	var record model.Record
	return team, record, games, activeRound, available, l.Db.Update(func(tx *bbolt.Tx) error {
		dao, err := l.GetLeegDAO(tx, update.LeegID)
		if err != nil {
			return err
		}
		leeg := dao.Leeg
		available = leeg.TeamsMap.NameAvailable(update.TeamID, update.Name)
		if !available {
			return nil
		}
		team, err = leeg.TeamsMap.RenameTeam(update.TeamID, update.Name)
		if err != nil {
			return err
		}

		games, err = dao.updateGamesForRenamedTeam(team.AsRef())
		if err != nil {
			return err
		}
		for _, roundRef := range leeg.Rounds {
			round, err := dao.getRoundByID(roundRef.ID)
			if err != nil {
				return err
			}
			round.AllTeams = round.AllTeams.Update(team.AsRef())
			round.UnplayedTeams = round.UnplayedTeams.Update(team.AsRef())
			err = dao.saveRound(round)
			if err != nil {
				return err
			}
			if round.IsActive {
				activeRound = round
			}
		}
		record = leeg.RecordsMap[team.ID]
		return dao.saveLeeg(leeg)
	})
}

func (l LeegServices) ResolveGame(leegID string, gameID string, winnerID string) (model.Game, []model.Team, []model.Team, model.RecordsMap, error) {
	var game model.Game
	var modifiedTeams []model.Team
	var allTeams []model.Team
	var recordsMap model.RecordsMap
	return game, allTeams, modifiedTeams, recordsMap, l.Db.Update(func(tx *bbolt.Tx) error {
		dao, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg := dao.Leeg

		game, err = dao.getGameByID(gameID)
		if err != nil {
			return err
		}
		teamA := leeg.TeamsMap[game.TeamA.ID]
		teamB := leeg.TeamsMap[game.TeamB.ID]

		teamAWins := teamA.ID == winnerID
		if teamAWins {
			game.Winner = teamA.AsRef()
		} else {
			game.Winner = teamB.AsRef()
		}

		err = dao.saveGame(game)
		if err != nil {
			return err
		}

		dao.setTeamRecords()

		allTeams = leeg.TeamsMap.AsList()
		modifiedTeams = append(modifiedTeams, teamA, teamB)
		recordsMap = leeg.RecordsMap

		err = dao.saveLeeg(leeg)
		return err
	})
}

func (l LeegServices) GetGame(leegID string, roundID string, gameID string) (model.Game, model.EntityRefList, error) {
	var game model.Game
	var teams model.EntityRefList
	return game, teams, l.Db.View(func(tx *bbolt.Tx) error {
		dao, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		round, err := dao.getRoundByID(roundID)
		if err != nil {
			return err
		}
		teams = round.AllTeams
		game, err = dao.getGameByID(gameID)
		return err
	})
}

func (l LeegServices) RematchGame(leegID string, roundID string, gameID string, teamA string, teamB string) (model.Game, model.RecordsMap, []model.Team, []model.Team, error) {
	var game model.Game
	var modifiedTeams []model.Team
	var allTeams []model.Team
	var recordsMap model.RecordsMap
	return game, recordsMap, modifiedTeams, allTeams, l.Db.Update(func(tx *bbolt.Tx) error {
		dao, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg := dao.Leeg

		round, err := dao.getRoundByID(roundID)
		if err != nil {
			return err
		}
		existingGame, err := dao.getGameByID(gameID)
		if err != nil {
			return err
		}
		teamAUpdated := teamA != existingGame.TeamA.ID
		teamBUpdated := teamB != existingGame.TeamB.ID

		if !teamAUpdated && !teamBUpdated {
			// no-op, so just return the existing game and no teams need to be updated
			game = existingGame
			return nil
		}

		originalWinner := existingGame.GetWinner()
		originalLoser := existingGame.GetLoser()

		round.UnplayedTeams = append(round.UnplayedTeams, existingGame.TeamA)
		round.UnplayedTeams = append(round.UnplayedTeams, existingGame.TeamB)

		leeg.MatchupMap.RemoveMatchup(existingGame) // this will pull the most recent instance each team from each other's history

		if existingGame.Complete() {
			// remove the recorded victory
			existingGame.Winner = model.EntityRef{}
			modifiedTeams = append(modifiedTeams, leeg.TeamsMap[originalWinner.ID])
			modifiedTeams = append(modifiedTeams, leeg.TeamsMap[originalLoser.ID])
			round.Wins--
		}

		newTeamA := leeg.TeamsMap[teamA].AsRef()
		existingGame.TeamA = newTeamA
		newTeamB := leeg.TeamsMap[teamB].AsRef()
		existingGame.TeamB = newTeamB

		round.Games = round.Games.Update(existingGame.AsRef())
		round.UnplayedTeams = round.UnplayedTeams.RemoveAll(newTeamA.ID)
		round.UnplayedTeams = round.UnplayedTeams.RemoveAll(newTeamB.ID)
		err = dao.saveRound(round)
		if err != nil {
			return err
		}

		leeg.MatchupMap.RecordMatchup(existingGame)

		err = dao.saveGame(existingGame)
		if err != nil {
			return err
		}
		game = existingGame

		err = dao.setTeamRecords()
		if err != nil {
			return err
		}

		recordsMap = leeg.RecordsMap
		allTeams = leeg.TeamsMap.AsList()

		err = dao.saveLeeg(leeg)
		if err != nil {
			return err
		}

		return nil
	})
}

func (l LeegServices) RecordMatchup(leegID string, roundID string, teamAID string, teamBID string, winner string) (model.Round, model.Game, []model.Team, model.RecordsMap, error) {
	var round model.Round
	var game model.Game
	var updatedTeams []model.Team
	var recordsMap model.RecordsMap
	return round, game, updatedTeams, recordsMap, l.Db.Update(func(tx *bbolt.Tx) error {
		dao, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg := dao.Leeg

		round, err = dao.getRoundByID(roundID)
		if err != nil {
			return err
		}
		if len(round.Games) >= round.GamesPerRound {
			return errors.New("round is already full, unable to RecordMatchup")
		}
		gameNumber := len(round.Games) + 1
		winnerRef := model.EntityRef{}

		teamA := leeg.TeamsMap[teamAID]
		teamB := leeg.TeamsMap[teamBID]

		if winner != "" {
			if winner == "teamA" {
				winnerRef = teamA.AsRef()
			} else {
				winnerRef = teamB.AsRef()
			}
			round.Wins++
			updatedTeams = append(updatedTeams, teamA, teamB)
		}

		game = model.Game{
			ID:          model.NewId(),
			Round:       round.AsRef(),
			GameNumber:  gameNumber,
			RoundNumber: round.RoundNumber,
			TeamA:       teamA.AsRef(),
			TeamB:       teamB.AsRef(),
			Winner:      winnerRef,
		}
		round.UnplayedTeams = round.UnplayedTeams.RemoveAll(teamAID)
		round.UnplayedTeams = round.UnplayedTeams.RemoveAll(teamBID)
		round.Games = append(round.Games, game.AsRef())

		leeg.MatchupMap.RecordMatchup(game)

		if len(round.Games) == round.GamesPerRound {
			err = dao.advanceRound()
			if err != nil {
				return err
			}
			round.IsActive = false
		}
		err = dao.saveGame(game)
		if err != nil {
			return err
		}

		if winnerRef.ID != "" {
			dao.setTeamRecords()
		}

		err = dao.saveRound(round)
		if err != nil {
			return err
		}
		err = dao.saveLeeg(leeg)
		if err != nil {
			return err
		}
		recordsMap = leeg.RecordsMap
		return nil
	})
}

func (l LeegServices) CreateRandomGame(leegID string, roundID string) (model.Round, model.Game, error) {
	var game model.Game
	var round model.Round
	return round, game, l.Db.Update(func(tx *bbolt.Tx) error {
		dao, err := l.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg := dao.Leeg

		round, err = dao.getRoundByID(roundID)
		if err != nil {
			return err
		}

		gameNumber := len(round.Games) + 1
		game, round.UnplayedTeams, err = newRandomMatchup(gameNumber, round.RoundNumber, round.UnplayedTeams, leeg.MatchupMap, l.Rando)
		if err != nil {
			return err
		}
		game.Round = round.AsRef()
		err = dao.saveGame(game)
		if err != nil {
			return err
		}

		round.Games = append(round.Games, game.AsRef())
		if len(round.Games) == round.GamesPerRound {
			err := dao.advanceRound()
			if err != nil {
				return err
			}
			round.IsActive = false
		}

		err = dao.saveRound(round)
		if err != nil {
			return err
		}
		leeg.MatchupMap.RecordMatchup(game)

		return nil
	})
}

func (l *LeegDAO) advanceRound() error {
	nextRoundRef := l.Leeg.GetNextRound()
	if nextRoundRef.ID == "" {
		// This Leeg is fully scheduled
		l.Leeg.Scheduled = true
	} else {
		// Next round becomes active
		nextRound, err := l.getRoundByID(nextRoundRef.ID)
		if err != nil {
			return err
		}

		nextRound.IsActive = true
		err = l.saveRound(nextRound)
		if err != nil {
			return err
		}
		l.Leeg.ActiveRound = nextRoundRef
		err = l.saveLeeg(l.Leeg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LeegDAO) setTeamRecords() error {

	l.Leeg.RecordsMap.Reset()

	gamesCursor := l.GamesBucket.Cursor()
	var game = model.Game{}
	if l.Leeg.RecordsMap == nil {
		l.Leeg.RecordsMap = model.RecordsMap{}
	}
	recordsMap := l.Leeg.RecordsMap

	for key, value := gamesCursor.First(); key != nil; key, value = gamesCursor.Next() {
		err := json.Unmarshal(value, &game)
		if err != nil {
			return err
		}
		if game.Complete() {
			teamA, found := l.Leeg.TeamsMap[game.TeamA.ID]
			if !found {
				return fmt.Errorf("no team with ID %v", game.TeamA.ID)
			}
			teamB, found := l.Leeg.TeamsMap[game.TeamB.ID]
			if !found {
				return fmt.Errorf("no team with ID %v", game.TeamB.ID)
			}

			teamARecord := recordsMap[teamA.ID]
			teamBRecord := recordsMap[teamB.ID]

			if game.Winner.ID == teamA.ID {
				teamARecord.Wins++
				teamBRecord.Losses++
			} else {
				teamARecord.Losses++
				teamBRecord.Wins++
			}
			recordsMap[teamA.ID] = teamARecord
			recordsMap[teamB.ID] = teamBRecord
		}
	}
	l.Leeg.RecordsMap = recordsMap
	return l.saveLeeg(l.Leeg)
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

	eligibleTeams = eligibleTeams.RemoveAll(game.TeamA.ID)
	eligibleTeams = eligibleTeams.RemoveAll(game.TeamB.ID)

	return game, eligibleTeams, nil
}

func (b LeegServices) GetLeegDAO(tx *bbolt.Tx, leegID string) (LeegDAO, error) {
	dao := LeegDAO{}

	leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
	if leegsBucket == nil {
		return dao, errors.New("failed to load leegs bucket")
	}

	leegBucket := leegsBucket.Bucket([]byte(leegID))
	if leegBucket == nil {
		return dao, fmt.Errorf("failed to load leeg bucket with id %v", leegID)
	}
	leegDataBucket := leegBucket.Bucket([]byte(dataBucketKey))
	if leegDataBucket == nil {
		return dao, errors.New("failed to retrieve leeg data bucket")
	}
	dao.DataBucket = leegDataBucket

	var leeg model.Leeg
	var leegBytes = leegDataBucket.Get([]byte(leegDataID))
	if leegBytes == nil {
		return dao, errors.New("failed to retrieve leeg data bytes")
	}
	err := json.Unmarshal(leegBytes, &leeg)
	if err != nil {
		return dao, err
	}
	dao.Leeg = leeg

	roundsBucket := leegBucket.Bucket([]byte(roundsBucketKey))
	if roundsBucket == nil {
		return dao, errors.New("failed to load rounds bucket for leeg")
	}
	dao.RoundsBucket = roundsBucket

	gameBucket := leegBucket.Bucket([]byte(gamesBucketKey))
	if gameBucket == nil {
		return dao, errors.New("failed to load games bucket for leeg")
	}
	dao.GamesBucket = gameBucket
	return dao, nil
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
			RecordsMap:     model.RecordsMap{},
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
		dao, err := b.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		round, err = dao.getRoundByID(roundID)
		if err != nil {
			return err
		}
		for _, gameRef := range round.Games {
			game, err := dao.getGameByID(gameRef.ID)
			if err != nil {
				return err
			}
			gamesByIDMap[game.ID] = game
		}
		return nil
	})
}

func (b LeegServices) CopyLeeg(leegID string) (model.Leeg, error) {
	var newLeeg model.Leeg
	return newLeeg, b.Db.Update(func(tx *bbolt.Tx) error {

		existingLeegDAO, err := b.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		existingLeeg := existingLeegDAO.Leeg

		newLeegID := model.NewId()

		leegsBucket := tx.Bucket([]byte(LeegsBucketKey))
		if leegsBucket == nil {
			return errors.New("failed to retrieve leegsBucket")
		}

		newLeegBucket, err := leegsBucket.CreateBucket([]byte(newLeegID))
		if err != nil {
			return err
		}
		newDataBucket, err := newLeegBucket.CreateBucket([]byte(dataBucketKey))
		if err != nil {
			return err
		}
		newRoundsBucket, err := newLeegBucket.CreateBucket([]byte(roundsBucketKey))
		if err != nil {
			return err
		}
		_, err = newLeegBucket.CreateBucket([]byte(gamesBucketKey))
		if err != nil {
			return err
		}

		newLeeg = model.Leeg{
			ID:             newLeegID,
			Name:           fmt.Sprintf("%v copy", existingLeeg.Name),
			TeamDescriptor: existingLeeg.TeamDescriptor,
			TeamsMap:       model.TeamsMap{},
		}
		var newTeamsList = model.EntityRefList{}

		for _, existingTeam := range existingLeeg.TeamsMap {
			newTeam := model.Team{
				ID:       model.NewId(),
				Name:     existingTeam.Name,
				ImageURL: existingTeam.ImageURL,
			}
			newLeeg.TeamsMap[newTeam.ID] = newTeam
			newTeamsList = append(newTeamsList, newTeam.AsRef())
		}

		for i, existingRoundRef := range existingLeeg.Rounds {
			existingRound, err := existingLeegDAO.getRoundByID(existingRoundRef.ID)
			if err != nil {
				return err
			}

			var round = model.Round{
				ID:            model.NewId(),
				RoundNumber:   existingRound.RoundNumber,
				LeegID:        newLeegID,
				Games:         model.EntityRefList{},
				GamesPerRound: existingLeeg.GamesPerRound(),
				UnplayedTeams: newTeamsList,
				AllTeams:      newTeamsList,
			}
			if i == 0 {
				newLeeg.ActiveRound = round.AsRef()
				round.IsActive = true
			}

			roundBytes, err := json.Marshal(round)
			if err != nil {
				return err
			}
			err = newRoundsBucket.Put([]byte(round.ID), roundBytes)
			if err != nil {
				return err
			}

			newLeeg.Rounds = append(newLeeg.Rounds, round.AsRef())
		}
		newLeegBytes, err := json.Marshal(newLeeg)
		if err != nil {
			return err
		}

		err = newDataBucket.Put([]byte(leegDataID), newLeegBytes)

		return err
	})

}

func (b LeegServices) GetLeeg(leegID string) (model.Leeg, error) {
	var leeg model.Leeg

	return leeg, b.Db.View(func(tx *bbolt.Tx) error {
		dao, err := b.GetLeegDAO(tx, leegID)
		if err != nil {
			return err
		}
		leeg = dao.Leeg
		return nil
	})
}

func (b LeegServices) GetLeegs() ([]model.EntityRef, error) {
	var leegs []model.EntityRef
	return leegs, b.Db.Update(func(tx *bbolt.Tx) error {

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
