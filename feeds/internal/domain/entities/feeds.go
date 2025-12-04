package entities

import (
	"time"

	"github.com/google/uuid"
)

type Feed struct {
	ID         uuid.UUID  `json:"id,omitempty"`
	Title      string     `json:"title,omitempty"`
	URL        string     `json:"url,omitempty"`
	Text       string     `json:"text,omitempty"`
	By         string     `json:"by,omitempty"`
	Type       string     `json:"type,omitempty"`
	Points     int        `json:"points,omitempty"`
	NCommnents int        `json:"n_comments,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}
