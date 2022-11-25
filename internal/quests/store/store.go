package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"github.com/superhorsy/quest-app-backend/internal/transport/http"
	"go.uber.org/zap"
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
	// ErrInvalidID si returned when the ID is not a valid UUID or is empty.
	ErrInvalidID       = errors.Error("invalid_id: id is invalid")
	ErrQuestNotDeleted = errors.Error("quest not deleted")
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

// GetQuest fetches quest by id
func (s *Store) GetQuest(ctx context.Context, id string) (*model.QuestWithSteps, error) {
	var q model.QuestWithSteps

	userId := ctx.Value(http.ContextUserIdKey).(string)

	if err := s.db.GetContext(ctx, &q, "SELECT * FROM quests WHERE id = $1 AND owner = $2", id, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound.Wrap(err)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == pqErrInvalidTextRepresentation && strings.Contains(pqErr.Error(), "uuid") {
				return nil, ErrInvalidID.Wrap(errors.ErrValidation.Wrap(err))
			}
		}

		return nil, errors.ErrUnknown.Wrap(err)
	}

	res, err := s.db.QueryxContext(ctx, `SELECT * FROM steps WHERE quest_id = $1`, id)

	if err = checkWriteError(err); err != nil {
		return nil, err
	}

	q.Steps = []model.Step{}
	var step model.Step
	for res.Next() {
		if err := res.StructScan(&step); err != nil {
			return nil, errors.ErrUnknown.Wrap(err)
		}
		q.Steps = append(q.Steps, step)
	}

	defer res.Close()

	return &q, nil
}

func (s *Store) saveQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error) {
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

	createdQuest := &model.QuestWithSteps{}

	if err := res.StructScan(&createdQuest); err != nil {
		return nil, errors.ErrUnknown.Wrap(err)
	}
	return createdQuest, nil
}

func (s *Store) UpdateQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error) {
	quest.UpdatedAt = timeNow()

	uId := ctx.Value(http.ContextUserIdKey).(string)
	res, err := s.db.NamedQueryContext(ctx,
		fmt.Sprintf(`UPDATE quests SET "name" = :name, description = :description, updated_at = :updated_at 
			WHERE id = :id AND "owner" = '%s' RETURNING *`, uId), quest)
	if err = checkWriteError(err); err != nil {
		return nil, err
	}
	if !res.Next() {
		return nil, errors.ErrUnknown
	}
	defer res.Close()

	updatedQuest := &model.QuestWithSteps{}

	if err := res.StructScan(&updatedQuest); err != nil {
		return nil, errors.ErrUnknown.Wrap(err)
	}

	// Delete saved steps
	_, err = s.db.ExecContext(ctx, "DELETE FROM steps WHERE quest_id = $1", updatedQuest.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound.Wrap(err)
		}
		return nil, errors.ErrUnknown.Wrap(err)
	}

	// Update steps if there are any
	if len(quest.Steps) == 0 {
		return updatedQuest, nil
	}

	// Remove IDs from steps if there were any
	for i, step := range quest.Steps {
		emptyId := ""
		if step.ID != &emptyId {
			quest.Steps[i].ID = &emptyId
		}
	}

	updatedQuest, err = s.updateSteps(ctx, updatedQuest, quest.Steps)
	if err = checkWriteError(err); err != nil {
		return nil, err
	}

	return updatedQuest, nil
}

func (s *Store) updateSteps(ctx context.Context, quest *model.QuestWithSteps, steps []model.Step) (*model.QuestWithSteps, error) {
	for i := range steps {
		steps[i].QuestId = quest.ID
		steps[i].CreatedAt = quest.UpdatedAt
		steps[i].UpdatedAt = quest.UpdatedAt
	}

	res, err := s.db.NamedQueryContext(ctx, `INSERT INTO
			steps(quest_id, sort,description,question_type,question_content,answer_type,answer_content,created_at,updated_at)
			VALUES (:quest_id, :sort,:description,:question_type,:question_content,:answer_type,:answer_content,:created_at,:updated_at)
			RETURNING *`, steps)

	if err = checkWriteError(err); err != nil {
		return nil, err
	}

	quest.Steps = []model.Step{}
	var step model.Step
	for res.Next() {
		if err := res.StructScan(&step); err != nil {
			return nil, errors.ErrUnknown.Wrap(err)
		}
		quest.Steps = append(quest.Steps, step)
	}

	defer res.Close()
	return quest, nil
}

func (s *Store) updateEmails(ctx context.Context, quest *model.QuestWithSteps, emails *[]model.Email) error {
	var args []map[string]interface{}
	for _, email := range *emails {
		arg := map[string]interface{}{
			"quest_id": quest.ID,
			"email":    email,
		}
		args = append(args, arg)
	}
	res, err := s.db.NamedQueryContext(ctx, `INSERT INTO quest_to_email(quest_id, email) VALUES (:quest_id, :email)`, args)
	if err = checkWriteError(err); err != nil {
		return err
	}
	defer res.Close()
	return nil
}

// InsertQuest will add a new quest to the database using the provided data.
func (s *Store) InsertQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error) {
	createdQuest, err := s.saveQuest(ctx, quest)
	if err != nil {
		return nil, err
	}

	if len(quest.Steps) != 0 {
		createdQuest, err = s.updateSteps(ctx, createdQuest, quest.Steps)
		if err != nil {
			return nil, err
		}
	}

	return createdQuest, nil
}

func (s *Store) DeleteQuest(ctx context.Context, id string) error {
	userId := ctx.Value(http.ContextUserIdKey).(string)
	res, err := s.db.ExecContext(ctx, "DELETE FROM quests WHERE id = $1 AND owner = $2", id, userId)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == pqErrInvalidTextRepresentation && strings.Contains(pqErr.Error(), "uuid") {
				return ErrInvalidID.Wrap(errors.ErrValidation.Wrap(err))
			}
		}

		return errors.ErrUnknown.Wrap(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return errors.ErrUnknown.Wrap(err)
	}
	if rows != 1 {
		return ErrQuestNotDeleted.Wrap(errors.ErrNotFound)
	}

	return nil
}

// GetQuestsByUser will get quests created by user
func (s *Store) GetQuestsByUser(ctx context.Context, uuid string, offset int, limit int) ([]model.Quest, error) {
	ownerClause := fmt.Sprintf("owner='%s'", uuid)
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	rows, err := s.db.QueryxContext(ctx, fmt.Sprintf("SELECT * FROM quests WHERE %s ORDER BY created_at ASC%s", ownerClause, limitClause))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound.Wrap(err)
		}

		return nil, errors.ErrUnknown.Wrap(err)
	}
	defer rows.Close()

	var quests []model.Quest

	for rows.Next() {
		var quest model.Quest
		if err := rows.StructScan(&quest); err != nil {
			logging.From(ctx).Error("failed to deserialize quest from database", zap.Error(err))
		} else {
			quests = append(quests, quest)
		}
	}

	return quests, nil
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
