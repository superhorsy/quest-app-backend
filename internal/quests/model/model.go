package model

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"time"
)

type QuestionType string

const (
	QuestionText  QuestionType = "text"
	QuestionSound QuestionType = "sound"
)

type AnswerType string

const (
	AnswerText AnswerType = "text"
)

type AnswerContent []string

type Step struct {
	ID              *string        `json:"id,omitempty" db:"id"`
	QuestId         *string        `json:"quest_id" db:"quest_id"`
	Sort            *int           `json:"sort,omitempty" db:"sort"`
	Description     *string        `json:"description,omitempty" db:"description"`
	QuestionType    *QuestionType  `json:"question_type" db:"question_type"`
	QuestionContent *string        `json:"question_content" db:"question_content"`
	AnswerType      *AnswerType    `json:"answer_type" db:"answer_type"`
	AnswerContent   *AnswerContent `json:"answer_content" db:"answer_content"`

	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

// Value Make the Attrs struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (a *AnswerContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan Make the Attrs struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (a *AnswerContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

type Email string

type Theme string

const (
	ThemeValentine Theme = "valentine"
	ThemeChristmas Theme = "christmas"
	ThemeBirthday  Theme = "birthday"
	ThemeHalloween Theme = "halloween"
	ThemeCommon    Theme = "common"
)

// QuestWithSteps represents a quest
type QuestWithSteps struct {
	ID          *string `json:"id" db:"id"`
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Owner       *string `json:"owner" db:"owner"`
	Theme       *Theme  `json:"theme" db:"theme"`
	Steps       []Step  `json:"steps"`

	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

// Quest represents a quest
type Quest struct {
	ID          *string `json:"id" db:"id"`
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Owner       *string `json:"owner" db:"owner"`

	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

type Status string

const (
	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusFinished   = "finished"
)

type Assignment struct {
	QuestId     string `json:"quest_id" db:"quest_id"`
	Email       string `json:"email" db:"email"`
	Status      Status `json:"status" db:"status"`
	CurrentStep int    `json:"current_step" db:"current_step"`
}

type Owner struct {
	ID       string `json:"id,omitempty" db:"id"`
	FullName string `json:"name,omitempty" db:"name"`
}

type QuestAvailable struct {
	QuestId          string `json:"quest_id" db:"quest_id"`
	QuestName        string `json:"quest_name" db:"quest_name"`
	QuestDescription string `json:"quest_description" db:"quest_description"`
	Status           Status `json:"status" db:"status"`
	CurrentStep      int    `json:"steps_current" db:"steps_current"`
	StepsCount       int    `json:"steps_count" db:"steps_count"`
	Owner            *Owner `json:"owner"`
}

type Meta struct {
	TotalCount int `json:"total_count,omitempty" db:"total_count"`
}

type SendQuestRequest struct {
	QuestId string `json:"quest_id" db:"quest_id"`
	Email   string `json:"email" db:"email"`
	Name    string `json:"name" db:"name"`
}
