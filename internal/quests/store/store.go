package store

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/quests/model"
)

const (
	// ErrInvalidEmail is returned when the email is not a valid address or is empty.
	ErrInvalidEmail = errors.Error("invalid_email: email is invalid")
	// ErrEmailAlreadyUsed is returned when the email address is already used via another user.
	ErrEmailAlreadyUsed = errors.Error("email_already_used: email is already in use")
	// ErrEmptyNickname is returned when the nickname is empty.
	ErrEmptyNickname = errors.Error("empty_nickname: nickname is empty")
	// ErrNicknameAlreadyUsed is returned when the nickname is already used via another user.
	ErrNicknameAlreadyUsed = errors.Error("nickname_already_used: nickname is already in use")
	// ErrEmptyPassword is returned when the password is empty.
	ErrEmptyPassword = errors.Error("empty_password: password is empty")
	// ErrEmptyCountry is returned when the country is empty.
	ErrEmptyCountry = errors.Error("empty_country: password is empty")
	// ErrInvalidID si returned when the ID is not a valid UUID or is empty.
	ErrInvalidID = errors.Error("invalid_id: id is invalid")
	// ErrUserNotUpdated is returned when a record can't be found to update.
	ErrUserNotUpdated = errors.Error("user_not_updated: user record wasn't updated")
	// ErrUserNotDeleted is returned when a record can't be found to delete.
	ErrUserNotDeleted = errors.Error("user_not_deleted: user record wasn't deleted")
	// ErrInvalidFilters is returned when the filters for finding a user are not valid.
	ErrInvalidFilters = errors.Error("invalid_filters: filters invalid for finding user")
)

const (
	pqErrInvalidTextRepresentation = "invalid_text_representation"
)

var timeNow = func() *time.Time {
	now := time.Now().UTC()
	return &now
}

// DB represents a type for interfacing with a postgres database.
type DB interface {
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
}

// Store provides functionality for working with a postgres database.
type Store struct {
	db DB
}

// New will instantiate a new instance of Store.
func New(db DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) saveQuest(ctx context.Context, quest *model.Quest) (*model.Quest, error) {
	quest.CreatedAt = timeNow()
	quest.UpdatedAt = quest.CreatedAt

	res, err := s.db.NamedQueryContext(ctx,
		`INSERT INTO 
		quests("name",description,"owner",created_at,updated_at) 
		VALUES (:name,:description,:owner,:created_at, :updated_at) 
		RETURNING *`, quest)
	if err = checkWriteError(err); err != nil {
		return nil, err
	}
	defer res.Close()

	if !res.Next() {
		return nil, errors.ErrUnknown
	}

	createdQuest := &model.Quest{}

	if err := res.StructScan(&createdQuest); err != nil {
		return nil, errors.ErrUnknown.Wrap(err)
	}
	return createdQuest, nil
}

func (s *Store) updateSteps(ctx context.Context, createdQuest *model.Quest, steps []model.Step) (*model.Quest, error) {
	for i, _ := range steps {
		steps[i].QuestId = createdQuest.ID
		steps[i].CreatedAt = createdQuest.CreatedAt
		steps[i].UpdatedAt = createdQuest.CreatedAt
	}

	res, err := s.db.NamedQueryContext(ctx, `INSERT INTO
			steps(quest_id, sort,description,question_type,question_content,answer_type,answer_content,created_at,updated_at)
			VALUES (:quest_id, :sort,:description,:question_type,:question_content,:answer_type,:answer_content,:created_at,:updated_at)
			RETURNING *`, steps)

	if err = checkWriteError(err); err != nil {
		return nil, err
	}

	createdQuest.Steps = []model.Step{}
	var step model.Step
	for res.Next() {
		if err := res.StructScan(&step); err != nil {
			return nil, errors.ErrUnknown.Wrap(err)
		}
		createdQuest.Steps = append(createdQuest.Steps, step)
	}

	defer res.Close()
	return createdQuest, nil
}

func (s *Store) updateEmails(ctx context.Context, quest *model.Quest, emails *[]model.Email) error {
	for _, email := range *emails {
		arg := map[string]interface{}{
			"quest_id": quest.ID,
			"email":    email,
		}
		res, err := s.db.NamedQueryContext(ctx, `INSERT INTO quest_to_email(quest_id, email) VALUES (:quest_id, :email)`, arg)
		if err = checkWriteError(err); err != nil {
			return err
		}
		defer res.Close()
	}
	return nil
}

// InsertQuest will add a new quest to the database using the provided data.
func (s *Store) InsertQuest(ctx context.Context, quest *model.Quest) (*model.Quest, error) {
	createdQuest, err := s.saveQuest(ctx, quest)
	if err != nil {
		return nil, err
	}

	createdQuest, err = s.updateSteps(ctx, createdQuest, quest.Steps)
	if err != nil {
		return nil, err
	}

	err = s.updateEmails(ctx, createdQuest, quest.Emails)
	if err != nil {
		return nil, err
	}
	createdQuest.Emails = quest.Emails

	return createdQuest, nil
}

//nolint:cyclop
func checkWriteError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code.Name() {
		case "string_data_right_truncation":
			return errors.ErrValidation.Wrap(err)
		case "check_violation":
			switch {
			case strings.Contains(pqErr.Error(), "email_check"):
				return ErrInvalidEmail.Wrap(errors.ErrValidation.Wrap(err))
			case strings.Contains(pqErr.Error(), "users_nickname_check"):
				return ErrEmptyNickname.Wrap(errors.ErrValidation.Wrap(err))
			case strings.Contains(pqErr.Error(), "users_password_check"):
				return ErrEmptyPassword.Wrap(errors.ErrValidation.Wrap(err))
			case strings.Contains(pqErr.Error(), "users_country_check"):
				return ErrEmptyCountry.Wrap(errors.ErrValidation.Wrap(err))
			default:
				return errors.ErrValidation.Wrap(err)
			}
		case "not_null_violation":
			switch {
			case strings.Contains(pqErr.Error(), "email"):
				return ErrInvalidEmail.Wrap(errors.ErrValidation.Wrap(err))
			case strings.Contains(pqErr.Error(), "nickname"):
				return ErrEmptyNickname.Wrap(errors.ErrValidation.Wrap(err))
			case strings.Contains(pqErr.Error(), "password"):
				return ErrEmptyPassword.Wrap(errors.ErrValidation.Wrap(err))
			case strings.Contains(pqErr.Error(), "country"):
				return ErrEmptyCountry.Wrap(errors.ErrValidation.Wrap(err))
			default:
				return errors.ErrValidation.Wrap(err)
			}
		case "unique_violation":
			if strings.Contains(pqErr.Error(), "email_unique") {
				return ErrEmailAlreadyUsed.Wrap(errors.ErrValidation.Wrap(err))
			} else if strings.Contains(pqErr.Error(), "nickname_unique") {
				return ErrNicknameAlreadyUsed.Wrap(errors.ErrValidation.Wrap(err))
			}
			return errors.ErrValidation.Wrap(err)
		case "invalid_text_representation":
			if strings.Contains(pqErr.Error(), "uuid") {
				return ErrInvalidID.Wrap(errors.ErrValidation.Wrap(err))
			}
		}
	}

	return errors.ErrUnknown.Wrap(err)
}
