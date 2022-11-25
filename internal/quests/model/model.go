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

// QuestWithSteps represents a quest
type QuestWithSteps struct {
	ID          *string `json:"id" db:"id"`
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Owner       *string `json:"owner" db:"owner"`
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
