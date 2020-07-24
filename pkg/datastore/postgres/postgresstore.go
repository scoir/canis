/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/scoir/canis/pkg/datastore"
)

const (
	tablePrefix = "t_"
)

// Provider represents a Postgres DB implementation of the storage.Provider interface
type Provider struct {
	dbURL    string
	adminDB  *sql.DB
	db       *sql.DB
	dbs      map[string]*sqlDBStore
	dbPrefix string
	sync.RWMutex
}

type sqlDBStore struct {
	db        *sql.DB
	tableName string
}

// NewProvider instantiates Provider
func NewProvider(config *Config) (*Provider, error) {

	if config == nil {
		return nil, errors.New("info for new postgres DB provider can't be empty")
	}

	db, err := sql.Open("postgres", config.String())
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	admindb, err := sql.Open("postgres", config.AdminString())
	if err != nil {
		return nil, fmt.Errorf("failed to open admin connection: %w", err)
	}

	p := &Provider{
		dbURL:   config.String(),
		db:      db,
		adminDB: admindb,
		dbs:     map[string]*sqlDBStore{},
	}
	pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))


	return p, nil
}

// OpenStore opens and returns new db for given name space.
func (p *Provider) OpenStore(name string) (datastore.Store, error) {
	p.Lock()
	defer p.Unlock()

	if name == "" {
		return nil, errors.New("store name is required")
	}

	fmt.Println("prefix", p.dbPrefix)

	if p.dbPrefix != "" {
		name = p.dbPrefix + "_" + name
	}

	// query the postgres db to see if the store exists
	statement := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '` + name + `');`

	row := p.adminDB.QueryRow(statement)

	var exists bool
	err := row.Scan(&exists)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println("CREATE DATABASE ", name, p.dbURL)

	if exists == false {
		statement = fmt.Sprintf("CREATE DATABASE %s;", name)
		_, err = p.adminDB.Exec(statement)
		if err != nil {
			return nil, err
		}
	}

	// connect to store
	newDBConn, err := sql.Open("postgres", p.dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create new connection %s: %w", p.dbURL, err)
	}

	fmt.Println("new conn")

	tableName := tablePrefix + name
	createTableStmt := `CREATE Table IF NOT EXISTS ` + tableName +
		` (key SERIAL NOT NULL, data JSONB, PRIMARY KEY (key));`

	_, err = newDBConn.Exec(createTableStmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create table %s: %w", name, err)
	}

	store := &sqlDBStore{
		db:        newDBConn,
		tableName: tableName,
	}

	p.dbs[name] = store

	return store, nil
}

// Close closes the provider.
func (p *Provider) Close() error {
	p.Lock()
	defer p.Unlock()

	for _, store := range p.dbs {
		err := store.db.Close()
		if err != nil {
			return fmt.Errorf("failed to close provider: %w", err)
		}
	}

	if err := p.db.Close(); err != nil {
		return err
	}

	p.dbs = make(map[string]*sqlDBStore)

	return nil
}

// CloseStore closes a previously opened store
func (p *Provider) CloseStore(name string) error {
	p.Lock()
	defer p.Unlock()

	if p.dbPrefix != "" {
		name = p.dbPrefix + "_" + name
	}

	store, exists := p.dbs[name]
	if !exists {
		return nil
	}

	delete(p.dbs, name)

	return store.db.Close()
}

// InsertDID todo
func (p *sqlDBStore) InsertDID(d *datastore.DID) error {
	_, err := p.db.Exec("INSERT INTO "+p.tableName+" (data) VALUES ($1)", d)
	if err != nil {
		return err
	}

	return nil
}

// ListDIDs todo
func (p *sqlDBStore) ListDIDs(c *datastore.DIDCriteria) (*datastore.DIDList, error) {
	if c == nil {
		c = &datastore.DIDCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	all := []*datastore.DID{}
	rows, err := p.db.Query(fmt.Sprintf("SELECT data FROM %s LIMIT %d OFFSET %d", p.tableName, c.PageSize, c.Start))
	if err != nil {
		return nil, err
	}

	var errs bool
	for rows.Next() {
		did := &datastore.DID{}
		err := rows.Scan(did)
		if err != nil {
			errs = true
			continue
		}

		all = append(all, did)
	}

	if errs {
		return nil, errors.New("scanning rows")
	}

	return &datastore.DIDList{
		Count: len(all),
		DIDs:  all,
	}, nil
}

// SetPublicDID todo
func (p *sqlDBStore) SetPublicDID(DID string) error {

	_, err := p.db.Exec(`UPDATE ` + p.tableName + ` SET data = jsonb_set(data, '{"Public"}', 'false', true);`)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(`UPDATE ` + p.tableName + ` SET data = jsonb_set(data, '{"Public"}', 'true', true) WHERE data ->> 'DID' = '` + DID + `';`)
	if err != nil {
		return err
	}

	return nil
}

// GetPublicDID todo
func (p *sqlDBStore) GetPublicDID() (*datastore.DID, error) {

	did := &datastore.DID{}
	row := p.db.QueryRow(`SELECT data FROM ` + p.tableName + ` WHERE data ->> 'Public' = 'true';`)

	err := row.Scan(did)
	if err != nil {
		return nil, err
	}

	return did, nil
}

// InsertSchema todo
func (p *sqlDBStore) InsertSchema(s *datastore.Schema) (string, error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	_, err := p.db.Exec(fmt.Sprintf(`INSERT INTO %s (data) VALUES ($1)`, p.tableName), s)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

// ListSchema todo
func (p *sqlDBStore) ListSchema(c *datastore.SchemaCriteria) (*datastore.SchemaList, error) {
	if c == nil {
		c = &datastore.SchemaCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	var all []*datastore.Schema
	rows, err := p.db.Query(fmt.Sprintf("SELECT data FROM %s LIMIT %d OFFSET %d", p.tableName, c.PageSize, c.Start))
	if err != nil {
		return nil, err
	}

	var errs bool
	for rows.Next() {
		s := &datastore.Schema{}
		err := rows.Scan(s)
		if err != nil {
			errs = true
			continue
		}

		all = append(all, s)
	}

	if errs {
		return nil, errors.New("scanning rows")
	}

	return &datastore.SchemaList{
		Count:  len(all),
		Schema: all,
	}, nil
}

// GetSchema todo
func (p *sqlDBStore) GetSchema(id string) (*datastore.Schema, error) {
	return nil, nil
}

//DeleteSchema todo
func (p *sqlDBStore) DeleteSchema(id string) error {
	return nil
}

// UpdateSchema todo
func (p *sqlDBStore) UpdateSchema(s *datastore.Schema) error {
	return nil
}

// InsertAgent todo
func (p *sqlDBStore) InsertAgent(s *datastore.Agent) (string, error) {
	return "", nil
}

// ListAgent todo
func (p *sqlDBStore) ListAgent(c *datastore.AgentCriteria) (*datastore.AgentList, error) {
	return nil, nil
}

// GetAgent todo
func (p *sqlDBStore) GetAgent(id string) (*datastore.Agent, error) {
	return nil, nil
}

// GetAgentByInvitation todo
func (p *sqlDBStore) GetAgentByInvitation(invitationID string) (*datastore.Agent, error) {
	return nil, nil
}

// DeleteAgent todo
func (p *sqlDBStore) DeleteAgent(id string) error {
	return nil
}

// UpdateAgent todo
func (p *sqlDBStore) UpdateAgent(s *datastore.Agent) error {
	return nil
}
