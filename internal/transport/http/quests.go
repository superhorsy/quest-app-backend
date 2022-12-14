package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/superhorsy/quest-app-backend/internal/core/helpers"
	questModel "github.com/superhorsy/quest-app-backend/internal/quests/model"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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
		offset, err = strconv.Atoi(limitQueryParam)
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
			handleError(ctx, w, errors.ErrUnknown.Wrap(err))
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
	quests, meta, err := s.quests.GetQuestsAvailable(ctx, *user.Email, offset, limit)
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

	if err := s.quests.DeleteQuest(ctx, id); err != nil {
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
				Name: sendRequest.Name,
				URL:  "https://questy.fun",
				IMG:  "https://wsrv.nl/?url=questy.fun/files/10d26a38-2fdf-4f48-adff-3e052e7466f5.png",
			}
			subject := fmt.Sprintf("Ваш друг %s отправил вам квест на Questy.fun!", user.FullName())
			err := helpers.SendEmail(sendRequest.Email, subject, "config/quest_invite.html", templateData)
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

	vars := mux.Vars(r)
	id := vars["id"]

	q, err := s.quests.GetQuest(ctx, id)
	if err != nil {
		logging.From(ctx).Error("failed to start q", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Check if q has any steps
	if len(q.Steps) == 0 {
		err = errors.ErrValidation.Wrap(errors.Error("can't start q: no steps found inside a q"))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, err)
		return
	}
	// Check if q is already started
	userId := ctx.Value(ContextUserIdKey)
	user, err := s.users.GetUser(ctx, userId.(string))
	if err != nil {
		logging.From(ctx).Error("failed to find user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	ass, err := s.quests.GetAssignment(ctx, *q.ID, user.Email)
	if err != nil {
		logging.From(ctx).Error("failed to find assignment", zap.Error(err))
		handleError(ctx, w, err)
		return
	}
	if ass.Status == questModel.StatusInProgress {
		err = errors.New("quest already started")
		logging.From(ctx).Error("failed to start quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}
	if ass.Status == questModel.StatusFinished {
		err = errors.New("quest already finished")
		logging.From(ctx).Error("failed to start quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Create linked list from steps
	ql := q.NewQuestLine(nil, questModel.StatusInProgress)

	// Save to DB
	if err := s.quests.UpdateAssignment(ctx, *q.ID, user.Email, *ql.List.Head.Value.Sort, questModel.StatusInProgress); err != nil {
		logging.From(ctx).Error("failed to start quest", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, ql)
}

func (s *Server) checkAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	q, err := s.quests.GetQuest(ctx, id)
	if err != nil {
		logging.From(ctx).Error("failed to start q", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Check if q has any steps
	if len(q.Steps) == 0 {
		err = errors.ErrValidation.Wrap(errors.Error("can't start quest: no steps found inside a quest"))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, err)
		return
	}
	// Check if q is in progress
	userId := ctx.Value(ContextUserIdKey)
	user, err := s.users.GetUser(ctx, userId.(string))
	if err != nil {
		logging.From(ctx).Error("failed to find user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	ass, err := s.quests.GetAssignment(ctx, *q.ID, user.Email)
	if err != nil {
		logging.From(ctx).Error("failed to find assignment", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	if ass.Status == questModel.StatusNotStarted {
		err = errors.New("quest not started")
		logging.From(ctx).Error("failed to check answer", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	if ass.Status == questModel.StatusFinished {
		err = errors.New("quest finished")
		logging.From(ctx).Error("failed to check answer", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Create linked list from steps
	ql := q.NewQuestLine(&ass.CurrentStep, ass.Status)

	req, err := parseBodyIntoStruct(r, questModel.CheckAnswerRequest{})
	if err != nil {
		handleError(ctx, w, err)
		return
	}

	isCorrect, err := checkIfAnswerCorrect(*req, ql.List.Head.Value)

	if isCorrect {
		// If it was last one
		if ql.IsLastStep() {
			ql.QuestStatus = questModel.StatusFinished
			ql.FinalMessage = q.FinalMessage
			if err = s.quests.UpdateAssignment(ctx, *q.ID, user.Email, ql.CurrentStep(), questModel.StatusFinished); err != nil {
				logging.From(ctx).Error("failed to start q", zap.Error(err))
				handleError(ctx, w, err)
				return
			}
		} else {
			// Switch to next question
			ql.Next()
			// Save to DB
			if err = s.quests.UpdateAssignment(ctx, *q.ID, user.Email, ql.CurrentStep(), questModel.StatusInProgress); err != nil {
				logging.From(ctx).Error("failed to start q", zap.Error(err))
				handleError(ctx, w, err)
				return
			}
		}
	}

	ql.IsQuestionAnswerCorrect = &isCorrect

	handleResponse(ctx, w, ql)
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]

	q, err := s.quests.GetQuest(ctx, id)
	if err != nil {
		logging.From(ctx).Error("failed to start q", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Check if q has any steps
	if len(q.Steps) == 0 {
		err = errors.ErrValidation.Wrap(errors.Error("can't get quest status: no steps found inside a quest"))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, err)
		return
	}
	// Check if q is in progress
	userId := ctx.Value(ContextUserIdKey)
	user, err := s.users.GetUser(ctx, userId.(string))
	if err != nil {
		logging.From(ctx).Error("failed to find user", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	ass, err := s.quests.GetAssignment(ctx, *q.ID, user.Email)
	if err != nil {
		logging.From(ctx).Error("failed to find assignment", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	// Create linked list from steps
	ql := q.NewQuestLine(&ass.CurrentStep, ass.Status)

	handleResponse(ctx, w, ql)
}

func checkIfAnswerCorrect(req questModel.CheckAnswerRequest, q questModel.Question) (bool, error) {
	answer := req.Answer
	for _, correctAnswer := range *q.AnswerContent {
		if strings.TrimSpace(strings.ToLower(answer)) == strings.TrimSpace(strings.ToLower(correctAnswer)) {
			return true, nil
		}
	}
	return false, nil
}
