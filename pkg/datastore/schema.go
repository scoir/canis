package datastore

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)


type SchemaCriteria struct {
	Start, PageSize int
	Name            string
}

type SchemaList struct {
	Count  int
	Schema []*Schema
}

// Value implements driver.Valuer
func (d SchemaList) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner
func (d *SchemaList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type Schema struct {
	ID         string
	Name       string
	Version    string
	Attributes []*Attribute
}

// Value implements driver.Valuer
func (d Schema) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner
func (d *Schema) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type Schemas []*Schema

// Value implements driver.Valuer
func (d Schemas) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner
func (d *Schemas) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}