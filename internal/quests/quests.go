package quests

import (
	"context"

	"github.com/superhorsy/quest-app-backend/internal/events"
	"github.com/superhorsy/quest-app-backend/internal/quests/model"
)

// Store represents a type for storing a user in a database.
type Store interface {
	InsertQuest(ctx context.Context, quest *model.Quest) (*model.Quest, error)
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

func New(s Store, e Events) *Quests {
	return &Quests{
		store:  s,
		events: e,
	}
}

func (q *Quests) CreateQuest(ctx context.Context, quest *model.Quest) (*model.Quest, error) {
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
