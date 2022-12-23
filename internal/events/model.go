package events

import (
	"github.com/superhorsy/quest-app-backend/internal/media/model"
	questModel "github.com/superhorsy/quest-app-backend/internal/quests/model"
	userModel "github.com/superhorsy/quest-app-backend/internal/users/model"
)

// EventType represents the type of event that occurred.
type EventType string

const (
	// EventTypeUserCreated is triggered after a user has been successfully created.
	EventTypeUserCreated EventType = "user_created"
	// EventTypeUserUpdated is triggered after a user has been successfully updated.
	EventTypeUserUpdated EventType = "user_updated"
	// EventTypeUserDeleted is triggered after a user has been successfully deleted.
	EventTypeUserDeleted EventType = "user_deleted"

	EventTypeQuestCreated EventType = "quest_created"
	EventTypeQuestUpdated EventType = "quest_updated"
	EventTypeQuestDeleted EventType = "quest_deleted"
	EventTypeQuestSent    EventType = "quest_sent"
)

// UserEvent represents an event that occurs on a user entity.
type UserEvent struct {
	EventType EventType       `json:"event_type"`
	ID        string          `json:"id"`
	User      *userModel.User `json:"user"`
}

type QuestEvent struct {
	EventType EventType                  `json:"event_type"`
	ID        string                     `json:"id"`
	Quest     *questModel.QuestWithSteps `json:"quest"`
}

type MediaEvent struct {
	EventType   EventType          `json:"event_type"`
	ID          string             `json:"id"`
	MediaRecord *model.MediaRecord `json:"media_record"`
}
