package model

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"sort"
	"time"
)

type QuestionType string

const (
	QuestionText  QuestionType = "text"
	QuestionQR    QuestionType = "qr"
	QuestionSound QuestionType = "sound"
	QuestionImage QuestionType = "image"
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
// simply decodes a JSON-encoded Value into the struct fields.
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
	ThemeValentain Theme = "valentain"
	ThemeChristmas Theme = "christmas"
	ThemeBirthday  Theme = "birthday"
	ThemeHalloween Theme = "halloween"
	ThemeCommon    Theme = "common"
)

// QuestWithSteps represents a quest
type QuestWithSteps struct {
	Quest
	Steps []Step `json:"steps"`
}

// Quest represents a quest
type Quest struct {
	ID          *string `json:"id" db:"id"`
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Owner       *string `json:"owner" db:"owner"`
	Theme       *Theme  `json:"theme" db:"theme"`

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
	Name        string `json:"name" db:"name"`
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

type CheckAnswerRequest struct {
	AnswerType AnswerType `json:"answer_type"`
	Answer     string     `json:"answer"`
}

type QuestLine struct {
	IsQuestionAnswerCorrect *bool       `json:"success,omitempty"`
	List                    *LinkedList `json:"current"`
	PreviousQuestions       []Question  `json:"previous"`
	QuestionCount           int         `json:"question_count"`
	QuestStatus             Status      `json:"quest_status"`
}

type Question struct {
	ID              *string        `json:"id,omitempty"`
	QuestId         *string        `json:"quest_id"`
	Sort            *int           `json:"sort,omitempty" `
	Description     *string        `json:"description,omitempty"`
	QuestionType    *QuestionType  `json:"question_type"`
	QuestionContent *string        `json:"question_content"`
	AnswerType      *AnswerType    `json:"answer_type"`
	AnswerContent   *AnswerContent `json:"-"`
}

// Node represents a node of linked list
type Node struct {
	Value Question
	next  *Node
}

func NewQuestLineFromSteps(steps []Step, currentStep *int, status Status) *QuestLine {
	sort.Slice(steps, func(i, j int) bool {
		return *steps[i].Sort < *steps[j].Sort
	})
	// get current step from list if not defined
	if currentStep == nil {
		currentStep = steps[0].Sort
	}
	ql := &QuestLine{}
	ll := LinkedList{}
	for _, s := range steps {
		q := &Question{
			ID:              s.ID,
			QuestId:         s.QuestId,
			Sort:            s.Sort,
			Description:     s.Description,
			QuestionType:    s.QuestionType,
			QuestionContent: s.QuestionContent,
			AnswerType:      s.AnswerType,
			AnswerContent:   s.AnswerContent,
		}
		if *s.Sort < *currentStep {
			prev := append(ql.PreviousQuestions, *q)
			ql.PreviousQuestions = prev
		} else {
			ll.Insert(*q)
		}
	}
	ql.List = &ll
	ql.QuestionCount = len(steps)
	ql.QuestStatus = status
	return ql
}

// LinkedList represents a linked list
type LinkedList struct {
	Head *Node
	len  int
}

func (l *LinkedList) MarshalJSON() ([]byte, error) {
	if l.Head == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(l.Head.Value)
}

// Insert inserts new node at the end of  from linked list
func (l *LinkedList) Insert(val Question) {
	n := Node{}
	n.Value = val
	if l.len == 0 {
		l.Head = &n
		l.len++
		return
	}
	ptr := l.Head
	for i := 0; i < l.len; i++ {
		if ptr.next == nil {
			ptr.next = &n
			l.len++
			return
		}
		ptr = ptr.next
	}
}

func (ql *QuestLine) Next() {
	// no action on last step
	if ql.List.Head.next == nil {
		return
	}

	ql.PreviousQuestions = append(ql.PreviousQuestions, ql.List.Head.Value)
	ql.List.Head = ql.List.Head.next
}

func (ql *QuestLine) CurrentStep() int {
	if ql.List.Head == nil {
		return -1
	}
	return *ql.List.Head.Value.Sort
}

func (ql *QuestLine) IsLastStep() bool {
	return ql.List.Head.next == nil
}
