package store

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/helpers"
	"github.com/superhorsy/quest-app-backend/internal/media/model"
)

// DB represents a type for interfacing with a database.
type DB interface {
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
}

// RecordStore provides functionality for working with a database.
type RecordStore struct {
	db DB
}

// New will instantiate a new instance of RecordStore.
func New(db DB) *RecordStore {
	return &RecordStore{
		db: db,
	}
}

func (s *RecordStore) InsertMedia(ctx context.Context, m *model.MediaRecord) (*model.MediaRecord, error) {
	m.CreatedAt = helpers.TimeNow()
	m.UpdatedAt = m.CreatedAt

	res, err := s.db.NamedQueryContext(ctx,
		`INSERT INTO 
		media(storage, type, filename, link, created_at, updated_at) 
		VALUES (:storage, :type, :filename, :link, :created_at, :updated_at) 
		RETURNING *`, m)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	if !res.Next() {
		return nil, errors.ErrUnknown
	}

	if err := res.StructScan(&m); err != nil {
		return nil, errors.ErrUnknown.Wrap(err)
	}

	return m, nil
}

func (s *RecordStore) UpdateMedia(ctx context.Context, m *model.MediaRecord) (*model.MediaRecord, error) {
	m.UpdatedAt = helpers.TimeNow()

	res, err := s.db.NamedQueryContext(ctx,
		`UPDATE media SET storage = :storage, type = :type, filename = :filename, link = :link, updated_at = :updated_at 
WHERE id = :id RETURNING *`, m)

	if err != nil {
		return nil, err
	}
	defer res.Close()

	if !res.Next() {
		return nil, errors.ErrUnknown
	}

	if err := res.StructScan(&m); err != nil {
		return nil, errors.ErrUnknown.Wrap(err)
	}

	return m, nil
}

func (s *RecordStore) GetMedia(ctx context.Context, id string) (*model.MediaRecord, error) {
	var m model.MediaRecord

	if err := s.db.GetContext(ctx, &m, "SELECT * FROM media WHERE id = $1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound.Wrap(err)
		}

		return nil, errors.ErrUnknown.Wrap(err)
	}

	return &m, nil
}
