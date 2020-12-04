package mongodb

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
)

func (r *mongoDBStore) InsertCredential(c *datastore.Credential) (string, error) {
	c.ID = uuid.New().String()
	_, err := r.db.Collection(CredentialC).InsertOne(context.Background(), c)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert credential")
	}

	return c.ID, nil
}

func (r *mongoDBStore) FindCredentialByOffer(offerID string) (*datastore.Credential, error) {
	c := &datastore.Credential{}
	err := r.db.Collection(CredentialC).FindOne(context.Background(),
		bson.M{"threadid": offerID, "systemstate": "offered"}).Decode(c)

	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed load offer").Error())
	}

	return c, nil
}

func (r *mongoDBStore) DeleteCredentialByOffer(offerID string) error {
	_, err := r.db.Collection(CredentialC).DeleteOne(context.Background(), bson.M{"threadid": offerID})
	if err != nil {
		return errors.Wrap(err, "unable to delete credential")
	}

	return nil
}

func (r *mongoDBStore) UpdateCredential(c *datastore.Credential) error {
	_, err := r.db.Collection(CredentialC).UpdateOne(context.Background(), bson.M{"id": c.ID}, c)
	if err != nil {
		return errors.Wrap(err, "unable to update credential")
	}

	return nil
}
