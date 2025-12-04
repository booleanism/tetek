package entities

import "github.com/google/uuid"

type HiddenFeed struct {
	ID     uuid.UUID `json:"id,omitempty"`
	To     string    `json:"to,omitempty"`
	FeedID uuid.UUID `json:"feed_id,omitempty"`
}
