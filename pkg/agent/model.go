package agent

import (
	"github.com/scoir/canis/pkg/datastore"
)

type Self struct {
	ID           string
	HasPublicDID bool
	PublicDID    string
}

func (r *Self) GetID() string {
	return r.ID
}

func SelfGen() datastore.Doc {
	return &Self{}
}
