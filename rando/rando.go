package rando

import (
	"errors"
	"leeg/model"
	"math/rand/v2"
)

type RandoConfig struct {
	RandomSeed uint64
	Random     *rand.Rand
}

func (r *RandoConfig) Init() {
	if r.RandomSeed != 0 {
		r.Random = rand.New(rand.NewPCG(r.RandomSeed, r.RandomSeed+1))
	}
}

func (r RandoConfig) RandFrom(min, max int) int {
	if r.Random != nil {
		return r.Random.IntN(max-min) + min
	} else {
		return rand.IntN(max+1-min) + min
	}
}

func (r RandoConfig) RandomEntity(entityList model.EntityRefList) (model.EntityRef, error) {
	if len(entityList) < 1 {
		return model.EntityRef{}, errors.New("can't select from an empty list")
	}
	return entityList[r.RandFrom(0, len(entityList)-1)], nil

}
