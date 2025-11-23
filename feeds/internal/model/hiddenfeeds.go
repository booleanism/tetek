package model

import "github.com/google/uuid"

type HiddenFeed struct {
	ID     string    `json:"id"`
	To     string    `json:"to"`
	FeedID uuid.UUID `json:"feed_id"`
}
