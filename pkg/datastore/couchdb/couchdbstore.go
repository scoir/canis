package couchdbstore

import (
	"context"
	"fmt"
	"sync"

	_ "github.com/go-kivik/couchdb" // The CouchDB driver
	"github.com/go-kivik/kivik"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
)

// Provider represents an CouchDB implementation of the storage.Provider interface
type Provider struct {
	hostURL       string
	couchDBClient *kivik.Client
	dbs           map[string]*CouchDBStore
	sync.RWMutex
}

const (
	blankHostErrMsg           = "hostURL for new CouchDB provider can't be blank"
	failToCloseProviderErrMsg = "failed to close provider"
	couchDBNotFoundErr        = "Not Found:"
)

// Option configures the couchdb provider
type Option func(opts *Provider)

// NewProvider instantiates Provider
func NewProvider(hostURL string) (*Provider, error) {
	if hostURL == "" {
		return nil, errors.New(blankHostErrMsg)
	}

	client, err := kivik.New("couch", hostURL)
	if err != nil {
		return nil, err
	}

	p := &Provider{hostURL: hostURL, couchDBClient: client, dbs: map[string]*CouchDBStore{}}
	return p, nil
}

// OpenStore opens an existing store with the given name and returns it.
func (p *Provider) OpenStore(name string) (datastore.Store, error) {
	p.Lock()
	defer p.Unlock()

	// Check cache first
	cachedStore, existsInCache := p.dbs[name]
	if existsInCache {
		return cachedStore, nil
	}

	err := p.couchDBClient.CreateDB(context.Background(), name)
	if err != nil {
		if err.Error() != "Precondition Failed: The database could not be created, the file already exists." {
			return nil, fmt.Errorf("failed to create db: %w", err)
		}
	}

	db := p.couchDBClient.DB(context.Background(), name)

	if db.Err() != nil {
		return nil, db.Err()
	}

	store := &CouchDBStore{db: db}

	p.dbs[name] = store

	return store, nil
}

// CloseStore closes a previously opened store.
func (p *Provider) CloseStore(name string) error {
	p.Lock()
	defer p.Unlock()

	store, exists := p.dbs[name]
	if !exists {
		return nil
	}

	delete(p.dbs, name)

	return store.db.Close(context.Background())
}

// Close closes the provider.
func (p *Provider) Close() error {
	p.Lock()
	defer p.Unlock()

	for _, store := range p.dbs {
		err := store.db.Close(context.Background())
		if err != nil {
			return fmt.Errorf(failToCloseProviderErrMsg+": %w", err)
		}
	}

	if err := p.couchDBClient.Close(context.Background()); err != nil {
		return err
	}

	p.dbs = make(map[string]*CouchDBStore)

	return nil
}

// CouchDBStore represents a CouchDB-backed database.
type CouchDBStore struct {
	db *kivik.DB
}

func (r *CouchDBStore) InsertDID(d *datastore.DID) error {
	if d.DID == "" {
		return errors.New("wat")
	}

	_, err := r.db.Put(context.Background(), d.DID, d)
	if err != nil {
		return err
	}

	return nil
}

func (r *CouchDBStore) ListDIDs(c *datastore.DIDCriteria) (*datastore.DIDList, error) {
	//todo or fail if no criteria?
	c = &datastore.DIDCriteria{
		Start:    0,
		PageSize: 10,
	}

	query := map[string]interface{}{}
	if c.Start > 0 {
		query["skip"] = c.Start
	}

	if c.PageSize > 0 {
		query["limit"] = c.PageSize
	}

	ctx := context.Background()
	var all []*datastore.DID
	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to find DIDs")
	}

	var errs bool
	for rows.Next() {
		did := &datastore.DID{}
		err := rows.ScanDoc(did)
		if err != nil {
			errs = true
			continue
		}

		all = append(all, did)
	}

	if errs {
		return nil, errors.New("scanning rows")
	}

	out := datastore.DIDList{
		Count: len(all),
		DIDs:  all,
	}

	return &out, nil
}

func (r *CouchDBStore) SetPublicDID(DID string) error {
	panic("implement me")
}

func (r *CouchDBStore) GetPublicDID() (*datastore.DID, error) {
	panic("implement me")
}

func (r *CouchDBStore) InsertSchema(s *datastore.Schema) (string, error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	_, err := r.db.Put(context.Background(), s.ID, s)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

func (r *CouchDBStore) ListSchema(c *datastore.SchemaCriteria) (*datastore.SchemaList, error) {
	//todo or fail if no criteria?
	c = &datastore.SchemaCriteria{
		Start:    0,
		PageSize: 10,
	}

	query := map[string]interface{}{}
	if c.Start > 0 {
		query["skip"] = c.Start
	}

	if c.PageSize > 0 {
		query["limit"] = c.PageSize
	}

	ctx := context.Background()
	var all []*datastore.Schema
	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to find DIDs")
	}

	var errs bool
	for rows.Next() {
		did := &datastore.Schema{}
		err := rows.ScanDoc(did)
		if err != nil {
			errs = true
			continue
		}

		all = append(all, did)
	}

	if errs {
		return nil, errors.New("scanning rows")
	}

	out := datastore.SchemaList{
		Count:  len(all),
		Schema: all,
	}

	return &out, nil
}

func (r *CouchDBStore) GetSchema(id string) (*datastore.Schema, error) {
	row := r.db.Get(context.Background(), id)

	s := &datastore.Schema{}
	err := row.ScanDoc(s)

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r *CouchDBStore) DeleteSchema(id string) error {
	panic("implement me")
}

func (r *CouchDBStore) UpdateSchema(s *datastore.Schema) error {
	panic("implement me")
}

func (r *CouchDBStore) InsertAgent(a *datastore.Agent) (string, error) {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	_, err := r.db.Put(context.Background(), a.ID, a)
	if err != nil {
		return "", err
	}

	return a.ID, nil
}

func (r *CouchDBStore) ListAgent(c *datastore.AgentCriteria) (*datastore.AgentList, error) {
	//todo or fail if no criteria?
	c = &datastore.AgentCriteria{
		Start:    0,
		PageSize: 10,
	}

	query := map[string]interface{}{}
	if c.Start > 0 {
		query["skip"] = c.Start
	}

	if c.PageSize > 0 {
		query["limit"] = c.PageSize
	}

	ctx := context.Background()
	var all []*datastore.Agent
	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to find DIDs")
	}

	var errs bool
	for rows.Next() {
		did := &datastore.Agent{}
		err := rows.ScanDoc(did)
		if err != nil {
			errs = true
			continue
		}

		all = append(all, did)
	}

	if errs {
		return nil, errors.New("scanning rows")
	}

	out := datastore.AgentList{
		Count:  len(all),
		Agents: all,
	}

	return &out, nil
}

func (r *CouchDBStore) GetAgent(id string) (*datastore.Agent, error) {
	row := r.db.Get(context.Background(), id)

	a := &datastore.Agent{}
	err := row.ScanDoc(a)

	if err != nil {
		return nil, err
	}

	return a, nil
}

func (r *CouchDBStore) GetAgentByInvitation(invitationID string) (*datastore.Agent, error) {
	panic("implement me")
}

func (r *CouchDBStore) DeleteAgent(id string) error {
	r.db.Delete(context.Background(), id, "")
	return nil
}

func (r *CouchDBStore) UpdateAgent(s *datastore.Agent) error {
	panic("implement me")
}
