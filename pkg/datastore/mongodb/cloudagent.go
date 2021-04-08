package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
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

func (r *mongoDBStore) InsertCloudAgentConnection(ac *datastore.CloudAgentConnection) error {

	if ac.CloudAgentID == "" {
		return errors.New("cloud agent ID is required")
	}

	if ac.ConnectionID == "" && ac.InvitationID == "" {
		return errors.New("either a connection ID or an invitation ID are required")
	}

	ac.LastUpdated = time.Now()

	_, err := r.db.Collection(CloudAgentConnectionC).InsertOne(context.Background(), ac)
	if err != nil {
		return errors.Wrap(err, "unable to insert agent")
	}

	return nil

}

func (r *mongoDBStore) UpdateCloudAgentConnection(ac *datastore.CloudAgentConnection) error {

	if ac.CloudAgentID == "" {
		return errors.New("cloud agent ID is required")
	}

	criteria := bson.M{"cloudagentid": ac.CloudAgentID}
	if ac.InvitationID != "" {
		criteria["invitationid"] = ac.InvitationID
	} else if ac.ConnectionID != "" {
		criteria["connectionid"] = ac.ConnectionID
	} else {
		return errors.New("either a connection ID or an invitation ID are required")
	}

	ac.LastUpdated = time.Now()

	_, err := r.db.Collection(CloudAgentConnectionC).UpdateOne(context.Background(), criteria, bson.M{"$set": ac})
	if err != nil {
		return errors.Wrap(err, "unable to update cloud agent")
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

func (r *mongoDBStore) GetCloudAgentConnection(a *datastore.CloudAgent, invitationID string) (*datastore.CloudAgentConnection, error) {
	ac := &datastore.CloudAgentConnection{}
	err := r.db.Collection(CloudAgentConnectionC).FindOne(context.Background(),
		bson.M{"cloudagentid": a.ID, "invitationid": invitationID}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to load cloud agent connection")
	}

	return ac, nil
}

func (r *mongoDBStore) GetCloudAgentConnectionForDIDs(myDID, theirDID string) (*datastore.CloudAgentConnection, error) {
	ac := &datastore.CloudAgentConnection{}
	err := r.db.Collection(CloudAgentConnectionC).FindOne(context.Background(),
		bson.M{"mydid": myDID, "theirdid": theirDID}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "failed load agent connection")
	}

	return ac, nil
}

func (r *mongoDBStore) InsertCloudAgentCredential(cred *datastore.CloudAgentCredential) error {
	_, err := r.db.Collection(CloudAgentCredentialC).InsertOne(context.Background(), cred)
	if err != nil {
		return errors.Wrap(err, "unable to insert agent")
	}

	return nil

}

func (r *mongoDBStore) UpdateCloudAgentCredential(cred *datastore.CloudAgentCredential) error {
	_, err := r.db.Collection(CloudAgentCredentialC).UpdateOne(context.Background(), bson.M{"id": cred.ID}, bson.M{"$set": cred})
	if err != nil {
		return errors.Wrap(err, "unable to insert agent")
	}

	return nil

}

func (r *mongoDBStore) ListCloudAgentCredentials(a *datastore.CloudAgent) ([]*datastore.CloudAgentCredential, error) {
	ctx := context.Background()
	var ac []*datastore.CloudAgentCredential
	results, err := r.db.Collection(CloudAgentCredentialC).Find(ctx,
		bson.M{"cloudagentid": a.ID})

	if err != nil {
		return nil, errors.Wrap(err, "unable to list cloud agent credentials")
	}

	err = results.All(ctx, &ac)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode cloud agent credentials")
	}

	return ac, nil
}

func (r *mongoDBStore) DeleteCloudAgentCredential(a *datastore.CloudAgent, id string) error {
	_, err := r.db.Collection(CloudAgentCredentialC).DeleteOne(context.Background(),
		bson.M{"cloudagentid": a.ID, "id": id})

	if err != nil {
		return errors.Wrap(err, "unable to delete cloud agent credential")
	}

	return nil
}

func (r *mongoDBStore) GetCloudAgentCredential(a *datastore.CloudAgent, id string) (*datastore.CloudAgentCredential, error) {
	ac := &datastore.CloudAgentCredential{}
	err := r.db.Collection(CloudAgentCredentialC).FindOne(context.Background(),
		bson.M{"cloudagentid": a.ID, "id": id}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to load cloud agent credential")
	}

	return ac, nil
}

func (r *mongoDBStore) GetCloudAgentCredentialFromThread(cloudAgentID string, thid string) (*datastore.CloudAgentCredential, error) {
	ac := &datastore.CloudAgentCredential{}
	err := r.db.Collection(CloudAgentCredentialC).FindOne(context.Background(),
		bson.M{"cloudagentid": cloudAgentID, "threadid": thid}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to load cloud agent credential")
	}

	return ac, nil
}

func (r *mongoDBStore) InsertCloudAgentProofRequest(cred *datastore.CloudAgentProofRequest) error {
	_, err := r.db.Collection(CloudAgentProofRequestC).InsertOne(context.Background(), cred)
	if err != nil {
		return errors.Wrap(err, "unable to insert agent")
	}

	return nil

}

func (r *mongoDBStore) UpdateCloudAgentProofRequest(cred *datastore.CloudAgentProofRequest) error {
	_, err := r.db.Collection(CloudAgentProofRequestC).UpdateOne(context.Background(), bson.M{"id": cred.ID}, bson.M{"$set": cred})
	if err != nil {
		return errors.Wrap(err, "unable to insert agent")
	}

	return nil

}

func (r *mongoDBStore) ListCloudAgentProofRequests(a *datastore.CloudAgent) ([]*datastore.CloudAgentProofRequest, error) {
	ctx := context.Background()
	var ac []*datastore.CloudAgentProofRequest
	results, err := r.db.Collection(CloudAgentProofRequestC).Find(ctx,
		bson.M{"cloudagentid": a.ID})

	if err != nil {
		return nil, errors.Wrap(err, "unable to list cloud agent ProofRequests")
	}

	err = results.All(ctx, &ac)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode cloud agent ProofRequests")
	}

	return ac, nil
}

func (r *mongoDBStore) DeleteCloudAgentProofRequest(a *datastore.CloudAgent, id string) error {
	_, err := r.db.Collection(CloudAgentProofRequestC).DeleteOne(context.Background(),
		bson.M{"cloudagentid": a.ID, "id": id})

	if err != nil {
		return errors.Wrap(err, "unable to delete cloud agent ProofRequest")
	}

	return nil
}

func (r *mongoDBStore) GetCloudAgentProofRequest(a *datastore.CloudAgent, id string) (*datastore.CloudAgentProofRequest, error) {
	ac := &datastore.CloudAgentProofRequest{}
	err := r.db.Collection(CloudAgentProofRequestC).FindOne(context.Background(),
		bson.M{"cloudagentid": a.ID, "id": id}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to load cloud agent ProofRequest")
	}

	return ac, nil
}

func (r *mongoDBStore) GetCloudAgentProofRequestFromThread(cloudAgentID string, thid string) (*datastore.CloudAgentProofRequest, error) {
	ac := &datastore.CloudAgentProofRequest{}
	err := r.db.Collection(CloudAgentProofRequestC).FindOne(context.Background(),
		bson.M{"cloudagentid": cloudAgentID, "threadid": thid}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to load cloud agent ProofRequest")
	}

	return ac, nil
}
