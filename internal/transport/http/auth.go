package http

import (
	"encoding/json"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/helpers"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"github.com/superhorsy/quest-app-backend/internal/users/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"regexp"
)

type LoginForm struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegistrationRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	Email     string `json:"email"`
}

type Validation struct {
	Value string
	Valid string
	Error error
}

type UserWithToken struct {
	*model.User
	AccessToken string `json:"access_token,omitempty"`
}

func validation(values []Validation) error {
	username := regexp.MustCompile(`^([A-Za-z0-9]{5,})+$`)
	email := regexp.MustCompile(`^[A-Za-z0-9]+[@]+[A-Za-z0-9]+[.][A-Za-z0-9]+$`)

	for i := 0; i < len(values); i++ {
		switch values[i].Valid {
		case "username":
			if !username.MatchString(values[i].Value) {
				return values[i].Error
			}
		case "email":
			if !email.MatchString(values[i].Value) {
				return values[i].Error
			}
		case "password":
			if len(values[i].Value) < 5 {
				return values[i].Error
			}
		}
	}
	return nil
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		logging.From(ctx).Error("failed to read request body", zap.Error(err))
		handleError(ctx, w, errors.ErrUnknown.Wrap(err))
		return
	}

	f := RegistrationRequest{}

	if err := json.Unmarshal(data, &f); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	err = validation(
		[]Validation{
			//{Value: f.Nickname, Valid: "username", Error: errors.New("Username can contain only numbers and english letters")},
			{Value: f.Email, Valid: "email", Error: errors.New("Invalid email")},
			//{Value: f.Password, Valid: "password", Error: errors.New("Password should be at least 5 letters long")},
		})
	if err != nil {
		logging.From(ctx).Error("validation failed", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	u := model.UserWithPass{}

	if err := json.Unmarshal(data, &u); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	createdUser, err := s.users.CreateUser(ctx, &u)
	if err != nil {
		logging.From(ctx).Error("failed to create user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	token := helpers.CreateJwtToken(*createdUser.ID, err)
	if err != nil {
		logging.From(ctx).Error("failed to create token", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, createdUser, token)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logging.From(ctx).Error("failed to read request body", zap.Error(err))
		handleError(ctx, w, errors.ErrUnknown.Wrap(err))
		return
	}

	f := LoginForm{}

	if err := json.Unmarshal(data, &f); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	user, err := s.users.GetUserByEmail(ctx, f.Email)
	if err != nil {
		logging.From(ctx).Error("failed to login user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(f.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword && err != nil {
		err = errors.New("Wrong password")
		logging.From(ctx).Error("failed to login user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}
	if err != nil {
		logging.From(ctx).Error("failed to login user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	token := helpers.CreateJwtToken(*user.ID, err)
	if err != nil {
		logging.From(ctx).Error("failed to create token", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, user, token)
}
