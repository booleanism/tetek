package model

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID  `json:"id"`
	Parent    uuid.UUID  `json:"parent"`
	Text      string     `json:"text"`
	By        string     `json:"by"`
	Points    int        `json:"points"`
	CreatedAt *time.Time `json:"created_at"`
	Child     []Comment  `json:"child"`
}
