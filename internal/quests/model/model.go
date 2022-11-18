package model

import "time"

type QuestionType string

const (
	QuestionText  QuestionType = "text"
	QuestionSound QuestionType = "sound"
)

type AnswerType string

const (
	AnswerText AnswerType = "text"
)

type Question struct {
	QuestionType QuestionType `json:"question_type,omitempty" db:"question_type"`
	Question     *string      `json:"question,omitempty" db:"question"`
}

type Answer struct {
	AnswerType AnswerType `json:"answer_type,omitempty" db:"answer_type"`
	Answer     *[]string  `json:"answer,omitempty" db:"answer"`
}

type Step struct {
	ID          *string  `json:"id,omitempty" db:"id"`
	Sort        *int     `json:"sort,omitempty" db:"sort"`
	Description *string  `json:"description,omitempty" db:"description"`
	Question    Question `json:"question"`
	Answer      Answer   `json:"answer"`
}

type Email string

// Quest represents a quest
type Quest struct {
	ID        *string    `json:"id" db:"id"`
	Owner     *string    `json:"owner" db:"owner"`
	Name      *string    `json:"name" db:"name"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	Steps     *[]Step    `json:"steps"`
	Emails    *[]Email   `json:"emails"`
}
