package model

import "github.com/google/uuid"

type EntityRef struct {
	ID       string     `json:"id"`
	Text     string     `json:"text"`
	Type     EntityType `json:"type"`
	ImageURL string     `json:"imageURL"`
}

type EntityRefList []EntityRef

func (e EntityRefList) WithID(id string) EntityRef {
	for _, entity := range e {
		if entity.ID == id {
			return entity
		}
	}
	return EntityRef{}
}

func (e EntityRefList) HasID(id string) bool {
	for _, entity := range e {
		if entity.ID == id {
			return true
		}
	}
	return false
}

func (e EntityRefList) Diff(oe EntityRefList) EntityRefList {
	diffList := EntityRefList{}

	for _, ref := range e {
		if !oe.HasID(ref.ID) {
			diffList = append(diffList, ref)
		}
	}
	return diffList
}
func (e EntityRefList) RemoveAll(id string) EntityRefList {
	newEntities := EntityRefList{}
	for _, ref := range e {
		if ref.ID != id {
			newEntities = append(newEntities, ref)
		}
	}
	return newEntities
}

func (e EntityRefList) RemoveFirst(id string) EntityRefList {
	newEntities := EntityRefList{}
	removedOne := false
	for _, ref := range e {
		if ref.ID != id || removedOne {
			newEntities = append(newEntities, ref)
		} else if !removedOne {
			removedOne = true
		}
	}
	return newEntities
}

func (e EntityRefList) Update(entity EntityRef) EntityRefList {
	newEntities := EntityRefList{}
	for _, ref := range e {
		if ref.ID == entity.ID {
			ref.Text = entity.Text
			ref.ImageURL = entity.ImageURL
		}
		newEntities = append(newEntities, ref)
	}
	return newEntities
}

type EntityType string

const LEEG EntityType = "leeg"
const TEAM EntityType = "team"
const GAME EntityType = "game"
const ROUND EntityType = "round"

const LEEG_ID = "leeg-id"

func NewId() string {
	return uuid.NewString()
}
