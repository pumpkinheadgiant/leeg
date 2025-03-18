package model

import (
	"fmt"
	"sort"
	"strings"
)

type Leeg struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	TeamDescriptor string        `json:"teamDescriptor"`
	TeamsMap       TeamsMap      `json:"teams"`
	Rounds         EntityRefList `json:"rounds"`
	ImageURL       string        `json:"imageURL"`
	MatchupMap     MatchupMap    `json:"matchupMap"`
	ActiveRound    EntityRef     `json:"activeRound"`
	Scheduled      bool          `json:"scheduled"`
}

func (l Leeg) AsRef() EntityRef {
	return EntityRef{ID: l.ID, Text: l.Name, ImageURL: l.ImageURL, Type: LEEG}
}

func (l Leeg) TotalRounds() int {
	return len(l.Rounds)
}

func (l Leeg) TeamList() EntityRefList {
	allTeams := EntityRefList{}
	for _, team := range l.TeamsMap {
		allTeams = append(allTeams, team.AsRef())
	}
	sort.Slice(allTeams, func(i, j int) bool {
		return allTeams[i].Text > allTeams[j].Text
	})
	return allTeams
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

func (l Leeg) GetRankedTeamsList() EntityRefList {
	teamsList := EntityRefList{}

	type kv struct {
		Key   string
		Value Team
	}
	var ss []kv
	for k, v := range l.TeamsMap {
		ss = append(ss, kv{Key: k, Value: v})
	}

	sort.Slice(ss, func(i, j int) bool {
		teamA := ss[i].Value
		teamB := ss[j].Value
		if teamA.Wins() == teamB.Wins() {
			if teamA.Losses() == teamB.Losses() {
				return teamA.Name < teamB.Name
			} else {
				return teamA.Losses() < teamB.Losses()
			}
		}
		return teamA.Wins() > teamB.Wins()
	})

	for _, team := range ss {
		teamsList = append(teamsList, team.Value.AsRef())
	}

	return teamsList
}

func (l Leeg) getCurrentRoundIdx() int {
	for i, round := range l.Rounds {
		if round.ID == l.ActiveRound.ID {
			return i
		}
	}
	return -1
}

type TeamsMap map[string]Team

func (t TeamsMap) NameAvailable(teamID string, name string) bool {
	for _, team := range t {
		if team.Name == name && team.ID != teamID {
			return false
		}
	}
	return true
}

func (t *TeamsMap) RenameTeam(teamID string, name string) (Team, error) {
	var updatedTeam Team
	for _, existingTeam := range *t {
		if existingTeam.ID == teamID {
			existingTeam.Name = name
			(*t)[teamID] = existingTeam
			updatedTeam = existingTeam

			break
		}
	}
	if updatedTeam.ID == "" {
		return Team{}, fmt.Errorf("no team with ID %v in leeg", teamID)
	}
	updatedRef := updatedTeam.AsRef()

	for _, existingTeam := range *t {
		if existingTeam.ID != teamID {
			existingTeam.TeamsDefeated = existingTeam.TeamsDefeated.Update(updatedRef)
			existingTeam.TeamsDefeatedBy = existingTeam.TeamsDefeatedBy.Update(updatedRef)
		}
	}
	return updatedTeam, nil
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

type TeamUpdateRequest struct {
	LeegID string
	TeamID string
	Name   string
}

type Round struct {
	ID            string        `json:"id"`
	LeegID        string        `json:"leegID"`
	RoundNumber   int           `json:"roundNumber"`
	Games         EntityRefList `json:"games"`
	IsActive      bool          `json:"isActive"`
	GamesPerRound int           `json:"gamesPerRound"`
	AllTeams      EntityRefList `json:"allTeams"`
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

func (g *Game) RenameTeam(teamRef EntityRef) bool {
	if g.TeamA.ID == teamRef.ID {
		g.TeamA = teamRef
		return true
	} else if g.TeamB.ID == teamRef.ID {
		g.TeamB = teamRef
		return true
	}
	return false
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
	(*m)[game.TeamA.ID] = append((*m)[game.TeamA.ID], game.TeamB)
	(*m)[game.TeamB.ID] = append((*m)[game.TeamB.ID], game.TeamA)
	return nil
}

type NavContextKey struct{}

type Nav struct {
	LeegID  string
	RoundID string
	GameID  string
}
