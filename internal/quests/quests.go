package quests

import (
	"context"
	"fmt"
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
	GetQuestsByUser(ctx context.Context, uuid string, offset int, limit int) ([]model.Quest, error)
	UpdateQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error)
	DeleteQuest(ctx context.Context, id string) error
	GetQuestsAvailable(ctx context.Context, uuid string, offset int, limit int) ([]model.QuestAvailable, *model.Meta, error)
	CreateAssignment(ctx context.Context, request model.SendQuestRequest) error
	GetAssignment(ctx context.Context, id string, email *string) (*model.Assignment, error)
	UpdateAssignment(ctx context.Context, questId string, email *string, currentStep int, status model.Status) error
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
	err := q.store.CreateAssignment(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (q *Quests) GetAssignment(ctx context.Context, questId string, email *string) (*model.Assignment, error) {
	ass, err := q.store.GetAssignment(ctx, questId, email)
	if err != nil {
		return nil, err
	}

	return ass, nil
}

func (q *Quests) UpdateAssignment(ctx context.Context, questId string, email *string, currentStep int, status model.Status) error {
	err := q.store.UpdateAssignment(ctx, questId, email, currentStep, status)
	if err != nil {
		return err
	}

	return nil
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

func (q *Quests) GetQuest(ctx context.Context, id string) (*model.QuestWithSteps, error) {
	quest, err := q.store.GetQuest(ctx, id)
	if err != nil {
		return nil, err
	}
	return quest, nil
}

// UpdateQuest updates quests. If there were any steps inside it deletes them and insert new regardless of already created steps
func (q *Quests) UpdateQuest(ctx context.Context, quest *model.QuestWithSteps) (*model.QuestWithSteps, error) {
	uId := ctx.Value(http.ContextUserIdKey).(string)
	if *quest.Owner != uId {
		return nil, errors.New(fmt.Sprintf("Bad owner ID %s for quest %s", *quest.Owner, *quest.ID))
	}
	createdQuest, err := q.store.UpdateQuest(ctx, quest)
	if err != nil {
		return nil, err
	}
	q.events.Produce(ctx, events.TopicQuests, events.QuestEvent{
		EventType: events.EventTypeQuestUpdated,
		ID:        *createdQuest.ID,
		Quest:     createdQuest,
	})

	return createdQuest, nil
}

func (q *Quests) GetQuestsByUser(ctx context.Context, ownerUuid string, offset int, limit int) ([]model.Quest, error) {
	quests, err := q.store.GetQuestsByUser(ctx, ownerUuid, offset, limit)
	if err != nil {
		return nil, err
	}

	return quests, nil
}

func (q *Quests) DeleteQuest(ctx context.Context, id string) error {
	err := q.store.DeleteQuest(ctx, id)
	if err != nil {
		return err
	}

	q.events.Produce(ctx, events.TopicUsers, events.UserEvent{
		EventType: events.EventTypeQuestDeleted,
		ID:        id,
	})

	return nil
}

func (q *Quests) GetQuestsAvailable(ctx context.Context, email string, offset int, limit int) ([]model.QuestAvailable, *model.Meta, error) {
	list, meta, err := q.store.GetQuestsAvailable(ctx, email, offset, limit)
	if err != nil {
		return nil, nil, err
	}
	return list, meta, nil
}
