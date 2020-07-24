package datastore

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type DIDCriteria struct {
	Start, PageSize int
}

type DIDs []*DID

// Value implements driver.Valuer
func (d DIDs) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner
func (d *DIDs) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type DID struct {
	DID, Verkey, Endpoint string
	Public                bool
}

// Value implements driver.Valuer
func (d DID) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner
func (d *DID) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type DIDList struct {
	Count int
	DIDs  []*DID
}

// Value implements driver.Valuer
func (d DIDList) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner
func (d *DIDList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}
