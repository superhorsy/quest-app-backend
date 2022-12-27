package quests

import (
	"context"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/events"
	"github.com/superhorsy/quest-app-backend/internal/quests/model"
	questStore "github.com/superhorsy/quest-app-backend/internal/quests/store"
	"github.com/superhorsy/quest-app-backend/internal/transport/http"
)

// Store represents a type for storing a user in a database.
type Store interface {
	InsertQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error)
	GetQuest(ctx context.Context, id string) (*model.QuestWithSteps, error)
	GetQuestsByUser(ctx context.Context, uuid string, offset int, limit int) ([]model.Quest, *model.Meta, error)
	UpdateQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error)
	DeleteQuest(ctx context.Context, id string) error
	GetQuestsAvailable(ctx context.Context, email string, offset int, limit int, finished bool) ([]model.QuestAvailable, *model.Meta, error)
	CreateAssignment(ctx context.Context, request model.SendQuestRequest) error
	GetAssignment(ctx context.Context, questId string, userId string) (*model.Assignment, error)
	UpdateAssignment(ctx context.Context, questId string, userId *string, currentStep int, status model.Status) error
}

// Events represents a type for producing events on user CRUD operations.
type Events interface {
	Produce(ctx context.Context, topic events.Topic, payload interface{})
}

// Quests provides functionality for CRUD operations on a quests.
type Quests struct {
	store  Store
	events Events
}

func (q *Quests) CreateAssignment(ctx context.Context, request model.SendQuestRequest) error {
	_, err := q.getQuestWithAuthCheck(ctx, request.QuestId)
	if err != nil {
		return err
	}
	return q.store.CreateAssignment(ctx, request)
}

func (q *Quests) GetAssignment(ctx context.Context, questId string) (*model.QuestLine, error) {
	quest, err := q.store.GetQuest(ctx, questId)
	if err != nil {
		return nil, err
	}
	// Check if q has any steps
	if len(quest.Steps) == 0 {
		return nil, errors.ErrValidation.Wrap(errors.Error("can't get quest status: no steps found inside a quest"))
	}
	userId := ctx.Value(http.ContextUserIdKey).(string)
	ass, err := q.store.GetAssignment(ctx, questId, userId)
	if err != nil {
		return nil, err
	}
	return quest.NewQuestLine(&ass.CurrentStep, ass.Status), err
}

func (q *Quests) StartQuest(ctx context.Context, questId string, userId *string) (*model.QuestLine, error) {
	quest, err := q.store.GetQuest(ctx, questId)
	if err != nil {
		return nil, err
	}

	// Check if q has any steps
	if len(quest.Steps) == 0 {
		return nil, errors.ErrValidation.Wrap(errors.Error("can't start q: no steps found inside a q"))
	}

	ass, err := q.store.GetAssignment(ctx, questId, *userId)
	if err != nil {
		return nil, err
	}
	if ass.Status == model.StatusInProgress {
		return nil, errors.New("quest already started")
	}
	if ass.Status == model.StatusFinished {
		return nil, errors.New("quest already finished")
	}

	// Create linked list from steps
	ql := quest.NewQuestLine(nil, model.StatusInProgress)

	// Save to DB
	err = q.store.UpdateAssignment(ctx, questId, userId, *ql.List.Head.Value.Sort, model.StatusInProgress)

	return ql, err
}
func (q *Quests) CheckAnswer(ctx context.Context, questId string, userId *string, answer *model.Answer) (*model.QuestLine, error) {
	quest, err := q.store.GetQuest(ctx, questId)
	if err != nil {
		return nil, err
	}

	// Check if q has any steps
	if len(quest.Steps) == 0 {
		return nil, errors.ErrValidation.Wrap(errors.Error("can't start q: no steps found inside a q"))
	}

	ass, err := q.store.GetAssignment(ctx, questId, *userId)
	if err != nil {
		return nil, err
	}
	if ass.Status != model.StatusInProgress {
		return nil, errors.New("quest not in progress")
	}

	// Create linked list from steps
	ql := quest.NewQuestLine(&ass.CurrentStep, ass.Status)
	isCorrect := ql.CheckIfAnswerCorrect(*answer)
	ql.IsQuestionAnswerCorrect = &isCorrect

	if !isCorrect {
		return ql, nil
	}

	if !ql.IsLastStep() {
		// Switch to next question
		ql.Next()
	} else {
		// If it was last one
		ql.QuestStatus = model.StatusFinished
		ql.FinalMessage = quest.FinalMessage
		ql.Rewards = quest.Rewards
	}

	return ql, q.store.UpdateAssignment(ctx, questId, userId, ql.CurrentStep(), ql.QuestStatus)
}

func New(s *questStore.Store, e Events) *Quests {
	return &Quests{
		store:  s,
		events: e,
	}
}

func (q *Quests) CreateQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error) {
	createdQuest, err := q.store.InsertQuest(ctx, quest)
	if err != nil {
		return nil, err
	}

	q.events.Produce(ctx, events.TopicQuests, events.QuestEvent{
		EventType: events.EventTypeUserCreated,
		ID:        *createdQuest.ID,
		Quest:     createdQuest,
	})

	return createdQuest, nil
}

func (q *Quests) getQuestWithAuthCheck(ctx context.Context, id string) (*model.QuestWithSteps, error) {
	quest, err := q.store.GetQuest(ctx, id)
	if err != nil {
		return &model.QuestWithSteps{}, err
	}
	uId := ctx.Value(http.ContextUserIdKey).(string)
	if *quest.Owner != uId {
		return nil, errors.ErrForbidden.Wrap(errors.Error("Пользователь не имеет доступа к квесту"))
	}
	return quest, nil
}

func (q *Quests) GetQuest(ctx context.Context, id string) (*model.QuestWithSteps, error) {
	return q.getQuestWithAuthCheck(ctx, id)
}

// UpdateQuest updates quests. If there were any steps inside it deletes them and insert new regardless of already created steps
func (q *Quests) UpdateQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error) {
	_, err := q.getQuestWithAuthCheck(ctx, *quest.ID)
	if err != nil {
		return nil, err
	}
	quest, err = q.store.UpdateQuest(ctx, quest)
	if err != nil {
		return nil, err
	}
	q.events.Produce(ctx, events.TopicQuests, events.QuestEvent{
		EventType: events.EventTypeQuestUpdated,
		ID:        *quest.ID,
		Quest:     quest,
	})
	return quest, nil
}

func (q *Quests) DeleteQuest(ctx context.Context, id string) error {
	_, err := q.getQuestWithAuthCheck(ctx, id)
	if err != nil {
		return err
	}
	err = q.store.DeleteQuest(ctx, id)
	if err != nil {
		return err
	}

	q.events.Produce(ctx, events.TopicQuests, events.QuestEvent{
		EventType: events.EventTypeQuestDeleted,
		ID:        id,
	})

	return nil
}

func (q *Quests) GetQuestsByUser(ctx context.Context, ownerUuid string, offset int, limit int) ([]model.Quest, *model.Meta, error) {
	return q.store.GetQuestsByUser(ctx, ownerUuid, offset, limit)
}

func (q *Quests) GetQuestsAvailable(ctx context.Context, email string, offset int, limit int, finished bool) ([]model.QuestAvailable, *model.Meta, error) {
	return q.store.GetQuestsAvailable(ctx, email, offset, limit, finished)
}
