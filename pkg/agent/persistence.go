package agent

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
)

type persistence interface {
	SaveSelf(s *Self) error
	GetSelf(id string) (*Self, error)
}

type persister struct {
	self datastore.Store
}

func newPersistence(prov datastore.Provider) (persistence, error) {
	ds, err := prov.OpenStore("self")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open agent self store")
	}
	return &persister{
		self: ds,
	}, nil
}

func (r *persister) SaveSelf(s *Self) error {
	fmt.Println("calling self.Update")
	_, err := r.self.Insert(s)
	if err != nil {
		return errors.Wrap(err, "unable to save agent identity")
	}

	return nil
}

func (r *persister) GetSelf(id string) (*Self, error) {
	out, err := r.self.Get(id, SelfGen)
	if err != nil {
		return nil, errors.Wrap(err, "agent identity not found")
	}

	s, ok := out.(*Self)
	if !ok {
		return nil, errors.New("unexpected error loading agent identity")
	}

	return s, nil
}
