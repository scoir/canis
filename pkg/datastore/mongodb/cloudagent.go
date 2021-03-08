package mongodb

import (
	"context"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/scoir/canis/pkg/datastore"
)

func (r *mongoDBStore) RegisterCloudAgent(externalID string, publicKey, nextKey []byte) (string, error) {
	a := &datastore.CloudAgent{
		ID:        uuid.New().String(),
		PublicKey: publicKey,
		NextKey:   nextKey,
	}

	_, err := r.db.Collection(CloudAgentC).InsertOne(context.Background(), a)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert agent")
	}

	return a.ID, nil
}

func (r *mongoDBStore) GetCloudAgent(ID string) (*datastore.CloudAgent, error) {
	a := &datastore.CloudAgent{}

	err := r.db.Collection(CloudAgentC).FindOne(context.Background(), bson.M{"id": ID}).Decode(a)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load cloud agent")
	}

	return a, nil
}

func (r *mongoDBStore) GetCloudAgentForDID(myDID string) (*datastore.CloudAgent, error) {
	ac := &datastore.CloudAgentConnection{}
	err := r.db.Collection(CloudAgentConnectionC).FindOne(context.Background(),
		bson.M{"mydid": myDID}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "failed load cloud agent connection")
	}

	return r.GetCloudAgent(ac.CloudAgentID)
}

func (r *mongoDBStore) UpdateCloudAgent(a *datastore.CloudAgent) error {
	_, err := r.db.Collection(CloudAgentC).UpdateOne(context.Background(), bson.M{"_id": a.ID}, bson.M{"$set": a})
	if err != nil {
		return errors.Wrap(err, "unable to update cloud agent")
	}

	return nil
}

func (r *mongoDBStore) InsertCloudAgentConnection(a *datastore.CloudAgent, conn *didexchange.Connection) error {
	ac := &datastore.CloudAgentConnection{
		CloudAgentID: a.ID,
		TheirLabel:   conn.TheirLabel,
		TheirDID:     conn.TheirDID,
		MyDID:        conn.MyDID,
		ConnectionID: conn.ConnectionID,
	}

	_, err := r.db.Collection(CloudAgentConnectionC).InsertOne(context.Background(), ac)
	if err != nil {
		return errors.Wrap(err, "unable to insert agent")
	}

	return nil

}

func (r *mongoDBStore) ListCloudAgentConnections(a *datastore.CloudAgent) ([]*datastore.CloudAgentConnection, error) {
	ctx := context.Background()
	var ac []*datastore.CloudAgentConnection
	results, err := r.db.Collection(CloudAgentConnectionC).Find(ctx,
		bson.M{"cloudagentid": a.ID})

	if err != nil {
		return nil, errors.Wrap(err, "unable to list cloud agent connections")
	}

	err = results.All(ctx, &ac)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode cloud agent connections")
	}

	return ac, nil
}

func (r *mongoDBStore) DeleteCloudAgentConnection(a *datastore.CloudAgent, externalID string) error {
	_, err := r.db.Collection(CloudAgentConnectionC).DeleteMany(context.Background(),
		bson.M{"cloudagentid": a.ID, "externalid": externalID})

	if err != nil {
		return errors.Wrap(err, "unable to delete cloud agent connection")
	}

	return nil
}

func (r *mongoDBStore) GetCloudAgentConnection(a *datastore.CloudAgent, externalID string) (*datastore.CloudAgentConnection, error) {
	ac := &datastore.CloudAgentConnection{}
	err := r.db.Collection(CloudAgentConnectionC).FindOne(context.Background(),
		bson.M{"cloudagentid": a.ID, "externalid": externalID}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to load cloud agent connection")
	}

	return ac, nil
}

func (r *mongoDBStore) GetCloudAgentConnectionForDID(a *datastore.CloudAgent, theirDID string) (*datastore.CloudAgentConnection, error) {
	ac := &datastore.CloudAgentConnection{}
	err := r.db.Collection(CloudAgentConnectionC).FindOne(context.Background(),
		bson.M{"cloudagentid": a.ID, "theirdid": theirDID}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "failed load agent connection")
	}

	return ac, nil
}
