package postgres

import (
	"fmt"
)

// Config todo
type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"postgres"`
	SSLMode  string `mapstructure:"sslmode"`
}

// String todo
func (r *Config) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/?sslmode=%s", r.User, r.Password, r.Host, r.Port, r.SSLMode)
}

// AdminString hate this
func (r *Config) AdminString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		r.Host, r.Port, r.User, r.Password, r.SSLMode)
}
