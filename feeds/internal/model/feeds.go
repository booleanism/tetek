package model

import (
	"time"

	"github.com/google/uuid"
)

type Feed struct {
	Id         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	Url        string     `json:"url"`
	Text       string     `json:"text"`
	By         string     `json:"by"`
	Type       string     `json:"type"`
	Points     int        `json:"points"`
	NCommnents int        `json:"n_comments"`
	Created_At *time.Time `json:"created_at"`
	Deleted_At *time.Time `json:"deleted_at"`
}
