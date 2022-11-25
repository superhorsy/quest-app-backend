package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	questModel "github.com/superhorsy/quest-app-backend/internal/quests/model"
	"io"
	"net/http"
	"strconv"

	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"go.uber.org/zap"
)

func (s *Server) createQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logging.From(ctx).Error("failed to read request body", zap.Error(err))
		handleError(ctx, w, errors.ErrUnknown.Wrap(err))
		return
	}

	q := questModel.QuestWithSteps{}

	if err := json.Unmarshal(data, &q); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	id := ctx.Value(ContextUserIdKey).(string)
	q.ID = &id

	createdQuest, err := s.quests.CreateQuest(ctx, &q)
	if err != nil {
		logging.From(ctx).Error("failed to create quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, createdQuest)
}

func (s *Server) updateQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logging.From(ctx).Error("failed to read request body", zap.Error(err))
		handleError(ctx, w, errors.ErrUnknown.Wrap(err))
		return
	}

	q := questModel.QuestWithSteps{}

	if err := json.Unmarshal(data, &q); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	q.ID = &id

	updatedQuest, err := s.quests.UpdateQuest(ctx, &q)
	if err != nil {
		logging.From(ctx).Error("failed to update quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, updatedQuest)
}

func (s *Server) getQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	quest, err := s.quests.GetQuest(ctx, id)
	if err != nil {
		logging.From(ctx).Error("failed to fetch quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, quest)
}

func (s *Server) getQuestsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 50
	limitQueryParam := r.URL.Query().Get("limit")
	if limitQueryParam != "" {
		var err error
		limit, err = strconv.Atoi(limitQueryParam)
		if err != nil {
			logging.From(ctx).Error("failed to read request body", zap.Error(err))
			handleError(ctx, w, errors.ErrUnknown.Wrap(err))
			return
		}
	}

	offset := 0
	offsetQueryParam := r.URL.Query().Get("offset")
	if offsetQueryParam != "" {
		var err error
		offset, err = strconv.Atoi(limitQueryParam)
		if err != nil {
			logging.From(ctx).Error("failed to read request body", zap.Error(err))
			handleError(ctx, w, errors.ErrUnknown.Wrap(err))
			return
		}
	}

	userId := ctx.Value(ContextUserIdKey)
	quests, err := s.quests.GetQuestsByUser(ctx, userId.(string), offset, limit)
	if err != nil {
		logging.From(ctx).Error("failed to fetch quests", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, quests)
}

func (s *Server) deleteQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.quests.DeleteQuest(ctx, id); err != nil {
		logging.From(ctx).Error("failed to delete quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, struct {
		Success bool `json:"success"`
	}{Success: true})
}
