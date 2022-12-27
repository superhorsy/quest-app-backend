package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/helpers"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	questModel "github.com/superhorsy/quest-app-backend/internal/quests/model"
	"go.uber.org/zap"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
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
	q.Owner = &id

	createdQuest, err := s.quests.CreateQuest(ctx, &q)
	if err != nil {
		logging.From(ctx).Error("failed to create quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, createdQuest)
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
		offset, err = strconv.Atoi(offsetQueryParam)
		if err != nil {
			logging.From(ctx).Error("failed to read request body", zap.Error(err))
			handleError(ctx, w, errors.ErrUnknown.Wrap(err))
			return
		}
	}

	userId := ctx.Value(ContextUserIdKey)
	quests, meta, err := s.quests.GetQuestsByUser(ctx, userId.(string), offset, limit)
	if err != nil {
		logging.From(ctx).Error("failed to fetch quests", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponseWithMeta(ctx, w, quests, meta)
}

func (s *Server) getAvailableQuests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 50
	limitQueryParam := r.URL.Query().Get("limit")
	if limitQueryParam != "" {
		var err error
		queryLimit, err := strconv.Atoi(limitQueryParam)
		if queryLimit < 1000 {
			limit = queryLimit
		}
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
		offset, err = strconv.Atoi(offsetQueryParam)
		if err != nil {
			logging.From(ctx).Error("failed to read request body", zap.Error(err))
			handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
			return
		}
	}

	finished := false
	finishedQueryParam := r.URL.Query().Get("finished")
	if finishedQueryParam != "" {
		var err error
		finished, err = strconv.ParseBool(finishedQueryParam)
		if err != nil {
			logging.From(ctx).Error("failed to read request body", zap.Error(err))
			handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
			return
		}
	}

	userId := ctx.Value(ContextUserIdKey)
	user, err := s.users.GetUser(ctx, userId.(string))
	if err != nil {
		logging.From(ctx).Error("failed to find user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}
	quests, meta, err := s.quests.GetQuestsAvailable(ctx, *user.Email, offset, limit, finished)
	if err != nil {
		logging.From(ctx).Error("failed to fetch quests", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponseWithMeta(ctx, w, quests, meta)
}

func (s *Server) deleteQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	err := s.quests.DeleteQuest(ctx, id)
	if err != nil {
		logging.From(ctx).Error("failed to delete quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, struct {
		Success bool `json:"success"`
	}{Success: true})
}

func (s *Server) sendQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	sendRequest, err := parseBodyIntoStruct(r, questModel.SendQuestRequest{})
	if err != nil {
		handleError(ctx, w, err)
		return
	}
	sendRequest.QuestId = id

	quest, err := s.quests.GetQuest(ctx, id)
	if err != nil {
		logging.From(ctx).Error("failed to send quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Check if quest has any steps
	if len(quest.Steps) == 0 {
		err = errors.ErrValidation.Wrap(errors.Error("can't send quest: no steps found inside a quest"))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Save to DB
	if err := s.quests.CreateAssignment(ctx, *sendRequest); err != nil {
		logging.From(ctx).Error("failed to save send quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Get user
	userId := ctx.Value(ContextUserIdKey)
	user, err := s.users.GetUser(ctx, userId.(string))
	if err != nil {
		logging.From(ctx).Error("failed to find user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	mailingEnabled := os.Getenv("MAILING_ENABLED")
	if mailingEnabled == "true" {
		// Send email
		go func() {
			templateData := struct {
				Name string
				URL  string
				IMG  string
			}{
				Name: html.EscapeString(sendRequest.Name),
				URL:  "https://questy.fun",
				IMG:  "https://wsrv.nl/?url=questy.fun/files/10d26a38-2fdf-4f48-adff-3e052e7466f5.png",
			}
			subject := fmt.Sprintf("Ваш друг %s отправил вам квест на Questy.fun!", user.FullName())
			err := helpers.SendEmail(sendRequest.Email, subject, "config/quest_invite.gohtml", templateData)
			if err != nil {
				logging.From(ctx).Error("failed to send email", zap.Error(err))
				return
			}
		}()
	}

	handleResponse(ctx, w, struct {
		Success bool `json:"success"`
	}{Success: true})
}

func (s *Server) updateQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	q, err := parseBodyIntoStruct(r, questModel.QuestWithSteps{})
	if err != nil {
		handleError(ctx, w, err)
		return
	}
	q.ID = &id

	updatedQuest, err := s.quests.UpdateQuest(ctx, q)
	if err != nil {
		logging.From(ctx).Error("failed to update quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, updatedQuest)
}

// Прохождение

func (s *Server) startQuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	questId := mux.Vars(r)["id"]

	userId := ctx.Value(ContextUserIdKey).(string)

	ql, err := s.quests.StartQuest(ctx, questId, &userId)
	if err != nil {
		logging.From(ctx).Error("failed to start quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, ql)
}

func (s *Server) checkAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	questId := mux.Vars(r)["id"]
	userId := ctx.Value(ContextUserIdKey).(string)
	answer, err := parseBodyIntoStruct(r, questModel.Answer{})
	if err != nil {
		handleError(ctx, w, err)
		return
	}

	ql, err := s.quests.CheckAnswer(ctx, questId, &userId, answer)
	if err != nil {
		logging.From(ctx).Error("failed to check answer", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, ql)
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	questId := mux.Vars(r)["id"]

	ql, err := s.quests.GetAssignment(ctx, questId)
	if err != nil {
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, ql)
}
