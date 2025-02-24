package model

import "github.com/google/uuid"

type EntityRef struct {
	ID    string     `json:"id"`
	Text  string     `json:"text"`
	Type  EntityType `json:"type"`
	Image string     `json:"imag"`
}

type EntityType string

const LeegType EntityType = "leeg"
const TeamType EntityType = "team"
const GameType EntityType = "game"

func NewId() string {
	return uuid.NewString()
}
