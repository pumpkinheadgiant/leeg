package model

import (
	"fmt"
	"strings"
)

type Leeg struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	TeamDescriptor string  `json:"teamDescriptor"`
	Teams          []Team  `json:"teams"`
	Rounds         []Round `json:"rounds"`
	ImageURL       string  `json:"imageURL"`
}

func (l Leeg) AsRef() EntityRef {
	return EntityRef{ID: l.ID, Text: l.Name, Image: l.ImageURL, Type: LeegType}
}

func (l Leeg) TotalRounds() int {
	return len(l.Rounds)
}

func (l Leeg) GamesPerRound() int {
	return len(l.Teams) / 2
}

type Team struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	ImageURL    string      `json:"imageURL"`
	Wins        int         `json:"wins"`
	Losses      int         `json:"losses"`
	TeamsPlayed []EntityRef `json:"teamsPlayed"`
}

func (t Team) AsRef() EntityRef {
	return EntityRef{ID: t.ID, Text: t.Name, Image: t.ImageURL, Type: TeamType}
}

type Round struct {
	Active        bool   `json:"active"`
	RoundNumber   int    `json:"roundNumber"`
	Games         []Game `json:"games"`
	GamesPerRound int    `json:"gamesPerRound"`
}

func (r Round) Complete() bool {
	return len(r.Games) == r.GamesPerRound
}

type Game struct {
	TeamA  EntityRef `json:"teamA"`
	TeamB  EntityRef `json:"teamB"`
	Winner EntityRef `json:"winner"`
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
