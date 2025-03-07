package model

import (
	"fmt"
	"strings"
)

type Leeg struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	TeamDescriptor string          `json:"teamDescriptor"`
	TeamsMap       map[string]Team `json:"teams"`
	Rounds         EntityRefList   `json:"rounds"`
	ImageURL       string          `json:"imageURL"`
	MatchupMap     MatchupMap      `json:"matchupMap"`
	ActiveRound    EntityRef       `json:"activeRound"`
	Scheduled      bool            `json:"scheduled"`
}

func (l Leeg) AsRef() EntityRef {
	return EntityRef{ID: l.ID, Text: l.Name, ImageURL: l.ImageURL, Type: LEEG}
}

func (l Leeg) TotalRounds() int {
	return len(l.Rounds)
}

func (l Leeg) GamesPerRound() int {
	return len(l.TeamsMap) / 2
}

func (l Leeg) GetNextRound() EntityRef {
	currentRoundIdx := l.getCurrentRoundIdx()
	if currentRoundIdx == -1 || currentRoundIdx+1 == len(l.Rounds) {
		return EntityRef{}
	}
	return l.Rounds[currentRoundIdx+1]
}

func (l Leeg) getCurrentRoundIdx() int {
	for i, round := range l.Rounds {
		if round.ID == l.ActiveRound.ID {
			return i
		}
	}
	return -1
}

type Team struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	ImageURL        string        `json:"imageURL"`
	TeamsDefeated   EntityRefList `json:"teamsDefeated"`
	TeamsDefeatedBy EntityRefList `json:"teamsDefeatedBy"`
}

func (t Team) Wins() int {
	return len(t.TeamsDefeated)
}

func (t Team) Losses() int {
	return len(t.TeamsDefeatedBy)
}

func (t Team) AsRef() EntityRef {
	return EntityRef{ID: t.ID, Text: t.Name, ImageURL: t.ImageURL, Type: TEAM}
}

type Round struct {
	ID            string        `json:"id"`
	LeegID        string        `json:"leegID"`
	RoundNumber   int           `json:"roundNumber"`
	Games         EntityRefList `json:"games"`
	IsActive      bool          `json:"isActive"`
	GamesPerRound int           `json:"gamesPerRound"`
	UnplayedTeams EntityRefList `json:"unplayedTeams"`
}

func (r Round) Scheduled() bool {
	return len(r.Games) == r.GamesPerRound
}

func (r Round) AsRef() EntityRef {
	return EntityRef{ID: r.ID, Type: ROUND, Text: fmt.Sprintf("Round %v", r.RoundNumber)}
}

type Game struct {
	ID          string    `json:"id"`
	Round       EntityRef `json:"round"`
	RoundNumber int       `json:"roundNumber"`
	GameNumber  int       `json:"gameNumber"`
	TeamA       EntityRef `json:"teamA"`
	TeamB       EntityRef `json:"teamB"`
	Winner      EntityRef `json:"winner"`
}

func (g Game) Complete() bool {
	return g.Winner.ID != ""
}

func (g Game) AsRef() EntityRef {
	var outcome = "TBD"
	if g.Winner.ID != "" {
		outcome = fmt.Sprintf("Winner: %v", g.Winner.Text)
	}
	return EntityRef{ID: g.ID, Text: fmt.Sprintf("Game %v. %v vs %v. Winner: %v", g.GameNumber, g.TeamA.Text, g.TeamB.Text, outcome)}
}

type LeegStatus struct {
	CurrentRound          int
	TotalRounds           int
	GamesRemainingInRound int
}

type LeegCreateRequest struct {
	Name           string
	TeamDescriptor string
	TeamCount      int
	RoundCount     int
}

func (l *LeegCreateRequest) ValidateAndNormalize() map[string]string {
	errors := map[string]string{}
	l.Name = strings.TrimSpace(l.Name)
	if len(l.Name) < 1 || len(l.Name) > 50 {
		errors["name"] = "please select a name with between 1 and 50 characters"
	}
	if l.TeamCount < 4 || l.TeamCount > 32 || l.TeamCount%2 != 0 {
		errors["teamCount"] = "please select an even number of between 4 and 32 teams"
	}
	if l.RoundCount < 1 || l.RoundCount > (l.TeamCount-1) {
		errors["roundCount"] = fmt.Sprintf("please select between 1 and %v (# of Teams -1) rounds", l.TeamCount-1)
	}
	l.TeamDescriptor = strings.TrimSpace(l.TeamDescriptor)
	if len(l.TeamDescriptor) < 1 || len(l.TeamDescriptor) > 20 {
		errors["teamDescriptor"] = "team descriptor should be between 1 and 20 characters"
	}
	return errors
}

type LeegPageData struct {
	Leeg        Leeg
	ActiveRound Round
}

type MatchupMap map[string]EntityRefList

func (m *MatchupMap) RecordMatchup(game Game) error {
	var teamAMatchups = (*m)[game.TeamA.ID]
	var teamBMatchups = (*m)[game.TeamB.ID]

	if teamAMatchups.HasID(game.TeamB.ID) || teamBMatchups.HasID(game.TeamA.ID) {
		return fmt.Errorf("%v and %v have already played", game.TeamA.Text, game.TeamB.Text)
	}
	(*m)[game.TeamA.ID] = append((*m)[game.TeamA.ID], game.TeamB)
	(*m)[game.TeamB.ID] = append((*m)[game.TeamB.ID], game.TeamA)
	return nil
}

type ContextKey struct{}

type Nav struct {
	LeegID  string
	RoundID string
	GameID  string
}
