/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdbstore

import (
	"context"
	"encoding/json"
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
	dbs           map[string]*couchDBStore
	sync.RWMutex
}

const (
	blankHostErrMsg           = "hostURL for new CouchDB provider can't be blank"
	failToCloseProviderErrMsg = "failed to close provider"
)

type Config struct {
	URL string `mapstructure:"url"`
}

// NewProvider instantiates Provider
func NewProvider(hostURL string) (*Provider, error) {
	if hostURL == "" {
		return nil, errors.New(blankHostErrMsg)
	}

	client, err := kivik.New("couch", hostURL)
	if err != nil {
		return nil, err
	}

	p := &Provider{hostURL: hostURL, couchDBClient: client, dbs: map[string]*couchDBStore{}}
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

	store := &couchDBStore{db: db}

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

	p.dbs = make(map[string]*couchDBStore)

	return nil
}

// couchDBStore represents a CouchDB-backed database.
type couchDBStore struct {
	db *kivik.DB
}

// InsertDID add DID to store
func (r *couchDBStore) InsertDID(d *datastore.DID) error {
	if d.DID == "" {
		return errors.New("malformed DID")
	}

	_, err := r.db.Put(context.Background(), d.DID, d)
	if err != nil {
		return err
	}

	return nil
}

// ListDIDs query DIDs
func (r *couchDBStore) ListDIDs(c *datastore.DIDCriteria) (*datastore.DIDList, error) {
	if c == nil {
		c = &datastore.DIDCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	//todo - replace this with a couchdb view
	ctx := context.Background()
	docs, err := r.db.AllDocs(ctx)
	if err != nil {
		return nil, err
	}
	for docs.Next() {
	}

	selector := map[string]interface{}{}
	query := map[string]interface{}{
		"selector": selector,
	}

	if c.Start > 0 {
		query["skip"] = c.Start
	}

	if c.PageSize > 0 {
		query["limit"] = c.PageSize
	}

	var all []*datastore.DID
	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "ListDIDs")
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
		Count: int(docs.TotalRows()),
		DIDs:  all,
	}

	return &out, nil
}

// SetPublicDID update single DID to public, unset remaining
func (r *couchDBStore) SetPublicDID(DID string) error {
	ctx := context.Background()
	query := map[string]interface{}{
		"selector": map[string]interface{}{
		},
	}

	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return err
	}

	var errs bool
	dids := []interface{}{}
	for rows.Next() {
		d := make(map[string]interface{})
		err = rows.ScanDoc(&d)
		if err != nil {
			errs = true
			break
		}

		d["Public"] = d["DID"] == DID
		dids = append(dids, d)
	}

	if errs {
		return errors.New("SetPublicDID failed")
	}

	_, err = r.db.BulkDocs(ctx, dids)
	if err != nil {
		return err
	}

	return nil
}

// GetPublicDID get public DID
func (r *couchDBStore) GetPublicDID() (*datastore.DID, error) {
	ctx := context.Background()
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"Public": true,
		},
		"limit": 1,
	}

	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "GetPublicDID")
	}

	did := &datastore.DID{}
	for rows.Next() {
		err = rows.ScanDoc(did)
	}

	return did, err
}

// InsertSchema add Schema to store
func (r *couchDBStore) InsertSchema(s *datastore.Schema) (string, error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	_, err := r.db.Put(context.Background(), s.ID, s)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

// ListSchema query schemas
func (r *couchDBStore) ListSchema(c *datastore.SchemaCriteria) (*datastore.SchemaList, error) {
	if c == nil {
		c = &datastore.SchemaCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	//todo - replace this with a couchdb view
	ctx := context.Background()
	docs, err := r.db.AllDocs(ctx)
	if err != nil {
		return nil, err
	}

	for docs.Next() {
	}

	selector := map[string]interface{}{}
	query := map[string]interface{}{
		"selector": selector,
	}

	if c.Start > 0 {
		query["skip"] = c.Start
	}

	if c.PageSize > 0 {
		query["limit"] = c.PageSize
	}

	var all []*datastore.Schema
	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "ListSchema")
	}

	var errs bool
	for rows.Next() {
		did := &datastore.Schema{}
		err := rows.ScanDoc(did)
		if err != nil {
			errs = true
			break
		}

		all = append(all, did)
	}

	if errs {
		return nil, errors.New("scanning rows")
	}

	out := datastore.SchemaList{
		Count:  int(docs.TotalRows()),
		Schema: all,
	}

	return &out, nil
}

// GetSchema return single Schema
func (r *couchDBStore) GetSchema(id string) (*datastore.Schema, error) {
	row := r.db.Get(context.Background(), id)

	s := &datastore.Schema{}
	err := row.ScanDoc(s)

	if err != nil {
		return nil, err
	}

	return s, nil
}

// DeleteSchema delete single schema
func (r *couchDBStore) DeleteSchema(id string) error {
	row := r.db.Get(context.Background(), id)
	a := make(map[string]string)
	err := row.ScanDoc(&a)

	if err != nil {
		return err
	}

	_, err = r.db.Delete(context.Background(), id, a["_rev"])
	return err
}

// UpdateSchema update single schema
func (r *couchDBStore) UpdateSchema(s *datastore.Schema) error {
	row := r.db.Get(context.Background(), s.ID)
	a := make(map[string]interface{})
	err := row.ScanDoc(&a)

	if err != nil {
		return err
	}

	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	m["_rev"] = a["_rev"]

	_, err = r.db.Put(context.Background(), s.ID, &m)
	if err != nil {
		return err
	}

	return nil
}

// InsertAgent add agent to store
func (r *couchDBStore) InsertAgent(a *datastore.Agent) (string, error) {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	_, err := r.db.Put(context.Background(), a.ID, a)
	if err != nil {
		return "", err
	}

	return a.ID, nil
}

// ListAgent query agents
func (r *couchDBStore) ListAgent(c *datastore.AgentCriteria) (*datastore.AgentList, error) {
	c = &datastore.AgentCriteria{
		Start:    0,
		PageSize: 10,
	}

	//todo - replace this with a couchdb view
	ctx := context.Background()
	docs, err := r.db.AllDocs(ctx)
	if err != nil {
		return nil, err
	}

	for docs.Next() {
	}

	selector := map[string]interface{}{}
	query := map[string]interface{}{
		"selector": selector,
	}

	if c.Start > 0 {
		query["skip"] = c.Start
	}

	if c.PageSize > 0 {
		query["limit"] = c.PageSize
	}

	var all []*datastore.Agent
	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to find agents")
	}

	var errs bool
	for rows.Next() {
		did := &datastore.Agent{}
		err := rows.ScanDoc(did)
		if err != nil {
			errs = true
			break
		}

		all = append(all, did)
	}

	if errs {
		return nil, errors.New("scanning rows")
	}

	out := datastore.AgentList{
		Count:  int(docs.TotalRows()),
		Agents: all,
	}

	return &out, nil
}

// GetAgent return single agent
func (r *couchDBStore) GetAgent(id string) (*datastore.Agent, error) {
	row := r.db.Get(context.Background(), id)

	a := &datastore.Agent{}
	err := row.ScanDoc(a)

	if err != nil {
		return nil, err
	}

	return a, nil
}

// GetAgentByInvitation return single agent
func (r *couchDBStore) GetAgentByInvitation(invitationID string) (*datastore.Agent, error) {
	ctx := context.Background()
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"InvitationID": invitationID,
		},
		"limit": 1,
	}

	rows, err := r.db.Find(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "GetAgentByInvitation")
	}

	a := &datastore.Agent{}
	for rows.Next() {
		err = rows.ScanDoc(a)
		if err != nil {
			return nil, err
		}
	}

	return a, err
}

// DeleteAgent delete single agent
func (r *couchDBStore) DeleteAgent(id string) error {
	row := r.db.Get(context.Background(), id)
	a := make(map[string]interface{})
	err := row.ScanDoc(&a)

	if err != nil {
		return err
	}

	_, err = r.db.Delete(context.Background(), id, a["_rev"].(string))
	return err
}

// UpdateAgent delete single agent
func (r *couchDBStore) UpdateAgent(a *datastore.Agent) error {
	row := r.db.Get(context.Background(), a.ID)
	t := make(map[string]interface{})
	err := row.ScanDoc(&t)

	if err != nil {
		return err
	}

	b, err := json.Marshal(a)
	if err != nil {
		return err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	m["_rev"] = t["_rev"]

	_, err = r.db.Put(context.Background(), a.ID, &m)
	if err != nil {
		return err
	}

	return nil
}
