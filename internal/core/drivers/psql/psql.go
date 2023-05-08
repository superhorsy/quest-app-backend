package psql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
)

const (
	// ErrConnect is returned when we cannot connect to the database.
	ErrConnect = errors.Error("failed to connect to db")
	// ErrClose is returned when we cannot close the database.
	ErrClose = errors.Error("failed to close db connection")
)

// Config represents the configuration for our database.
type Config struct {
	DSN string `env:"DSN" validate:"required"`
}

// Driver provides an implementation for connecting to a database.
type Driver struct {
	cfg Config
	db  *sqlx.DB
}

// New instantiates an instance of the Driver.
func New(cfg Config) *Driver {
	return &Driver{
		cfg: cfg,
	}
}

// Connect connects to the database.
func (d *Driver) Connect(_ context.Context) error {
	db, err := sqlx.Connect("postgres", d.cfg.DSN)
	if err != nil {
		return ErrConnect.Wrap(err)
	}

	d.db = db

	return nil
}

// Close closes the database connection.
func (d *Driver) Close(_ context.Context) error {
	if err := d.db.Close(); err != nil {
		return ErrClose.Wrap(err)
	}

	return nil
}

// GetDB returns the underlying database connection.
func (d *Driver) GetDB() *sqlx.DB {
	return d.db
}
