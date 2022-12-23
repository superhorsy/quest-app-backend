package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"github.com/superhorsy/quest-app-backend/internal/users/model"
	"go.uber.org/zap"
)

type searchUsersRequest struct {
	Filters []model.Filter `json:"filters"`
	Offset  int64          `json:"offset"`
	Limit   int64          `json:"limit"`
}

type deletedUserResponse struct {
	Success bool `json:"success"`
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logging.From(ctx).Error("failed to read request body", zap.Error(err))
		handleError(ctx, w, errors.ErrUnknown.Wrap(err))
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
		// TODO deal with different error types that affect the error response from the generic types
		logging.From(ctx).Error("failed to create user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, createdUser)
}

func (s *Server) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId := ctx.Value(ContextUserIdKey)

	u, err := s.users.GetUser(ctx, userId.(string))
	if err != nil {
		logging.From(ctx).Error("failed to get user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, u)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId := ctx.Value(ContextUserIdKey).(string)

	u, err := s.users.GetUser(ctx, userId)
	if err != nil {
		logging.From(ctx).Error("failed to get user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	req, err := parseBodyIntoStruct(r, model.UserWithPass{})
	if err != nil {
		handleError(ctx, w, err)
		return
	}
	req.ID = &userId
	// Email can not be changed on profile update
	req.Email = u.Email

	updateUser, err := s.users.UpdateUser(ctx, req)
	if err != nil {
		logging.From(ctx).Error("failed to update user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, updateUser)
}

func (s *Server) searchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logging.From(ctx).Error("failed to read request body", zap.Error(err))
		handleError(ctx, w, errors.ErrUnknown.Wrap(err))
		return
	}

	req := searchUsersRequest{}

	if err := json.Unmarshal(data, &req); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	users, err := s.users.FindUsers(ctx, req.Filters, req.Offset, req.Limit)
	if err != nil {
		logging.From(ctx).Error("failed to find users", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, users)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.users.DeleteUser(ctx, id); err != nil {
		logging.From(ctx).Error("failed to delete user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, deletedUserResponse{Success: true})
}
