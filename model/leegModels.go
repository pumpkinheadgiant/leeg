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
	RecordsMap     RecordsMap    `json:"recordsMap"`
}

func (l Leeg) AsRef() EntityRef {
	return EntityRef{ID: l.ID, Text: l.Name, ImageURL: l.ImageURL, Type: LEEG}
}

func (l Leeg) TotalRounds() int {
	return len(l.Rounds)
}

type Record struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
}

type RecordsMap map[string]Record

func (r *RecordsMap) Reset() {
	for k := range *r {
		delete(*r, k)
	}
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
	var kvs []kv
	for k, v := range l.TeamsMap {
		kvs = append(kvs, kv{Key: k, Value: v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		teamA := kvs[i].Value
		teamB := kvs[j].Value
		teamARecord := l.RecordsMap[teamA.ID]
		teamBRecord := l.RecordsMap[teamB.ID]
		if teamARecord.Wins == teamBRecord.Wins {
			if teamARecord.Losses == teamBRecord.Losses {
				return teamA.Name < teamB.Name
			} else {
				return teamARecord.Losses < teamBRecord.Losses
			}
		}
		return teamARecord.Wins > teamBRecord.Wins
	})

	for _, team := range kvs {
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

func (t TeamsMap) AsList() []Team {
	var teamList = TeamList{}
	for _, team := range t {
		teamList = append(teamList, team)
	}
	return teamList
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

	return updatedTeam, nil
}

type Team struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"imageURL"`
}

func (t Team) AsRef() EntityRef {
	return EntityRef{ID: t.ID, Text: t.Name, ImageURL: t.ImageURL, Type: TEAM}
}

type TeamList []Team

func (t TeamList) AsEntityList() EntityRefList {
	var list EntityRefList
	for _, team := range t {
		list = append(list, team.AsRef())
	}
	return list
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
	Wins          int           `json:"wins"`
	IsActive      bool          `json:"isActive"`
	GamesPerRound int           `json:"gamesPerRound"`
	AllTeams      EntityRefList `json:"allTeams"`
	UnplayedTeams EntityRefList `json:"unplayedTeams"`
}

func (r Round) SortedTeams() EntityRefList {

	type kv struct {
		Key   string
		Value EntityRef
	}

	var kvs []kv
	for _, v := range r.AllTeams {
		kvs = append(kvs, kv{Key: v.Text, Value: v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value.Text < kvs[j].Value.Text
	})
	sortedTeams := EntityRefList{}

	for _, team := range kvs {
		sortedTeams = append(sortedTeams, team.Value)
	}
	return sortedTeams
}

func (r Round) Scheduled() bool {
	return len(r.Games) == r.GamesPerRound
}

func (r Round) Complete() bool {
	return r.Scheduled() && r.Wins == r.GamesPerRound
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

func (g Game) GetWinner() EntityRef {
	return g.Winner
}

func (g Game) GetLoser() EntityRef {
	if g.TeamA.ID == g.Winner.ID {
		return g.TeamB
	} else {
		return g.TeamA
	}
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

func (m *MatchupMap) RecordMatchup(game Game) {
	(*m)[game.TeamA.ID] = append((*m)[game.TeamA.ID], game.TeamB)
	(*m)[game.TeamB.ID] = append((*m)[game.TeamB.ID], game.TeamA)
}

func (m *MatchupMap) RemoveMatchup(game Game) {
	(*m)[game.TeamA.ID] = (*m)[game.TeamA.ID].RemoveFirst(game.TeamB.ID)
	(*m)[game.TeamB.ID] = (*m)[game.TeamB.ID].RemoveFirst(game.TeamA.ID)
}

type NavContextKey struct{}

type Nav struct {
	LeegID  string
	RoundID string
	GameID  string
}
