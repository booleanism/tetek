package model

import (
	"time"

	"github.com/google/uuid"
)

type Feed struct {
	ID         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	Text       string     `json:"text"`
	By         string     `json:"by"`
	Type       string     `json:"type"`
	Points     int        `json:"points"`
	NCommnents int        `json:"n_comments"`
	CreatedAt  *time.Time `json:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}
