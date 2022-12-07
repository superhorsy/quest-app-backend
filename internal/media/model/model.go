package model

import (
	"time"
)

type MediaRecord struct {
	ID        string     `json:"id" db:"id"`
	Storage   string     `json:"-" db:"storage"`
	Type      MediaType  `json:"type" db:"type"`
	Filename  string     `json:"filename" db:"filename"`
	Link      string     `json:"link" db:"link"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

type MediaType string

const (
	Sound MediaType = "sound"
	Image MediaType = "image"
)
