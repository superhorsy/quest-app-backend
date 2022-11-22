package http

import (
	"encoding/json"
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

	q := questModel.Quest{}

	if err := json.Unmarshal(data, &q); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	createdQuest, err := s.quests.CreateQuest(ctx, &q)
	if err != nil {
		// TODO deal with different error types that affect the error response from the generic types
		logging.From(ctx).Error("failed to create quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, createdQuest)
}

func (s *Server) getQuestsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Add("Content-Type", "application/json")

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		err := errors.New("uuid not found")
		logging.From(ctx).Error("failed to create quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	limit := 0
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

	quests, err := s.quests.GetQuestsByUser(ctx, uuid, offset, limit)
	if err != nil {
		logging.From(ctx).Error("failed to fetch quests", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, quests)
}
