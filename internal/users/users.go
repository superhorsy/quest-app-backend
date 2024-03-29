//go:generate mockgen -destination=./mocks/users_mock.go -package mocks github.com/superhorsy/quest-app-backend/internal/users Store,Events

package users

import (
	"context"

	"github.com/superhorsy/quest-app-backend/internal/core/helpers"

	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"github.com/superhorsy/quest-app-backend/internal/events"
	"github.com/superhorsy/quest-app-backend/internal/users/model"
	"go.uber.org/zap"
)

const (
	// ErrInvalidFilterValue is returned when a filter value is empty.
	ErrInvalidFilterValue = errors.Error("invalid_filter_value: invalid filter value")
	// ErrInvalidFilterMatchType is returned when a filter match type is not found in the supported enum list.
	ErrInvalidFilterMatchType = errors.Error("invalid_filter_match_type: invalid filter match type")
	// ErrInvalidFilterField is returned when a filter field is not found in the supported enum list.
	ErrInvalidFilterField = errors.Error("invalid_filter_field: invalid filter field")
)

// Store represents a type for storing a user in a database.
type Store interface {
	InsertUser(ctx context.Context, user *model.UserWithPass) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.UserWithPass) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUsers(ctx context.Context, filters []model.Filter, offset, limit int64) ([]*model.User, error)
	DeleteUser(ctx context.Context, id string) error
}

// Events represents a type for producing events on user CRUD operations.
type Events interface {
	Produce(ctx context.Context, topic events.Topic, payload interface{})
}

// Users provides functionality for CRUD operations on a user.
type Users struct {
	store  Store
	events Events
}

// Quests provides functionality for CRUD operations on quests.
type Quests struct {
	store  Store
	events Events
}

func (u *Users) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := u.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// New will instantiate a new instance of Users.
func New(s Store, e Events) *Users {
	return &Users{
		store:  s,
		events: e,
	}
}

// CreateUser will try to create a user in our database with the provided data if it represents a unique new user.
func (u *Users) CreateUser(ctx context.Context, user *model.UserWithPass) (*model.User, error) {
	passwordHash := helpers.HashAndSalt([]byte(*user.Password))
	user.Password = &passwordHash

	// Not much validation needed before storing in the database as the database itself is handling most of that
	// if we were to use something else you would probably want to add validation of inputs here
	createdUser, err := u.store.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Assuming our events producer has some guarantees on retries and recovering from failures
	// we shouldn't need to worry about failures to send the event here and assume it will be sent eventually.
	// In an ideal world our producer would in the case of failing to send an event have some sort
	// of recovery mechanism to ensure that we don't lose any events. Such as picking up failed events on
	// a later run, and retrying them.
	// We shouldn't need to fail the whole process if we can't produce an event right now.
	u.events.Produce(ctx, events.TopicUsers, events.UserEvent{
		EventType: events.EventTypeUserCreated,
		ID:        *createdUser.ID,
		User:      createdUser,
	})

	return createdUser, nil
}

// UpdateUser will try to update an existing user in our database with the provided data.
func (u *Users) UpdateUser(ctx context.Context, user *model.UserWithPass) (*model.User, error) {
	if user.Password != nil && *user.Password != "" {
		passwordHash := helpers.HashAndSalt([]byte(*user.Password))
		user.Password = &passwordHash
	}

	updatedUser, err := u.store.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	u.events.Produce(ctx, events.TopicUsers, events.UserEvent{
		EventType: events.EventTypeUserUpdated,
		ID:        *updatedUser.ID,
		User:      updatedUser,
	})

	return updatedUser, nil
}

// GetUser will try to get an existing user in our database with the provided id.
func (u *Users) GetUser(ctx context.Context, id string) (*model.User, error) {
	user, err := u.store.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindUsers will retrieve a list of users based on matching all the provided filters and using pagination if limit is gt 0.
func (u *Users) FindUsers(ctx context.Context, filters []model.Filter, offset, limit int64) ([]*model.User, error) {
	// Validate filters before searching with them
	// TODO may want to return details of error instead of just logging
	for i, f := range filters {
		if f.Value == "" {
			err := ErrInvalidFilterValue.Wrap(errors.ErrValidation)
			logging.From(ctx).Error("empty filter value provided", zap.Error(err), zap.Int("index", i))
			return nil, err
		}

		switch f.MatchType {
		case model.MatchTypeEqual, model.MatchTypeLike:
			// noop
		default:
			err := ErrInvalidFilterMatchType.Wrap(errors.ErrValidation)
			logging.From(ctx).Error("match type not supported", zap.Error(err), zap.String("match_type", string(f.MatchType)), zap.Int("index", i))
			return nil, err
		}

		switch f.Field {
		case model.FieldFirstName, model.FieldLastName, model.FieldNickname, model.FieldEmail:
		// noop
		default:
			err := ErrInvalidFilterField.Wrap(errors.ErrValidation)
			logging.From(ctx).Error("filter field not supported", zap.Error(err), zap.String("field", string(f.Field)), zap.Int("index", i))
			return nil, err
		}
	}

	users, err := u.store.FindUsers(ctx, filters, offset, limit)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUser will try to delete an existing user in our database with the provided id.
func (u *Users) DeleteUser(ctx context.Context, id string) error {
	err := u.store.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	u.events.Produce(ctx, events.TopicUsers, events.UserEvent{
		EventType: events.EventTypeUserDeleted,
		ID:        id,
	})

	return nil
}
