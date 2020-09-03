/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/scoir/canis/pkg/datastore"
)

type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"postgres"`
	SSLMode  string `mapstructure:"sslmode"`
	DB       string `mapstructure:"database"`
}

// Provider represents a Postgres DB implementation of the storage.Provider interface
type Provider struct {
	dbURL    string
	adminURL string
	conns    map[string]*postgresDBStore
	sync.RWMutex
}

func (p *Provider) GetAriesProvider() (storage.Provider, error) {
	panic("implement me")
}

type postgresDBStore struct {
	pool      *pgxpool.Pool
	tableName string
}

func (p *postgresDBStore) Insert(d datastore.Doc) (string, error) {
	panic("implement me")
}

func (p *postgresDBStore) List(c datastore.Criteria, gen datastore.DocGen, start, pageSize int) (*datastore.DocList, error) {
	panic("implement me")
}

func (p *postgresDBStore) Get(id string, gen datastore.DocGen) (datastore.Doc, error) {
	panic("implement me")
}

func (p *postgresDBStore) Delete(id string) error {
	panic("implement me")
}

func (p *postgresDBStore) Update(d datastore.Doc) error {
	panic("implement me")
}

// NewProvider instantiates Provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		return nil, errors.New("config missing")
	}

	p := &Provider{
		dbURL: fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			config.User, config.Password, config.Host, config.Port, config.DB, config.SSLMode),
		adminURL: fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s",
			config.User, config.Password, config.Host, config.Port, config.SSLMode),
		conns: map[string]*postgresDBStore{},
	}

	return p, nil
}

// OpenStore opens and returns new connection pool for given name space.
func (p *Provider) OpenStore(name string) (datastore.Store, error) {
	p.Lock()
	defer p.Unlock()

	if name == "" {
		return nil, errors.New("store name is required")
	}

	pool, err := pgxpool.Connect(context.Background(), p.dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create new connection %s: %w", p.dbURL, err)
	}

	createTableStmt := `CREATE Table IF NOT EXISTS ` + name +
		` (key SERIAL NOT NULL, data JSONB, PRIMARY KEY (key));`

	_, err = pool.Exec(context.Background(), createTableStmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create table %s: %w", name, err)
	}

	store := &postgresDBStore{
		pool:      pool,
		tableName: name,
	}

	p.conns[name] = store

	return store, nil
}

// Close closes the provider.
func (p *Provider) Close() error {
	p.Lock()
	defer p.Unlock()

	// timeout?
	for _, store := range p.conns {
		store.pool.Close()
	}

	p.conns = make(map[string]*postgresDBStore)

	return nil
}

// CloseStore closes a previously opened store
func (p *Provider) CloseStore(name string) error {
	p.Lock()
	defer p.Unlock()

	store, exists := p.conns[name]
	if !exists {
		return nil
	}

	delete(p.conns, name)

	store.pool.Close()
	return nil
}

// InsertDID add DID to store
func (p *postgresDBStore) InsertDID(d *datastore.DID) error {

	_, err := p.pool.Exec(context.Background(), fmt.Sprintf(`INSERT INTO %s (data) VALUES ($1)`, p.tableName), d)
	if err != nil {
		return err
	}

	return nil
}

// ListDIDs query DIDs
func (p *postgresDBStore) ListDIDs(c *datastore.DIDCriteria) (*datastore.DIDList, error) {
	if c == nil {
		c = &datastore.DIDCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	var count int
	err := p.pool.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", p.tableName)).Scan(&count)
	if err != nil {
		return nil, err
	}

	var all []*datastore.DID
	rows, err := p.pool.Query(context.Background(), fmt.Sprintf("SELECT data FROM %s LIMIT %d OFFSET %d", p.tableName, c.PageSize, c.Start))
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
		Count: count,
		DIDs:  all,
	}, nil
}

// SetPublicDID update single DID to public, unset remaining
func (p *postgresDBStore) SetPublicDID(DID string) error {

	_, err := p.pool.Exec(context.Background(),
		fmt.Sprintf(`UPDATE %s SET data = jsonb_set(data, '{"Public"}', 'false', true);`, p.tableName))
	if err != nil {
		return err
	}

	_, err = p.pool.Exec(context.Background(),
		fmt.Sprintf(`UPDATE %s SET data = jsonb_set(data, '{"Public"}', 'true', true) WHERE data -> 'DID' -> 'DIDVal' ->> 'DID' = '%s';`,
			p.tableName, DID))
	if err != nil {
		return err
	}

	return nil
}

// GetPublicDID get public DID
func (p *postgresDBStore) GetPublicDID() (*datastore.DID, error) {

	did := &datastore.DID{}
	row := p.pool.QueryRow(context.Background(),
		fmt.Sprintf(`SELECT data FROM %s WHERE data ->> 'Public' = 'true';`, p.tableName))

	err := row.Scan(did)
	if err != nil {
		return nil, err
	}

	return did, nil
}

// InsertSchema add Schema to store
func (p *postgresDBStore) InsertSchema(s *datastore.Schema) (string, error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	_, err := p.pool.Exec(context.Background(),
		fmt.Sprintf(`INSERT INTO %s (data) VALUES ($1)`, p.tableName), s)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

// ListSchema query schemas
func (p *postgresDBStore) ListSchema(c *datastore.SchemaCriteria) (*datastore.SchemaList, error) {
	if c == nil {
		c = &datastore.SchemaCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	var count int
	err := p.pool.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", p.tableName)).Scan(&count)
	if err != nil {
		return nil, err
	}

	var all []*datastore.Schema
	rows, err := p.pool.Query(context.Background(),
		fmt.Sprintf("SELECT data FROM %s LIMIT %d OFFSET %d", p.tableName, c.PageSize, c.Start))
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
		Count:  count,
		Schema: all,
	}, nil
}

// GetSchema return single Schema
func (p *postgresDBStore) GetSchema(id string) (*datastore.Schema, error) {
	s := &datastore.Schema{}
	err := p.pool.QueryRow(context.Background(),
		fmt.Sprintf(`SELECT data FROM %s WHERE data ->> 'ID' = '%s';`, p.tableName, id)).Scan(s)

	if err != nil {
		return nil, err
	}

	return s, nil
}

// DeleteSchema delete single schema
func (p *postgresDBStore) DeleteSchema(id string) error {
	t, err := p.pool.Exec(context.Background(),
		fmt.Sprintf(`DELETE FROM %s WHERE data ->> 'ID' = '%s';`, p.tableName, id))
	if err != nil {
		return err
	}

	if t.RowsAffected() != 1 {
		return errors.New("no schema deleted")
	}

	return nil
}

// UpdateSchema update single schema
func (p *postgresDBStore) UpdateSchema(s *datastore.Schema) error {
	_, err := p.pool.Exec(context.Background(),
		fmt.Sprintf(`UPDATE %s SET data = $1 WHERE data ->> 'ID' = '%s';`,
			p.tableName, s.ID), s)
	if err != nil {
		return err
	}

	return nil
}

// InsertAgent add agent to store
func (p *postgresDBStore) InsertAgent(s *datastore.Agent) (string, error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	_, err := p.pool.Exec(context.Background(),
		fmt.Sprintf(`INSERT INTO %s (data) VALUES ($1)`, p.tableName), s)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

// ListAgent query agents
func (p *postgresDBStore) ListAgent(c *datastore.AgentCriteria) (*datastore.AgentList, error) {
	if c == nil {
		c = &datastore.AgentCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	var count int
	err := p.pool.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", p.tableName)).Scan(&count)
	if err != nil {
		return nil, err
	}

	var all []*datastore.Agent
	rows, err := p.pool.Query(context.Background(),
		fmt.Sprintf("SELECT data FROM %s LIMIT %d OFFSET %d", p.tableName, c.PageSize, c.Start))
	if err != nil {
		return nil, err
	}

	var errs bool
	for rows.Next() {
		s := &datastore.Agent{}
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

	return &datastore.AgentList{
		Count:  count,
		Agents: all,
	}, nil
}

// GetAgent return single agent
func (p *postgresDBStore) GetAgent(id string) (*datastore.Agent, error) {
	a := &datastore.Agent{}
	err := p.pool.QueryRow(context.Background(),
		fmt.Sprintf(`SELECT data FROM %s WHERE data ->> 'ID' = '%s';`, p.tableName, id)).Scan(a)

	if err != nil {
		return nil, err
	}

	return a, nil
}

// GetAgentByInvitation return single agent
func (p *postgresDBStore) GetAgentByInvitation(invitationID string) (*datastore.Agent, error) {
	a := &datastore.Agent{}
	err := p.pool.QueryRow(context.Background(),
		fmt.Sprintf(`SELECT data FROM %s WHERE data ->> 'InvitationID' = '%s';`, p.tableName, invitationID)).Scan(a)

	if err != nil {
		return nil, err
	}

	return a, nil
}

// DeleteAgent delete single agent
func (p *postgresDBStore) DeleteAgent(id string) error {
	t, err := p.pool.Exec(context.Background(),
		fmt.Sprintf(`DELETE FROM %s WHERE data ->> 'ID' = '%s';`, p.tableName, id))
	if err != nil {
		return err
	}

	if t.RowsAffected() != 1 {
		return errors.New("no agent deleted")
	}

	return nil
}

// UpdateAgent delete single agent
func (p *postgresDBStore) UpdateAgent(s *datastore.Agent) error {
	_, err := p.pool.Exec(context.Background(),
		fmt.Sprintf(`UPDATE %s SET data = $1 WHERE data ->> 'ID' = '%s';`,
			p.tableName, s.ID), s)
	if err != nil {
		return err
	}

	return nil
}
